package domain

import (
	"time"
)

type Product struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`
	Name        string  `gorm:"size:255;not null"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
