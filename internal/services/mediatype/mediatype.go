// Package mediatype classifies files as movies or music by extension, so the
// library scanner and the upload-complete handler agree on the same rules.
package mediatype

import (
	"path/filepath"
	"strings"
)

type Kind string

const (
	KindMovie Kind = "movie"
	KindMusic Kind = "music"
)

var movieExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".mov": true, ".avi": true,
	".m4v": true, ".webm": true, ".wmv": true, ".flv": true,
}

var musicExtensions = map[string]bool{
	".mp3": true, ".flac": true, ".m4a": true, ".aac": true, ".ogg": true, ".wav": true,
}

// Detect classifies a filename by extension. ok is false for unrecognized
// extensions.
func Detect(filename string) (kind Kind, ok bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	if movieExtensions[ext] {
		return KindMovie, true
	}
	if musicExtensions[ext] {
		return KindMusic, true
	}
	return "", false
}

func IsMovie(filename string) bool {
	k, ok := Detect(filename)
	return ok && k == KindMovie
}

func IsMusic(filename string) bool {
	k, ok := Detect(filename)
	return ok && k == KindMusic
}
