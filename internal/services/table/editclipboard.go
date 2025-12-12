// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// editclipboard.go contains functions for handling cell cut/copy/paste/highlighting.

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Selection margins
var selStartRow, selStartCol, selEndRow, selEndCol int32

// Clear selected rows
func clearSelectionRange() {
	selStartRow, selStartCol, selEndRow, selEndCol = 0, 0, 0, 0
}

// Returns the selected cell
func getSelectionRange(table *tview.Table) (r1, c1, r2, c2 int32) {
	activeViewport := GetActiveViewport()
	if activeViewport == nil {
		return 1, 1, 1, 1
	}

	if selStartRow != 0 || selStartCol != 0 || selEndRow != 0 || selEndCol != 0 {
		r1, r2 = utils.MinMax(selStartRow, selEndRow)
		c1, c2 = utils.MinMax(selStartCol, selEndCol)

		if r1 == 0 {
			r1 = activeViewport.TopRow
			r2 = activeViewport.TopRow + activeViewport.ViewRows - 1
		}

		if c1 == 0 {
			c1 = activeViewport.LeftCol
			c2 = activeViewport.LeftCol + activeViewport.ViewCols - 1
		}

		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	absRow, absCol := activeViewport.ToAbsolute(visualRow, visualCol)

	if absRow <= 0 || absCol <= 0 {
		return 1, 1, 1, 1
	}

	if absRow == 0 && absCol > 0 {
		return activeViewport.TopRow, absCol, activeViewport.TopRow + activeViewport.ViewRows - 1, absCol
	}

	if absCol == 0 && absRow > 0 {
		return absRow, activeViewport.LeftCol, absRow, activeViewport.LeftCol + activeViewport.ViewCols - 1
	}

	return absRow, absCol, absRow, absCol
}

// Highlights the selected cells
func highlightRange(table *tview.Table, absR1, absC1, absR2, absC2 int32) {
	activeViewport := GetActiveViewport()
	if activeViewport == nil {
		return
	}

	selStartRow, selStartCol = absR1, absC1
	selEndRow, selEndCol = absR2, absC2

	actualR1, actualC1, actualR2, actualC2 := absR1, absC1, absR2, absC2

	if absR1 == 0 {
		actualR1 = activeViewport.TopRow
		actualR2 = activeViewport.TopRow + activeViewport.ViewRows - 1
	}

	if absC1 == 0 {
		actualC1 = activeViewport.LeftCol
		actualC2 = activeViewport.LeftCol + activeViewport.ViewCols - 1
	}

	if actualR1 > actualR2 {
		actualR1, actualR2 = actualR2, actualR1
	}
	if actualC1 > actualC2 {
		actualC1, actualC2 = actualC2, actualC1
	}

	for r := 1; r < table.GetRowCount(); r++ {
		for c := 1; c < table.GetColumnCount(); c++ {
			currentCell := table.GetCell(r, c)
			if currentCell == nil {
				continue
			}
			if ref := currentCell.GetReference(); ref != nil {
				if cc, ok := ref.(*cell.Cell); ok {
					currentCell.SetBackgroundColor(cc.BgColor.ToTCellColor())
				}
			}
		}
	}

	for absR := actualR1; absR <= actualR2; absR++ {
		for absC := actualC1; absC <= actualC2; absC++ {
			if activeViewport.IsVisible(absR, absC) {
				visualR, visualC := activeViewport.ToRelative(absR, absC)
				currentCell := table.GetCell(int(visualR), int(visualC))
				if currentCell != nil {
					currentCell.SetBackgroundColor(tcell.ColorDarkGray)
				}
			}
		}
	}

	if absR1 == 0 {
		for c := absC1; c <= absC2; c++ {
			if activeViewport.IsVisible(0, c) {
				_, visualC := activeViewport.ToRelative(0, c)
				if headerCell := table.GetCell(0, int(visualC)); headerCell != nil {
					headerCell.SetBackgroundColor(tcell.ColorBlue)
				}
			}
		}
	}

	if absC1 == 0 {
		for r := absR1; r <= absR2; r++ {
			if activeViewport.IsVisible(r, 0) {
				visualR, _ := activeViewport.ToRelative(r, 0)
				if headerCell := table.GetCell(int(visualR), 0); headerCell != nil {
					headerCell.SetBackgroundColor(tcell.ColorBlue)
				}
			}
		}
	}
}

// Copies to clipboard the cell selection
func copySelection(table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	if selStartRow == 0 && selStartCol == 0 && selEndRow == 0 && selEndCol == 0 {
		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		row, col := activeViewport.ToAbsolute(visualRow, visualCol)

		if row == 0 && col > 0 {
			selStartRow, selStartCol = 0, col
			selEndRow, selEndCol = 0, col
		} else if col == 0 && row > 0 {
			selStartRow, selStartCol = row, 0
			selEndRow, selEndCol = row, 0
		} else {
			selStartRow, selStartCol = row, col
			selEndRow, selEndCol = row, col
		}
	}

	r1, c1, r2, c2 := getSelectionRange(table)

	clipboard = [][]*cell.Cell{}

	for r := r1; r <= r2; r++ {
		rowSlice := []*cell.Cell{}
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r), int(c)}
			if cellData, exists := activeData[key]; exists {
				clone := *cellData
				clone.Row = 0
				clone.Column = 0
				clone.Dependents = []*string{}
				rowSlice = append(rowSlice, &clone)
			} else {
				emptyCell := cell.NewCell(0, 0, "")
				rowSlice = append(rowSlice, emptyCell)
			}
		}
		clipboard = append(clipboard, rowSlice)
	}

	clearSelectionRange()
}

