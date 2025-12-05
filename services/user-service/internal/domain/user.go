package domain

import (
	"time"
)

type UserStatus string

const (
	StatusPending UserStatus = "PENDING"
	StatusActive  UserStatus = "ACTIVE"
)

type User struct {
	ID                uint       `gorm:"primaryKey;autoIncrement"`
	Name              string     `gorm:"size:255;not null"`
	Surname           string     `gorm:"size:255;not null"`
	Email             string     `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash      string     `gorm:"size:255;not null"`
	Status            UserStatus `gorm:"type:text;not null;default:PENDING"`
	VerificationToken string     `gorm:"size:64"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
