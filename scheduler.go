// Schedule periodic tasks such as full backups, incremental backups, cleanup, and remote uploads using cron.
package main

import (
	"log"
	"time"
)

var (
	fullBackupTicker        *time.Ticker
	incrementalBackupTicker *time.Ticker
	cleanupTicker           *time.Ticker
	rcloneUploadTicker      *time.Ticker
)

func fullBackupJob() {
	log.Println("Starting scheduled full backup...")
	err := PerformFullBackup("", "Scheduled full backup")
	if err != nil {
		log.Printf("Scheduled full backup failed: %v", err)
	} else {
		log.Println("Scheduled full backup completed successfully.")
	}
}
func incrementalBackupJob() {
	log.Println("Starting scheduled incremental backup...")
	err := PerformIncrementalBackup("", "Scheduled incremental backup")
	if err != nil {
		log.Printf("Scheduled incremental backup failed: %v", err)
	} else {
		log.Println("Scheduled incremental backup completed successfully.")
	}
}
func cleanupJob() {
	log.Println("Starting scheduled cleanup of old backups...")
	backups, err := tracker.GetOldBackups()
	if err != nil {
		log.Printf("Scheduled cleanup failed: %v", err)
		return
	}
	for _, backup := range backups {
		err := DeleteLocalBackup(backup)
		if err != nil {
			log.Printf("Failed to delete local backup %s: %v", backup.GetBackupPath(), err)
		}
	}
}
func rcloneUploadJob() {
	log.Println("Starting scheduled rclone upload of pending backups...")
	backups, err := tracker.GetPendingUploads()
	if err != nil {
		log.Printf("Scheduled rclone upload failed: %v", err)
		return
	}
	for _, backup := range backups {
		err := UploadToRClone(backup.BackupTime, config.DefaultRCloneRemote, backup.IsIncrementalBackup())
		if err != nil {
			log.Printf("Failed to upload backup %s: %v", backup.GetBackupPath(), err)
			continue
		}
		tracker.UpdateBackupStatus(backup.BackupTime, Uploaded)
	}
}

// Initialize and start scheduled jobs.
func InitializeJobs() {
	fullBackupTicker = time.NewTicker(config.FullBackupInterval)
	incrementalBackupTicker = time.NewTicker(config.IncrementalBackupInterval)
	cleanupTicker = time.NewTicker(config.CleanupInterval)
	rcloneUploadTicker = time.NewTicker(config.RcloneUploadInterval)

	go func() {
		for range fullBackupTicker.C {
			fullBackupJob()
		}
	}()
	go func() {
		for range incrementalBackupTicker.C {
			incrementalBackupJob()
		}
	}()
	go func() {
		for range cleanupTicker.C {
			cleanupJob()
		}
	}()
	go func() {
		for range rcloneUploadTicker.C {
			rcloneUploadJob()
		}
	}()
}

// Stop all scheduled jobs.
func StopJobs() {
	fullBackupTicker.Stop()
	incrementalBackupTicker.Stop()
	cleanupTicker.Stop()
	rcloneUploadTicker.Stop()
}
