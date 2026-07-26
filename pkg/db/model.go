package db

import (
	"time"

	"gorm.io/gorm"
)

// Plugin 插件
type Plugin struct {
	ID           int64  `gorm:"primaryKey"`
	Name         string `gorm:"size:255;index"`
	AssemblyName string `gorm:"size:255;uniqueIndex"`
	Description  string
	// Version 最新版本号
	Version       string `gorm:"size:64"`
	DownloadCount int64  `gorm:"not null;default:0"`
	RatingCount   int64  `gorm:"not null;default:0"`
	RatingScore   float64 `gorm:"not null;default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// PluginVersion 插件版本
type PluginVersion struct {
	ID        int64  `gorm:"primaryKey"`
	PluginID  int64  `gorm:"uniqueIndex:idx_plugin_version"`
	Version   string `gorm:"size:64;uniqueIndex:idx_plugin_version"`
	FilePath  string `gorm:"size:512"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PluginRating 插件评分(用IP区分用户)
type PluginRating struct {
	ID        int64  `gorm:"primaryKey"`
	PluginID  int64  `gorm:"uniqueIndex:idx_plugin_ip"`
	IP        string `gorm:"size:64;uniqueIndex:idx_plugin_ip"`
	Score     int32  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
