package models

type Artist struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null;uniqueIndex" json:"name"`
	ImagePath string `json:"image_path"`
}
