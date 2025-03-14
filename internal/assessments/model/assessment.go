package models

import (
	"time"

	"gorm.io/gorm"
)

type Assessment struct {
	ID           uint               `json:"id" gorm:"primaryKey"`
	Title        string             `json:"title" gorm:"size:255;not null"`
	Subject      string             `json:"subject" gorm:"size:100;not null"`
	Description  string             `json:"description" gorm:"type:text"`
	Duration     int                `json:"duration" gorm:"not null"` // in minutes
	Status       string             `json:"status" gorm:"size:50;not null;default:Draft"`
	DueDate      *time.Time         `json:"dueDate" gorm:"type:date"`
	CreatedByID  uint               `json:"createdById" gorm:"not null"`
	CreatedBy    models.User        `json:"createdBy" gorm:"foreignKey:CreatedByID"`
	PassingScore float64            `json:"passingScore" gorm:"not null;default:70"`
	Questions    []Question         `json:"questions" gorm:"foreignKey:AssessmentID"`
	Settings     AssessmentSettings `json:"settings" gorm:"foreignKey:AssessmentID"`
	CreatedAt    time.Time          `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time          `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt     `json:"-" gorm:"index"`
}

type AssessmentSettings struct {
	ID                          uint      `json:"id" gorm:"primaryKey"`
	AssessmentID                uint      `json:"assessmentId" gorm:"uniqueIndex;not null"`
	RandomizeQuestions          bool      `json:"randomizeQuestions" gorm:"default:false"`
	ShowResults                 bool      `json:"showResults" gorm:"default:true"`
	AllowRetake                 bool      `json:"allowRetake" gorm:"default:false"`
	MaxAttempts                 int       `json:"maxAttempts" gorm:"default:1"`
	TimeLimitEnforced           bool      `json:"timeLimitEnforced" gorm:"default:true"`
	RequireWebcam               bool      `json:"requireWebcam" gorm:"default:false"`
	PreventTabSwitching         bool      `json:"preventTabSwitching" gorm:"default:false"`
	RequireIdentityVerification bool      `json:"requireIdentityVerification" gorm:"default:false"`
	CreatedAt                   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
