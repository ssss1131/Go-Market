package domain

import (
	"time"
)

type Product struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`
	UserID      uint    `gorm:"not null"`
	Name        string  `gorm:"size:255;not null"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Categories  []Category `gorm:"many2many:product_categories;"`
}

type Category struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:255;not null"`
	Slug      string `gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time
}
