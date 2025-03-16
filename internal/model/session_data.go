package models

import "time"

type SessionData struct {
	UserID       uint
	AssessmentID uint
	Action       string
	Timestamp    time.Time
	UserAgent    string
	Details      string
}
