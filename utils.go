// Handle time formatting and parsing for backup timestamps.
package main

import (
	"os/exec"
	"time"
)

func FormatBackupTime(t time.Time) string {
	return t.Format("20060102_1504")
}

func FormatFullBackupDir(t time.Time) string {
	return backupPath + "db_" + FormatBackupTime(t)
}

func FormatIncrementalBackupDir(t time.Time) string {
	return backupPath + "db_" + FormatBackupTime(t) + "_inc"
}

func RunSubprocess(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
