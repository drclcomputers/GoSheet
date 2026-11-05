package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// DefaultSaveFormat is the extension offered first in the save dialog.
var DefaultSaveFormat = ".gsheet"

// Settings are the values from the Options screen. They apply to new cells,
// the save dialog, and the recent-file list, including after a restart.
type Settings struct {
	DecimalPoints int    `json:"decimalPoints"`
	Thousands     string `json:"thousands"`
	DecimalSep    string `json:"decimalSep"`
	Currency      string `json:"currency"`
	ColumnWidth   int    `json:"columnWidth"`
	SaveFormat    string `json:"saveFormat"`
	RecentFiles   int    `json:"recentFiles"`
}

func DefaultSettings() Settings {
	return Settings{
		DecimalPoints: 2,
		Thousands:     ",",
		DecimalSep:    ".",
		Currency:      "$",
		ColumnWidth:   10,
		SaveFormat:    ".gsheet",
		RecentFiles:   10,
	}
}

func configPath() string {
	if p := os.Getenv("GOSHEET_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".gosheet", "options.json")
}

// LoadSettings reads the saved options. A missing or broken file is ignored
// and the built-in defaults are used instead.
func LoadSettings() Settings {
	s := DefaultSettings()
	path := configPath()
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			var got Settings
			if json.Unmarshal(data, &got) == nil {
				s = got
			}
		}
	}
	s = s.normalized()
	s.apply()
	return s
}

// SaveSettings stores the options and applies them immediately.
func SaveSettings(s Settings) error {
	s = s.normalized()
	s.apply()
	path := configPath()
	if path == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s Settings) normalized() Settings {
	if s.DecimalPoints < 0 {
		s.DecimalPoints = 0
	}
	if s.DecimalPoints > 8 {
		s.DecimalPoints = 8
	}
	s.Thousands = oneRune(s.Thousands, ",")
	s.DecimalSep = oneRune(s.DecimalSep, ".")
	if s.DecimalSep == "Ø" || s.DecimalSep == s.Thousands {
		s.DecimalSep = "."
		if s.Thousands == "." {
			s.Thousands = ","
		}
	}
	s.Currency = oneRune(s.Currency, "$")
	if s.Currency == "Ø" {
		s.Currency = "$"
	}
	if s.ColumnWidth < 4 {
		s.ColumnWidth = 4
	}
	if s.ColumnWidth > 40 {
		s.ColumnWidth = 40
	}
	if s.RecentFiles < 1 {
		s.RecentFiles = 1
	}
	if s.RecentFiles > 20 {
		s.RecentFiles = 20
	}
	if s.SaveFormat == "" || !stringsHasDot(s.SaveFormat) {
		s.SaveFormat = ".gsheet"
	}
	return s
}

func (s Settings) apply() {
	DEFAULT_CELL_DECIMAL_POINTS = int32(s.DecimalPoints)
	DEFAULT_CELL_THOUSANDS_SEPARATOR = settingRune(s.Thousands)
	DEFAULT_CELL_DECIMAL_SEPARATOR = settingRune(s.DecimalSep)
	DEFAULT_CELL_FINANCIAL_SIGN = settingRune(s.Currency)
	DEFAULT_CELL_MIN_WIDTH = int32(s.ColumnWidth)
	if DEFAULT_CELL_MAX_WIDTH < DEFAULT_CELL_MIN_WIDTH {
		DEFAULT_CELL_MAX_WIDTH = DEFAULT_CELL_MIN_WIDTH
	}
	DEFAULT_RECENT_FILES_NUMBER = s.RecentFiles
	DefaultSaveFormat = s.SaveFormat
	UpdateNrCellsOnScrn()
}

func settingRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	if r == 'Ø' || r == utf8.RuneError {
		return 0
	}
	return r
}

func oneRune(s, fallback string) string {
	if s == "" {
		return fallback
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size == 0 {
		return fallback
	}
	return string(r)
}

func stringsHasDot(s string) bool {
	return len(s) > 1 && s[0] == '.'
}
