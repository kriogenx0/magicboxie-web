// Package web embeds the built React frontend (web/dist) into the Go binary
// so a single binary can serve both the API and the static site.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// FileSystem returns an http.FileSystem rooted at the embedded dist/ directory.
func FileSystem() (http.FileSystem, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
