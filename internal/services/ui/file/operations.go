// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// fileUI.go provides functions to display file-related dialogs in the UI.

package file

import (
	"fmt"
	"os"
	"path/filepath"

	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"
	"gosheet/internal/services/ui/cellui"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Helper functions (add these to fileUI.go)
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func showOverwriteConfirmation(app *tview.Application, table *tview.Table, filename string, format FileFormatUI, shouldExit bool, globalData map[[2]int]*cell.Cell) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("File '%s' already exists.\n\nDo you want to overwrite it?", filepath.Base(filename))).
		AddButtons([]string{"Overwrite", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Overwrite" {
				performSave(app, table, filename, format, shouldExit, globalData)
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	
	modal.SetBackgroundColor(tcell.ColorDarkRed).SetBorderColor(tcell.ColorRed)
	app.SetRoot(modal, true).SetFocus(modal)
}

func performSave(app *tview.Application, table *tview.Table, filename string, format FileFormatUI, shouldExit bool, globalData map[[2]int]*cell.Cell) {
	err := format.SaveFunc(table, filename, globalData)
	
	if err != nil {
		ShowErrorModal(app, table, fmt.Sprintf("Failed to save file:\n%s", err.Error()))
		return
	}

	if format.Extension == ".json" || format.Extension == ".gsheet" || format.Extension == ".txt" {
		fileop.AddToRecentFiles(filename)
	}
	
	message := fmt.Sprintf("File saved successfully!\n\nLocation:\n%s", filename)
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if shouldExit {
				app.Stop()
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	
	modal.SetBackgroundColor(tcell.ColorDarkGreen).SetBorderColor(tcell.ColorGreen)
	app.SetRoot(modal, true).SetFocus(modal)
}


// Modal used for showing that a cell is uneditable
func ShowUneditableModal(app *tview.Application, table *tview.Table, row, col int32, RecordCellEdit func(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error, RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Cell %s%d is marked as uneditable.\nDo you wish to continue editing anyway?", utils.ColumnName(col), row)).
		AddButtons([]string{"Yes", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Yes":
				cellui.EditCellDialog(app, table, row, col, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)
			case "Cancel":
				app.SetRoot(table, true).SetFocus(table)
			}
		})

	modal.SetBorder(true).SetTitle(" Confirm Edit ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}
