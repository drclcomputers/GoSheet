// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpers.go provides helper functions for managing sheets

package sheetmanager

import (
	"fmt"

	"github.com/rivo/tview"
)

// updateSheetList refreshes the sheet list with enhanced display
func updateSheetList(list *tview.List, callbacks SheetManagerCallbacks) {
	list.Clear()
	sheets := callbacks.GetSheets()

	for i, sheet := range sheets {
		icon := "  "
		if sheet.IsActive {
			icon = "* "
		}

		mainText := icon + sheet.Name
		secondaryText := fmt.Sprintf("    %d cells", sheet.CellCount)
		if sheet.IsActive {
			secondaryText = fmt.Sprintf("    active  ·  %d cells", sheet.CellCount)
		}

		list.AddItem(
			mainText,
			secondaryText,
			0,
			func(idx int) func() {
				return func() {
					callbacks.SwitchToSheet(idx)
					callbacks.UpdateTabBar()
					callbacks.UpdateTableTitle()
					callbacks.RenderActiveSheet()
				}
			}(i),
		)
	}
}

// getWorkbookInfoText returns enhanced workbook information
func getWorkbookInfoText(info WorkbookInfo) string {
	fileName := info.FileName
	if fileName == "" {
		fileName = "[#5c6b7a]Untitled[-]"
	} else {
		fileName = "[#e8eef4]" + fileName + "[-]"
	}

	statusColor := "#7dcea0"
	statusText := "Saved"
	if info.HasChanges {
		statusColor = "#f0d090"
		statusText = "Modified"
	}

	return fmt.Sprintf(
		"[#f0d090::b]Workbook[-]\n\n"+
			"[#8aa0b8]File[-]\n  %s\n"+
			"[#8aa0b8]Status[-]  [%s]%s[-]\n\n"+
			"[#8aa0b8]Sheets[-]  [#e8eef4]%d[-]\n"+
			"[#8aa0b8]Open[-]    [#e8eef4]%s[-]\n"+
			"[#8aa0b8]Cells[-]   [#e8eef4]%d[-]",
		fileName,
		statusColor, statusText,
		info.TotalSheets,
		info.ActiveSheet,
		info.TotalCells,
	)
}

// getSheetInfoText returns enhanced info for a specific sheet
func getSheetInfoText(sheet SheetInfo, index, total int) string {
	statusColor := "#8aa0b8"
	statusText := "Not the open sheet"
	if sheet.IsActive {
		statusColor = "#f0d090"
		statusText = "Open"
	}

	return fmt.Sprintf(
		"[#f0d090::b]%s[-]\n\n"+
			"[#8aa0b8]Status[-]  [%s]%s[-]\n"+
			"[#8aa0b8]Cells[-]   [#e8eef4]%d[-]\n"+
			"[#8aa0b8]Order[-]   [#e8eef4]%d[-] of [#e8eef4]%d[-]",
		sheet.Name,
		statusColor, statusText,
		sheet.CellCount,
		index+1,
		total,
	)
}
