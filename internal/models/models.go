package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"type:varchar(100);not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Wallet struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex;not null"`
	Balance   int64     `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Transaction struct {
	ID         uint      `gorm:"primaryKey"`
	SenderID   uint      `gorm:"index;not null"`
	ReceiverID uint      `gorm:"index;not null"`
	Sender     User      `gorm:"foreignKey:SenderID;references:ID"`
	Receiver   User      `gorm:"foreignKey:ReceiverID;references:ID"`
	Amount     int64     `gorm:"not null"`
	Status     string    `gorm:"type:varchar(20);not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type RevokedToken struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"type:text;uniqueIndex;not null"`
	RevokedAt time.Time `gorm:"autoCreateTime"`
}
