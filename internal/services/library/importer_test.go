package library

import (
	"context"
	"testing"

	"magicbox/internal/db"
	"magicbox/internal/models"
	"magicbox/internal/services/tmdb"
)

func TestApplyTMDBDetailsPrefersTMDBTitle(t *testing.T) {
	im := &Importer{}
	movie := &models.Movie{Title: "Filename Title"}
	details := &tmdb.MovieDetails{ID: 123, Title: "Official TMDB Title"}

	im.applyTMDBDetails(context.Background(), movie, details)

	if movie.Title != details.Title {
		t.Fatalf("movie title = %q, want %q", movie.Title, details.Title)
	}
}

func TestApplyTMDBDetailsKeepsTitleWhenTMDBTitleIsEmpty(t *testing.T) {
	im := &Importer{}
	movie := &models.Movie{Title: "Filename Title"}
	details := &tmdb.MovieDetails{ID: 123, Title: "  "}

	im.applyTMDBDetails(context.Background(), movie, details)

	if movie.Title != "Filename Title" {
		t.Fatalf("movie title = %q, want original title", movie.Title)
	}
}

// RetryStuck exists because a movie interrupted mid-import (crash, restart)
// is left at "probing" forever otherwise: ImportNewFile only ever looks at
// files with no row at all, so a stuck row is invisible to a plain rescan.
func TestRetryStuckSkipsAMovieWhoseFileNoLongerExists(t *testing.T) {
	tmpDir := t.TempDir()
	database, err := db.Open(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	im := NewImporter(database, tmpDir, tmpDir, tmdb.NewClient(""))

	stuck := &models.Movie{
		Title:         "Ghost File",
		SourceRelpath: "does-not-exist.mp4",
		Status:        models.MovieStatusProbing,
	}
	if err := database.Create(stuck).Error; err != nil {
		t.Fatalf("creating stuck movie: %v", err)
	}

	retried, err := im.RetryStuck(context.Background())
	if err != nil {
		t.Fatalf("RetryStuck: %v", err)
	}
	if retried != 0 {
		t.Fatalf("retried = %d, want 0 (file doesn't exist, nothing to retry against)", retried)
	}

	var reloaded models.Movie
	if err := database.First(&reloaded, stuck.ID).Error; err != nil {
		t.Fatalf("reloading movie: %v", err)
	}
	if reloaded.Status != models.MovieStatusProbing {
		t.Fatalf("status = %q, want unchanged %q", reloaded.Status, models.MovieStatusProbing)
	}
}

func TestRetryStuckIgnoresAMovieThatsAlreadyReady(t *testing.T) {
	tmpDir := t.TempDir()
	database, err := db.Open(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	im := NewImporter(database, tmpDir, tmpDir, tmdb.NewClient(""))

	ready := &models.Movie{Title: "Fine", SourceRelpath: "fine.mp4", Status: models.MovieStatusReady}
	if err := database.Create(ready).Error; err != nil {
		t.Fatalf("creating ready movie: %v", err)
	}

	retried, err := im.RetryStuck(context.Background())
	if err != nil {
		t.Fatalf("RetryStuck: %v", err)
	}
	if retried != 0 {
		t.Fatalf("retried = %d, want 0 (nothing was stuck)", retried)
	}
}
