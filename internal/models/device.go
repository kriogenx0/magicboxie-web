package models

import "time"

// Device tracks a magicboxie-device Pi that has checked in via
// POST /devices/register. There is no pairing/approval step -- any device
// that knows the server's address can register -- so this exists purely for
// visibility (e.g. confirming a device really is checking in), not access
// control.
type Device struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   string    `gorm:"not null;uniqueIndex" json:"device_id"`
	LastSeenAt time.Time `gorm:"not null" json:"last_seen_at"`
}
