// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// startMenuUI.go provides an interactive menu when opening the app.

package file

import (
	"fmt"
	"os"

	"gosheet/internal/services/fileop"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const newSheetButtonLabel = `
┌──┬──┬──┬──┬──┐
├──┼──┼──┼──┼──┤
├──┼──┼──┼──┼──┤
├──┼──┼──┼──┼──┤
├──┼──┼──┼──┼──┤
└──┴──┴──┴──┴──┘
NEW WORKBOOK
`

const openFileButtonLabel = `
┌────────────┐
│ ▄▄▄        │
│ ┌────────┐ │
│ │ ────── │ │
│ └────────┘ │
└────────────┘
OPEN
`

const optionsButtonLabel = `
┌─────────────┐
│ ──●──────── │
│ ────────●── │
│ ───●─────── │
│ ────────●── │
└─────────────┘
OPTIONS
`

var selectedFile string = ""

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
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))

	newSheetBtn := tview.NewTextView().
		SetText(newSheetButtonLabel).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorLightGreen)
	newSheetBtn.SetBorder(true).
		SetBorderColor(tcell.NewRGBColor(90, 180, 140)).
		SetBackgroundColor(tcell.NewRGBColor(16, 28, 26))
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
		SetTextColor(tcell.NewRGBColor(168, 212, 220))
	openFileBtn.SetBorder(true).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(14, 22, 32))
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
		SetTextColor(tcell.NewRGBColor(240, 208, 144))
	optionsBtn.SetBorder(true).
		SetBorderColor(tcell.NewRGBColor(160, 130, 70)).
		SetBackgroundColor(tcell.NewRGBColor(28, 24, 16))
	optionsBtn.SetScrollable(false)

	var recentList *tview.List
	optionsBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			ShowOptions(app, container, func() { fillRecent(app, container, recentList) })
		}
		return event
	})

	leftPanel.
		AddItem(title, 0, 1, false).
		AddItem(newSheetBtn, 0, 3, true).
		AddItem(openFileBtn, 0, 3, true).
		AddItem(optionsBtn, 0, 3, true)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	recentTitle := tview.NewTextView().
		SetText("[::b]Recent Files[::-]").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	recentTitle.SetBorder(true).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))

	recentList = tview.NewList()
	fillRecent(app, container, recentList)

	recentList.SetBorder(true).SetBorderColor(tcell.NewRGBColor(70, 160, 170))
	recentList.SetSelectedBackgroundColor(tcell.NewRGBColor(28, 110, 128))
	recentList.SetSelectedTextColor(tcell.ColorWhite)
	recentList.SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))

	rightPanel.
		AddItem(recentTitle, 3, 0, false).
		AddItem(recentList, 0, 1, true)

	container.
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 3, false)

	container.SetBorder(true).
		SetTitle(" GoSheet  ·  arrows or tab to move  ·  Enter to choose  ·  Esc quits ").
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(10, 12, 18))

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
		case event.Key() == tcell.KeyLeft && (event.Modifiers() == 0 || utils.HasAlt(event.Modifiers())),
			event.Key() == tcell.KeyBacktab:
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
		case event.Key() == tcell.KeyRight && (event.Modifiers() == 0 || utils.HasAlt(event.Modifiers())),
			event.Key() == tcell.KeyTab:
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

func fillRecent(app *tview.Application, container tview.Primitive, recentList *tview.List) {
	if recentList == nil {
		return
	}
	recentList.Clear()
	filenames, locations := fileop.GetRecentFileList()
	if len(filenames) == 0 {
		recentList.AddItem("No recent files", "", 0, nil)
		recentList.SetSelectedTextColor(tcell.ColorGray)
		return
	}
	recentList.SetSelectedTextColor(tcell.ColorWhite)
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
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		}
		recentList.AddItem(filename, secondaryText, shortcut, func() {
			if _, err := os.Stat(loc); os.IsNotExist(err) {
				ShowErrorModal(app, container, fmt.Sprintf("File not found: %s", loc))
			} else {
				selectFile(app, loc)
			}
		})
	}
}

func ShowErrorModal(app *tview.Application, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Error\n\n%s", message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			sheetview.Back(app, returnTo, returnTo)
		})

	uitheme.StyleModalWarn(modal, "Error")
	app.SetRoot(modal, true).SetFocus(modal)
}
