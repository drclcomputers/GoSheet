// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// history.go implements undo/redo functionality with formula dependency management

package table

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"github.com/rivo/tview"
)

type ActionType int32

const (
	ActionEditCell ActionType = iota
	ActionDeleteCells
	ActionPasteCells
	ActionInsertRow
	ActionDeleteRow
	ActionInsertColumn
	ActionDeleteColumn
	ActionFormatCells
)

type Action struct {
	Type     ActionType
	Row      int32
	Col      int32
	OldCell  *cell.Cell
	NewCell  *cell.Cell
	OldCells [][]*cell.Cell
	NewCells [][]*cell.Cell
	Data     any
}

type History struct {
	undoStack []*Action
	redoStack []*Action
	maxSize   int
}

// InitHistory initializes history for a sheet (called in NewSheet)
func InitHistory(maxSize int) *History {
	return &History{
		undoStack: make([]*Action, 0, maxSize),
		redoStack: make([]*Action, 0, maxSize),
		maxSize:   maxSize,
	}
}

// RecordAction records an action to the active sheet's history
func RecordAction(action *Action) {
	history := GetActiveHistory()
	if history == nil {
		return
	}

	history.undoStack = append(history.undoStack, action)

	if len(history.undoStack) > history.maxSize {
		newStack := make([]*Action, history.maxSize)
	    copy(newStack, history.undoStack[1:])
	    history.undoStack = newStack
	}

	history.redoStack = nil
}

// RecordCellEdit records an edited cell
func RecordCellEdit(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell) {
	action := &Action{
		Type:    ActionEditCell,
		Row:     row,
		Col:     col,
		OldCell: oldCell.Clone(),
		NewCell: newCell.Clone(),
	}

	MarkAsModified(table)

	RecordAction(action)
}

// RecordMultiCellAction records multiple edited cells
func RecordMultiCellAction(actionType ActionType, r1, c1, r2, c2 int32, oldCells, newCells [][]*cell.Cell) {
	action := &Action{
		Type:     actionType,
		Row:      r1,
		Col:      c1,
		OldCells: cloneCellGrid(oldCells),
		NewCells: cloneCellGrid(newCells),
		Data:     map[string]int32{"r2": r2, "c2": c2},
	}
	RecordAction(action)
}

// restoreCell restores a cell to data map
func restoreCell(table *tview.Table, row, col int32, cellData *cell.Cell) {
	if cellData == nil {
		return
	}

	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	cellData.Row = row
	cellData.Column = col

	if cellData.IsFormula() {
		newCellRef := utils.FormatCellRef(row, col)

		clearOldDependencies(table, cellData)

		for _, depRef := range cellData.DependsOn {
			depCell, err := GetCellByRef(table, *depRef)
			if err != nil {
				continue
			}

			if !contains(depCell.Dependents, newCellRef) {
				depCell.Dependents = append(depCell.Dependents, &newCellRef)
			}
		}

		cellData.ClearFlag(cell.FlagEvaluated)

		if err := EvaluateCell(table, cellData); err != nil {
			*cellData.Display = "#ERROR!"
		}
	}

	key := [2]int{int(row), int(col)}
	activeData[key] = cellData

	if activeViewport.IsVisible(row, col) {
		visualR, visualC := activeViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}

	RecalculateCell(table, cellData)
}

// clearCell removes a cell from data map
func clearCell(table *tview.Table, row, col int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	key := [2]int{int(row), int(col)}

	if oldCell, exists := activeData[key]; exists {
		cellRef := utils.FormatCellRef(row, col)
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

	newCell := cell.NewCell(row, col, "")
	activeData[key] = newCell

	if activeViewport.IsVisible(row, col) {
		visualR, visualC := activeViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), newCell.ToTViewCell())
	}
}

