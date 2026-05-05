// RClone integration for uploading and downloading backups.
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"
)

func UploadToRClone(track DatabaseTrack, remote string) error {
	if remote == "" {
		remote = config.DefaultRCloneRemote
	}

	log.Printf("Uploading backup %s to remote %s\n", track.BackupTime.Format(time.DateTime), remote)
	output, err := RunSubprocess(
		"rclone",
		"--config",
		"rclone.conf",
		"copy",
		track.GetBackupPath(),
		BuildRCloneRemotePath(remote, track.GetRelativeBackupPath()),
	)
	if err != nil {
		return fmt.Errorf("Failed to upload backup to rclone remote: %v, output: %s", err, output)
	}
	log.Printf("Backup %s uploaded to remote %s successfully.\n", track.BackupTime.Format(time.DateTime), remote)
	return nil
}

func DownloadFromRClone(remote string, backupName string) error {
	if remote == "" {
		remote = config.DefaultRCloneRemote
	}

	destinationPath := filepath.Join(downloadedBackupPath, filepath.Base(NormalizeBackupRelativePath(backupName)))
	candidates := ResolveBackupRelativePaths(backupName)
	if len(candidates) == 0 {
		return fmt.Errorf("backup_name is required")
	}
	var lastErr error
	for _, candidate := range candidates {
		log.Printf("Downloading backup %s from remote %s\n", candidate, remote)
		output, err := RunSubprocess(
			"rclone",
			"--config",
			"rclone.conf",
			"copy",
			BuildRCloneRemotePath(remote, candidate),
			destinationPath,
		)
		if err == nil {
			log.Printf("Backup %s downloaded from remote %s successfully.\n", candidate, remote)
			return nil
		}
		lastErr = fmt.Errorf("Failed to download backup from rclone remote path %s: %v, output: %s", candidate, err, output)
	}

	return lastErr
}
