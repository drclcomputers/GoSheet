// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// sheettabs.go provides sheet tab bar and quick navigation

package table

import (
	"fmt"
	"gosheet/internal/services/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowSheetContextMenu shows quick options for sheet management
func ShowSheetContextMenu(app *tview.Application, table *tview.Table) {
	if globalWorkbook == nil {
		return
	}
	
	sheet := globalWorkbook.GetActiveSheet()
	
	modal := tview.NewModal().
		SetText(fmt.Sprintf(
			"[yellow::b]Sheet: %s[::-]\n\n"+
				"Quick Actions:",
			sheet.Name,
		)).
		AddButtons([]string{
			"Rename",
			"Duplicate",
			"Delete",
			"Add New",
			"Manager",
			"Cancel",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Rename":
				showQuickRenameDialog(app, table)
			case "Duplicate":
				quickDuplicateSheet(app, table)
			case "Delete":
				if len(globalWorkbook.Sheets) <= 1 {
					ui.ShowWarningModal(app, table, "Cannot delete the last sheet!")
				} else {
					confirmDeleteSheet(app, table)
				}
			case "Add New":
				quickAddSheet(app, table)
			case "Manager":
				ShowSheetManagerDialog(app, table)
			default:
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	
	modal.SetBorder(true).
		SetTitle(" Sheet Options ").
		SetBorderColor(tcell.ColorYellow)
	
	app.SetRoot(modal, true).SetFocus(modal)
}

// showQuickRenameDialog shows a quick rename dialog
func showQuickRenameDialog(app *tview.Application, table *tview.Table) {
	sheet := globalWorkbook.GetActiveSheet()
	
	form := tview.NewForm()
	nameInput := tview.NewInputField().
		SetLabel("Sheet Name: ").
		SetText(sheet.Name).
		SetFieldWidth(30)
	
	form.AddFormItem(nameInput).
		AddButton("Rename", func() {
			newName := nameInput.GetText()
			if newName == "" {
				ui.ShowWarningModal(app, form, "Sheet name cannot be empty!")
				return
			}
			
			if err := RenameSheetByIndex(globalWorkbook.ActiveSheet, newName); err != nil {
				ui.ShowWarningModal(app, form, err.Error())
				return
			}
			
			updateTableTitle(table)
			MarkAsModified(table)
			app.SetRoot(table, true).SetFocus(table)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(table, true).SetFocus(table)
		})
	
	form.SetBorder(true).
		SetTitle(" Rename Sheet ").
		SetBorderColor(tcell.ColorYellow)
	
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(table, true).SetFocus(table)
			return nil
		}
		return event
	})
	
	app.SetRoot(form, true).SetFocus(form)
}

// quickAddSheet quickly adds a new sheet
func quickAddSheet(app *tview.Application, table *tview.Table) {
	newName := fmt.Sprintf("Sheet%d", len(globalWorkbook.Sheets)+1)
	
	if err := AddSheetWithName(newName); err != nil {
		ui.ShowWarningModal(app, table, err.Error())
		return
	}
	
	globalWorkbook.SwitchToSheet(len(globalWorkbook.Sheets) - 1)
	
	sheet := globalWorkbook.GetActiveSheet()
	RenderVisible(table, sheet.Viewport, sheet.Data)
	updateTableTitle(table)
	MarkAsModified(table)
	
	app.SetRoot(table, true).SetFocus(table)
}

// quickDuplicateSheet duplicates the current sheet
func quickDuplicateSheet(app *tview.Application, table *tview.Table) {
	if err := DuplicateSheetByIndex(globalWorkbook.ActiveSheet); err != nil {
		ui.ShowWarningModal(app, table, err.Error())
		return
	}
	
	globalWorkbook.SwitchToSheet(len(globalWorkbook.Sheets) - 1)
	
	sheet := globalWorkbook.GetActiveSheet()
	RenderVisible(table, sheet.Viewport, sheet.Data)
	updateTableTitle(table)
	MarkAsModified(table)
	
	app.SetRoot(table, true).SetFocus(table)
}

// confirmDeleteSheet confirms and deletes the current sheet
func confirmDeleteSheet(app *tview.Application, table *tview.Table) {
	sheet := globalWorkbook.GetActiveSheet()
	
	modal := tview.NewModal().
		SetText(fmt.Sprintf(
			"Delete sheet '%s'?\n\n"+
				"This will delete %d cells and cannot be undone!",
			sheet.Name,
			len(sheet.Data),
		)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := DeleteSheetByIndex(globalWorkbook.ActiveSheet); err != nil {
					ui.ShowWarningModal(app, table, err.Error())
					app.SetRoot(table, true).SetFocus(table)
					return
				}
				
				sheet := globalWorkbook.GetActiveSheet()
				RenderVisible(table, sheet.Viewport, sheet.Data)
				updateTableTitle(table)
				MarkAsModified(table)
			}
			app.SetRoot(table, true).SetFocus(table)
		})
	
	modal.SetBackgroundColor(tcell.ColorDarkRed).
		SetBorderColor(tcell.ColorRed)
	
	app.SetRoot(modal, true).SetFocus(modal)
}

// SwitchSheet switches to a different sheet by index
func SwitchSheet(app *tview.Application, table *tview.Table, index int) {
	if err := globalWorkbook.SwitchToSheet(index); err != nil {
		return
	}
	
	sheet := globalWorkbook.GetActiveSheet()
	RenderVisible(table, sheet.Viewport, sheet.Data)
	updateTableTitle(table)
}
