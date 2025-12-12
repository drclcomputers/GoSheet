// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// selection.go contains functions for handling cell selection.

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/ui/cellui"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Variables for Shift+Arrow selection
var anchorRow, anchorCol int32
var isSelecting bool

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
		
		rows := table.GetRowCount()
		cols := table.GetColumnCount()

		for r := range rows {
			for c := range cols {
				tvCell := table.GetCell(r, c)
				if tvCell == nil {
					continue
				}
				if ref := tvCell.GetReference(); ref != nil {
					if cellData, ok := ref.(*cell.Cell); ok {
						tvCell.SetBackgroundColor(cellData.BgColor.ToTCellColor())
					}
				} else {
					tvCell.SetBackgroundColor(utils.ColorOptions["Black"].ToTCellColor())
				}
			}
		}

		// Highlight selected column
		if row == 0 && column > 0 {
			for r := 1; r < rows; r++ {
				cell := table.GetCell(r, column)
				if cell != nil {
					cell.SetBackgroundColor(tcell.ColorDarkGray)
				}
			}
		}

		// Highlight selected row
		if column == 0 && row > 0 {
			for c := 1; c < cols; c++ {
				cell := table.GetCell(row, c)
				if cell != nil {
					cell.SetBackgroundColor(tcell.ColorDarkGray)
				}
			}
		}

		// Highlight column and row headers for selected cell
		if row > 0 && column > 0 {
			headerCol := table.GetCell(0, column)
			headerRow := table.GetCell(row, 0)

			if headerCol != nil {
				headerCol.SetBackgroundColor(tcell.ColorBlue)
			}
			if headerRow != nil {
				headerRow.SetBackgroundColor(tcell.ColorBlue)
			}
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
