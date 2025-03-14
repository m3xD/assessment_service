package models

import (
	"time"

	"gorm.io/gorm"
)

type Question struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	AssessmentID  uint             `json:"assessmentId" gorm:"not null;index"`
	Type          string           `json:"type" gorm:"size:50;not null"` // multiple-choice, true-false, essay
	Text          string           `json:"text" gorm:"type:text;not null"`
	Options       []QuestionOption `json:"options" gorm:"foreignKey:QuestionID"`
	CorrectAnswer string           `json:"correctAnswer" gorm:"size:255"`
	Points        float64          `json:"points" gorm:"not null;default:1"`
	CreatedAt     time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt   `json:"-" gorm:"index"`
}

type QuestionOption struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	QuestionID uint      `json:"questionId" gorm:"not null;index"`
	OptionID   string    `json:"optionId" gorm:"size:50;not null"` // a, b, c, d or custom identifier
	Text       string    `json:"text" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
