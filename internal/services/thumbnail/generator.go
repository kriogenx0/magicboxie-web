// Package thumbnail generates a fallback "poster" GIF from random frames of
// a movie when no TMDB match was found.
package thumbnail

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
)

const frameCount = 6

// Generate builds an animated GIF from frameCount random frames spread
// across the middle 80% of the video's runtime (avoiding black intro/credit
// sequences), each held for ~2 seconds, and writes the result to destPath.
func Generate(ctx context.Context, sourcePath string, durationSeconds float64, destPath string) error {
	if durationSeconds <= 0 {
		return fmt.Errorf("thumbnail: unknown duration for %q", sourcePath)
	}

	lo := durationSeconds * 0.10
	hi := durationSeconds * 0.90

	tmpDir, err := os.MkdirTemp("", "magicbox-thumb-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	for i := 0; i < frameCount; i++ {
		ts := lo + rand.Float64()*(hi-lo)
		framePath := filepath.Join(tmpDir, fmt.Sprintf("frame-%02d.png", i))
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-ss", fmt.Sprintf("%.2f", ts),
			"-i", sourcePath,
			"-vframes", "1",
			"-vf", "scale=342:-1",
			framePath,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("thumbnail: extracting frame %d: %w: %s", i, err, out)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	framePattern := filepath.Join(tmpDir, "frame-%02d.png")
	palettePath := filepath.Join(tmpDir, "palette.png")

	paletteCmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-framerate", "0.5",
		"-i", framePattern,
		"-vf", "palettegen",
		palettePath,
	)
	if out, err := paletteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("thumbnail: generating palette: %w: %s", err, out)
	}

	gifCmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-framerate", "0.5",
		"-i", framePattern,
		"-i", palettePath,
		"-lavfi", "paletteuse",
		destPath,
	)
	if out, err := gifCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("thumbnail: assembling gif: %w: %s", err, out)
	}

	return nil
}
