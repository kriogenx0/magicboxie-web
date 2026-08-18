package models

import "time"

// Movie statuses.
const (
	MovieStatusPending        = "pending"
	MovieStatusProbing        = "probing"
	MovieStatusNeedsTranscode = "needs_transcode"
	MovieStatusTranscoding    = "transcoding"
	MovieStatusReady          = "ready"
	MovieStatusError          = "error"
)

type Movie struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Title             string    `gorm:"not null" json:"title"`
	Year              int       `json:"year"`
	TMDBID            *int      `json:"tmdb_id"`
	Overview          string    `json:"overview"`
	RuntimeMinutes    int       `json:"runtime_minutes"`
	GenresJSON        string    `json:"-"`
	CastJSON          string    `json:"-"`
	PosterPath        string    `json:"poster_path"`
	PosterIsGenerated bool      `json:"poster_is_generated"`
	BackdropPath      string    `json:"backdrop_path"`
	OriginalFilename  string    `gorm:"not null" json:"original_filename"`
	SourceRelpath     string    `gorm:"not null;uniqueIndex" json:"-"` // relpath under movies_dir at time of discovery; kept for scan dedup even after the original file is deleted post-transcode
	PlayableRelpath   string    `json:"-"`
	FileSizeBytes     int64     `json:"file_size_bytes"`
	DurationSeconds   float64   `json:"duration_seconds"`
	VideoCodec        string    `json:"video_codec"`
	AudioCodec        string    `json:"audio_codec"`
	Container         string    `json:"container"`
	Status            string    `gorm:"not null;default:pending" json:"status"`
	NeedsReview       bool      `json:"needs_review"`
	SyncEnabled       bool      `json:"sync_enabled"`
	ErrorMessage      string    `json:"error_message,omitempty"`
	AddedAt           time.Time `gorm:"not null;autoCreateTime" json:"added_at"`
	UpdatedAt         time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}
