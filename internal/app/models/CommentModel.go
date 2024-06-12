package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	Content    string `gorm:"type:text;not null" json:"content"`
	UserID     uint   `gorm:"not null" json:"user_id"`
	Login      string `gorm:"type:varchar(20);not null" json:"login"`
	PetitionID uint   `gorm:"not null" json:"petition_id"`
}
