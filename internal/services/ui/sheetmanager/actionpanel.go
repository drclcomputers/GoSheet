// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// actionpanel.go provides the action panel on the right of the sheet manager dialog

package sheetmanager

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createActionPanel creates an enhanced button panel with icons and descriptions
func createActionPanel(app *tview.Application, table *tview.Table,
	callbacks SheetManagerCallbacks, list *tview.List, infoPanel *tview.TextView) *tview.Flex {
	
	actionPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	actionPanel.SetBorder(true).
		SetTitle(" Actions ").
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleAlign(tview.AlignLeft)

	createActionBtn := func(icon, label, shortcut, description string, color tcell.Color, action func()) *tview.Box {
		btn := tview.NewBox().
			SetBorder(true).
			SetBorderColor(color)
		
		btn.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			tview.Print(screen, fmt.Sprintf(" %s %s", icon, label), x+1, y+1, width-2, tview.AlignLeft, color)
			
			tview.Print(screen, shortcut, x+width-len(shortcut)-2, y+1, len(shortcut), tview.AlignRight, tcell.ColorYellow)
			
			if len(description) > 0 {
				tview.Print(screen, description, x+1, y+2, width-2, tview.AlignLeft, tcell.ColorGray)
			}
			
			return x + 1, y + 3, width - 2, height - 3
		})

		btn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEnter {
				action()
				return nil
			}
			return event
		})

		return btn
	}

	newSheetBtn := createActionBtn(
		"", "New Sheet", "Alt+N",
		"Create a blank sheet",
		tcell.ColorGreen,
		func() { showAddSheetDialog(app, callbacks, actionPanel, list, infoPanel) },
	)

	renameBtn := createActionBtn(
		"", "Rename", "Alt+R",
		"Change sheet name",
		tcell.ColorBlue,
		func() { showRenameSheetFromManager(app, callbacks, actionPanel, list, infoPanel) },
	)

	duplicateBtn := createActionBtn(
		"", "Duplicate", "Alt+C",
		"Copy entire sheet",
		tcell.ColorLightBlue,
		func() { duplicateSheetFromManager(app, table, callbacks, list, infoPanel) },
	)

	moveBtn := createActionBtn(
		"", "Move/Reorder", "Alt+M",
		"Change sheet position",
		tcell.ColorYellow,
		func() { showMoveSheetDialog(app, callbacks, actionPanel, list, infoPanel) },
	)

	deleteBtn := createActionBtn(
		"", "Delete", "Alt+D",
		"Remove sheet permanently",
		tcell.ColorRed,
		func() { confirmDeleteSheetFromManager(app, callbacks, actionPanel, list, infoPanel) },
	)

	switchBtn := createActionBtn(
		"", "Switch To", "Enter",
		"Open selected sheet",
		tcell.ColorLightBlue,
		func() { switchToSelectedSheet(app, table, callbacks, list) },
	)

	exitBtn := createActionBtn(
		"", "Exit", "Esc",
		"Exit sheet manager",
		tcell.ColorLightBlue,
		func() { app.SetRoot(table, true).SetFocus(table) },
	)

	actionPanel.
		AddItem(newSheetBtn, 4, 0, false).
		AddItem(renameBtn, 4, 0, false).
		AddItem(duplicateBtn, 4, 0, false).
		AddItem(moveBtn, 4, 0, false).
		AddItem(deleteBtn, 4, 0, false).
		AddItem(switchBtn, 4, 0, false).
		AddItem(exitBtn, 4, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	return actionPanel
}
