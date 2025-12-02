// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// navigationUI.go provides navigation, find, and replace functionality

package ui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func SplitAlphaNumeric(s string) (letters, numbers string) {
	for i, ch := range s {
		if ch >= '0' && ch <= '9' {
			return s[:i], s[i:]
		}
	}
	return s, ""
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

// Find
var lastSearchTerm string
var lastSearchRow, lastSearchCol int32  // Changed to int32 for absolute coordinates
var lastCaseSensitive bool
var lastMatchWholeWord bool

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
			app.SetRoot(table, true).SetFocus(table)
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
			app.SetRoot(table, true).SetFocus(table)
		}
	})

	form.AddButton("Clear", func(){
		lastSearchTerm = ""
		lastCaseSensitive = false
		lastMatchWholeWord = false
		lastSearchRow, lastSearchCol = 0, 0

		app.SetRoot(table, true).SetFocus(table)
	})
	
	form.AddButton("Close", func() {
		lastSearchRow, lastSearchCol = 0, 0
		app.SetRoot(table, true).SetFocus(table)
	})

	form.SetBorder(true).SetTitle(" Find ").SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			lastSearchRow, lastSearchCol = 0, 0
			app.SetRoot(table, true).SetFocus(table)
			return nil
		}
		return event
	})

	app.SetRoot(form, true).SetFocus(form)
}

func findNext(table *tview.Table, searchTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
	// Start from last position or current selection
	var startRow, startCol int32
	
	if lastSearchRow == 0 && lastSearchCol == 0 {
		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		startRow, startCol = globalViewport.ToAbsolute(visualRow, visualCol)
	} else {
		startRow, startCol = lastSearchRow, lastSearchCol
		startCol++
	}
	
	maxRow, maxCol := int32(utils.MAX_ROWS), int32(utils.MAX_COLS)
	
	for r := startRow; r <= maxRow; r++ {
		colStart := int32(1)
		if r == startRow {
			colStart = startCol
		}
		
		for c := colStart; c <= maxCol; c++ {
			if matchCellInData(globalData, r, c, searchTerm, caseSensitive, matchWholeWord) {
				navigateToCell(table, r, c, globalViewport, globalData, RenderVisible)
				lastSearchRow, lastSearchCol = r, c
				return true
			}
		}
	}
	
	for r := int32(1); r < startRow; r++ {
		for c := int32(1); c <= maxCol; c++ {
			if matchCellInData(globalData, r, c, searchTerm, caseSensitive, matchWholeWord) {
				navigateToCell(table, r, c, globalViewport, globalData, RenderVisible)
				lastSearchRow, lastSearchCol = r, c
				return true
			}
		}
	}
	
	for c := int32(1); c < startCol; c++ {
		if matchCellInData(globalData, startRow, c, searchTerm, caseSensitive, matchWholeWord) {
			navigateToCell(table, startRow, c, globalViewport, globalData, RenderVisible)
			lastSearchRow, lastSearchCol = startRow, c
			return true
		}
	}
	
	return false
}

func findPrevious(table *tview.Table, searchTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
	var startRow, startCol int32
	
	if lastSearchRow == 0 && lastSearchCol == 0 {
		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		startRow, startCol = globalViewport.ToAbsolute(visualRow, visualCol)
	} else {
		startRow, startCol = lastSearchRow, lastSearchCol
		startCol--
		if startCol < 1 {
			startCol = utils.MAX_COLS
			startRow--
		}
	}
	
	maxRow, maxCol := int32(utils.MAX_ROWS), int32(utils.MAX_COLS)
	
	for r := startRow; r >= 1; r-- {
		colEnd := maxCol
		if r == startRow {
			colEnd = startCol
		}
		
		for c := colEnd; c >= 1; c-- {
			if matchCellInData(globalData, r, c, searchTerm, caseSensitive, matchWholeWord) {
				navigateToCell(table, r, c, globalViewport, globalData, RenderVisible)
				lastSearchRow, lastSearchCol = r, c
				return true
			}
		}
	}
	
	for r := maxRow; r > startRow; r-- {
		for c := maxCol; c >= 1; c-- {
			if matchCellInData(globalData, r, c, searchTerm, caseSensitive, matchWholeWord) {
				navigateToCell(table, r, c, globalViewport, globalData, RenderVisible)
				lastSearchRow, lastSearchCol = r, c
				return true
			}
		}
	}
	
	return false
}

func navigateToCell(table *tview.Table, absRow, absCol int32, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	globalViewport.TopRow = absRow
	globalViewport.LeftCol = absCol
	
	RenderVisible(table, globalViewport, globalData)
	
	table.Select(1, 1)
}

