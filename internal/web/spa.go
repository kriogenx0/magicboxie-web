package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterSPA serves the embedded frontend, falling back to index.html for
// unknown browser route so client-side routing survives a page refresh.
func RegisterSPA(router *gin.Engine, fsys http.FileSystem) {
	fileServer := http.FileServer(fsys)

	router.NoRoute(func(c *gin.Context) {
		if isAPIPath(c.Request.URL.Path) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		reqPath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if reqPath == "" {
			reqPath = "index.html"
		}

		if f, err := fsys.Open(reqPath); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

// isAPIPath prevents unknown Jellyfin endpoints from receiving index.html
// with a misleading 200 response. Generated clients otherwise attempt to
// decode the HTML as JSON and surface an opaque data-format error.
func isAPIPath(path string) bool {
	for _, prefix := range []string{
		"/api/", "/Audio/", "/Branding/", "/Devices/", "/DisplayPreferences/",
		"/Items/", "/Library/", "/LiveTv/", "/MediaSegments/", "/MusicGenres/",
		"/Persons/", "/Playlists/", "/QuickConnect/", "/Sessions/", "/Shows/",
		"/System/", "/UserImage", "/Users/", "/UserViews", "/Videos/", "/socket",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
