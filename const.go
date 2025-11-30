// Define constants used across the application.
package main

import "time"

const (
	defaultMysqlUser        = "backup"
	defaultMysqlPassword    = "password"
	defaultMysqlHost        = "localhost"
	defaultMysqlPort        = 3306
	defaultParallel         = 4
	defaultLocalBackupCount = 3
	defaultRCloneRemote     = "onedrive:"

	configFileName       = "config.json"
	sqliteDBPath         = "/data/data.db"
	backupPath           = "/backup/"
	downloadedBackupPath = "/downloaded_backup/"

	HttpPort = 32400

	fullBackupInterval        = 12 * time.Hour
	incrementalBackupInterval = 30 * time.Minute
	cleanupInterval           = 1 * time.Hour
	rcloneUploadInterval      = 15 * time.Minute
)