// Pastes from clipboard
func pasteSelection(app *tview.Application, table *tview.Table, visualTargetRow, visualTargetCol int32) {
	if len(clipboard) == 0 {
		return
	}

	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	targetRow, targetCol := activeViewport.ToAbsolute(visualTargetRow, visualTargetCol)

	if targetRow == 0 {
		targetRow = 1
	}
	if targetCol == 0 {
		targetCol = 1
	}

	for r, rowSlice := range clipboard {
		for c := range rowSlice {
			destRow := targetRow + int32(r)
			destCol := targetCol + int32(c)

			if destRow == 0 || destCol == 0 {
				continue
			}

			key := [2]int{int(destRow), int(destCol)}
			if cellData, exists := activeData[key]; exists {
				if !cellData.HasFlag(cell.FlagEditable) {
					modal := tview.NewModal().
						SetText(fmt.Sprintf("Target cell %s%d and potential others are uneditable.\nDo you wish to overwrite uneditable cells and continue pasting?", utils.ColumnName(destCol), destRow)).
						AddButtons([]string{"Yes", "Cancel"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "Yes" {
								performPaste(table, clipboard, targetRow, targetCol)
							}
							app.SetRoot(table, true).SetFocus(table)
						})
					modal.SetBorder(true).SetTitle("Confirm Paste").SetTitleAlign(tview.AlignCenter)
					app.SetRoot(modal, true).SetFocus(modal)
					return
				}
			}
		}
	}

	performPaste(table, clipboard, targetRow, targetCol)
	clearSelectionRange()
}

// Perform the actual pasting process
func performPaste(table *tview.Table, clipboard [][]*cell.Cell, targetRow, targetCol int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	r1, c1 := targetRow, targetCol
	r2 := targetRow + int32(len(clipboard)) - 1
	c2 := targetCol + int32(len(clipboard[0])) - 1
	
	oldCells := captureCellRange(r1, c1, r2, c2)
	
	for r, rowSlice := range clipboard {
		for c, srcCell := range rowSlice {
			destRow := targetRow + int32(r)
			destCol := targetCol + int32(c)

			if destRow == 0 || destCol == 0 {
				continue
			}

			newCell := *srcCell
			newCell.Row = destRow
			newCell.Column = destCol
			
			if newCell.IsFormula() {
				newCell.ClearFlag(cell.FlagEvaluated)
				newCell.Dependents = []*string{}
				
				newCellRef := utils.FormatCellRef(destRow, destCol)
				
				for _, depRef := range newCell.DependsOn {
					depCell, err := GetCellByRef(table, *depRef)
					if err != nil {
						continue
					}
					
					if !contains(depCell.Dependents, newCellRef) {
						depCell.Dependents = append(depCell.Dependents, &newCellRef)
					}
				}
				
				if err := EvaluateCell(table, &newCell); err != nil {
					*newCell.Display = "#ERROR!"
				}
			}
			
			key := [2]int{int(destRow), int(destCol)}
			activeData[key] = &newCell
			
			if activeViewport.IsVisible(destRow, destCol) {
				visualR, visualC := activeViewport.ToRelative(destRow, destCol)
				table.SetCell(int(visualR), int(visualC), newCell.ToTViewCell())
			}
		}
	}
	
	newCells := captureCellRange(r1, c1, r2, c2)
	RecordMultiCellAction(ActionPasteCells, r1, c1, r2, c2, oldCells, newCells)
}

