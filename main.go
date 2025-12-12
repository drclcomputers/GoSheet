// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// A simple terminal-based spreadsheet application using tview.
// main.go is the entry point of the application.

package main

import (
	"fmt"
	"gosheet/internal/services/fileop"
	"gosheet/internal/services/table"
	"gosheet/internal/services/ui/file"
	"gosheet/internal/utils"
	"os"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"flag"
)

// main is the entry point of the application, where the tview application is initialized, it is checked for command-line arguments to open a file or create a new table
func main() {
	//runtime.MemProfileRate = 1

	utils.UpdateNrCellsOnScrn()

	app := tview.NewApplication()
	
	defer func() {
	    if r := recover(); r != nil {
	        fmt.Fprintf(os.Stderr, "Application crashed: %v\n", r)
	        fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
	        os.Exit(1)
	    }
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	    if event.Key() == tcell.KeyCtrlC {
	        modal := tview.NewModal().
	            SetText("Ctrl+C detected. Exiting...\nUnsaved edits will be lost.").
	            AddButtons([]string{"OK"}).
	            SetDoneFunc(func(buttonIndex int, buttonLabel string) {
	                app.Stop()
	            })
	        app.SetRoot(modal, true).SetFocus(modal)
	        return nil
	    }
	    return event
	})
	
	var filename string

	flag.StringVar(&filename, "file", "", "Path to .gsheet/.json/.txt file to open")
	flag.Parse()

	if filename == "" && len(flag.Args()) > 0 {
	    filename = flag.Args()[0]
	}

	var t *tview.Table
	if filename != "" {
		fileop.AddToRecentFiles(filename)
		t = table.OpenTable(app, filename)
	} else {
		filename = file.StartMenuUI(app)

		if filename == "QUIT" {
        	return
    	}

		if filename == "THERE_IS_NO_FILE_SELECTED" {
			t = table.NewTable(app)
		} else {
			fileop.AddToRecentFiles(filename)
			t = table.OpenTable(app, filename)
		}
    }


	if t != nil { t.Select(1, 1) } else { return }

	app.SetRoot(t, true).SetFocus(t)

	if err := app.Run(); err != nil {
		panic(err)
	}

	if len(utils.TOBEPRINTED) > 0 { fmt.Println(utils.TOBEPRINTED) }


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
