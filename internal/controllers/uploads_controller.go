package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"magicbox/internal/services/library"
	"magicbox/internal/services/mediatype"
	"magicbox/internal/services/music"
	"magicbox/internal/services/upload"
)

type UploadsController struct {
	manager       *upload.Manager
	moviesDir     string
	musicDir      string
	importer      *library.Importer
	musicImporter *music.Importer
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

	kind, ok := mediatype.Detect(session.OriginalFilename)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized file extension"})
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
		go func() {
			relpath, err := filepath.Rel(uc.moviesDir, destPath)
			if err != nil {
				log.Printf("upload complete: computing relpath failed: %v", err)
				return
			}
			if _, err := uc.importer.ImportNewFile(context.Background(), relpath, destPath); err != nil {
				log.Printf("upload complete: import failed for %q: %v", relpath, err)
			}
		}()
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
