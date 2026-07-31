package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterSPA serves the embedded frontend, falling back to index.html for
// any unknown non-/api path so client-side routing survives a page refresh.
func RegisterSPA(router *gin.Engine, fsys http.FileSystem) {
	fileServer := http.FileServer(fsys)

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
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
