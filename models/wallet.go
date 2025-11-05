package models

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	Address string `gorm:"primaryKey"`
	Balance int32 `gorm:"not null;default:0"`
}