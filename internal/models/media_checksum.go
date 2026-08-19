package models

import "time"

// MediaChecksum is a content-addressed record of every completed upload.
// SHA-256 is the primary key so the database atomically rejects concurrent
// attempts to upload identical bytes, even when filenames differ.
type MediaChecksum struct {
	SHA256           string    `gorm:"primaryKey;size:64"`
	OriginalFilename string    `gorm:"not null"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime"`
}
