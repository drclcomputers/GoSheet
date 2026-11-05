// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// compat.go provides hooks for the UI layer to save workbooks.

package fileop

// GetWorkbookForSaveFunc is a hook set by the table package to provide workbook sheets for saving.
var GetWorkbookForSaveFunc func() (sheets []SheetInfo, activeSheet int, hasWorkbook bool)

// GetWorkbookForSave returns the current workbook sheets for saving.
func GetWorkbookForSave() ([]SheetInfo, int, bool) {
	if GetWorkbookForSaveFunc == nil {
		return nil, 0, false
	}
	return GetWorkbookForSaveFunc()
}
