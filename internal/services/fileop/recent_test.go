package fileop

import (
	"os"
	"path/filepath"
	"testing"

	"gosheet/internal/utils"
)

func TestRecentFilesLimitAndClear(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	old := utils.DEFAULT_RECENT_FILES_NUMBER
	utils.DEFAULT_RECENT_FILES_NUMBER = 2
	t.Cleanup(func() { utils.DEFAULT_RECENT_FILES_NUMBER = old })

	var paths []string
	for _, name := range []string{"a.gsheet", "b.gsheet", "c.gsheet"} {
		p := filepath.Join(home, name)
		if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
		AddToRecentFiles(p)
	}

	names, locs := GetRecentFileList()
	if len(names) != 2 || len(locs) != 2 {
		t.Fatalf("got %d names", len(names))
	}
	if locs[0] != paths[2] {
		t.Fatalf("newest first, got %s", locs[0])
	}

	if err := ClearRecentFiles(); err != nil {
		t.Fatal(err)
	}
	names, _ = GetRecentFileList()
	if len(names) != 0 {
		t.Fatalf("cleared list still has %d", len(names))
	}
}
