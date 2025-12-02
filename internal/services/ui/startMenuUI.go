// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// startMenuUI.go provides an interactive menu when opening the app.

package ui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const newSheetButtonLabel = `
┌──┬──┬──┬──┬──┐
├──┼──┼──┼──┼──┤
├──┼──┼──┼──┼──┤
├──┼──┼──┼──┼──┤
└──┴──┴──┴──┴──┘
NEW WORKBOOK
`

const openFileButtonLabel = `
╔═══════════╗
║           ║
║   OPEN    ║
║           ║
╚═══════════╝
`

const optionsButtonLabel = `
╔═══════════╗
║           ║
║  OPTIONS  ║
║           ║
╚═══════════╝
`

const recentFilesPath = ".gosheet/recent.cf"

var selectedFile string = ""

// getRecentFileList reads recent files from a config file
func getRecentFileList() ([]string, []string) {
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

// selectFile sets the selected file and stops the app to return control
func selectFile(app *tview.Application, filename string) {
	selectedFile = filename
	app.Stop()
}

// StartMenuUI displays the startup menu and returns the selected filename
func StartMenuUI(app *tview.Application) string {
	selectedFile = "" 
	
	container := tview.NewFlex()
	
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	
	title := tview.NewTextView().
		SetText(fmt.Sprintf("[::b]GoSheet[::-] [yellow]v%s[::-]", utils.VER)).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	title.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue)
	
	newSheetBtn := tview.NewTextView().
		SetText(newSheetButtonLabel).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorLightGreen)
	newSheetBtn.SetBorder(true).
		SetBorderColor(tcell.ColorGreen)
	newSheetBtn.SetScrollable(false)

	newSheetBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			selectFile(app, "THERE_IS_NO_FILE_SELECTED")
		}
		return event
	})

	openFileBtn := tview.NewTextView().
		SetText(openFileButtonLabel).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorLightBlue)
	openFileBtn.SetBorder(true).
		SetBorderColor(tcell.ColorBlue)
	openFileBtn.SetScrollable(false)

	openFileBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			ShowUnifiedFileDialog(app, container, "open", nil, nil, nil, nil, nil, "")
		}
		return event
	})

	optionsBtn := tview.NewTextView().
		SetText(optionsButtonLabel).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorYellow)
	optionsBtn.SetBorder(true).
		SetBorderColor(tcell.ColorYellow)
	optionsBtn.SetScrollable(false)

	optionsBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			ShowWarningModal(app, container, "Options menu coming soon!")
		}
		return event
	})

	leftPanel.
		AddItem(title, 0, 1, false).
		AddItem(newSheetBtn, 0, 2, true).
		AddItem(openFileBtn, 0, 2, true).
		AddItem(optionsBtn, 0, 2, true)
	
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	
	recentTitle := tview.NewTextView().
		SetText("[::b]Recent Files[::-]").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	recentTitle.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue)
	
	filenames, locations := getRecentFileList()
	recentList := tview.NewList()
	
	if len(filenames) == 0 {
		recentList.AddItem("No recent files", "", 0, nil)
		recentList.SetSelectedTextColor(tcell.ColorGray)
	} else {
		for i, filename := range filenames {
			loc := locations[i]
			
			info, _ := os.Stat(loc)
			var sizeStr, modStr string
			if info != nil {
				size := info.Size()
				if size < 1024 {
					sizeStr = fmt.Sprintf("%d B", size)
				} else if size < 1024*1024 {
					sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
				} else {
					sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
				}
				modStr = info.ModTime().Format("15:04 - Jan 02, 2006")
			}
			
			secondaryText := fmt.Sprintf("%s  •  %s  •  %s", utils.PrettyPath(loc, "recentfiles"), modStr, sizeStr)
			
			recentList.AddItem(filename, secondaryText, rune('1'+i), func() {
				if _, err := os.Stat(loc); os.IsNotExist(err) {
					ShowErrorModal(app, container, fmt.Sprintf("File not found: %s", loc))
				} else {
					selectFile(app, loc)
				}
			})
		}
	}
	
	recentList.SetBorder(true).SetBorderColor(tcell.ColorLightBlue)
	recentList.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	recentList.SetSelectedTextColor(tcell.ColorWhite)
	
	rightPanel.
		AddItem(recentTitle, 3, 0, false).
		AddItem(recentList, 0, 1, true)
	
	container.
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 3, false)
	
	container.SetBorder(true).
		SetTitle(" Welcome to GoSheet | Use Ctrl+←/→ to navigate around the menus | Esc to quit ").
		SetBorderColor(tcell.ColorLightBlue)
	
	leftItems := []tview.Primitive{newSheetBtn, openFileBtn, optionsBtn}
	currentLeftItem := 0
	inLeftPanel := true
	
	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEscape || event.Rune() == 'q' || event.Rune() == 'Q':
			selectedFile = "QUIT"
			app.Stop()
			return nil
		case event.Rune() == 'n' && event.Modifiers()&tcell.ModAlt != 0:
			selectFile(app, "THERE_IS_NO_FILE_SELECTED")
			return nil
		case event.Rune() == 'o' && event.Modifiers()&tcell.ModAlt != 0:
			ShowUnifiedFileDialog(app, container, "open", nil, nil, nil, nil, nil, "")
			return nil
		case event.Key() == tcell.KeyLeft && event.Modifiers()&tcell.ModCtrl != 0:
			if inLeftPanel {
				currentLeftItem--
				if currentLeftItem < 0 {
					inLeftPanel = false
					app.SetFocus(recentList)
				} else {
					app.SetFocus(leftItems[currentLeftItem])
				}
			} else {
				inLeftPanel = true
				currentLeftItem = len(leftItems) - 1
				app.SetFocus(leftItems[currentLeftItem])
			}
			return nil
		case event.Key() == tcell.KeyRight && event.Modifiers()&tcell.ModCtrl != 0:
			if inLeftPanel {
				currentLeftItem++
				if currentLeftItem >= len(leftItems) {
					inLeftPanel = false
					app.SetFocus(recentList)
				} else {
					app.SetFocus(leftItems[currentLeftItem])
				}
			} else {
				inLeftPanel = true
				currentLeftItem = 0
				app.SetFocus(leftItems[currentLeftItem])
			}
			return nil
		}
		return event
	})
	
	for i, item := range leftItems {
		idx := i
		if tv, ok := item.(*tview.TextView); ok {
			tv.SetFocusFunc(func() {
				currentLeftItem = idx
				inLeftPanel = true
			})
		}
	}
	
	recentList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		inLeftPanel = false
	})
	
	app.SetRoot(container, true).SetFocus(newSheetBtn)
	if err := app.Run(); err != nil {
		return "QUIT"
	}
	
	return selectedFile
}

func ShowErrorModal(app *tview.Application, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Error\n\n%s", message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(returnTo, true).SetFocus(returnTo)
		})
	
	modal.SetBackgroundColor(tcell.ColorDarkRed).SetBorderColor(tcell.ColorRed)
	app.SetRoot(modal, true).SetFocus(modal)
}

