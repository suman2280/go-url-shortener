package domain

import (
	"time"

	"gorm.io/gorm"
)

type UrlMapping struct {
	ShortCode     string         `gorm:"primaryKey;size:10;uniqueIndex"`
	LongUrl       string         `gorm:"not null;size:2048"`
	ClickCount    int64          `gorm:"default:0"`
	ExpiresAt     *time.Time     `gorm:"index"`
	IsCustomAlias bool           `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (UrlMapping) TableName() string {
	return "url_mappings"
}
