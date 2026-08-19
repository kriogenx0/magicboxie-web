package db

import (
	"fmt"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"magicbox/internal/models"
)

// Open opens (creating if necessary) the SQLite database under dataDir and
// runs AutoMigrate for all known models.
func Open(dataDir string) (*gorm.DB, error) {
	dsn := filepath.Join(dataDir, "magicbox.sqlite")

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database at %q: %w", dsn, err)
	}

	if err := db.AutoMigrate(
		&models.Movie{},
		&models.Job{},
		&models.UploadSession{},
		&models.MediaChecksum{},
		&models.Artist{},
		&models.Album{},
		&models.Track{},
		&models.Device{},
	); err != nil {
		return nil, fmt.Errorf("running auto-migrations: %w", err)
	}

	return db, nil
}
