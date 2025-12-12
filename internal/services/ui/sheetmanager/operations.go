// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// operations.go provides functions that manage sheets

package sheetmanager

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showAddSheetDialog shows enhanced dialog to add a new sheet
func showAddSheetDialog(app *tview.Application, 
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkGreen)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	sheets := callbacks.GetSheets()
	defaultName := fmt.Sprintf("Sheet%d", len(sheets)+1)
	
	nameInput := tview.NewInputField().
		SetLabel("Sheet Name: ").
		SetText(defaultName).
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	form.AddFormItem(nameInput).
		AddButton("Create", func() {
			name := strings.TrimSpace(nameInput.GetText())
			if name == "" {
				ShowWarningModal(app, form, "Sheet name cannot be empty!")
				return
			}

			if err := callbacks.AddSheet(name); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle(" + Add New Sheet ").
		SetBorderColor(tcell.ColorGreen).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// showRenameSheetFromManager shows enhanced rename dialog
func showRenameSheetFromManager(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	sheet := sheets[selectedIndex]

	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkBlue)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	nameInput := tview.NewInputField().
		SetLabel(" New Name: ").
		SetText(sheet.Name).
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	form.AddFormItem(nameInput).
		AddButton("Rename", func() {
			newName := strings.TrimSpace(nameInput.GetText())
			if newName == "" {
				ShowWarningModal(app, form, "Sheet name cannot be empty!")
				return
			}

			if err := callbacks.RenameSheet(selectedIndex, newName); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			callbacks.UpdateTableTitle()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("x Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle( " Rename Sheet ").
		SetBorderColor(tcell.ColorBlue).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// confirmDeleteSheetFromManager shows enhanced deletion confirmation
func confirmDeleteSheetFromManager(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	sheets := callbacks.GetSheets()
	if len(sheets) <= 1 {
		ShowWarningModal(app, returnTo, "Cannot delete the last sheet!\n\nA workbook must have at least one sheet.")
		return
	}

	selectedIndex := list.GetCurrentItem()
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	sheet := sheets[selectedIndex]

	modal := tview.NewModal().
		SetText(fmt.Sprintf(
			"[red::b]⚠️  DELETE SHEET[::-]\n\n"+
				"Are you sure you want to delete:\n"+
				"[yellow]'%s'[-]?\n\n"+
				"[white]This will permanently remove:[-]\n"+
				"  • [white]%d[-] cells with data\n"+
				"  • All formulas and formatting\n"+
				"  • All undo/redo history\n\n"+
				"[red::b]This action cannot be undone![::-]",
			sheet.Name,
			sheet.CellCount,
		)).
		AddButtons([]string{"X Delete", "x Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if strings.Contains(buttonLabel, "Delete") {
				if err := callbacks.DeleteSheet(selectedIndex); err != nil {
					ShowWarningModal(app, returnTo, err.Error())
					app.SetRoot(returnTo, true).SetFocus(list)
					return
				}

				updateSheetList(list, callbacks)
				callbacks.UpdateTabBar()
				callbacks.RenderActiveSheet()
				callbacks.UpdateTableTitle()
				infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
				callbacks.MarkAsModified()
			}
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	modal.SetBackgroundColor(tcell.ColorDarkRed).
		SetBorderColor(tcell.ColorRed)

	app.SetRoot(modal, true).SetFocus(modal)
}

// duplicateSheetFromManager duplicates with visual feedback
func duplicateSheetFromManager(app *tview.Application, table *tview.Table,
	callbacks SheetManagerCallbacks, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	if err := callbacks.DuplicateSheet(selectedIndex); err != nil {
		ShowWarningModal(app, table, "X "+err.Error())
		return
	}

	updateSheetList(list, callbacks)
	callbacks.UpdateTabBar()
	infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
	callbacks.MarkAsModified()

	list.SetCurrentItem(len(sheets))
}

// showMoveSheetDialog shows enhanced reorder dialog
func showMoveSheetDialog(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkGoldenrod)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	positions := make([]string, len(sheets))
	for i := range sheets {
		if i == selectedIndex {
			positions[i] = fmt.Sprintf("Position %d (current)", i+1)
		} else {
			positions[i] = fmt.Sprintf("Position %d", i+1)
		}
	}

	form.AddDropDown("Move to:", positions, selectedIndex, nil).
		AddButton("Move", func() {
			newPos, _ := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
			
			if newPos == selectedIndex {
				app.SetRoot(returnTo, true).SetFocus(list)
				return
			}

			if err := callbacks.MoveSheet(selectedIndex, newPos); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			list.SetCurrentItem(newPos)
			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle(" Move/Reorder Sheet ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// switchToSelectedSheet switches with smooth feedback
func switchToSelectedSheet(app *tview.Application, table *tview.Table, 
	callbacks SheetManagerCallbacks, list *tview.List) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	if err := callbacks.SwitchToSheet(selectedIndex); err != nil {
		return
	}

	callbacks.UpdateTabBar()
	callbacks.UpdateTableTitle()
	callbacks.RenderActiveSheet()
	
	app.SetRoot(table, true).SetFocus(table)
}


