package library

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	yearRe = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

	// Tags commonly found after the title in scene-release-style filenames.
	// Matching starts at the year (if present) or at the first recognized
	// tag, and everything from that point on is discarded.
	junkTagRe = regexp.MustCompile(`(?i)\b(1080p|2160p|720p|480p|4k|uhd|hdr|bluray|blu-ray|bdrip|brrip|dvdrip|webrip|web-dl|webdl|hdtv|remux|x264|x265|h264|h265|hevc|aac|ac3|dts|truehd|atmos|repack|proper|extended|unrated|multi|dual)\b`)

	sepRe = regexp.MustCompile(`[._]+`)
	wsRe  = regexp.MustCompile(`\s+`)
)

// ParseFilename extracts a best-guess Title and Year from a movie filename,
// stripping the extension plus common release-group/quality/codec tags.
// Examples handled:
//
//	"Movie.Title.2015.1080p.BluRay.x264-GROUP.mkv" -> "Movie Title", 2015
//	"Movie Title (2015).mkv"                       -> "Movie Title", 2015
//	"Some.Movie.mkv"                                -> "Some Movie", 0
func ParseFilename(filename string) (title string, year int) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = sepRe.ReplaceAllString(name, " ")

	if m := yearRe.FindStringIndex(name); m != nil {
		if y, err := strconv.Atoi(name[m[0]:m[1]]); err == nil {
			year = y
		}
		name = name[:m[0]]
	} else if m := junkTagRe.FindStringIndex(name); m != nil {
		name = name[:m[0]]
	}

	name = strings.Trim(name, " -([")
	name = wsRe.ReplaceAllString(name, " ")
	title = strings.TrimSpace(name)

	return title, year
}
