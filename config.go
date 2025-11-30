// Load configuration from config.json.
package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	MysqlUser           string
	MysqlPassword       string
	MysqlHost           string
	MysqlPort           int
	Parallel            int
	LocalBackupCount    int
	DefaultRCloneRemote string
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
