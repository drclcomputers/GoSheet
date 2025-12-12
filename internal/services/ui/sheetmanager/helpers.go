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
		icon := ""
		badge := ""
		
		if sheet.IsActive {
			icon = ""
			badge = " [yellow::b]> ACTIVE[::-]"
		}

		mainText := fmt.Sprintf(" %s  %s%s", icon, sheet.Name, badge)
		secondaryText := fmt.Sprintf("   └─ %d cells with data", sheet.CellCount)

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
		fileName = "[gray]Untitled Workbook[-]"
	} else {
		fileName = "[white]" + fileName + "[-]"
	}

	statusIcon := ""
	statusColor := "green"
	statusText := "Saved"
	
	if info.HasChanges {
		statusIcon = "o"
		statusColor = "yellow"
		statusText = "Modified"
	}

	return fmt.Sprintf(
		"[::b]WORKBOOK OVERVIEW[::-]\n"+
			"[gray]━━━━━━━━━━━━━━━━━━━━[-]\n"+
			"[lightblue]File:[-]\n  %s\n"+
			"[lightblue]Status:[-]  [%s]%s %s[-]\n"+
			"[lightblue]Structure:[-]\n"+
			"  • Sheets: [white]%d[-]\n"+
			"  • Active: [white]%s[-]\n"+
			"  • Total Cells: [white]%d[-]",
		fileName,
		statusColor, statusIcon, statusText,
		info.TotalSheets,
		info.ActiveSheet,
		info.TotalCells,
	)
}

// getSheetInfoText returns enhanced info for a specific sheet
func getSheetInfoText(sheet SheetInfo, index, total int) string {
	statusIcon := "o"
	statusColor := "gray"
	statusText := "Inactive"
	
	if sheet.IsActive {
		statusIcon = ">"
		statusColor = "yellow"
		statusText = "Active Sheet"
	}

	return fmt.Sprintf(
		"[::b]SHEET DETAILS[::-]\n"+
			"[gray]━━━━━━━━━━━━━━━━━━━━[-]\n"+
			"[lightblue]Name:[-]  [white::b]%s[::-]\n"+
			"[lightblue]Status:[-]  [%s]%s %s[-]\n"+
			"[lightblue]Content:[-]\n"+
			"  • Cells: [white]%d[-]\n"+
			"  • Position: [white]%d[-] of [white]%d[-]",
		sheet.Name,
		statusColor, statusIcon, statusText,
		sheet.CellCount,
		index+1,
		total,
	)
}
