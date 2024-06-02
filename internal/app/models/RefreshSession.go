package models

import "time"

type RefreshSession struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"type:char(36);not null"`
	RefreshToken string    `gorm:"type:text;not null"`
	UA           string    `gorm:"type:varchar(200);not null"`
	IP           string    `gorm:"type:varchar(15);not null"`
	ExpiresIn    int64     `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;"`
}
