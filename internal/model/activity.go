package models

import "time"

type Activity struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"not null;index"`
	User         User      `json:"-" gorm:"foreignKey:UserID"`
	Action       string    `json:"action" gorm:"size:100;not null"` // LOGIN, ASSESSMENT_START, ASSESSMENT_SUBMIT, etc.
	AssessmentID *uint     `json:"assessmentId" gorm:"index"`
	Details      string    `json:"details" gorm:"type:text"`
	IPAddress    string    `json:"ipAddress" gorm:"size:50"`
	UserAgent    string    `json:"userAgent" gorm:"type:text"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

type SuspiciousActivity struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"not null;index"`
	User         User      `json:"-" gorm:"foreignKey:UserID"`
	AssessmentID uint      `json:"assessmentId" gorm:"not null;index"`
	AttemptID    uint      `json:"attemptId" gorm:"not null;index"`
	Type         string    `json:"type" gorm:"size:100;not null"` // TAB_SWITCHING, FACE_NOT_DETECTED, etc.
	Details      string    `json:"details" gorm:"type:text"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null"`
	Severity     string    `json:"severity" gorm:"size:50;not null;default:MEDIUM"` // LOW, MEDIUM, HIGH
	Reviewed     bool      `json:"reviewed" gorm:"not null;default:false"`
	ImageData    []byte    `json:"-" gorm:"type:bytea"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
