// Track backups in SQLite.
package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
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
	// Relative backup path.
	Path string
}

func (track DatabaseTrack) IsFullBackup() bool {
	return track.Type == "full"
}

func (track DatabaseTrack) IsIncrementalBackup() bool {
	return track.Type == "incremental"
}

func (track DatabaseTrack) GetBackupPath() string {
	return FormatBackupLocalPath(track.GetRelativeBackupPath())
}

func (track DatabaseTrack) GetRelativeBackupPath() string {
	if track.Path != "" {
		return NormalizeBackupRelativePath(track.Path)
	}
	if track.IsFullBackup() {
		return FormatBackupRelativePath(track.BackupTime, false)
	}
	return FormatBackupRelativePath(track.BackupTime, true)
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
		log.Fatalln(err)
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
		comment TEXT,
		backup_path TEXT NOT NULL DEFAULT ''
	);
	`)
	if err != nil {
		return err
	}

	err = ensureBackupPathColumn(db)
	if err != nil {
		return err
	}

	return migrateBackupPaths(db)
}

func (t *Tracker) Close() error {
	return t.DB.Close()
}

// Track a new backup in the database.
func (t *Tracker) TrackBackup(track DatabaseTrack) error {
	_, err := t.Exec(
		"INSERT INTO backups (backup_time, status, type, comment, backup_path) VALUES (?, ?, ?, ?, ?)",
		track.BackupTime.Format(time.RFC3339),
		track.Status,
		track.Type,
		track.Comment,
		track.GetRelativeBackupPath(),
	)
	return err
}

// Update the status of a backup.
func (t *Tracker) UpdateBackupStatus(backupTime time.Time, status Status) error {
	_, err := t.Exec("UPDATE backups SET status = ? WHERE backup_time = ?", status, backupTime.Format(time.RFC3339))
	return err
}

// Get the last backup time.
func (t *Tracker) GetLastBackup() (DatabaseTrack, error) {
	var track DatabaseTrack
	var backupTimeStr string
	err := t.QueryRow("SELECT id, backup_time, status, type, comment, backup_path FROM backups ORDER BY backup_time DESC LIMIT 1").Scan(&track.ID, &backupTimeStr, &track.Status, &track.Type, &track.Comment, &track.Path)
	if err != nil {
		return DatabaseTrack{}, err
	}
	track.BackupTime, err = time.Parse(time.RFC3339, backupTimeStr)
	if err != nil {
		return DatabaseTrack{}, err
	}
	return track, nil
}

// Get old full backups that exceed the local backup count and not uploaded.
func (t *Tracker) GetOldBackups() ([]DatabaseTrack, error) {
	rows, err := t.Query("SELECT id, backup_time, status, type, comment, backup_path FROM backups WHERE type = 'full' AND status = 0 ORDER BY backup_time ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []DatabaseTrack
	var allBackups []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment, &bt.Path)
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
	rows, err := t.Query("SELECT id, backup_time, status, type, comment, backup_path FROM backups WHERE type = 'incremental' AND backup_time > ? AND backup_time < ? ORDER BY backup_time ASC", parentTrack.BackupTime.Format(time.RFC3339), nextParentTimeStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incTracks []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment, &bt.Path)
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
	rows, err := t.Query("SELECT id, backup_time, status, type, comment, backup_path FROM backups WHERE status = 0 ORDER BY backup_time ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []DatabaseTrack
	for rows.Next() {
		var bt DatabaseTrack
		var backupTimeStr string
		err := rows.Scan(&bt.ID, &backupTimeStr, &bt.Status, &bt.Type, &bt.Comment, &bt.Path)
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

func ensureBackupPathColumn(db *Tracker) error {
	rows, err := db.Query("PRAGMA table_info(backups)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasBackupPath := false
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		err = rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk)
		if err != nil {
			return err
		}
		if name == "backup_path" {
			hasBackupPath = true
			break
		}
	}
	if hasBackupPath {
		return nil
	}

	_, err = db.Exec("ALTER TABLE backups ADD COLUMN backup_path TEXT NOT NULL DEFAULT ''")
	return err
}

func migrateBackupPaths(db *Tracker) error {
	rows, err := db.Query("SELECT id, backup_time, type, backup_path FROM backups ORDER BY backup_time ASC")
	if err != nil {
		return err
	}
	defer rows.Close()

	type backupMigration struct {
		id   int
		time time.Time
		typ  string
		path string
	}

	var migrations []backupMigration
	for rows.Next() {
		var migration backupMigration
		var backupTimeStr string
		err = rows.Scan(&migration.id, &backupTimeStr, &migration.typ, &migration.path)
		if err != nil {
			return err
		}
		migration.time, err = time.Parse(time.RFC3339, backupTimeStr)
		if err != nil {
			return err
		}
		migrations = append(migrations, migration)
	}

	for _, migration := range migrations {
		isIncremental := migration.typ == "incremental"
		newRelativePath := FormatBackupRelativePath(migration.time, isIncremental)
		currentRelativePath := NormalizeBackupRelativePath(migration.path)
		moveCandidates := []string{
			currentRelativePath,
			FormatLegacyBackupRelativePath(migration.time, isIncremental),
		}
		for _, oldRelativePath := range moveCandidates {
			if oldRelativePath == "" || oldRelativePath == newRelativePath {
				continue
			}
			oldPath := FormatBackupLocalPath(oldRelativePath)
			newPath := FormatBackupLocalPath(newRelativePath)
			err = moveLocalBackupIfExists(oldPath, newPath)
			if err != nil {
				return err
			}
		}

		if currentRelativePath != newRelativePath {
			_, err = db.Exec("UPDATE backups SET backup_path = ? WHERE id = ?", newRelativePath, migration.id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func moveLocalBackupIfExists(oldPath string, newPath string) error {
	if oldPath == newPath {
		return nil
	}

	_, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	_, err = os.Stat(newPath)
	if err == nil {
		log.Printf("Skip moving local backup because target already exists: %s", newPath)
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(filepath.Dir(newPath), 0755)
	if err != nil {
		return err
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	log.Printf("Moved local backup from %s to %s", oldPath, newPath)
	return nil
}
