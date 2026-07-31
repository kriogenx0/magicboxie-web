package models

type Album struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ArtistID  uint   `gorm:"not null;index;uniqueIndex:idx_artist_title" json:"artist_id"`
	Title     string `gorm:"not null;uniqueIndex:idx_artist_title" json:"title"`
	Year      int    `json:"year"`
	CoverPath string `json:"cover_path"`
}
