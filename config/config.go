package config

import "time"

type AppConfig struct {
	DisableThumbCache bool
	CacheDir          string
	UpdateInterval    time.Duration
	MaxWidth          int
	MaxHeight         int
}

var DefaultConfig = AppConfig{
	DisableThumbCache: false,
	CacheDir:          "geonow-cache",
	UpdateInterval:    time.Minute * 16,
	MaxWidth:          10000,
	MaxHeight:         10000,
}
