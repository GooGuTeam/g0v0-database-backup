// Handle time formatting and parsing for backup timestamps.
package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const backupDateLayout = "0601/02" // YYMM/DD

func FormatBackupTime(t time.Time) string {
	return t.Format("20060102_1504")
}

func FormatBackupDatePath(t time.Time) string {
	return t.Format(backupDateLayout)
}

func FormatFullBackupName(t time.Time) string {
	return "db_" + FormatBackupTime(t)
}

func FormatIncrementalBackupName(t time.Time) string {
	return FormatFullBackupName(t) + "_inc"
}

func FormatBackupRelativePath(t time.Time, isIncremental bool) string {
	backupName := FormatFullBackupName(t)
	if isIncremental {
		backupName = FormatIncrementalBackupName(t)
	}
	return filepath.ToSlash(filepath.Join(FormatBackupDatePath(t), backupName))
}

func FormatLegacyBackupRelativePath(t time.Time, isIncremental bool) string {
	if isIncremental {
		return FormatIncrementalBackupName(t)
	}
	return FormatFullBackupName(t)
}

func NormalizeBackupRelativePath(path string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(path))
	normalized = strings.TrimPrefix(normalized, backupPath)
	normalized = strings.TrimPrefix(normalized, strings.TrimPrefix(backupPath, "/"))
	normalized = strings.TrimPrefix(normalized, remoteBackupBasePath)
	normalized = strings.TrimPrefix(normalized, "/")
	return normalized
}

func FormatBackupLocalPath(relativePath string) string {
	return filepath.Join(backupPath, filepath.FromSlash(NormalizeBackupRelativePath(relativePath)))
}

func FormatRemoteBackupPath(relativePath string) string {
	return remoteBackupBasePath + NormalizeBackupRelativePath(relativePath)
}

func FormatFullBackupDir(t time.Time) string {
	return FormatBackupLocalPath(FormatBackupRelativePath(t, false))
}

func FormatIncrementalBackupDir(t time.Time) string {
	return FormatBackupLocalPath(FormatBackupRelativePath(t, true))
}

func BuildRCloneRemotePath(remote string, relativePath string) string {
	return remote + FormatRemoteBackupPath(relativePath)
}

func ResolveBackupRelativePaths(backupName string) []string {
	backupName = strings.TrimSpace(backupName)
	if backupName == "" {
		return nil
	}

	normalized := NormalizeBackupRelativePath(backupName)
	if strings.Contains(normalized, "/") {
		return []string{normalized}
	}

	layouts := []string{
		"db_20060102_1504",
		"db_20060102_1504_inc",
	}
	for _, layout := range layouts {
		backupTime, err := time.Parse(layout, backupName)
		if err == nil {
			isIncremental := strings.HasSuffix(backupName, "_inc")
			return []string{
				FormatBackupRelativePath(backupTime, isIncremental),
				FormatLegacyBackupRelativePath(backupTime, isIncremental),
			}
		}
	}

	return []string{normalized}
}

func RunSubprocess(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
