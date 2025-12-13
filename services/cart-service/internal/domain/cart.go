package domain

import "time"

type CartItem struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	UserID    uint `gorm:"not null;index"`
	ProductID uint `gorm:"not null;index"`
	Quantity  int  `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
