// Perform backup operations using xtrabackup.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Creates a full backup using xtrabackup.
func CreateFullBackup(backupTime time.Time) error {
	targetDir := FormatFullBackupDir(backupTime)
	err := os.MkdirAll(filepath.Dir(targetDir), 0755)
	if err != nil {
		return fmt.Errorf("failed to prepare full backup directory: %v", err)
	}

	log.Printf("Creating full backup %s\n", backupTime.Format(time.DateTime))
	output, err := RunSubprocess(
		"xtrabackup",
		"--backup",
		"--datadir=/var/lib/mysql",
		"--user="+config.MysqlUser,
		"--password="+config.MysqlPassword,
		"--host="+config.MysqlHost,
		"--port="+strconv.Itoa(config.MysqlPort),
		"--target-dir="+targetDir,
		"--parallel="+strconv.Itoa(config.Parallel),
		"--compress=zstd",
		"--compress-threads="+strconv.Itoa(config.Parallel))
	if err != nil {
		return fmt.Errorf("Failed to create full backup: %v, output: %s", err, output)
	}
	log.Printf("Full backup %s created successfully.\n", backupTime.Format(time.DateTime))
	return nil
}

// Creates an incremental backup using xtrabackup.
func CreateIncrementalBackup(backupTime time.Time, baseTrack DatabaseTrack) error {
	targetDir := FormatIncrementalBackupDir(backupTime)
	err := os.MkdirAll(filepath.Dir(targetDir), 0755)
	if err != nil {
		return fmt.Errorf("failed to prepare incremental backup directory: %v", err)
	}

	log.Printf("Creating incremental backup %s\n", backupTime.Format(time.DateTime))
	output, err := RunSubprocess(
		"xtrabackup",
		"--backup",
		"--datadir=/var/lib/mysql",
		"--user="+config.MysqlUser,
		"--password="+config.MysqlPassword,
		"--host="+config.MysqlHost,
		"--port="+strconv.Itoa(config.MysqlPort),
		"--target-dir="+targetDir,
		"--incremental-basedir="+baseTrack.GetBackupPath(),
		"--parallel="+strconv.Itoa(config.Parallel),
		"--compress=zstd",
		"--compress-threads="+strconv.Itoa(config.Parallel))
	if err != nil {
		return fmt.Errorf("Failed to create incremental backup: %v, output: %s", err, output)
	}
	log.Printf("Incremental backup %s on %s created successfully.\n", backupTime.Format(time.DateTime), baseTrack.BackupTime.Format(time.DateTime))
	return nil
}
