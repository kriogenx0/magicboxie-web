// Package music imports audio files into the library by reading their own
// embedded tags (ID3v2/MP4/Vorbis comments) rather than an external API --
// unlike movies, real-world music files almost always already carry
// accurate title/artist/album/track metadata and cover art.
package music

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"gorm.io/gorm"

	"magicbox/internal/models"
	"magicbox/internal/services/media"
	"magicbox/internal/services/mediatype"
)

type Importer struct {
	db       *gorm.DB
	musicDir string
	dataDir  string
}

func NewImporter(db *gorm.DB, musicDir, dataDir string) *Importer {
	return &Importer{db: db, musicDir: musicDir, dataDir: dataDir}
}

// Scan walks musicDir for audio files not already known (by FileRelpath)
// and imports each one. Errors on individual files are logged and skipped
// rather than aborting the whole scan.
func (im *Importer) Scan(ctx context.Context) (imported int, err error) {
	err = filepath.WalkDir(im.musicDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !mediatype.IsMusic(d.Name()) {
			return nil
		}

		relpath, err := filepath.Rel(im.musicDir, path)
		if err != nil {
			return nil
		}

		var count int64
		im.db.Model(&models.Track{}).Where("file_relpath = ?", relpath).Count(&count)
		if count > 0 {
			return nil
		}

		if err := im.importFile(ctx, relpath, path); err != nil {
			log.Printf("music: failed to import %q: %v", relpath, err)
			return nil
		}
		imported++
		return nil
	})
	return imported, err
}

// ImportNewFile imports absPath (relpath under musicDir) if it isn't
// already known. Used by the chunked-upload completion handler.
func (im *Importer) ImportNewFile(ctx context.Context, relpath, absPath string) (bool, error) {
	var count int64
	im.db.Model(&models.Track{}).Where("file_relpath = ?", relpath).Count(&count)
	if count > 0 {
		return false, nil
	}
	if err := im.importFile(ctx, relpath, absPath); err != nil {
		return false, err
	}
	return true, nil
}

func (im *Importer) importFile(ctx context.Context, relpath, absPath string) error {
	var artistName, albumTitle, title string
	var trackNum, discNum, year int
	var coverData []byte

	if f, err := os.Open(absPath); err == nil {
		if m, terr := tag.ReadFrom(f); terr == nil {
			artistName = m.Artist()
			albumTitle = m.Album()
			title = m.Title()
			trackNum, _ = m.Track()
			discNum, _ = m.Disc()
			year = m.Year()
			if pic := m.Picture(); pic != nil {
				coverData = pic.Data
			}
		}
		f.Close()
	}

	if artistName == "" {
		artistName = "Unknown Artist"
	}
	if albumTitle == "" {
		albumTitle = "Unknown Album"
	}
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
	}

	info, err := media.Probe(ctx, absPath)
	if err != nil {
		return fmt.Errorf("probing audio file: %w", err)
	}

	var artist models.Artist
	if err := im.db.Where("name = ?", artistName).
		FirstOrCreate(&artist, models.Artist{Name: artistName}).Error; err != nil {
		return fmt.Errorf("finding/creating artist: %w", err)
	}

	var album models.Album
	if err := im.db.Where("artist_id = ? AND title = ?", artist.ID, albumTitle).
		FirstOrCreate(&album, models.Album{ArtistID: artist.ID, Title: albumTitle, Year: year}).Error; err != nil {
		return fmt.Errorf("finding/creating album: %w", err)
	}

	if album.CoverPath == "" && len(coverData) > 0 {
		if err := im.cacheAlbumCover(album, coverData); err != nil {
			log.Printf("music: caching cover art for album %d failed: %v", album.ID, err)
		}
	}

	track := models.Track{
		AlbumID:         album.ID,
		Title:           title,
		TrackNumber:     trackNum,
		DiscNumber:      discNum,
		DurationSeconds: info.DurationSeconds,
		FileRelpath:     relpath,
		Codec:           info.AudioCodec,
		Bitrate:         info.BitrateBps,
	}
	return im.db.Create(&track).Error
}

func (im *Importer) cacheAlbumCover(album models.Album, data []byte) error {
	relpath := fmt.Sprintf("albums/%d.jpg", album.ID)
	fullPath := filepath.Join(im.dataDir, "images", relpath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return err
	}
	return im.db.Model(&models.Album{}).Where("id = ?", album.ID).Update("cover_path", relpath).Error
}