// Undo action
func Undo(table *tview.Table) bool {
	history := GetActiveHistory()
	if history == nil || len(history.undoStack) == 0 {
		return false
	}

	action := history.undoStack[len(history.undoStack)-1]
	history.undoStack = history.undoStack[:len(history.undoStack)-1]

	switch action.Type {
	case ActionEditCell:
		if action.OldCell != nil {
			clearCell(table, action.Row, action.Col)
			restoreCell(table, action.Row, action.Col, action.OldCell)
		}

	case ActionDeleteCells:
		if action.OldCells != nil {
			data := action.Data.(map[string]int32)
			r2, c2 := data["r2"], data["c2"]
			for r := action.Row; r <= int32(r2); r++ {
				for c := action.Col; c <= int32(c2); c++ {
					clearCell(table, r, c)
				}
			}
			restoreCellGrid(table, action.Row, action.Col, action.OldCells)
		}

	case ActionPasteCells, ActionFormatCells:
		if action.OldCells != nil {
			restoreCellGrid(table, action.Row, action.Col, action.OldCells)
		}

	case ActionInsertRow:
		deleteRow(nil, table, action.Row)
		RecalculateAllFormulas(table)

	case ActionDeleteRow:
		insertRow(nil, table, action.Row)
		if action.OldCells != nil {
			restoreCellGrid(table, action.Row, 1, action.OldCells)
		}

	case ActionInsertColumn:
		deleteCol(nil, table, action.Col)
		RecalculateAllFormulas(table)

	case ActionDeleteColumn:
		insertCol(nil, table, action.Col)
		if action.OldCells != nil {
			restoreCellGrid(table, 1, action.Col, action.OldCells)
		}
	}

	history.redoStack = append(history.redoStack, action)

	return true
}

// Redo action
func Redo(table *tview.Table) bool {
	history := GetActiveHistory()
	if history == nil || len(history.redoStack) == 0 {
		return false
	}

	action := history.redoStack[len(history.redoStack)-1]
	history.redoStack = history.redoStack[:len(history.redoStack)-1]

	switch action.Type {
	case ActionEditCell:
		if action.NewCell != nil {
			clearCell(table, action.Row, action.Col)
			restoreCell(table, action.Row, action.Col, action.NewCell)
		}

	case ActionDeleteCells:
		if action.NewCells != nil {
			data := action.Data.(map[string]int32)
			r2, c2 := data["r2"], data["c2"]
			for r := action.Row; r <= int32(r2); r++ {
				for c := action.Col; c <= int32(c2); c++ {
					clearCell(table, r, c)
				}
			}
		}

	case ActionPasteCells, ActionFormatCells:
		if action.NewCells != nil {
			restoreCellGrid(table, action.Row, action.Col, action.NewCells)
		}

	case ActionInsertRow:
		insertRow(nil, table, action.Row)
		RecalculateAllFormulas(table)

	case ActionDeleteRow:
		deleteRow(nil, table, action.Row)
		RecalculateAllFormulas(table)

	case ActionInsertColumn:
		insertCol(nil, table, action.Col)
		RecalculateAllFormulas(table)

	case ActionDeleteColumn:
		deleteCol(nil, table, action.Col)
		RecalculateAllFormulas(table)
	}

	history.undoStack = append(history.undoStack, action)

	return true
}

// Clones the grid
func cloneCellGrid(grid [][]*cell.Cell) [][]*cell.Cell {
	if grid == nil {
		return nil
	}
	clone := make([][]*cell.Cell, len(grid))
	for i, row := range grid {
		clone[i] = make([]*cell.Cell, len(row))
		for j, c := range row {
			if c != nil {
				clone[i][j] = c.Clone()
			}
		}
	}
	return clone
}

// Restores the grid
func restoreCellGrid(table *tview.Table, startRow, startCol int32, grid [][]*cell.Cell) {
	for r, row := range grid {
		for c, cellData := range row {
			if cellData != nil {
				actualRow := startRow + int32(r)
				actualCol := startCol + int32(c)
				restoreCell(table, actualRow, actualCol, cellData)
			}
		}
	}

	for r, row := range grid {
		for c, cellData := range row {
			if cellData != nil && cellData.IsFormula() {
				actualRow := startRow + int32(r)
				actualCol := startCol + int32(c)
				if actualRow < int32(table.GetRowCount()) && actualCol < int32(table.GetColumnCount()) {
					activeData := GetActiveSheetData()
					if activeData == nil {
						continue
					}
					key := [2]int{int(actualRow), int(actualCol)}
					if restoredCell, exists := activeData[key]; exists {
						if restoredCell.IsFormula() {
							RecalculateCell(table, restoredCell)
						}
					}
				}
			}
		}
	}
}