func matchCellInData(globalData map[[2]int]*cell.Cell, row, col int32, searchTerm string, caseSensitive, matchWholeWord bool) bool {
	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		return false
	}

	cellValue := cellData.Display
	if cellValue == nil {
		cellValue = cellData.RawValue
	}
	if cellValue == nil {
		return false
	}

	if matchWholeWord {
		for word := range strings.FieldsSeq(*cellValue) {
			if caseSensitive {
				if word == searchTerm {
					return true
				}
			} else {
				if strings.EqualFold(word, searchTerm) {
					return true
				}
			}
		}
		return false
	}

	searchValue := *cellValue
	if !caseSensitive {
		searchValue = strings.ToLower(searchValue)
		searchTerm = strings.ToLower(searchTerm)
	}

	return strings.Contains(searchValue, searchTerm)
}

func FindNextQuick(table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
	if lastSearchTerm == "" {
		return false
	}
	return findNext(table, lastSearchTerm, lastCaseSensitive, lastMatchWholeWord, globalData, globalViewport, RenderVisible)
}

func FindPreviousQuick(table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
	if lastSearchTerm == "" {
		return false
	}
	return findPrevious(table, lastSearchTerm, lastCaseSensitive, lastMatchWholeWord, globalData, globalViewport, RenderVisible)
}

// Replace functions
func ReplaceDialog(app *tview.Application, table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) {
	form := tview.NewForm()
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
					app.SetRoot(form, true).SetFocus(form)
				}
			})
		modal.SetBorder(true).SetTitle(" Confirm Replace All ").SetTitleAlign(tview.AlignCenter)
		app.SetRoot(modal, true).SetFocus(modal)
	})

	form.AddButton("Close", func() {
		app.SetRoot(table, true).SetFocus(table)
	})

	form.SetBorder(true).SetTitle(" Replace ").SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(table, true).SetFocus(table)
			return nil
		}
		return event
	})

	app.SetRoot(form, true).SetFocus(form)
}

func replaceInCell(app *tview.Application, table *tview.Table, row, col int32, findTerm, replaceTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) bool {
	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		return false
	}

	if !cellData.HasFlag(cell.FlagEditable) {
		ShowWarningModal(app, table, "Cell is not editable")
		return false
	}

	if !matchCellInData(globalData, row, col, findTerm, caseSensitive, matchWholeWord) {
		return false
	}

	cellValue := cellData.RawValue
	if cellValue == nil {
		return false
	}
	
	if matchWholeWord {
		aux := replaceWholeWord(*cellValue, findTerm, replaceTerm, caseSensitive)
		cellData.RawValue = &aux
	} else {
		if caseSensitive {
			aux := strings.ReplaceAll(*cellValue, findTerm, replaceTerm)
			cellData.RawValue = &aux
		} else {
			aux := replaceInsensitive(*cellValue, findTerm, replaceTerm)
			cellData.RawValue = &aux
		}
	}
	
	cellData.Display = cellData.RawValue
	globalData[key] = cellData
	
	if globalViewport.IsVisible(row, col) {
		visualR, visualC := globalViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}
	
	return true
}

func countMatches(globalData map[[2]int]*cell.Cell, searchTerm string, caseSensitive, matchWholeWord bool) int32 {
	count := int32(0)
	
	for _, cellData := range globalData {
		if matchCellInData(globalData, cellData.Row, cellData.Column, searchTerm, caseSensitive, matchWholeWord) {
			count++
		}
	}
	
	return count
}

func replaceAll(table *tview.Table, findTerm, replaceTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) int {
	count := 0

	for key, cellData := range globalData {
		if !cellData.HasFlag(cell.FlagEditable) {
			continue
		}

		if matchCellInData(globalData, cellData.Row, cellData.Column, findTerm, caseSensitive, matchWholeWord) {
			if cellData.RawValue != nil {
				if matchWholeWord {
					*cellData.RawValue = replaceWholeWord(*cellData.RawValue, findTerm, replaceTerm, caseSensitive)
				} else {
					if caseSensitive {
						*cellData.RawValue = strings.ReplaceAll(*cellData.RawValue, findTerm, replaceTerm)
					} else {
						*cellData.RawValue = replaceInsensitive(*cellData.RawValue, findTerm, replaceTerm)
					}
				}	
				cellData.Display = cellData.RawValue
				globalData[key] = cellData
				count++
			}
		}
	}
	
	RenderVisible(table, globalViewport, globalData)
	
	return count
}

func replaceInsensitive(s, old, new string) string {
	lowerOld := strings.ToLower(old)
	
	result := ""
	i := 0
	for i < len(s) {
		if i+len(old) <= len(s) && strings.ToLower(s[i:i+len(old)]) == lowerOld {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

func replaceWholeWord(s, old, new string, caseSensitive bool) string {
	words := strings.Fields(s)
	result := make([]string, len(words))
	
	for i, word := range words {
		if caseSensitive {
			if word == old {
				result[i] = new
			} else {
				result[i] = word
			}
		} else {
			if strings.EqualFold(word, old) {
				result[i] = new
			} else {
				result[i] = word
			}
		}
	}
	
	return strings.Join(result, " ")
}
