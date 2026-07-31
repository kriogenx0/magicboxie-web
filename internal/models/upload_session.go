package models

import "time"

const (
	UploadStatusInProgress = "in_progress"
	UploadStatusCompleted  = "completed"
	UploadStatusAborted    = "aborted"
)

type UploadSession struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	OriginalFilename string    `gorm:"not null" json:"original_filename"`
	TotalSizeBytes   int64     `gorm:"not null" json:"total_size_bytes"`
	ReceivedBytes    int64     `gorm:"not null;default:0" json:"received_bytes"`
	TempRelpath      string    `gorm:"not null" json:"-"`
	Status           string    `gorm:"not null;default:in_progress" json:"status"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}
