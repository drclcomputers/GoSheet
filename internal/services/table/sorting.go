// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// sorting.go provides sorting functionality for spreadsheet columns

package table

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// Sorts the column according to ascending
func SortColumn(app *tview.Application, table *tview.Table, ascending bool) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(visualRow, visualCol)

	var sortCol, startRow, endRow int32
	
	if selStartRow == 0 && selStartCol == 0 && selEndRow == 0 && selEndCol == 0 {
		if row == 0 && col > 0 {
			selStartRow, selStartCol = 0, col
			selEndRow, selEndCol = 0, col
		} else if col == 0 && row > 0 {
			return
		} else if row > 0 && col > 0 {
			selStartRow, selStartCol = 0, col
			selEndRow, selEndCol = 0, col
		}
	}
	
	r1, c1, r2, _ := getSelectionRange(table)
	
	sortCol = c1
	startRow = r1
	endRow = r2
	
	type sortableCell struct {
		cell     *cell.Cell
		sortKey  any
		isEmpty  bool
	}
	
	var cells []sortableCell
	var emptyCells []sortableCell
	
	// Get cells from data map using absolute coordinates
	for r := startRow; r <= endRow; r++ {
		key := [2]int{int(r), int(sortCol)}
		
		if cellData, exists := activeData[key]; exists {
			value := strings.TrimSpace(*cellData.RawValue)
			if value == "" {
				emptyCells = append(emptyCells, sortableCell{cell: cellData, isEmpty: true})
				continue
			}
			cells = append(cells, sortableCell{
				cell:    cellData,
				sortKey: getSortKey(cellData),
			})
		} else {
			emptyCells = append(emptyCells, sortableCell{cell: nil, isEmpty: true})
		}
	}
	
	sort.Slice(cells, func(i, j int) bool {
		less := compareSortKeys(cells[i].sortKey, cells[j].sortKey)
		if ascending {
			return less
		}
		return !less
	})
	
	cells = append(cells, emptyCells...)
	
	for i, cellData := range cells {
		targetRow := startRow + int32(i)
		key := [2]int{int(targetRow), int(sortCol)}
		
		if cellData.cell != nil {
			newCell := *cellData.cell
			newCell.Row = targetRow
			newCell.Column = sortCol
			activeData[key] = &newCell
			
			if activeViewport.IsVisible(targetRow, sortCol) {
				visualR, visualC := activeViewport.ToRelative(targetRow, sortCol)
				table.SetCell(int(visualR), int(visualC), newCell.ToTViewCell())
			}
		} else {
			emptyCell := cell.NewCell(targetRow, sortCol, "")
			activeData[key] = emptyCell
			
			if activeViewport.IsVisible(targetRow, sortCol) {
				visualR, visualC := activeViewport.ToRelative(targetRow, sortCol)
				table.SetCell(int(visualR), int(visualC), emptyCell.ToTViewCell())
			}
		}
	}
	
	clearSelectionRange()
}

func getSortKey(c *cell.Cell) any {
	value := strings.TrimSpace(*c.RawValue)
	
	if value == "" {
		return "" 
	}
	
	switch strings.ToLower(*c.Type) {
	case "number", "financial":
		normalized := strings.ReplaceAll(value, string(c.ThousandsSeparator), "")
		normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
		if num, err := strconv.ParseFloat(normalized, 64); err == nil {
			return num
		}
		return value
		
	case "date":
		formats := []string{
			"2006-01-02",
			"01/02/2006",
			"02/01/2006",
			"2006/01/02",
			"Jan 2, 2006",
			"2 Jan 2006",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Unix() 
			}
		}
		return value
		
	case "time":
		formats := []string{
			"15:04:05",
			"15:04",
			"3:04 PM",
			"3:04:05 PM",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Unix()
			}
		}
		return value 
		
	default:
		return strings.ToLower(value)
	}
}

func compareSortKeys(a, b any) bool {
	if a == nil || a == "" {
		return false
	}
	if b == nil || b == "" {
		return true
	}
	
	if aNum, aOk := a.(float64); aOk {
		if bNum, bOk := b.(float64); bOk {
			return aNum < bNum
		}
		return true
	}
	
	if aTime, aOk := a.(int64); aOk {
		if bTime, bOk := b.(int64); bOk {
			return aTime < bTime
		}
		return true
	}
	
	aStr, aOk := a.(string)
	bStr, bOk := b.(string)
	if aOk && bOk {
		return aStr < bStr
	}
	
	return false
}

func ShowSortDialog(app *tview.Application, table *tview.Table) {
	modal := tview.NewModal().
		SetText("Sort selected column/range").
		AddButtons([]string{"Ascending", "Descending", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Ascending":
				SortColumn(app, table, true)
			case "Descending":
				SortColumn(app, table, false)
			}
			app.SetRoot(table, true).SetFocus(table)
		})
	
	modal.SetBorder(true).SetTitle(" Sort ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}
