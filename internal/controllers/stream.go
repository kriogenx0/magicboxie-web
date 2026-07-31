package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ServeFileRange serves filePath with full HTTP Range support via the
// stdlib's http.ServeContent, which needs only an io.ReadSeeker + modtime to
// emit 206 Partial Content responses. This is the mechanism behind every
// streaming/download endpoint (movies, and tracks from M7 onward).
func ServeFileRange(c *gin.Context, filePath, downloadName string) {
	f, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stat file"})
		return
	}

	if c.Query("download") == "1" {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	}

	http.ServeContent(c.Writer, c.Request, downloadName, stat.ModTime(), f)
}

// serveImageFile serves a cached image file (poster/backdrop/album art)
// with long-lived caching -- these are immutable once cached (see
// internal/services/library/images.go), so there's no need to revalidate.
func serveImageFile(c *gin.Context, fullPath string) {
	if _, err := os.Stat(fullPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(fullPath)
}
