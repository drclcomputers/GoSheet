package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "options.json")
	t.Setenv("GOSHEET_CONFIG", path)

	oldPoints := DEFAULT_CELL_DECIMAL_POINTS
	oldThou := DEFAULT_CELL_THOUSANDS_SEPARATOR
	oldDec := DEFAULT_CELL_DECIMAL_SEPARATOR
	oldSign := DEFAULT_CELL_FINANCIAL_SIGN
	oldWidth := DEFAULT_CELL_MIN_WIDTH
	oldMax := DEFAULT_CELL_MAX_WIDTH
	oldRecent := DEFAULT_RECENT_FILES_NUMBER
	oldFormat := DefaultSaveFormat
	t.Cleanup(func() {
		DEFAULT_CELL_DECIMAL_POINTS = oldPoints
		DEFAULT_CELL_THOUSANDS_SEPARATOR = oldThou
		DEFAULT_CELL_DECIMAL_SEPARATOR = oldDec
		DEFAULT_CELL_FINANCIAL_SIGN = oldSign
		DEFAULT_CELL_MIN_WIDTH = oldWidth
		DEFAULT_CELL_MAX_WIDTH = oldMax
		DEFAULT_RECENT_FILES_NUMBER = oldRecent
		DefaultSaveFormat = oldFormat
		UpdateNrCellsOnScrn()
	})

	s := Settings{
		DecimalPoints: 4,
		Thousands:     "Ø",
		DecimalSep:    ",",
		Currency:      "€",
		ColumnWidth:   16,
		SaveFormat:    ".xlsx",
		RecentFiles:   3,
		AutosaveMins:  3,
	}
	if err := SaveSettings(s); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if DEFAULT_CELL_DECIMAL_POINTS != 4 || DEFAULT_CELL_THOUSANDS_SEPARATOR != 0 || DEFAULT_CELL_DECIMAL_SEPARATOR != ',' || DEFAULT_CELL_FINANCIAL_SIGN != '€' {
		t.Fatalf("applied %+v seps %q %q %q", s, string(DEFAULT_CELL_THOUSANDS_SEPARATOR), string(DEFAULT_CELL_DECIMAL_SEPARATOR), string(DEFAULT_CELL_FINANCIAL_SIGN))
	}
	if DEFAULT_CELL_MIN_WIDTH != 16 || DEFAULT_RECENT_FILES_NUMBER != 3 || DefaultSaveFormat != ".xlsx" {
		t.Fatalf("width %d recent %d format %s", DEFAULT_CELL_MIN_WIDTH, DEFAULT_RECENT_FILES_NUMBER, DefaultSaveFormat)
	}

	DEFAULT_CELL_DECIMAL_POINTS = 2
	got := LoadSettings()
	if got.DecimalPoints != 4 || got.ColumnWidth != 16 || got.SaveFormat != ".xlsx" || got.RecentFiles != 3 || got.AutosaveMins != 3 {
		t.Fatalf("reloaded %+v", got)
	}
	if DEFAULT_CELL_DECIMAL_POINTS != 4 {
		t.Fatal("reload did not apply")
	}
}

func TestSettingsRejectsUnsafeValues(t *testing.T) {
	s := Settings{DecimalPoints: 99, Thousands: ",", DecimalSep: ",", Currency: "", ColumnWidth: 1, RecentFiles: 0, SaveFormat: "xlsx", AutosaveMins: 61}
	s = s.normalized()
	if s.DecimalPoints != 8 || s.ColumnWidth != 4 || s.RecentFiles != 1 || s.SaveFormat != ".gsheet" || s.AutosaveMins != 60 {
		t.Fatalf("normalized %+v", s)
	}
	if s.DecimalSep == s.Thousands {
		t.Fatal("separators collided")
	}
	s.apply()
	if DEFAULT_CELL_MIN_WIDTH < 1 || DEFAULT_VIEWPORT_COLS < 1 || DEFAULT_VIEWPORT_ROWS < 1 {
		t.Fatalf("viewport %d x %d width %d", DEFAULT_VIEWPORT_COLS, DEFAULT_VIEWPORT_ROWS, DEFAULT_CELL_MIN_WIDTH)
	}
	DefaultSettings().apply()
}

func TestSettingsDefaultAutosaveInterval(t *testing.T) {
	if got := DefaultSettings().AutosaveMins; got != 5 {
		t.Fatalf("default autosave interval = %d, want 5", got)
	}
	if got := (Settings{}).normalized().AutosaveMins; got != 5 {
		t.Fatalf("normalized autosave interval = %d, want 5", got)
	}
}
