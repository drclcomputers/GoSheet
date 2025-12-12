// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// formatpainter.go provides functions used by the Format Painter feature

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

var formatSourceCell *cell.Cell

// Copy the cell style to the clipboard
func robCopyCellFormat(table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	visualRow, visualCol := table.GetSelection()
	absRow, absCol := activeViewport.ToAbsolute(int32(visualRow), int32(visualCol))

	if absRow == 0 || absCol == 0 {
		return
	}

	key := [2]int{int(absRow), int(absCol)}
	if cellData, exists := activeData[key]; exists {
		formatSourceCell = &cell.Cell{
			Color:               cellData.Color,
			BgColor:             cellData.BgColor,
			Flags:               cellData.Flags,
			Align:               cellData.Align,
			DecimalPoints:       cellData.DecimalPoints,
			ThousandsSeparator:  cellData.ThousandsSeparator,
			DecimalSeparator:    cellData.DecimalSeparator,
			FinancialSign:       cellData.FinancialSign,
			Valrule: 			 cellData.Valrule,
			Valrulemsg: 		 cellData.Valrulemsg,
			MaxWidth: 			 cellData.MaxWidth,
			MinWidth: 			 cellData.MinWidth,
		}
	}
}

// Paste clipboard style to selected cell/s
func imitatePasteCellFormat(app *tview.Application, table *tview.Table, visualTargetRow, visualTargetCol int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	if formatSourceCell == nil {
		return
	}

	targetRow, targetCol := activeViewport.ToAbsolute(visualTargetRow, visualTargetCol)

	r1, c1, r2, c2 := getSelectionRange(table)
	
	if r1 == r2 && c1 == c2 && targetRow != 0 && targetCol != 0 {
		r1, c1, r2, c2 = targetRow, targetCol, targetRow, targetCol
	}

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if r == 0 || c == 0 {
				continue
			}

			key := [2]int{int(r), int(c)}
			if cellData, exists := activeData[key]; exists {
				if !cellData.HasFlag(cell.FlagEditable) {
					modal := tview.NewModal().
						SetText(fmt.Sprintf("Target cell %s%d and potential others are uneditable.\nDo you wish to overwrite formatting on uneditable cells?", 
							utils.ColumnName(int32(c)), r)).
						AddButtons([]string{"Yes", "Cancel"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "Yes" {
								performActualFormatPaste(table, formatSourceCell, r1, c1, r2, c2)
							}
							app.SetRoot(table, true).SetFocus(table)
						})
					modal.SetBorder(true).SetTitle("Confirm Format Paste").SetTitleAlign(tview.AlignCenter)
					app.SetRoot(modal, true).SetFocus(modal)
					return
				}
			}
		}
	}

	performActualFormatPaste(table, formatSourceCell, r1, c1, r2, c2)
	clearSelectionRange()
}

// Function that actually pastes cells
func performActualFormatPaste(table *tview.Table, sourceFormat *cell.Cell, r1, c1, r2, c2 int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}
	
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if r == 0 || c == 0 {
				continue
			}
			
			targetCell := cell.GetOrCreateCell(table, r, c, activeData)
			
			targetCell.Color = sourceFormat.Color
			targetCell.BgColor = sourceFormat.BgColor
			targetCell.Flags = sourceFormat.Flags
			targetCell.Align = sourceFormat.Align
			targetCell.DecimalPoints = sourceFormat.DecimalPoints
			targetCell.ThousandsSeparator = sourceFormat.ThousandsSeparator
			targetCell.DecimalSeparator = sourceFormat.DecimalSeparator
			targetCell.FinancialSign = sourceFormat.FinancialSign
			targetCell.Valrule = sourceFormat.Valrule
			targetCell.Valrulemsg = sourceFormat.Valrulemsg
			targetCell.MaxWidth = sourceFormat.MaxWidth
			targetCell.MinWidth = sourceFormat.MinWidth

			if targetCell.Type != nil && (*targetCell.Type == "number" || *targetCell.Type == "financial") {
				if targetCell.RawValue != nil {
					normalized := strings.ReplaceAll(*targetCell.RawValue, string(targetCell.ThousandsSeparator), "")
					normalized = strings.TrimPrefix(normalized, string(targetCell.FinancialSign))
					
					if val, err := strconv.ParseFloat(normalized, 64); err == nil {
						formatted := utils.FormatWithCommas(val, targetCell.ThousandsSeparator, 
							targetCell.DecimalSeparator, targetCell.DecimalPoints, targetCell.FinancialSign)
						
						if *targetCell.Type == "financial" {
							formatted = fmt.Sprintf("%c%s", targetCell.FinancialSign, formatted)
						}
						*targetCell.Display = formatted
					}
				}
			}
			
			if activeViewport.IsVisible(r, c) {
				visualR, visualC := activeViewport.ToRelative(r, c)
				table.SetCell(int(visualR), int(visualC), targetCell.ToTViewCell())
			}
		}
	}
}
