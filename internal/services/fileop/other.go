// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// other.go provides other functions used by the file service

package fileop

import (
	"bytes"
	"gosheet/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

const recentFilesPath = ".gosheet/recent.cf"

// getRecentFileList reads recent files from a config file
func GetRecentFileList() ([]string, []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}, []string{}
	}

	recentFile := filepath.Join(homeDir, recentFilesPath)
	data, err := os.ReadFile(recentFile)
	if err != nil {
		return []string{}, []string{}
	}

	lines := strings.Split(string(data), "\n")
	var filenames, locations []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if _, err := os.Stat(line); err == nil {
			locations = append(locations, line)
			filenames = append(filenames, filepath.Base(line))
		}
	}

	limit := utils.DEFAULT_RECENT_FILES_NUMBER
	if limit < 1 {
		limit = 1
	}
	if len(filenames) > limit {
		filenames = filenames[:limit]
		locations = locations[:limit]
	}

	return filenames, locations
}

// AddToRecentFiles adds a file to the recent files list (exported for use in main)
func AddToRecentFiles(filepathtodir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	recentFile := filepath.Join(homeDir, recentFilesPath)

	dir := filepath.Dir(recentFile)
	os.MkdirAll(dir, 0755)

	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		if err := os.WriteFile(recentFile, []byte(""), 0644); err != nil {
			return
		}
	}

	normalizedNew, err := filepath.Abs(filepathtodir)
	if err != nil {
		normalizedNew = filepathtodir
	}

	existing := []string{}
	if data, err := os.ReadFile(recentFile); err == nil {
		lines := strings.SplitN(string(data), "\n", bytes.Count(data, []byte{'\n'})+1)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			normalizedExisting, err := filepath.Abs(line)
			if err != nil {
				normalizedExisting = line
			}
			if normalizedExisting != normalizedNew {
				existing = append(existing, line)
			}
		}
	}

	existing = append([]string{filepathtodir}, existing...)

	limit := utils.DEFAULT_RECENT_FILES_NUMBER
	if limit < 1 {
		limit = 1
	}
	if len(existing) > limit {
		existing = existing[:limit]
	}

	os.WriteFile(recentFile, []byte(strings.Join(existing, "\n")), 0644)
}

// TrimRecentFiles shortens the saved list to the current limit.
func TrimRecentFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	recentFile := filepath.Join(homeDir, recentFilesPath)
	data, err := os.ReadFile(recentFile)
	if err != nil {
		return
	}
	limit := utils.DEFAULT_RECENT_FILES_NUMBER
	if limit < 1 {
		limit = 1
	}
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kept = append(kept, line)
		if len(kept) >= limit {
			break
		}
	}
	_ = os.WriteFile(recentFile, []byte(strings.Join(kept, "\n")), 0644)
}

// ClearRecentFiles removes every path from the recent-file list.
func ClearRecentFiles() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	recentFile := filepath.Join(homeDir, recentFilesPath)
	if err := os.MkdirAll(filepath.Dir(recentFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(recentFile, []byte(""), 0644)
}
