package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"magicbox/internal/models"
)

type AudioController struct {
	db       *gorm.DB
	musicDir string
}

func NewAudioController(db *gorm.DB, musicDir string) *AudioController {
	return &AudioController{db: db, musicDir: musicDir}
}

// Stream serves a track's file with Range support at Jellyfin's
// conventional /Audio/{itemId}/stream path. Tracks are never transcoded
// (see internal/services/music), so this always serves the original file.
func (ac *AudioController) Stream(c *gin.Context) {
	id, ok := parseItemID(c.Param("itemId"), "track")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var track models.Track
	if err := ac.db.First(&track, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	fullPath := filepath.Join(ac.musicDir, track.FileRelpath)
	ServeFileRange(c, fullPath, filepath.Base(track.FileRelpath))
}
