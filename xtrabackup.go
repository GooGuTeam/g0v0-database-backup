// Perform backup operations using xtrabackup.
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// Creates a full backup using xtrabackup.
func CreateFullBackup(backupTime time.Time) error {
	log.Printf("Creating full backup %s\n", backupTime.Format(time.DateTime))
	output, err := RunSubprocess(
		"xtrabackup",
		"--backup",
		"--datadir=/var/lib/mysql",
		"--user="+config.MysqlUser,
		"--password="+config.MysqlPassword,
		"--host="+config.MysqlHost,
		"--port="+strconv.Itoa(config.MysqlPort),
		"--target-dir="+FormatFullBackupDir(backupTime),
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
func CreateIncrementalBackup(backupTime time.Time, lastBackupTime time.Time) error {
	log.Printf("Creating incremental backup %s\n", backupTime.Format(time.DateTime))
	output, err := RunSubprocess(
		"xtrabackup",
		"--backup",
		"--datadir=/var/lib/mysql",
		"--user="+config.MysqlUser,
		"--password="+config.MysqlPassword,
		"--host="+config.MysqlHost,
		"--port="+strconv.Itoa(config.MysqlPort),
		"--target-dir="+FormatIncrementalBackupDir(backupTime),
		"--incremental-basedir="+FormatFullBackupDir(lastBackupTime),
		"--parallel="+strconv.Itoa(config.Parallel),
		"--compress=zstd",
		"--compress-threads="+strconv.Itoa(config.Parallel))
	if err != nil {
		return fmt.Errorf("Failed to create incremental backup: %v, output: %s", err, output)
	}
	log.Printf("Incremental backup %s on %s created successfully.\n", backupTime.Format(time.DateTime), lastBackupTime.Format(time.DateTime))
	return nil
}
