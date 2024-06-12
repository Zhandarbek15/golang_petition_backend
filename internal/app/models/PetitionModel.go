package models

import "gorm.io/gorm"

type Petition struct {
	gorm.Model
	UserID       uint   `gorm:"not null" json:"user_id"`
	Title        string `gorm:"type:varchar(100);not null" json:"title"`
	Description  string `gorm:"type:text;not null" json:"description"`
	TargetByVote uint   `gorm:"type:int;not null" json:"target_by_vote"`
	CurrentVotes uint   `gorm:"type:int;not null" json:"current_votes"`
	Recipient    string `gorm:"type:varchar(100);" json:"recipient"`
}

type PetitionUpdate struct {
	Title        string `json:"title" binding:"omitempty"`
	Description  string `json:"description" binding:"omitempty"`
	TargetByVote uint   `json:"target_by_vote" binding:"omitempty"`
	CurrentVotes uint   `json:"current_votes" binding:"omitempty"`
	Recipient    string `json:"recipient" binding:"omitempty"`
}
