// Track backups in SQLite.
package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Status int

const (
	// A backup is saved locally, waiting to be uploaded.
	Saved Status = iota
	// A backup is already uploaded to remote storage.
	Uploaded
	// A backup is already uploaded and old local backup file is deleted.
	Archived
)

type DatabaseTrack struct {
	// The primary key ID.
	ID int
	// The backup time, saved in ISO 8601 format.
	BackupTime time.Time
	// The status of this backup.
	Status Status
	// The type of this backup, full or incremental.
	Type string
	// Optional comment.
	Comment string
}

func (track DatabaseTrack) IsFullBackup() bool {
	return track.Type == "full"
}

func (track DatabaseTrack) IsIncrementalBackup() bool {
	return track.Type == "incremental"
}

func (track DatabaseTrack) GetBackupPath() string {
	if track.IsFullBackup() {
		return FormatFullBackupDir(track.BackupTime)
	} else {
		return FormatIncrementalBackupDir(track.BackupTime)
	}
}

type Tracker struct {
	*sql.DB
}

var tracker *Tracker

func InitializeTracker() {
	db, err := sql.Open("sqlite3", sqliteDBPath)
	if err != nil {
		log.Fatalln(err)
	}
	tracker = &Tracker{db}
	err = initializeTrackingDB(tracker)
	if err != nil {
		log.Fatalln()
	}
}

// Initialize the tracking database.
func initializeTrackingDB(db *Tracker) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		backup_time TEXT NOT NULL,
		status INTEGER NOT NULL,
		type TEXT NOT NULL,
		comment TEXT
	);
	`)
	return err
}

func (t *Tracker) Close() error {
	return t.DB.Close()
}

// Track a new backup in the database.
func (t *Tracker) TrackBackup(backupTime time.Time, status Status, backupType string, comment string) error {
	_, err := t.Exec("INSERT INTO backups (backup_time, status, type, comment) VALUES (?, ?, ?, ?)", backupTime.Format(time.RFC3339), status, backupType, comment)
	return err
}

// Update the status of a backup.
func (t *Tracker) UpdateBackupStatus(backupTime time.Time, status Status) error {
	_, err := t.Exec("UPDATE backups SET status = ? WHERE backup_time = ?", status, backupTime.Format(time.RFC3339))
	return err
}

// Get the last backup time.
func (t *Tracker) GetLastBackupTime() (time.Time, error) {
	var backupTimeStr string
	err := t.QueryRow("SELECT backup_time FROM backups WHERE type = 'full' ORDER BY backup_time DESC LIMIT 1").Scan(&backupTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	backupTime, err := time.Parse(time.RFC3339, backupTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	return backupTime, nil
}

// Get the last full backup time.
func (t *Tracker) GetLastFullBackupTime() (time.Time, error) {
	var backupTimeStr string
	err := t.QueryRow("SELECT backup_time FROM backups WHERE type = 'full' ORDER BY backup_time DESC LIMIT 1").Scan(&backupTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	backupTime, err := time.Parse(time.RFC3339, backupTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	return backupTime, nil
}

// Get old full backups that exceed the local backup count and not uploaded.
func (t *Tracker) GetOldBackups() ([]DatabaseTrack, error) {
	rows, err := t.Query("SELECT id, backup_time, status, type, comment FROM backups WHERE type = 'full' AND status = 0 ORDER BY backup_time ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []DatabaseTrack
	var allBackups []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment)
		if err != nil {
			return nil, err
		}
		bt.BackupTime, err = time.Parse(time.RFC3339, backupTimeStr)
		if err != nil {
			return nil, err
		}
		allBackups = append(allBackups, bt)
	}
	if len(allBackups) <= config.LocalBackupCount {
		return []DatabaseTrack{}, nil
	}
	backups = allBackups[:len(allBackups)-config.LocalBackupCount]
	return backups, nil
}

// Get incremental backups associated with a full backup.
func (t *Tracker) GetIncrementalTracks(parentTrack DatabaseTrack) ([]DatabaseTrack, error) {
	var nextParentTimeStr string
	err := t.QueryRow("SELECT backup_time FROM backups WHERE type = 'full' AND backup_time > ? ORDER BY backup_time ASC LIMIT 1", parentTrack.BackupTime.Format(time.RFC3339)).Scan(&nextParentTimeStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if nextParentTimeStr == "" {
		nextParentTimeStr = time.Now().Format(time.RFC3339)
	}
	rows, err := t.Query("SELECT id, backup_time, status, type, comment FROM backups WHERE type = 'incremental' AND backup_time > ? AND backup_time < ? ORDER BY backup_time ASC", parentTrack.BackupTime.Format(time.RFC3339), nextParentTimeStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incTracks []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment)
		if err != nil {
			return nil, err
		}
		bt.BackupTime, err = time.Parse(time.RFC3339, backupTimeStr)
		if err != nil {
			return nil, err
		}
		incTracks = append(incTracks, bt)
	}
	return incTracks, nil
}

// Get backups that are not yet uploaded.
func (t *Tracker) GetPendingUploads() ([]DatabaseTrack, error) {
	rows, err := t.Query("SELECT id, backup_time, status, type, comment FROM backups WHERE status = 0 ORDER BY backup_time ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment)
		if err != nil {
			return nil, err
		}
		bt.BackupTime, err = time.Parse(time.RFC3339, backupTimeStr)
		if err != nil {
			return nil, err
		}
		backups = append(backups, bt)
	}
	return backups, nil
}
