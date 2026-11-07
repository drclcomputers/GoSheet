// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// selection.go contains functions for handling cell selection.

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/ui/cellui"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Variables for Shift+Arrow selection
var anchorRow, anchorCol int32
var isSelecting bool
var lastHeadRow, lastHeadCol int

var (
	headerBG = tcell.NewRGBColor(32, 40, 58)
	headerFG = tcell.NewRGBColor(168, 188, 214)
	selectBG = tcell.NewRGBColor(28, 110, 128)
	sheetBG  = tcell.NewRGBColor(12, 14, 22)
)

func resetHeaderTrack() {
	lastHeadRow, lastHeadCol = 0, 0
}

func paintHeader(table *tview.Table, row, col int, color tcell.Color) {
	if row > 0 {
		if h := table.GetCell(row, 0); h != nil {
			h.SetBackgroundColor(color)
		}
	}
	if col > 0 {
		if h := table.GetCell(0, col); h != nil {
			h.SetBackgroundColor(color)
		}
	}
}

// SelectInTable configures the given tview.Table with selection and input handling
func SelectInTable(app *tview.Application, table *tview.Table, vp *utils.Viewport, data map[[2]int]*cell.Cell) *tview.Table {
	table.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 && column == 0 {
			table.Select(1, 1)
			return
		}

		activeViewport := GetActiveViewport()
		if activeViewport == nil {
			return
		}

		absRow, absCol := activeViewport.ToAbsolute(int32(row), int32(column))

		if lastHeadRow != 0 || lastHeadCol != 0 {
			paintHeader(table, lastHeadRow, lastHeadCol, headerBG)
		}
		if row > 0 && column > 0 {
			paintHeader(table, row, column, selectBG)
			lastHeadRow, lastHeadCol = row, column
			activeData := GetActiveSheetData()
			addr, raw := currentCellRaw(activeData, absRow, absCol)
			sheetview.SetBar(addr, raw)
		} else {
			lastHeadRow, lastHeadCol = 0, 0
		}

		posCell := table.GetCell(0, 0)
		if posCell != nil {
			if !isSelecting {
				if row > 0 && column > 0 {
					posCell.SetText(fmt.Sprintf("%s%d", utils.ColumnName(absCol), absRow))
					//posCell.SetText(fmt.Sprintf("%s%d [%d cells]", utils.ColumnName(absCol), absRow, len(globalData)))
				} else if row == 0 && column > 0 {
					posCell.SetText(fmt.Sprintf("Col: %s", utils.ColumnName(absCol)))
				} else if column == 0 && row > 0 {
					posCell.SetText(fmt.Sprintf("Row: %d", absRow))
				} else {
					posCell.SetText("")
				}
			} else {
				if anchorRow == 0 && absRow == 0 {
					c1, c2 := utils.MinMax(anchorCol, absCol)
					if c1 == c2 {
						posCell.SetText(fmt.Sprintf("Col: %s", utils.ColumnName(c1)))
					} else {
						posCell.SetText(fmt.Sprintf("%s:%s", utils.ColumnName(c1), utils.ColumnName(c2)))
					}
				} else if anchorCol == 0 && absCol == 0 {
					r1, r2 := utils.MinMax(anchorRow, absRow)
					if r1 == r2 {
						posCell.SetText(fmt.Sprintf("Row: %d", r1))
					} else {
						posCell.SetText(fmt.Sprintf("%d:%d", r1, r2))
					}
				} else {
					posCell.SetText(fmt.Sprintf("%s%d:%s%d",
						utils.ColumnName(anchorCol), anchorRow,
						utils.ColumnName(absCol), absRow))
				}
			}
		}
	})

	table.SetSelectedFunc(func(row, col int) {
		if row == 0 && col > 0 {
			return
		}

		if col == 0 && row > 0 {
			return
		}

		if row == 0 || col == 0 {
			return
		}

		activeData := GetActiveSheetData()
		activeViewport := GetActiveViewport()

		if activeData == nil || activeViewport == nil {
			return
		}

		absRow, absCol := activeViewport.ToAbsolute(int32(row), int32(col))

		key := [2]int{int(absRow), int(absCol)}
		c, exists := activeData[key]

		if !exists {
			c = cell.NewCell(absRow, absCol, "")
			activeData[key] = c
		}

		if !c.HasFlag(cell.FlagEditable) {
			cellui.ShowUneditableModal(app, table, absRow, absCol, RecordCellEdit, EvaluateCell, RecalculateCell, activeData, activeViewport)
			return
		}

		cellui.EditCellDialog(app, table, absRow, absCol, RecordCellEdit, EvaluateCell, RecalculateCell, activeData, activeViewport)
	})

	table = InputCaptureService(app, table, vp, data)

	return table
}
