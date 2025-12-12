// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// findnext.go provides functons for finding the searched text

package navigation

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/rivo/tview"
)

func findNext(table *tview.Table, searchTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
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

func FindNextQuick(table *tview.Table, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) bool {
	if lastSearchTerm == "" {
		return false
	}
	return findNext(table, lastSearchTerm, lastCaseSensitive, lastMatchWholeWord, globalData, globalViewport, RenderVisible)
}
