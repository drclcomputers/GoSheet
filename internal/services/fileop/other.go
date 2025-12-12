// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// other.go provides other functions used by the file service

package fileop

import (
	"bytes"
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

	if len(existing) > 10 {
		existing = existing[:10]
	}

	os.WriteFile(recentFile, []byte(strings.Join(existing, "\n")), 0644)
}


