package fileop

import (
	"os"
	"path/filepath"
	"testing"

	"gosheet/internal/services/cell"
)

func TestSaveRecoveryWorkbookAtomically(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "recovery.gsheet")
	value := "recovered"
	sheets := []SheetInfo{
		{
			Name: "First",
			Rows: 2,
			Cols: 2,
			GlobalData: map[[2]int]*cell.Cell{
				{1, 1}: {RawValue: &value, Row: 1, Column: 1},
			},
		},
		{Name: "Second", Rows: 1, Cols: 1, GlobalData: map[[2]int]*cell.Cell{}},
	}
	if err := SaveRecoveryWorkbook(filename, sheets, 1); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("recovery permissions = %o, want 600", info.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Dir(filename))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(filename) {
		t.Fatalf("temporary files remain: %v", entries)
	}

	workbook, err := OpenWorkbook(filename)
	if err != nil {
		t.Fatal(err)
	}
	if workbook.ActiveSheet != 1 || len(workbook.Sheets) != 2 {
		t.Fatalf("restored workbook metadata: active=%d sheets=%d", workbook.ActiveSheet, len(workbook.Sheets))
	}
	if got := *workbook.Sheets[0].Cells[0].RawValue; got != value {
		t.Fatalf("restored cell value = %q, want %q", got, value)
	}
}

func TestSaveRecoveryWorkbookRequiresNativeExtension(t *testing.T) {
	if err := SaveRecoveryWorkbook(filepath.Join(t.TempDir(), "recovery.json"), nil, 0); err == nil {
		t.Fatal("expected extension error")
	}
}
