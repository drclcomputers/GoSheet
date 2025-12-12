// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// goto.go provides a goto dialog

package navigation

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


func navigateToCell(table *tview.Table, absRow, absCol int32, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	globalViewport.TopRow = absRow
	globalViewport.LeftCol = absCol
	
	RenderVisible(table, globalViewport, globalData)
	
	table.Select(1, 1)
}

// Quick navigation
func GoToCellModal(app *tview.Application, table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	form := tview.NewForm()
	form.AddInputField("Cell Address (e.g. AA56):", "", 20, nil, nil)
	
	form.AddButton("Go", func() {
		input := strings.TrimSpace(strings.ToUpper(form.GetFormItem(0).(*tview.InputField).GetText()))
		unformattedPartString, unformattedPartNumber := SplitAlphaNumeric(input)

		if unformattedPartNumber == "" || unformattedPartString == "" {
			ShowWarningModal(app, table, "Invalid format! Example: AA56")
			return
		}

		partColumn := utils.ColumnNumber(unformattedPartString)
		if partColumn == 0 {
			ShowWarningModal(app, table, "Invalid column!")
			return
		}
		if partColumn >= (1<<31 - 1) {
			ShowWarningModal(app, table, "Inputted column exceeds int32 limit!")
			return
		}

		partRowAux, err := strconv.Atoi(unformattedPartNumber)
		if err != nil {
			ShowWarningModal(app, table, "Invalid row number!")
			return
		}
		if partRowAux >= (1<<31 - 1) {
			ShowWarningModal(app, table, "Inputted row exceeds int32 limit!")
			return
		}
		partRow := int32(partRowAux)

		globalViewport.TopRow = partRow
		globalViewport.LeftCol = int32(partColumn)
		
		RenderVisible(table, globalViewport, globalData)
		
		table.Select(1, 1)
		
		app.SetRoot(table, true).SetFocus(table)
	})
	
	form.AddButton("Cancel", func() {
		app.SetRoot(table, true).SetFocus(table)
	})

	form.SetBorder(true).SetTitle(" Go To Cell ").SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(table, true).SetFocus(table)
			return nil
		}
		return event
	})

	app.SetRoot(form, true).SetFocus(form)
}
