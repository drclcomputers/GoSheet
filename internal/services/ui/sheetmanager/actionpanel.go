// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// actionpanel.go provides the action panel on the right of the sheet manager dialog

package sheetmanager

import (
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createActionPanel creates an enhanced button panel with icons and descriptions
func createActionPanel(app *tview.Application, table *tview.Table,
	callbacks SheetManagerCallbacks, list *tview.List, infoPanel *tview.TextView) *tview.Flex {

	actionPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	uitheme.Paint(actionPanel, "Do")
	actionPanel.SetTitleAlign(tview.AlignLeft)

	createActionBtn := func(label, shortcut, description string, action func()) *tview.Box {
		btn := tview.NewBox().
			SetBorder(true).
			SetBorderColor(uitheme.Line).
			SetBackgroundColor(uitheme.Ink)

		btn.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			tview.Print(screen, " "+label, x+1, y+1, width-len(shortcut)-4, tview.AlignLeft, uitheme.Text)
			tview.Print(screen, shortcut, x+width-len(shortcut)-2, y+1, len(shortcut), tview.AlignRight, uitheme.Amber)
			if description != "" && height > 3 {
				tview.Print(screen, " "+description, x+1, y+2, width-2, tview.AlignLeft, uitheme.Label)
			}
			return x + 1, y + 3, width - 2, height - 3
		})
		btn.SetFocusFunc(func() { btn.SetBorderColor(uitheme.Amber) })
		btn.SetBlurFunc(func() { btn.SetBorderColor(uitheme.Line) })

		btn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEnter {
				action()
				return nil
			}
			return event
		})

		return btn
	}

	newSheetBtn := createActionBtn("New sheet", "Alt+N", "Blank sheet",
		func() { showAddSheetDialog(app, callbacks, actionPanel, list, infoPanel) })
	renameBtn := createActionBtn("Rename", "Alt+R", "Change the name",
		func() { showRenameSheetFromManager(app, callbacks, actionPanel, list, infoPanel) })
	duplicateBtn := createActionBtn("Duplicate", "Alt+C", "Copy this sheet",
		func() { duplicateSheetFromManager(app, table, callbacks, list, infoPanel) })
	moveBtn := createActionBtn("Move", "Alt+M", "Change the order",
		func() { showMoveSheetDialog(app, callbacks, actionPanel, list, infoPanel) })
	deleteBtn := createActionBtn("Delete", "Alt+D", "Remove this sheet",
		func() { confirmDeleteSheetFromManager(app, callbacks, actionPanel, list, infoPanel) })
	switchBtn := createActionBtn("Open", "Enter", "Switch to it",
		func() { switchToSelectedSheet(app, table, callbacks, list) })
	exitBtn := createActionBtn("Back", "Esc", "Return to the grid",
		func() { sheetview.Show(app) })

	actionPanel.
		AddItem(newSheetBtn, 4, 0, false).
		AddItem(renameBtn, 4, 0, false).
		AddItem(duplicateBtn, 4, 0, false).
		AddItem(moveBtn, 4, 0, false).
		AddItem(deleteBtn, 4, 0, false).
		AddItem(switchBtn, 4, 0, false).
		AddItem(exitBtn, 4, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(uitheme.Panel), 0, 1, false)

	return actionPanel
}
