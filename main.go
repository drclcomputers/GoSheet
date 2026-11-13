// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// A simple terminal-based spreadsheet application using tview.
// main.go is the entry point of the application.

package main

import (
	"fmt"
	"gosheet/internal/services/fileop"
	"gosheet/internal/services/recovery"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/table"
	"gosheet/internal/services/ui/file"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"flag"
)

// main is the entry point of the application, where the tview application is initialized, it is checked for command-line arguments to open a file or create a new table
func main() {
	//runtime.MemProfileRate = 1

	settings := utils.LoadSettings()
	fileop.TrimRecentFiles()
	utils.UpdateNrCellsOnScrn()
	if utils.DEFAULT_VIEWPORT_ROWS > 6 {
		utils.DEFAULT_VIEWPORT_ROWS -= 2
	}

	app := tview.NewApplication()
	var autosaver *recovery.Autosaver

	defer func() {
		if r := recover(); r != nil {
			if autosaver != nil {
				autosaver.Stop()
			}
			fmt.Fprintf(os.Stderr, "\nGoSheet crashed: %v\n%s\n", r, debug.Stack())
			os.Exit(1)
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			modal := tview.NewModal().
				SetText("Ctrl+C detected. Exiting...\nUnsaved edits may be restored from the autosave recovery file.").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.Stop()
				})
			uitheme.StyleModalWarn(modal, "Exit")
			app.SetRoot(modal, true).SetFocus(modal)
			return nil
		}
		return event
	})

	var filename string

	flag.StringVar(&filename, "file", "", "Path to .gsheet/.json/.txt/.csv file to open")
	flag.Parse()

	if filename == "" && len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}

	var t *tview.Table
	var err error
	recoveryPath, recoveryPathErr := utils.RecoveryFilePath()
	if recoveryPathErr != nil {
		fmt.Fprintf(os.Stderr, "Autosave recovery is unavailable: %v\n", recoveryPathErr)
	}
	if recoveryPathErr == nil {
		if _, statErr := os.Stat(recoveryPath); statErr == nil {
			restore, promptErr := promptForRecovery(app, recoveryPath)
			if promptErr != nil {
				fmt.Fprintf(os.Stderr, "Could not ask about recovery data: %v\n", promptErr)
				return
			}
			if restore {
				t, err = table.OpenTable(app, recoveryPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not restore recovery data from %s: %v\n", recoveryPath, err)
					return
				}
				table.SetCurrentFilename(t, "")
				table.MarkAsModified(t)
			} else if err := os.Remove(recoveryPath); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Could not discard recovery data: %v\n", err)
				return
			}
		} else if !os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "Could not check recovery data: %v\n", statErr)
			recoveryPath = ""
		}
	}

	if t == nil && filename != "" {
		fileop.AddToRecentFiles(filename)
		t, err = table.OpenTable(app, filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", filename, err)
			os.Exit(1)
		}
	} else if t == nil {
		filename = file.StartMenuUI(app)

		if filename == "QUIT" {
			return
		}

		if filename == "THERE_IS_NO_FILE_SELECTED" {
			t = table.NewTable(app)
		} else {
			fileop.AddToRecentFiles(filename)
			t, err = table.OpenTable(app, filename)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	if t == nil {
		return
	}

	if recoveryPath != "" {
		autosaver = recovery.NewAutosaver(
			app,
			recoveryPath,
			time.Duration(settings.AutosaveMins)*time.Minute,
			func() (recovery.Snapshot, bool) {
				sheets, activeSheet, dirty := table.CaptureRecoverySnapshot()
				return recovery.Snapshot{Sheets: sheets, ActiveSheet: activeSheet}, dirty
			},
			func(err error) {
				fmt.Fprintf(os.Stderr, "Autosave failed: %v\n", err)
			},
		)
		table.SetRecoveryClearHook(func() {
			if err := autosaver.Clear(); err != nil {
				fmt.Fprintf(os.Stderr, "Could not clear recovery data: %v\n", err)
			}
		})
	}

	sheetview.Mount(app, t)
	t.Select(1, 1)
	sheetview.Show(app)

	if err := app.Run(); err != nil {
		panic(err)
	}
	if autosaver != nil {
		autosaver.Stop()
		if !table.HasUnsavedChanges() {
			if err := autosaver.Clear(); err != nil {
				fmt.Fprintf(os.Stderr, "Could not clear recovery data: %v\n", err)
			}
		}
		table.SetRecoveryClearHook(nil)
	}

	if len(utils.TOBEPRINTED) > 0 {
		fmt.Println(utils.TOBEPRINTED)
	}

	/*
	   // Print memory stats
	   var m runtime.MemStats
	   runtime.ReadMemStats(&m)
	   fmt.Printf("\nMemory Stats:\n")
	   fmt.Printf("Alloc = %v MB\n", m.Alloc/1024/1024)
	   fmt.Printf("TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024)
	   fmt.Printf("Sys = %v MB\n", m.Sys/1024/1024)
	   fmt.Printf("NumGC = %v\n", m.NumGC)
	*/
}

func promptForRecovery(app *tview.Application, path string) (bool, error) {
	var restore bool
	decided := false
	modal := tview.NewModal().
		SetText(fmt.Sprintf("An autosaved workbook was found:\n%s\n\nRecover it?", path)).
		AddButtons([]string{"Recover", "Discard"}).
		SetDoneFunc(func(_ int, label string) {
			decided = true
			restore = label == "Recover"
			app.Stop()
		})
	uitheme.StyleModalWarn(modal, "Workbook recovery")
	app.SetRoot(modal, true).SetFocus(modal)
	if err := app.Run(); err != nil {
		return false, err
	}
	if !decided {
		return false, fmt.Errorf("recovery prompt closed without a choice")
	}
	return restore, nil
}
