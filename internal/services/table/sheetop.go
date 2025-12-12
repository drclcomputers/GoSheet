// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// sheetop.go provides core sheet management operations

package table

import (
	"fmt"
	"gosheet/internal/services/ui/sheetmanager"

	"github.com/rivo/tview"
)

// GetSheets returns information about all sheets
func GetSheets() []sheetmanager.SheetInfo {
	if globalWorkbook == nil {
		return []sheetmanager.SheetInfo{}
	}

	sheets := make([]sheetmanager.SheetInfo, len(globalWorkbook.Sheets))
	for i, sheet := range globalWorkbook.Sheets {
		sheets[i] = sheetmanager.SheetInfo{
			Name:      sheet.Name,
			CellCount: len(sheet.Data),
			IsActive:  i == globalWorkbook.ActiveSheet,
		}
	}
	return sheets
}

// GetWorkbookInfo returns information about the workbook
func GetWorkbookInfo() sheetmanager.WorkbookInfo {
	if globalWorkbook == nil {
		return sheetmanager.WorkbookInfo{}
	}

	totalCells := 0
	for _, sheet := range globalWorkbook.Sheets {
		totalCells += len(sheet.Data)
	}

	activeSheetName := ""
	if activeSheet := globalWorkbook.GetActiveSheet(); activeSheet != nil {
		activeSheetName = activeSheet.Name
	}

	return sheetmanager.WorkbookInfo{
		TotalSheets: len(globalWorkbook.Sheets),
		ActiveSheet: activeSheetName,
		TotalCells:  totalCells,
		FileName:    globalWorkbook.CurrentFile,
		HasChanges:  globalWorkbook.HasChanges,
	}
}

// AddSheetWithName adds a new sheet with the given name
func AddSheetWithName(name string) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	for _, sheet := range globalWorkbook.Sheets {
		if sheet.Name == name {
			return fmt.Errorf("a sheet with this name already exists")
		}
	}

	globalWorkbook.AddSheet(name)
	return nil
}

// RenameSheetByIndex renames a sheet at the given index
func RenameSheetByIndex(index int, newName string) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	if index < 0 || index >= len(globalWorkbook.Sheets) {
		return fmt.Errorf("invalid sheet index")
	}

	for i, sheet := range globalWorkbook.Sheets {
		if i != index && sheet.Name == newName {
			return fmt.Errorf("a sheet with this name already exists")
		}
	}

	return globalWorkbook.RenameSheet(index, newName)
}

// DeleteSheetByIndex deletes a sheet at the given index
func DeleteSheetByIndex(index int) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	return globalWorkbook.DeleteSheet(index)
}

// DuplicateSheetByIndex duplicates a sheet at the given index
func DuplicateSheetByIndex(index int) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	if index < 0 || index >= len(globalWorkbook.Sheets) {
		return fmt.Errorf("invalid sheet index")
	}

	sourceSheet := globalWorkbook.Sheets[index]

	newName := fmt.Sprintf("%s (Copy)", sourceSheet.Name)
	counter := 1
	for {
		exists := false
		for _, sheet := range globalWorkbook.Sheets {
			if sheet.Name == newName {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		counter++
		newName = fmt.Sprintf("%s (Copy %d)", sourceSheet.Name, counter)
	}

	newSheet := NewSheet(newName)

	for key, cellData := range sourceSheet.Data {
		newSheet.Data[key] = cellData.Clone()
	}

	newSheet.Viewport.TopRow = sourceSheet.Viewport.TopRow
	newSheet.Viewport.LeftCol = sourceSheet.Viewport.LeftCol
	newSheet.Viewport.ViewRows = sourceSheet.Viewport.ViewRows
	newSheet.Viewport.ViewCols = sourceSheet.Viewport.ViewCols

	globalWorkbook.Sheets = append(globalWorkbook.Sheets, newSheet)
	globalWorkbook.HasChanges = true

	return nil
}

// MoveSheetByIndex moves a sheet from one position to another
func MoveSheetByIndex(fromIndex, toIndex int) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	if fromIndex < 0 || fromIndex >= len(globalWorkbook.Sheets) {
		return fmt.Errorf("invalid from index")
	}

	if toIndex < 0 || toIndex >= len(globalWorkbook.Sheets) {
		return fmt.Errorf("invalid to index")
	}

	sheet := globalWorkbook.Sheets[fromIndex]
	globalWorkbook.Sheets = append(
		globalWorkbook.Sheets[:fromIndex],
		globalWorkbook.Sheets[fromIndex+1:]...,
	)

	newSheets := make([]*Sheet, 0, len(globalWorkbook.Sheets)+1)
	newSheets = append(newSheets, globalWorkbook.Sheets[:toIndex]...)
	newSheets = append(newSheets, sheet)
	newSheets = append(newSheets, globalWorkbook.Sheets[toIndex:]...)
	globalWorkbook.Sheets = newSheets

	if globalWorkbook.ActiveSheet == fromIndex {
		globalWorkbook.ActiveSheet = toIndex
	} else if fromIndex < globalWorkbook.ActiveSheet && toIndex >= globalWorkbook.ActiveSheet {
		globalWorkbook.ActiveSheet--
	} else if fromIndex > globalWorkbook.ActiveSheet && toIndex <= globalWorkbook.ActiveSheet {
		globalWorkbook.ActiveSheet++
	}

	globalWorkbook.HasChanges = true
	return nil
}

// SwitchToSheetByIndex switches to a sheet at the given index
func SwitchToSheetByIndex(index int) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	return globalWorkbook.SwitchToSheet(index)
}

// UpdateTableTitleView updates the table title
func UpdateTableTitleView(table *tview.Table) {
	if table != nil {
		updateTableTitle(table)
	}
}

// MarkAsModifiedView marks the workbook as modified
func MarkAsModifiedView(table *tview.Table) {
	if table != nil {
		MarkAsModified(table)
	}
}

// RenderActiveSheetView renders the active sheet
func RenderActiveSheetView(table *tview.Table) {
	if globalWorkbook == nil {
		return
	}

	sheet := globalWorkbook.GetActiveSheet()
	if sheet != nil {
		RenderVisible(table, sheet.Viewport, sheet.Data)
	}
}

// GetSheetManagerCallbacks returns callbacks for the sheet manager sheetmanager
func GetSheetManagerCallbacks(table *tview.Table) sheetmanager.SheetManagerCallbacks {
	return sheetmanager.SheetManagerCallbacks{
		GetSheets:      GetSheets,
		GetActiveIndex: func() int {
			if globalWorkbook == nil {
				return -1
			}
			return globalWorkbook.ActiveSheet
		},
		GetWorkbookInfo:   GetWorkbookInfo,
		AddSheet:          AddSheetWithName,
		RenameSheet:       RenameSheetByIndex,
		DeleteSheet:       DeleteSheetByIndex,
		DuplicateSheet:    DuplicateSheetByIndex,
		MoveSheet:         MoveSheetByIndex,
		SwitchToSheet:     SwitchToSheetByIndex,
		UpdateTableTitle:  func() { UpdateTableTitleView(table) },
		MarkAsModified:    func() { MarkAsModifiedView(table) },
		RenderActiveSheet: func() { RenderActiveSheetView(table) },
		UpdateTabBar:      func() {},
	}
}

// ShowSheetManagerDialog shows the sheet manager dialog
func ShowSheetManagerDialog(app *tview.Application, table *tview.Table) {
	callbacks := GetSheetManagerCallbacks(table)
	sheetmanager.ShowSheetManager(app, table, callbacks)
}
