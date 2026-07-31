package models

import "time"

type Track struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AlbumID         uint      `gorm:"not null;index" json:"album_id"`
	Title           string    `gorm:"not null" json:"title"`
	TrackNumber     int       `json:"track_number"`
	DiscNumber      int       `json:"disc_number"`
	DurationSeconds float64   `json:"duration_seconds"`
	FileRelpath     string    `gorm:"not null" json:"-"`
	Codec           string    `json:"codec"`
	Bitrate         int       `json:"bitrate"`
	AddedAt         time.Time `gorm:"not null;autoCreateTime" json:"added_at"`
}
