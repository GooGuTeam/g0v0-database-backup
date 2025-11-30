// Handle HTTP requests from external services.
// Content-Type: application/json
// Accept: application/json
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// POST /full
// Trigger a full backup.
// Response: 204 No Content on success, 400 Bad Request on invalid input, 500 Internal Server Error on failure.
// Request body:
//
//	drive (string): The rclone drive name to upload the backup to.
//	comment (string, optional): An optional comment for the backup.
func HandleFullBackup(w http.ResponseWriter, r *http.Request) {
	type FullBackupRequest struct {
		Drive   string `json:"drive,omitempty"`
		Comment string `json:"comment,omitempty"`
	}
	var req FullBackupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Comment == "" {
		req.Comment = "Manual full backup"
	}
	log.Printf("Received full backup request: drive=%s, comment=%s", req.Drive, req.Comment)
	err = PerformFullBackup(req.Drive, req.Comment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Full backup failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /incremental
// Trigger an incremental backup.
// Response: 204 No Content on success, 400 Bad Request on invalid input, 500 Internal Server Error on failure.
// Request body:
//
//	drive (string): The rclone drive name to upload the backup to.
//	comment (string, optional): An optional comment for the backup.
func HandleIncrementalBackup(w http.ResponseWriter, r *http.Request) {
	type IncrementalBackupRequest struct {
		Drive   string `json:"drive,omitempty"`
		Comment string `json:"comment,omitempty"`
	}
	var req IncrementalBackupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Comment == "" {
		req.Comment = "Manual incremental backup"
	}
	log.Printf("Received incremental backup request: drive=%s, comment=%s", req.Drive, req.Comment)
	err = PerformIncrementalBackup(req.Drive, req.Comment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Incremental backup failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /download
// Download a backup from rclone.
// Response: 204 No Content on success, 400 Bad Request on invalid input, 500 Internal Server Error on failure.
// Request body:
//
//	drive (string): The rclone drive name to download the backup from.
//	backup_name (string): The backup filename.
func HandleDownloadBackup(w http.ResponseWriter, r *http.Request) {
	type DownloadBackupRequest struct {
		Drive      string `json:"drive,omitempty"`
		BackupName string `json:"backup_name"`
	}
	var req DownloadBackupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Received download backup request: drive=%s, backup_name=%s", req.Drive, req.BackupName)
	go func() {
		err = DownloadFromRClone(req.Drive, req.BackupName)
		if err != nil {
			log.Printf("Download backup failed: %v", err)
		} else {
			log.Printf("Download backup %s completed successfully.", req.BackupName)
		}
	}()
	w.WriteHeader(http.StatusNoContent)
}
