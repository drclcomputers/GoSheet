// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// commentsUI.go contains functions for managing cell comments/notes

package ui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Shows the dialog for modifying cell comments
func ShowCommentDialog(app *tview.Application, table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := globalViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 || col == 0 {
		return
	}

	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		cellData = cell.NewCell(row, col, "")
		globalData[key] = cellData
	}

	if cellData.Notes == nil {
		emptyStr := ""
		cellData.Notes = &emptyStr
	}

	form := tview.NewForm()
	
	form.AddTextArea("Comment:", *cellData.Notes, 0, 0, 500, func(text string) {
	})

	textAreaItem := form.GetFormItem(0).(*tview.TextArea)
	textAreaItem.SetOffset(0, 0)

	form.AddButton("Save", func() { 
		saveComment(app, table, form, cellData, row, col, globalData, globalViewport)
	}).SetButtonsAlign(tview.AlignCenter)

	form.AddButton("Delete", func() { 
		deleteComment(app, table, cellData, row, col, globalData, globalViewport) 
	}).SetButtonsAlign(tview.AlignCenter)
	
	form.AddButton("Cancel", func() {
		app.SetRoot(table, true).SetFocus(table)
	}).SetButtonsAlign(tview.AlignCenter)

	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Cell Comment - %s%d ", utils.ColumnName(col), row)).
		SetTitleAlign(tview.AlignCenter)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case (event.Rune() == 's' || event.Rune() == 'S') && event.Modifiers()&tcell.ModAlt != 0:
			saveComment(app, table, form, cellData, row, col, globalData, globalViewport)
		case (event.Rune() == 'd' || event.Rune() == 'D') && event.Modifiers()&tcell.ModAlt != 0:
			deleteComment(app, table, cellData, row, col, globalData, globalViewport)
		case event.Key() == tcell.KeyEscape:
			app.SetRoot(table, true).SetFocus(table)
			return nil 
		}
		return event
	})

	app.SetRoot(form, true).SetFocus(form)
}

func saveComment(app *tview.Application, table *tview.Table, form *tview.Form, cellData *cell.Cell, row, col int32, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	textAreaItem := form.GetFormItem(0).(*tview.TextArea)
	commentText := textAreaItem.GetText()
	
	if cellData.Notes == nil {
		cellData.Notes = new(string)
	}
	*cellData.Notes = commentText

	key := [2]int{int(row), int(col)}
	globalData[key] = cellData
	
	if globalViewport.IsVisible(row, col) {
		visualR, visualC := globalViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}
	
	app.SetRoot(table, true).SetFocus(table)
}

func deleteComment(app *tview.Application, table *tview.Table, cellData *cell.Cell, row, col int32, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	if cellData.Notes == nil {
		app.SetRoot(table, true).SetFocus(table)
		return
	}

	confirmModal := tview.NewModal().
		SetText("Delete this comment?").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				emptyStr := ""
				cellData.Notes = &emptyStr
				
				key := [2]int{int(row), int(col)}
				globalData[key] = cellData
				
				if globalViewport.IsVisible(row, col) {
					visualR, visualC := globalViewport.ToRelative(row, col)
					table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
				}
			}
			app.SetRoot(table, true).SetFocus(table)
		})
	confirmModal.SetBorder(true).
		SetTitle(" Confirm Delete ").
		SetTitleAlign(tview.AlignCenter)
	app.SetRoot(confirmModal, true).SetFocus(confirmModal)
}
