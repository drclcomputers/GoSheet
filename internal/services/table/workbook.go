// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// workbook.go implements the type definition and base functions for creating and managing workbooks

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

type Sheet struct {
	Name     string
	Data     map[[2]int]*cell.Cell
	Viewport *utils.Viewport
	History  *History
}

type Workbook struct {
	Sheets      []*Sheet
	ActiveSheet int
	CurrentFile string
	HasChanges  bool
}

func NewWorkbook() *Workbook {
	return &Workbook{
		Sheets: []*Sheet{
			NewSheet("Sheet1"),
		},
		ActiveSheet: 0,
		HasChanges:  false,
	}
}

// NewSheet creates a new sheet
func NewSheet(name string) *Sheet {
	return &Sheet{
		Name: name,
		Data: make(map[[2]int]*cell.Cell),
		Viewport: &utils.Viewport{
			TopRow:   1,
			LeftCol:  1,
			ViewRows: utils.DEFAULT_VIEWPORT_ROWS,
			ViewCols: utils.DEFAULT_VIEWPORT_COLS,
		},
		History: &History{
			undoStack: make([]*Action, 0, 100),
			redoStack: make([]*Action, 0, 100),
			maxSize:   100,
		},
	}
}

// GetActiveSheet returns the currently active sheet
func (wb *Workbook) GetActiveSheet() *Sheet {
	if wb.ActiveSheet >= 0 && wb.ActiveSheet < len(wb.Sheets) {
		return wb.Sheets[wb.ActiveSheet]
	}
	return nil
}

// GetActiveSheetData returns the data map for the active sheet
func GetActiveSheetData() map[[2]int]*cell.Cell {
	if globalWorkbook == nil {
		return nil
	}
	sheet := globalWorkbook.GetActiveSheet()
	if sheet == nil {
		return nil
	}
	return sheet.Data
}

// GetActiveViewport returns the viewport for the active sheet
func GetActiveViewport() *utils.Viewport {
	if globalWorkbook == nil {
		return nil
	}
	sheet := globalWorkbook.GetActiveSheet()
	if sheet == nil {
		return nil
	}
	return sheet.Viewport
}

// GetActiveHistory returns the history for the active sheet
func GetActiveHistory() *History {
	if globalWorkbook == nil {
		return nil
	}
	sheet := globalWorkbook.GetActiveSheet()
	if sheet == nil {
		return nil
	}
	return sheet.History
}

// AddSheet adds a new sheet to the workbook
func (wb *Workbook) AddSheet(name string) {
	wb.Sheets = append(wb.Sheets, NewSheet(name))
	wb.HasChanges = true
}

// DeleteSheet removes a sheet from the workbook
func (wb *Workbook) DeleteSheet(index int) error {
	if len(wb.Sheets) <= 1 {
		return fmt.Errorf("cannot delete the last sheet")
	}
	if index < 0 || index >= len(wb.Sheets) {
		return fmt.Errorf("invalid sheet index")
	}

	wb.Sheets = append(wb.Sheets[:index], wb.Sheets[index+1:]...)

	// Adjust active sheet if necessary
	if wb.ActiveSheet >= len(wb.Sheets) {
		wb.ActiveSheet = len(wb.Sheets) - 1
	}

	wb.HasChanges = true
	return nil
}

// RenameSheet renames a sheet
func (wb *Workbook) RenameSheet(index int, newName string) error {
	if index < 0 || index >= len(wb.Sheets) {
		return fmt.Errorf("invalid sheet index")
	}
	wb.Sheets[index].Name = newName
	wb.HasChanges = true
	return nil
}

// SwitchToSheet changes the active sheet
func (wb *Workbook) SwitchToSheet(index int) error {
	if index < 0 || index >= len(wb.Sheets) {
		return fmt.Errorf("invalid sheet index")
	}
	wb.ActiveSheet = index
	return nil
}
