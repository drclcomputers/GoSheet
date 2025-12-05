// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// compat.go provides backward compatibility with existing code

package fileop

import (
	"gosheet/internal/services/cell"
	"github.com/rivo/tview"
	"fmt"
)

// Legacy function wrappers for backward compatibility
// These allow existing code to continue working while using the new system

// OpenTable loads first sheet only (legacy compatibility)
func OpenTable(filename string) ([]*cell.Cell, error) {
	result, err := OpenWorkbook(filename)
	if err != nil {
		return nil, err
	}

	if len(result.Sheets) == 0 {
		return nil, nil
	}

	return result.Sheets[0].Cells, nil
}

// SaveTable saves single sheet in native format (legacy compatibility)
func SaveTable(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbook(sheets, 0, filename)
}

// SaveTableAsJSON saves in JSON format (legacy compatibility)
func SaveTableAsJSON(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAs(sheets, 0, filename, FormatJSON)
}

// SaveTableAsCSV exports as CSV (legacy compatibility)
func SaveTableAsCSV(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAs(sheets, 0, filename, FormatCSV)
}

// SaveTableAsHTML exports as HTML (legacy compatibility)
func SaveTableAsHTML(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAs(sheets, 0, filename, FormatHTML)
}

// SaveTableAsTXT exports as TXT (legacy compatibility)
func SaveTableAsTXT(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAs(sheets, 0, filename, FormatTXT)
}

// SaveTableAsExcel exports as XLSX
func SaveTableAsExcel(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAs(sheets, 0, filename, FormatXLSX)
}

// SaveTableAsPDF placeholder (not implemented)
func SaveTableAsPDF(table *tview.Table, filename string) error {
	return fmt.Errorf("PDF export not yet implemented")
}

