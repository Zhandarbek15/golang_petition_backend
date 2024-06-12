package models

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	Login      string `gorm:"type:varchar(20);not null;index" json:"login"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	PetitionID uint   `gorm:"not null;index" json:"petition_id"`
}
