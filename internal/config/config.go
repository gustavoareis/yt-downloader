// Package config holds application-wide constants and configuration values.
package config

const (
	AppName    = "YouTube Downloader"
	AppVersion = "1.0.0"

	// DownloadsDir is the output folder, relative to the working directory (project root).
	DownloadsDir = "downloads"

	// UI limits.
	MaxLogLines   = 12
	MaxLineLength = 80
	LogBoxWidth   = 62
	URLDisplayLen = 55
	URLTruncLen   = 60
)
