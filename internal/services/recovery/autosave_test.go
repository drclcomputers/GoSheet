package recovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"
)

func TestAutosaverPersistsAndClearsSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovery.gsheet")
	value := "saved"
	a := NewAutosaver(nil, path, time.Hour, nil, nil)
	defer a.Stop()

	a.persist(Snapshot{
		Sheets: []fileop.SheetInfo{{
			Name: "Sheet1",
			GlobalData: map[[2]int]*cell.Cell{
				{1, 1}: {RawValue: &value, Row: 1, Column: 1},
			},
		}},
	})
	a.workers.Wait()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("recovery snapshot was not written: %v", err)
	}

	if err := a.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("recovery snapshot remains after clear: %v", err)
	}
}

func TestAutosaverClearInvalidatesPendingSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovery.gsheet")
	value := "stale"
	a := NewAutosaver(nil, path, time.Hour, nil, nil)
	defer a.Stop()

	a.persist(Snapshot{
		Sheets: []fileop.SheetInfo{{
			Name: "Sheet1",
			GlobalData: map[[2]int]*cell.Cell{
				{1, 1}: {RawValue: &value, Row: 1, Column: 1},
			},
		}},
	})
	if err := a.Clear(); err != nil {
		t.Fatal(err)
	}
	a.workers.Wait()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stale recovery snapshot was written after clear: %v", err)
	}
}
