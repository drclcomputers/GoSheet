// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dialog.go provides the main dialog for managing sheets

package sheetmanager

import (
	"gosheet/internal/services/sheetview"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowSheetManager displays the comprehensive sheet management dialog
func ShowSheetManager(app *tview.Application, table *tview.Table, callbacks SheetManagerCallbacks) {
	list := tview.NewList().
		SetSelectedBackgroundColor(tcell.NewRGBColor(28, 110, 128)).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.NewRGBColor(232, 236, 244)).
		SetSecondaryTextColor(tcell.NewRGBColor(138, 160, 184)).
		ShowSecondaryText(true)
	list.SetBorder(true).
		SetTitle(" Sheets ").
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(14, 17, 26)).
		SetTitleAlign(tview.AlignLeft)

	updateSheetList(list, callbacks)

	infoPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
	infoPanel.SetBorder(true).
		SetTitle(" This sheet ").
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(18, 22, 34)).
		SetTitleAlign(tview.AlignLeft)

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 3, true).
		AddItem(infoPanel, 12, 0, false)

	mainContent := tview.NewFlex()

	hint := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [#f0d090]Enter[-] open    [#f0d090]Alt+N[-] new    [#f0d090]Alt+R[-] rename    [#f0d090]Alt+C[-] duplicate    [#f0d090]Alt+M[-] move    [#f0d090]Alt+D[-] delete    [#f0d090]Esc[-] back")
	hint.SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(hint, 1, 0, false)
	mainLayout.SetBackgroundColor(tcell.NewRGBColor(10, 12, 18))

	mainLayout.SetBorder(true).
		SetTitle(" Sheets ").
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetTitleAlign(tview.AlignCenter)

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
