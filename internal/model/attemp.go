package models

import (
	"time"

	"gorm.io/gorm"
)

type Attempt struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"userId" gorm:"not null;index"`
	User         User           `json:"-" gorm:"foreignKey:UserID"`
	AssessmentID uint           `json:"assessmentId" gorm:"not null;index"`
	Assessment   Assessment     `json:"-" gorm:"foreignKey:AssessmentID"`
	StartedAt    time.Time      `json:"startedAt" gorm:"not null"`
	EndedAt      *time.Time     `json:"endedAt"`
	SubmittedAt  *time.Time     `json:"submittedAt"`
	Score        *float64       `json:"score"`
	Duration     *int           `json:"duration"`                                           // in minutes
	Status       string         `json:"status" gorm:"size:50;not null;default:In Progress"` // In Progress, Completed, Expired
	Answers      []Answer       `json:"answers" gorm:"foreignKey:AttemptID"`
	CreatedAt    time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Feedback     string         `json:"feedback"`
}

type Answer struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	AttemptID  uint      `json:"attemptId" gorm:"not null;index"`
	QuestionID uint      `json:"questionId" gorm:"not null"`
	Answer     string    `json:"answer" gorm:"type:text"`
	IsCorrect  *bool     `json:"isCorrect"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type AttemptUpdateDTO struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
	Answers  []struct {
		ID        uint `json:"id"`
		IsCorrect bool `json:"isCorrect"`
	}
}
