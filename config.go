// Load configuration from config.json.
package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	MysqlUser           string `json:"mysql_user"`
	MysqlPassword       string `json:"mysql_password"`
	MysqlHost           string `json:"mysql_host"`
	MysqlPort           int    `json:"mysql_port"`
	Parallel            int    `json:"parallel"`
	LocalBackupCount    int    `json:"local_backup_count"`
	DefaultRCloneRemote string `json:"default_rclone_remote"`
}

var config Config

func InitializeConfig() {
	configFile, err := os.ReadFile(configFileName)
	if err != nil {
		log.Fatalln("Failed to read config.toml from current directory:", err)
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
}
