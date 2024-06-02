package models

import (
	"gorm.io/gorm"
	"time"
)

type UserModel struct {
	gorm.Model
	Login     string    `gorm:"type:varchar(20);unique;not null" json:"login" binding:"required"`
	Password  string    `gorm:"type:varchar(255);not null" json:"password" binding:"required"`
	Role      string    `gorm:"type:varchar(20);not null" json:"role" binding:"required,oneof=User Admin"`
	FirstName string    `gorm:"type:varchar(20);not null" json:"first_name"`
	LastName  string    `gorm:"type:varchar(20);not null" json:"last_name"`
	Email     string    `gorm:"type:varchar(50);not null" json:"email" binding:"email"`
	BirthDate time.Time `gorm:"type:date;not null" json:"birth_date"`
	Status    string    `gorm:"type:varchar(20);not null" json:"status" binding:"oneof=Active Passive"`
}

type UserUpdate struct {
	Login     string    `json:"login" binding:"omitempty"`
	Password  string    `json:"password" binding:"omitempty"`
	Role      string    `json:"role" binding:"omitempty,oneof=User Admin"`
	FirstName string    `json:"first_name" binding:"omitempty"`
	LastName  string    `json:"last_name" binding:"omitempty"`
	Email     string    `json:"email" binding:"email" binding:"omitempty"`
	BirthDate time.Time `json:"birth_date" binding:"omitempty"`
	Status    string    `json:"status" binding:"omitempty,oneof=Active Passive"`
}
