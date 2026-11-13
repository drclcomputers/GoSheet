// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dialog.go provides the main dialog for managing sheets

package sheetmanager

import (
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowSheetManager displays the comprehensive sheet management dialog
func ShowSheetManager(app *tview.Application, table *tview.Table, callbacks SheetManagerCallbacks) {
	list := uitheme.StyleList(tview.NewList(), "Sheets").
		ShowSecondaryText(true)
	list.SetTitleAlign(tview.AlignLeft)

	updateSheetList(list, callbacks)

	infoPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetTextColor(uitheme.Text).
		SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
	uitheme.Paint(infoPanel, "This sheet")
	infoPanel.SetTitleAlign(tview.AlignLeft)

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 3, true).
		AddItem(infoPanel, 12, 0, false)

	mainContent := tview.NewFlex()

	hint := tview.NewTextView().
		SetDynamicColors(true).
		SetText(uitheme.Markup(" {val}Enter{end} open    {val}Alt+N{end} new    {val}Alt+R{end} rename    {val}Alt+C{end} duplicate    {val}Alt+M{end} move    {val}Alt+D{end} delete    {val}Esc{end} back")).
		SetTextColor(uitheme.Text)
	hint.SetBackgroundColor(uitheme.Ink)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(hint, 1, 0, false)
	uitheme.Paint(mainLayout, "Sheets")
	mainLayout.SetTitleAlign(tview.AlignCenter)

	actionPanel := createActionPanel(app, table, callbacks, list, infoPanel)

	mainContent.
		AddItem(leftPanel, 0, 2, true).
		AddItem(actionPanel, 45, 0, false)

	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			sheetview.Show(app)
			return nil
		case tcell.KeyEnter:
			switchToSelectedSheet(app, table, callbacks, list)
			return nil
		}

		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case 'n', 'N':
				showAddSheetDialog(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'r', 'R':
				showRenameSheetFromManager(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'd', 'D':
				confirmDeleteSheetFromManager(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'm', 'M':
				showMoveSheetDialog(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'c', 'C':
				duplicateSheetFromManager(app, table, callbacks, list, infoPanel)
				return nil
			case 's', 'S':
				switchToSelectedSheet(app, table, callbacks, list)
				return nil
			}
		}

		return event
	})

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		sheets := callbacks.GetSheets()
		if index >= 0 && index < len(sheets) {
			infoPanel.SetText(getSheetInfoText(sheets[index], index, len(sheets)))
		}
	})

	app.SetRoot(mainLayout, true).SetFocus(list)
}
