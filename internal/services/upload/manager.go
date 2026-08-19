// Package upload implements simple chunked file uploads: the client PUTs
// byte-range chunks against a session, and the server appends them straight
// to a temp file on disk (never buffering a whole file in memory). This is
// deliberately not a full resumable-upload protocol (no tus.io) -- just
// enough to tolerate a dropped connection on a multi-GB movie file.
package upload

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"magicbox/internal/models"
)

const DefaultChunkSizeBytes = 8 * 1024 * 1024 // advisory only; client may use any chunk size

var (
	ErrOffsetMismatch = errors.New("upload: chunk offset does not match received_bytes")
	ErrIncomplete     = errors.New("upload: not all bytes have been received yet")
)

type Manager struct {
	db         *gorm.DB
	stagingDir string // <movies_dir>/.magicbox/uploads-tmp
}

// Checksum calculates the SHA-256 digest of the fully received staging file.
func (m *Manager) Checksum(session *models.UploadSession) (string, error) {
	f, err := os.Open(m.tempPath(session))
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ClaimChecksum atomically records content ownership. false means an earlier
// upload already claimed the same bytes, including a concurrent request.
func (m *Manager) ClaimChecksum(session *models.UploadSession, checksum string) (bool, error) {
	record := models.MediaChecksum{SHA256: checksum, OriginalFilename: session.OriginalFilename}
	result := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	return result.RowsAffected == 1, result.Error
}

func (m *Manager) HasChecksum(checksum string) (bool, error) {
	var count int64
	err := m.db.Model(&models.MediaChecksum{}).Where("sha256 = ?", checksum).Count(&count).Error
	return count > 0, err
}

func (m *Manager) SetMovieChecksum(sourceRelpath, checksum string) {
	if err := m.db.Model(&models.Movie{}).Where("source_relpath = ?", sourceRelpath).Update("content_sha256", checksum).Error; err != nil {
		fmt.Printf("upload: recording movie checksum failed: %v\n", err)
	}
}

// Abort removes an uncompleted staging file and marks its session aborted.
func (m *Manager) Abort(session *models.UploadSession) error {
	if err := os.Remove(m.tempPath(session)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return m.db.Model(session).Update("status", models.UploadStatusAborted).Error
}

func NewManager(db *gorm.DB, moviesDir string) (*Manager, error) {
	staging := filepath.Join(moviesDir, ".magicbox", "uploads-tmp")
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return nil, fmt.Errorf("creating upload staging dir: %w", err)
	}
	return &Manager{db: db, stagingDir: staging}, nil
}

// Create starts a new upload session for a file of the given size.
func (m *Manager) Create(filename string, sizeBytes int64) (*models.UploadSession, error) {
	id := uuid.NewString()
	session := &models.UploadSession{
		ID:               id,
		OriginalFilename: filename,
		TotalSizeBytes:   sizeBytes,
		TempRelpath:      id + ".part",
		Status:           models.UploadStatusInProgress,
	}

	f, err := os.Create(m.tempPath(session))
	if err != nil {
		return nil, fmt.Errorf("creating temp upload file: %w", err)
	}
	f.Close()

	if err := m.db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) Get(id string) (*models.UploadSession, error) {
	var session models.UploadSession
	if err := m.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// WriteChunk appends length bytes from r at the given offset. The offset
// must equal the session's current received_bytes -- this keeps the
// protocol simple (strictly sequential chunks) while still letting a client
// resync via Get() and resume after a dropped connection.
func (m *Manager) WriteChunk(session *models.UploadSession, offset int64, length int64, r io.Reader) (int64, error) {
	if offset != session.ReceivedBytes {
		return session.ReceivedBytes, ErrOffsetMismatch
	}

	f, err := os.OpenFile(m.tempPath(session), os.O_WRONLY, 0o644)
	if err != nil {
		return session.ReceivedBytes, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return session.ReceivedBytes, err
	}

	buf := make([]byte, 32*1024)
	written, err := io.CopyBuffer(f, io.LimitReader(r, length), buf)
	if err != nil {
		return session.ReceivedBytes, err
	}

	session.ReceivedBytes += written
	if err := m.db.Model(session).Updates(map[string]interface{}{
		"received_bytes": session.ReceivedBytes,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		return session.ReceivedBytes, err
	}

	return session.ReceivedBytes, nil
}

// Complete validates the upload is fully received and moves the assembled
// file into destDir under its original filename (de-duplicating the name if
// something is already there), returning the new absolute path.
func (m *Manager) Complete(session *models.UploadSession, destDir string) (string, error) {
	if session.ReceivedBytes != session.TotalSizeBytes {
		return "", ErrIncomplete
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}

	destPath := uniqueDestPath(destDir, session.OriginalFilename)
	if err := os.Rename(m.tempPath(session), destPath); err != nil {
		return "", fmt.Errorf("moving upload into library: %w", err)
	}

	if err := m.db.Model(session).Update("status", models.UploadStatusCompleted).Error; err != nil {
		return destPath, err
	}
	return destPath, nil
}

func (m *Manager) tempPath(session *models.UploadSession) string {
	return filepath.Join(m.stagingDir, session.TempRelpath)
}

// uniqueDestPath avoids clobbering an existing file by appending " (2)",
// " (3)", etc. before the extension.
func uniqueDestPath(dir, filename string) string {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]

	candidate := filepath.Join(dir, filename)
	for i := 2; ; i++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
	}
}
