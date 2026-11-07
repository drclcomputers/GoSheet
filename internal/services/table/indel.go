// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// indel.go contains functions for inserting and deleting rows/columns.

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/ui"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"
	"maps"

	"github.com/rivo/tview"
)

var totalRows, totalCols int32

// DELETE FUNCTIONS
func deleteRowCol(app *tview.Application, table *tview.Table) {
	activeViewport := GetActiveViewport()

	if activeViewport == nil {
		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 && col > 0 {
		deleteCol(app, table, col)
	} else if col == 0 && row > 0 {
		deleteRow(app, table, row)
	} else {
		ui.ShowWarningModal(app, table, "Select a column or a row before deleting it.")
	}
}

func deleteCol(app *tview.Application, table *tview.Table, col int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()

	if activeData == nil || activeViewport == nil {
		return
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you wish to delete col %s?", utils.ColumnName(col))).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()

				if activeData == nil || activeViewport == nil {
					sheetview.Show(app)
					return
				}

				keysToDelete := make([][2]int, 0)
				keysToUpdate := make(map[[2]int]*cell.Cell)

				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(c) == col {
						keysToDelete = append(keysToDelete, key)
					} else if int32(c) > col {
						cellData.Column--
						keysToUpdate[[2]int{r, c - 1}] = cellData
						keysToDelete = append(keysToDelete, key)
					}
				}

				for _, key := range keysToDelete {
					delete(activeData, key)
				}
				maps.Copy(activeData, keysToUpdate)

				RenderVisible(table, activeViewport, activeData)
				sheetview.Show(app)
			} else {
				sheetview.Show(app)
			}
		})
	uitheme.StyleModalWarn(modal, "Delete column")
	app.SetRoot(modal, true).SetFocus(modal)
}

func deleteRow(app *tview.Application, table *tview.Table, row int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you wish to delete row %d?", row)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()

				if activeData == nil || activeViewport == nil {
					sheetview.Show(app)
					return
				}

				keysToDelete := make([][2]int, 0)
				keysToUpdate := make(map[[2]int]*cell.Cell)

				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(r) == row {
						keysToDelete = append(keysToDelete, key)
					} else if int32(r) > row {
						cellData.Row--
						keysToUpdate[[2]int{r - 1, c}] = cellData
						keysToDelete = append(keysToDelete, key)
					}
				}

				for _, key := range keysToDelete {
					delete(activeData, key)
				}
				maps.Copy(activeData, keysToUpdate)

				RenderVisible(table, activeViewport, activeData)
				sheetview.Show(app)
			} else {
				sheetview.Show(app)
			}
		})
	uitheme.StyleModalWarn(modal, "Delete row")
	app.SetRoot(modal, true).SetFocus(modal)
}

// INSERT FUNCTIONS
func insertRowCol(app *tview.Application, table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()

	if activeData == nil || activeViewport == nil {
		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 && col > 0 {
		insertCol(app, table, col)
	} else if col == 0 && row > 0 {
		insertRow(app, table, row)
	} else {
		ui.ShowWarningModal(app, table, "Select a column or row header to insert before it.")
	}
}

func insertCol(app *tview.Application, table *tview.Table, col int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Insert a new column before %s?", utils.ColumnName(col))).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()

				if activeData == nil || activeViewport == nil {
					sheetview.Show(app)
					return
				}

				keysToDelete := make([][2]int, 0)
				keysToUpdate := make(map[[2]int]*cell.Cell)

				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(c) >= col {
						cellData.Column++
						keysToUpdate[[2]int{r, c + 1}] = cellData
						keysToDelete = append(keysToDelete, key)
					}
				}

				for _, key := range keysToDelete {
					delete(activeData, key)
				}
				maps.Copy(activeData, keysToUpdate)

				RenderVisible(table, activeViewport, activeData)
				sheetview.Show(app)
			} else {
				sheetview.Show(app)
			}
		})
	uitheme.StyleModal(modal, "Insert column")
	app.SetRoot(modal, true).SetFocus(modal)
}

func insertRow(app *tview.Application, table *tview.Table, row int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Insert a new row before row %d?", row)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()

				if activeData == nil || activeViewport == nil {
					sheetview.Show(app)
					return
				}

				keysToDelete := make([][2]int, 0)
				keysToUpdate := make(map[[2]int]*cell.Cell)

				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(r) >= row {
						cellData.Row++
						keysToUpdate[[2]int{r + 1, c}] = cellData
						keysToDelete = append(keysToDelete, key)
					}
				}

				for _, key := range keysToDelete {
					delete(activeData, key)
				}
				maps.Copy(activeData, keysToUpdate)

				RenderVisible(table, activeViewport, activeData)
				sheetview.Show(app)
			} else {
				sheetview.Show(app)
			}
		})
	uitheme.StyleModal(modal, "Insert row")
	app.SetRoot(modal, true).SetFocus(modal)
}
