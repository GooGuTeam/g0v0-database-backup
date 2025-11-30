// Some high-level functions to perform backups, track their status, clean old backups, and upload to remote storage.
package main

import (
	"log"
	"os"
	"time"
)

// A high-level function to perform a full backup and handle tracking and uploading.
func PerformFullBackup(drive string, comment string) error {
	backupTime := time.Now()
	err := CreateFullBackup(backupTime)
	if err != nil {
		log.Println(err)
		return err
	}

	err = tracker.TrackBackup(backupTime, Saved, "full", comment)
	if err != nil {
		return err
	}
	go func() {
		err = UploadToRClone(backupTime, drive, false)
		if err != nil {
			log.Println(err)
			return
		}
		tracker.UpdateBackupStatus(backupTime, Uploaded)
	}()

	fullBackupTicker.Reset(fullBackupInterval)
	return nil
}

// A high-level function to perform an incremental backup and handle tracking and uploading.
func PerformIncrementalBackup(drive string, comment string) error {
	backupTime := time.Now()
	lastBackupTime, err := tracker.GetLastBackupTime()
	if err != nil {
		return err
	}

	err = CreateIncrementalBackup(backupTime, lastBackupTime)
	if err != nil {
		log.Println(err)
		return err
	}

	err = tracker.TrackBackup(backupTime, Saved, "incremental", comment)
	if err != nil {
		return err
	}
	go func() {
		err = UploadToRClone(backupTime, drive, true)
		if err != nil {
			log.Println(err)
			return
		}
		tracker.UpdateBackupStatus(backupTime, Uploaded)
	}()

	incrementalBackupTicker.Reset(incrementalBackupInterval)
	return nil
}

// Delete old backup.
func DeleteLocalBackup(track DatabaseTrack) error {
	err := os.RemoveAll(track.GetBackupPath())
	if err != nil {
		return err
	}
	tracker.UpdateBackupStatus(track.BackupTime, Archived)
	log.Printf("Deleted local backup %s\n", track.GetBackupPath())
	// If this is a full backup, also delete incremental backups based on it.
	if track.IsFullBackup() {
		incrementalTracks, err := tracker.GetIncrementalTracks(track)
		if err != nil {
			return err
		}
		for _, incTrack := range incrementalTracks {
			err := os.RemoveAll(incTrack.GetBackupPath())
			if err != nil {
				return err
			}
			tracker.UpdateBackupStatus(incTrack.BackupTime, Archived)
			log.Printf("Deleted local incremental backup %s\n", incTrack.GetBackupPath())
		}
	}
	return nil
}
