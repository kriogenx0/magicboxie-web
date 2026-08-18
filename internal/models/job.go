package models

import "time"

const (
	JobTypeTranscode = "transcode"
	JobTypeThumbnail = "thumbnail_images"

	JobStatusQueued    = "queued"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

type Job struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	MovieID         uint       `gorm:"not null;index" json:"movie_id"`
	Type            string     `gorm:"not null" json:"type"`
	Status          string     `gorm:"not null;default:queued" json:"status"`
	ProgressPercent float64    `gorm:"not null;default:0" json:"progress_percent"`
	LogTail         string     `json:"log_tail,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	CreatedAt       time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
}
