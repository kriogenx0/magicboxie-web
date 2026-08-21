package library

import (
	"context"
	"os"
	"path/filepath"

	"magicboxie/internal/services/tmdb"
)

// cacheImage downloads a TMDB image (skipping if tmdbImagePath is empty) and
// writes it to destPath, creating parent directories as needed. Images are
// immutable once matched, so callers only need to do this once per movie.
func cacheImage(ctx context.Context, client *tmdb.Client, size, tmdbImagePath, destPath string) error {
	if tmdbImagePath == "" {
		return nil
	}
	data, err := client.DownloadImage(ctx, size, tmdbImagePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0o644)
}
