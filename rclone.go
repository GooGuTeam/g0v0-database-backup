// RClone integration for uploading and downloading backups.
package main

import (
	"fmt"
	"log"
	"time"
)

func UploadToRClone(backupTime time.Time, remote string, isIncremental bool) error {
	var path string
	if isIncremental {
		path = FormatIncrementalBackupDir(backupTime)
	} else {
		path = FormatFullBackupDir(backupTime)
	}

	if remote == "" {
		remote = config.DefaultRCloneRemote
	}

	log.Printf("Uploading backup %s to remote %s\n", backupTime.Format(time.DateTime), remote)
	output, err := RunSubprocess(
		"rclone",
		"--config",
		"rclone.conf",
		"copy",
		path,
		remote+path,
	)
	if err != nil {
		return fmt.Errorf("Failed to upload backup to rclone remote: %v, output: %s", err, output)
	}
	log.Printf("Backup %s uploaded to remote %s successfully.\n", backupTime.Format(time.DateTime), remote)
	return nil
}

func DownloadFromRClone(remote string, backupName string) error {
	if remote == "" {
		remote = config.DefaultRCloneRemote
	}

	log.Printf("Downloading backup %s from remote %s\n", backupName, remote)
	output, err := RunSubprocess(
		"rclone",
		"--config",
		"rclone.conf",
		"copy",
		remote+"/backup/"+backupName,
		downloadedBackupPath+backupName,
	)
	if err != nil {
		return fmt.Errorf("Failed to download backup from rclone remote: %v, output: %s", err, output)
	}
	log.Printf("Backup %s downloaded from remote %s successfully.\n", backupName, remote)
	return nil
}
