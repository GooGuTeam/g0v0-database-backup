// Load configuration from config.json.
package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Config struct {
	MysqlUser           string `json:"mysql_user"`
	MysqlPassword       string `json:"mysql_password"`
	MysqlHost           string `json:"mysql_host"`
	MysqlPort           int    `json:"mysql_port"`
	Parallel            int    `json:"parallel"`
	LocalBackupCount    int    `json:"local_backup_count"`
	DefaultRCloneRemote string `json:"default_rclone_remote"`

	FullBackupIntervalStr        string `json:"full_backup_interval"`
	IncrementalBackupIntervalStr string `json:"incremental_backup_interval"`
	CleanupIntervalStr           string `json:"cleanup_interval"`
	RcloneUploadIntervalStr      string `json:"rclone_upload_interval"`

	FullBackupInterval        time.Duration `json:"-"`
	IncrementalBackupInterval time.Duration `json:"-"`
	CleanupInterval           time.Duration `json:"-"`
	RcloneUploadInterval      time.Duration `json:"-"`
}

var config Config

func InitializeConfig() {
	configFile, err := os.ReadFile(configFileName)
	if err != nil {
		log.Fatalln("Failed to read config.json from current directory:", err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalln("Failed to parse config.json:", err)
	}

	if config.MysqlUser == "" {
		config.MysqlUser = defaultMysqlUser
	}
	if config.MysqlPassword == "" {
		config.MysqlPassword = defaultMysqlPassword
	}
	if config.MysqlHost == "" {
		config.MysqlHost = defaultMysqlHost
	}
	if config.MysqlPort == 0 {
		config.MysqlPort = defaultMysqlPort
	}
	if config.Parallel == 0 {
		config.Parallel = defaultParallel
	}
	if config.LocalBackupCount == 0 {
		config.LocalBackupCount = defaultLocalBackupCount
	}
	if config.DefaultRCloneRemote == "" {
		config.DefaultRCloneRemote = defaultRCloneRemote
	}

	if config.FullBackupIntervalStr != "" {
		config.FullBackupInterval, err = time.ParseDuration(config.FullBackupIntervalStr)
		if err != nil {
			log.Fatalf("Invalid full_backup_interval: %v", err)
		}
	} else {
		config.FullBackupInterval = defaultFullBackupInterval
	}

	if config.IncrementalBackupIntervalStr != "" {
		config.IncrementalBackupInterval, err = time.ParseDuration(config.IncrementalBackupIntervalStr)
		if err != nil {
			log.Fatalf("Invalid incremental_backup_interval: %v", err)
		}
	} else {
		config.IncrementalBackupInterval = defaultIncrementalBackupInterval
	}

	if config.CleanupIntervalStr != "" {
		config.CleanupInterval, err = time.ParseDuration(config.CleanupIntervalStr)
		if err != nil {
			log.Fatalf("Invalid cleanup_interval: %v", err)
		}
	} else {
		config.CleanupInterval = defaultCleanupInterval
	}

	if config.RcloneUploadIntervalStr != "" {
		config.RcloneUploadInterval, err = time.ParseDuration(config.RcloneUploadIntervalStr)
		if err != nil {
			log.Fatalf("Invalid rclone_upload_interval: %v", err)
		}
	} else {
		config.RcloneUploadInterval = defaultRcloneUploadInterval
	}
}
