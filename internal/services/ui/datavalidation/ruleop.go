// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// ruleop.go provides some functions for managing data validation rules

package datavalidation

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strings"

	"github.com/rivo/tview"
)

func deleteRule(cellData *cell.Cell, table *tview.Table, row, col int32, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	emptyStr := ""
	cellData.Valrule = &emptyStr
	cellData.Valrulemsg = &emptyStr

	key := [2]int{int(row), int(col)}
	globalData[key] = cellData

	if globalViewport.IsVisible(row, col) {
		visualR, visualC := globalViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}
}

func saveRule(app *tview.Application, table *tview.Table, cellData *cell.Cell, ruleText string, row, col int32, form tview.Primitive, returnTo tview.Primitive, focus tview.Primitive, container *tview.Flex, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	ruleText = strings.TrimSpace(ruleText)

	if ruleText == "" || ValidateValidationRule(ruleText, cellData) {
		cellData.Valrule = &ruleText

		key := [2]int{int(row), int(col)}
		globalData[key] = cellData

		if globalViewport.IsVisible(row, col) {
			visualR, visualC := globalViewport.ToRelative(row, col)
			table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
		}

		app.SetRoot(returnTo, true).SetFocus(focus)
	} else {
		showValidationErrorModal(app, container, form, "Invalid validation rule!\n\nMake sure:\n- You use 'THIS' instead of cell references (e.g., A1)\n- The syntax is correct\n- The rule returns true/false")
	}
}