// Cuts to clipboard the selected cells
func cutSelection(app *tview.Application, table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	if selStartRow == 0 && selStartCol == 0 && selEndRow == 0 && selEndCol == 0 {
		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		row, col := activeViewport.ToAbsolute(visualRow, visualCol)

		if row == 0 && col > 0 {
			selStartRow, selStartCol = 0, col
			selEndRow, selEndCol = 0, col
		} else if col == 0 && row > 0 {
			selStartRow, selStartCol = row, 0
			selEndRow, selEndCol = row, 0
		} else {
			selStartRow, selStartCol = row, col
			selEndRow, selEndCol = row, col
		}
	}

	r1, c1, r2, c2 := getSelectionRange(table)

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r), int(c)}
			if cellData, exists := activeData[key]; exists {
				if !cellData.HasFlag(cell.FlagEditable) {
					modal := tview.NewModal().
						SetText(fmt.Sprintf("Selected cell %s%d and potential others are uneditable.\nDo you wish to overwrite all other uneditable cells and continue cutting?", utils.ColumnName(int32(c)), r)).
						AddButtons([]string{"Yes", "Cancel"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "Yes" {
								copySelection(table)
								clearCutCells(table, r1, c1, r2, c2)
							}
							app.SetRoot(table, true).SetFocus(table)
						})
					modal.SetBorder(true).SetTitle("Confirm Cut").SetTitleAlign(tview.AlignCenter)
					app.SetRoot(modal, true).SetFocus(modal)
					return
				}
			}
		}
	}

	copySelection(table)
	clearCutCells(table, r1, c1, r2, c2)
	clearSelectionRange()
}

// clearCutCells removes cells
func clearCutCells(table *tview.Table, r1, c1, r2, c2 int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r), int(c)}
			
			if oldCell, exists := activeData[key]; exists {
				cellRef := utils.FormatCellRef(int32(r), int32(c))
				for _, depRef := range oldCell.DependsOn {
					depCell, err := GetCellByRef(table, *depRef)
					if err != nil {
						continue
					}
					depCell.Dependents = removeFromSlice(depCell.Dependents, cellRef)
				}
				
				for _, dependentRef := range oldCell.Dependents {
					dependentCell, err := GetCellByRef(table, *dependentRef)
					if err != nil {
						continue
					}
					if dependentCell.IsFormula() {
						RecalculateCell(table, dependentCell)
					}
				}
			}
			
			newCell := cell.NewCell(int32(r), int32(c), "")
			activeData[key] = newCell
			
			if activeViewport.IsVisible(int32(r), int32(c)) {
				visualR, visualC := activeViewport.ToRelative(int32(r), int32(c))
				table.SetCell(int(visualR), int(visualC), newCell.ToTViewCell())
			}
		}
	}
}

// Delete the selected cells
func deleteSelection(app *tview.Application, table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	if selStartRow == 0 && selStartCol == 0 && selEndRow == 0 && selEndCol == 0 {
		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
		row, col := activeViewport.ToAbsolute(visualRow, visualCol)

		if row == 0 && col > 0 {
			selStartRow, selStartCol = 0, col
			selEndRow, selEndCol = 0, col
		} else if col == 0 && row > 0 {
			selStartRow, selStartCol = row, 0
			selEndRow, selEndCol = row, 0
		} else {
			selStartRow, selStartCol = row, col
			selEndRow, selEndCol = row, col
		}
	}

	r1, c1, r2, c2 := getSelectionRange(table)	

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r), int(c)}
			if cellData, exists := activeData[key]; exists {
				if !cellData.HasFlag(cell.FlagEditable) {
					modal := tview.NewModal().
						SetText(fmt.Sprintf("Selected cell %s%d and potential others are uneditable.\nDo you wish to overwrite all other uneditable cells and continue deleting?", utils.ColumnName(int32(c)), r)).
						AddButtons([]string{"Yes", "Cancel"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "Yes" {
    							performDelete(table, r1, c1, r2, c2)
                        	}	
							app.SetRoot(table, true).SetFocus(table)
						})
					modal.SetBorder(true).SetTitle("Confirm Delete").SetTitleAlign(tview.AlignCenter)
                	app.SetRoot(modal, true).SetFocus(modal)	
					return
				}
			}
		}
	}


	performDelete(table, r1, c1, r2, c2)
	clearSelectionRange()
}

// performDelete deletes cells
func performDelete(table *tview.Table, r1, c1, r2, c2 int32) {
	oldCells := captureCellRange(r1, c1, r2, c2)
	
	clearCutCells(table, r1, c1, r2, c2)
	
	newCells := captureCellRange(r1, c1, r2, c2)
	
	RecordMultiCellAction(ActionDeleteCells, r1, c1, r2, c2, oldCells, newCells)
}

// captureCellRange captures a range of cells
func captureCellRange(r1, c1, r2, c2 int32) [][]*cell.Cell {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return nil
	}

	var cells [][]*cell.Cell
	
	for r := r1; r <= r2; r++ {
		var row []*cell.Cell
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r), int(c)}
			if cellData, exists := activeData[key]; exists {
				row = append(row, cellData.Clone())
			} else {
				row = append(row, cell.NewCell(r, c, ""))
			}
		}
		cells = append(cells, row)
	}
	
	return cells
}
