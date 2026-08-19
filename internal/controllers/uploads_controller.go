package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"magicbox/internal/models"
	"magicbox/internal/services/library"
	"magicbox/internal/services/mediatype"
	"magicbox/internal/services/music"
	"magicbox/internal/services/upload"
)

var sha256Pattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

type UploadsController struct {
	manager       *upload.Manager
	moviesDir     string
	musicDir      string
	importer      *library.Importer
	musicImporter *music.Importer
}

// Direct receives an entire file as one streaming request. This is the
// browser-friendly path; the chunked endpoints remain available for clients
// that explicitly need resumability.
func (uc *UploadsController) Direct(c *gin.Context) {
	filename := filepath.Base(strings.TrimSpace(c.Query("filename")))
	if filename == "." || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}
	if _, ok := mediatype.Detect(filename); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized file extension"})
		return
	}
	if c.Request.ContentLength <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Length header is required"})
		return
	}

	session, err := uc.manager.Create(filename, c.Request.ContentLength)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
		return
	}
	hasher := sha256.New()
	if _, err := uc.manager.WriteChunk(session, 0, c.Request.ContentLength, io.TeeReader(c.Request.Body, hasher)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write upload"})
		return
	}
	uc.completeSession(c, session, hex.EncodeToString(hasher.Sum(nil)))
}

func (uc *UploadsController) ChecksumStatus(c *gin.Context) {
	checksum := strings.ToLower(c.Param("sha256"))
	if !sha256Pattern.MatchString(checksum) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SHA-256 checksum"})
		return
	}
	exists, err := uc.manager.HasChecksum(checksum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check upload checksum"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func NewUploadsController(manager *upload.Manager, moviesDir, musicDir string, importer *library.Importer, musicImporter *music.Importer) *UploadsController {
	return &UploadsController{manager: manager, moviesDir: moviesDir, musicDir: musicDir, importer: importer, musicImporter: musicImporter}
}

type createUploadRequest struct {
	Filename  string `json:"filename" binding:"required"`
	SizeBytes int64  `json:"size_bytes" binding:"required"`
}

func (uc *UploadsController) Create(c *gin.Context) {
	var req createUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename and size_bytes are required"})
		return
	}
	if _, ok := mediatype.Detect(req.Filename); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized file extension"})
		return
	}

	session, err := uc.manager.Create(req.Filename, req.SizeBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"upload_id":        session.ID,
		"chunk_size_bytes": upload.DefaultChunkSizeBytes,
	})
}

func (uc *UploadsController) Status(c *gin.Context) {
	session, err := uc.manager.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"received_bytes":   session.ReceivedBytes,
		"total_size_bytes": session.TotalSizeBytes,
		"status":           session.Status,
	})
}

// Chunk appends a byte-range chunk at ?offset=N. The offset must equal the
// session's current received_bytes; on mismatch the client should GET the
// session to resync and retry from the correct offset.
func (uc *UploadsController) Chunk(c *gin.Context) {
	session, err := uc.manager.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	offset, err := strconv.ParseInt(c.Query("offset"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset query param must be an integer"})
		return
	}

	length := c.Request.ContentLength
	if length <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Length header is required"})
		return
	}

	received, err := uc.manager.WriteChunk(session, offset, length, c.Request.Body)
	if err != nil {
		if errors.Is(err, upload.ErrOffsetMismatch) {
			c.JSON(http.StatusConflict, gin.H{"error": "offset mismatch", "received_bytes": received})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write chunk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received_bytes": received})
}

func (uc *UploadsController) Complete(c *gin.Context) {
	session, err := uc.manager.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	uc.completeSession(c, session, "")
}

func (uc *UploadsController) completeSession(c *gin.Context, session *models.UploadSession, checksum string) {
	kind, ok := mediatype.Detect(session.OriginalFilename)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized file extension"})
		return
	}

	if checksum == "" {
		var err error
		checksum, err = uc.manager.Checksum(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to checksum upload"})
			return
		}
	}
	claimed, err := uc.manager.ClaimChecksum(session, checksum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record upload checksum"})
		return
	}
	if !claimed {
		if err := uc.manager.Abort(session); err != nil {
			log.Printf("duplicate upload: cleaning staging file failed: %v", err)
		}
		c.JSON(http.StatusConflict, gin.H{"error": "this file has already been uploaded", "checksum": checksum})
		return
	}

	destDir := uc.moviesDir
	if kind == mediatype.KindMusic {
		destDir = uc.musicDir
	}

	destPath, err := uc.manager.Complete(session, destDir)
	if err != nil {
		if errors.Is(err, upload.ErrIncomplete) {
			c.JSON(http.StatusConflict, gin.H{"error": "upload is not fully received yet"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete upload"})
		return
	}

	switch kind {
	case mediatype.KindMovie:
		go func(checksum string) {
			relpath, err := filepath.Rel(uc.moviesDir, destPath)
			if err != nil {
				log.Printf("upload complete: computing relpath failed: %v", err)
				return
			}
			if _, err := uc.importer.ImportNewFile(context.Background(), relpath, destPath); err != nil {
				log.Printf("upload complete: import failed for %q: %v", relpath, err)
				return
			}
			uc.manager.SetMovieChecksum(relpath, checksum)
		}(checksum)
	case mediatype.KindMusic:
		go func() {
			relpath, err := filepath.Rel(uc.musicDir, destPath)
			if err != nil {
				log.Printf("upload complete: computing relpath failed: %v", err)
				return
			}
			if _, err := uc.musicImporter.ImportNewFile(context.Background(), relpath, destPath); err != nil {
				log.Printf("upload complete: music import failed for %q: %v", relpath, err)
			}
		}()
	}

	c.JSON(http.StatusAccepted, gin.H{"kind": string(kind), "status": "importing"})
}
