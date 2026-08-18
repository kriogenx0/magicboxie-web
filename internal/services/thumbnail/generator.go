// Package thumbnail generates still-frame poster candidates from a movie when
// no TMDB match was found.
package thumbnail

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var candidatePositions = []float64{0.10, 0.30, 0.50, 0.70, 0.90}

type Candidate struct {
	Index            int
	TimestampSeconds float64
	Path             string
}

// GenerateCandidates extracts a few deterministic JPEG frames spread across
// the movie. Each file is kept so the user can choose the best poster later.
func GenerateCandidates(ctx context.Context, sourcePath string, durationSeconds float64, destDir string) ([]Candidate, error) {
	if durationSeconds <= 0 {
		return nil, fmt.Errorf("thumbnail: unknown duration for %q", sourcePath)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}

	candidates := make([]Candidate, 0, len(candidatePositions))
	for i, position := range candidatePositions {
		timestamp := durationSeconds * position
		path := filepath.Join(destDir, fmt.Sprintf("%d.jpg", i))
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-y", "-v", "error",
			"-ss", fmt.Sprintf("%.2f", timestamp),
			"-i", sourcePath,
			"-frames:v", "1",
			"-vf", "scale=342:-1",
			"-q:v", "2",
			path,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("thumbnail: extracting candidate %d: %w: %s", i, err, out)
		}
		candidates = append(candidates, Candidate{Index: i, TimestampSeconds: timestamp, Path: path})
	}

	return candidates, nil
}
