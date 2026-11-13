// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dialogs.go provides the find/find&replace dialogs

package navigation

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func FindDialog(app *tview.Application, table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	form := tview.NewForm()
	form.AddInputField("Find:", lastSearchTerm, 40, nil, nil)
	form.AddCheckbox("Case sensitive", lastCaseSensitive, func(checked bool) {
		lastCaseSensitive = checked
	})
	form.AddCheckbox("Match whole word", lastMatchWholeWord, func(checked bool) {
		lastMatchWholeWord = checked
	})

	form.AddButton("Find Next", func() {
		searchTerm := form.GetFormItem(0).(*tview.InputField).GetText()
		if searchTerm == "" {
			return
		}

		lastSearchTerm = searchTerm
		found := findNext(table, searchTerm, lastCaseSensitive, lastMatchWholeWord, globalData, globalViewport, RenderVisible)

		if !found {
			ShowWarningModal(app, table, "No matches found")
			lastSearchRow, lastSearchCol = 0, 0
		} else {
			sheetview.Show(app)
		}
	})

	form.AddButton("Find Previous", func() {
		searchTerm := form.GetFormItem(0).(*tview.InputField).GetText()
		if searchTerm == "" {
			return
		}

		lastSearchTerm = searchTerm
		found := findPrevious(table, searchTerm, lastCaseSensitive, lastMatchWholeWord, globalData, globalViewport, RenderVisible)

		if !found {
			ShowWarningModal(app, table, "No matches found")
			lastSearchRow, lastSearchCol = 0, 0
		} else {
			sheetview.Show(app)
		}
	})

	form.AddButton("Clear", func() {
		lastSearchTerm = ""
		lastCaseSensitive = false
		lastMatchWholeWord = false
		lastSearchRow, lastSearchCol = 0, 0

		sheetview.Show(app)
	})

	form.AddButton("Close", func() {
		lastSearchRow, lastSearchCol = 0, 0
		sheetview.Show(app)
	})

	uitheme.StyleForm(form, "Find")
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			lastSearchRow, lastSearchCol = 0, 0
			sheetview.Show(app)
			return nil
		}
		return event
	})

	dialog := uitheme.CenteredModal(form, 64, 12)
	app.SetRoot(dialog, true).SetFocus(form)
}

// Replace functions
func ReplaceDialog(app *tview.Application, table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	form := tview.NewForm()
	var dialog *tview.Flex
	form.AddInputField("Find:", "", 40, nil, nil)
	form.AddInputField("Replace with:", "", 40, nil, nil)

	var caseSensitive bool
	var matchWholeWord bool
	form.AddCheckbox("Case sensitive", false, func(checked bool) {
		caseSensitive = checked
	})
	form.AddCheckbox("Match whole word", false, func(checked bool) {
		matchWholeWord = checked
	})

	form.AddButton("Replace", func() {
		findTerm := form.GetFormItem(0).(*tview.InputField).GetText()
		replaceTerm := form.GetFormItem(1).(*tview.InputField).GetText()

		if findTerm == "" {
			return
		}

		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		absRow, absCol := globalViewport.ToAbsolute(visualRow, visualCol)

		if replaceInCell(app, table, absRow, absCol, findTerm, replaceTerm, caseSensitive, matchWholeWord, globalData, globalViewport) {
			ShowWarningModal(app, table, fmt.Sprintf("Replaced occurrence/s in cell %s%d!", utils.ColumnName(absCol), absRow))
		} else {
			ShowWarningModal(app, table, "No matches found in selected cell!")
		}
	})

	form.AddButton("Replace & Find Next", func() {
		findTerm := form.GetFormItem(0).(*tview.InputField).GetText()
		replaceTerm := form.GetFormItem(1).(*tview.InputField).GetText()

		if findTerm == "" {
			return
		}

		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		absRow, absCol := globalViewport.ToAbsolute(visualRow, visualCol)

		if replaceInCell(app, table, absRow, absCol, findTerm, replaceTerm, caseSensitive, matchWholeWord, globalData, globalViewport) {
			lastSearchTerm = findTerm
			lastCaseSensitive = caseSensitive
			lastMatchWholeWord = matchWholeWord
			lastSearchRow, lastSearchCol = absRow, absCol

			if !findNext(table, findTerm, caseSensitive, matchWholeWord, globalData, globalViewport, RenderVisible) {
				ShowWarningModal(app, table, "No more matches found")
				lastSearchRow, lastSearchCol = 0, 0
			}
		} else {
			lastSearchTerm = findTerm
			lastCaseSensitive = caseSensitive
			lastMatchWholeWord = matchWholeWord
			lastSearchRow, lastSearchCol = absRow, absCol

			if !findNext(table, findTerm, caseSensitive, matchWholeWord, globalData, globalViewport, RenderVisible) {
				ShowWarningModal(app, table, "No matches found")
				lastSearchRow, lastSearchCol = 0, 0
			}
		}
	})

	form.AddButton("Replace All", func() {
		findTerm := form.GetFormItem(0).(*tview.InputField).GetText()
		replaceTerm := form.GetFormItem(1).(*tview.InputField).GetText()

		if findTerm == "" {
			return
		}

		matchCount := countMatches(globalData, findTerm, caseSensitive, matchWholeWord)

		if matchCount == 0 {
			ShowWarningModal(app, table, "No matches found")
			return
		}

		modal := tview.NewModal().
			SetText(fmt.Sprintf("Replace %d occurrence(s)?", matchCount)).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					count := replaceAll(table, findTerm, replaceTerm, caseSensitive, matchWholeWord, globalData, globalViewport, RenderVisible)
					ShowWarningModal(app, table, fmt.Sprintf("Replaced %d occurrence(s)", count))
				} else {
					app.SetRoot(dialog, true).SetFocus(form)
				}
			})
		uitheme.StyleModal(modal, "Replace all")
		app.SetRoot(modal, true).SetFocus(modal)
	})

	form.AddButton("Close", func() {
		sheetview.Show(app)
	})

	uitheme.StyleForm(form, "Replace")
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			sheetview.Show(app)
			return nil
		}
		return event
	})

	dialog = uitheme.CenteredModal(form, 76, 14)
	app.SetRoot(dialog, true).SetFocus(form)
}
