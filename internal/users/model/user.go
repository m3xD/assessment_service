package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"` // Password hash
	Role      string         `json:"role" gorm:"size:50;not null"`
	Status    string         `json:"status" gorm:"size:50;not null;default:Active"`
	Phone     string         `json:"phone" gorm:"size:20"`
	Address   string         `json:"address" gorm:"size:255"`
	LastLogin *time.Time     `json:"lastLogin" gorm:"type:timestamp"`
	CreatedAt time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"not null"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Token     string    `json:"-" gorm:"size:255;not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
