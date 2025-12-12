// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// errors.go provides various error modals

package cellui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Modal used for showing that a cell is uneditable
func ShowUneditableModal(app *tview.Application, table *tview.Table, row, col int32, RecordCellEdit func(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error, RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Cell %s%d is marked as uneditable.\nDo you wish to continue editing anyway?", utils.ColumnName(col), row)).
		AddButtons([]string{"Yes", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Yes":
				EditCellDialog(app, table, row, col, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)
			case "Cancel":
				app.SetRoot(table, true).SetFocus(table)
			}
		})

	modal.SetBorder(true).SetTitle(" Confirm Edit ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}


// Shows a modal in case of a type mismatch
func ShowTypeErrorModal(app *tview.Application, form *tview.Flex, c *cell.Cell, leftForm *tview.Form) {
	message := "Invalid value for cell type.\n\nPlease enter a value matching the expected format or change the cell type."

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			app.SetRoot(form, true).SetFocus(leftForm)
		})

	modal.SetBorder(true)
	modal.SetTitle(" Type Mismatch ⚠️")
	modal.SetTitleAlign(tview.AlignCenter)

	app.SetRoot(modal, true).SetFocus(modal)
}

// Shows an error modal according to the type of formula error
func ShowFormulaErrorModal(app *tview.Application, form *tview.Flex, errorCode string, errorMessage string, leftForm *tview.Form) {
	var title string
	var message string

	switch errorCode {
	case "#ARGS!":
		title = " Argument Error ⚠️"
		message = fmt.Sprintf("Formula argument error:\n\n%s\n\nCheck the number and types of arguments in your formula.", errorMessage)
	case "#DIV/0!":
		title = " Division by Zero ⚠️"
		message = "Cannot divide by zero.\n\nCheck your formula for division operations."
	case "#VALUE!":
		title = " Value Error ⚠️"
		message = fmt.Sprintf("Invalid value in formula:\n\n%s\n\nCheck data types and values.", errorMessage)
	case "#CIRC!":
		title = " Circular Reference ⚠️"
		message = "Circular dependency detected.\n\nA cell cannot reference itself directly or indirectly."
	case "#REF!":
		title = " Reference Error ⚠️"
		message = "Invalid cell reference.\n\nCheck that all referenced cells exist."
	case "#ERROR!":
		title = " Formula Error ⚠️"
		message = fmt.Sprintf("An error occurred in the formula:\n\n%s", errorMessage)
	default:
		title = " Error ⚠️"
		message = fmt.Sprintf("An error occurred:\n\n%s", errorMessage)
	}

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			app.SetRoot(form, true).SetFocus(leftForm)
		})

	modal.SetBorder(true)
	modal.SetTitle(title)
	modal.SetTitleAlign(tview.AlignCenter)
	modal.SetBackgroundColor(tcell.ColorBlack)

	app.SetRoot(modal, true).SetFocus(modal)
}
