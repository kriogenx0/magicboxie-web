package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"magicbox/internal/models"
)

type VideosController struct {
	db        *gorm.DB
	moviesDir string
}

func NewVideosController(db *gorm.DB, moviesDir string) *VideosController {
	return &VideosController{db: db, moviesDir: moviesDir}
}

// Stream serves a movie's playable file with Range support at Jellyfin's
// conventional /Videos/{itemId}/stream path. Used by the web <video> tag,
// AVPlayer (magicbox-appletv), and any client downloading the file directly
// (magicbox-device's home-sync, magicbox-ios).
func (vc *VideosController) Stream(c *gin.Context) {
	id, ok := parseItemID(c.Param("itemId"), "movie")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var movie models.Movie
	if err := vc.db.First(&movie, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if movie.PlayableRelpath == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "movie is not ready to stream yet"})
		return
	}

	fullPath := filepath.Join(vc.moviesDir, movie.PlayableRelpath)
	// Use the currently playable file's own extension for Content-Type
	// detection -- OriginalFilename is a historical record and may have a
	// stale extension (e.g. .mkv) once a transcode has replaced the file.
	ServeFileRange(c, fullPath, filepath.Base(movie.PlayableRelpath))
}
