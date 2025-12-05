// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// excel_handler.go handles .xlsx Excel format using excelize library

package fileop

import (
	"fmt"
	"strconv"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/xuri/excelize/v2"
)

// ExcelFormatHandler handles .xlsx files
type ExcelFormatHandler struct{}

// SupportsFormat returns whether this handler supports the format
func (h *ExcelFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatXLSX
}

// Read reads an Excel .xlsx file
func (h *ExcelFormatHandler) Read(filename string) (*WorkbookResult, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %v", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	result := &WorkbookResult{
		Sheets:      make([]SheetResult, 0, len(sheetList)),
		ActiveSheet: f.GetActiveSheetIndex(),
		Version:     utils.FILEVER,
	}

	for _, sheetName := range sheetList {
		cells, rows, cols, err := h.readSheet(f, sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to read sheet %s: %v", sheetName, err)
		}

		result.Sheets = append(result.Sheets, SheetResult{
			Name:  sheetName,
			Cells: cells,
			Rows:  rows,
			Cols:  cols,
		})
	}

	return result, nil
}

// Write writes workbook to Excel .xlsx format
func (h *ExcelFormatHandler) Write(filename string, sheets []SheetInfo, activeSheet int) error {
	f := excelize.NewFile()
	defer f.Close()

	// Delete default sheet
	f.DeleteSheet("Sheet1")

	for i, sheet := range sheets {
		sheetName := sheet.Name
		if sheetName == "" {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		// Create sheet
		index, err := f.NewSheet(sheetName)
		if err != nil {
			return fmt.Errorf("failed to create sheet %s: %v", sheetName, err)
		}

		// Set as active if needed
		if i == activeSheet {
			f.SetActiveSheet(index)
		}

		// Write cells
		if err := h.writeSheet(f, sheetName, sheet); err != nil {
			return fmt.Errorf("failed to write sheet %s: %v", sheetName, err)
		}
	}

	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save Excel file: %v", err)
	}

	return nil
}

// readSheet reads a single sheet from Excel file
func (h *ExcelFormatHandler) readSheet(f *excelize.File, sheetName string) ([]*cell.Cell, int32, int32, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, 0, 0, err
	}

	var cells []*cell.Cell
	var maxRow, maxCol int32

	for rowIdx, row := range rows {
		rowNum := int32(rowIdx + 1)
		if rowNum > maxRow {
			maxRow = rowNum
		}

		for colIdx, cellValue := range row {
			colNum := int32(colIdx + 1)
			if colNum > maxCol {
				maxCol = colNum
			}

			if cellValue == "" {
				continue
			}

			// Get cell coordinates
			cellCoord, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)

			// Try to get formula
			formula, _ := f.GetCellFormula(sheetName, cellCoord)
			
			rawValue := cellValue
			displayValue := cellValue
			typeValue := "string"
			emptyStr := ""
			autotype := "auto"

			// If cell has formula, use it
			if formula != "" {
				rawValue = "$=" + formula
				typeValue = "formula"
			} else {
				// Auto-detect type
				if utils.IsNumber(cellValue, utils.DEFAULT_CELL_FINANCIAL_SIGN) {
					typeValue = "number"
				} else if isValid, format := utils.IsValidDateTime(cellValue); isValid {
					typeValue = "datetime"
					autotype = format
				}
			}

			c := &cell.Cell{
				Row:      rowNum,
				Column:   colNum,
				MaxWidth: utils.DEFAULT_CELL_MAX_WIDTH,
				MinWidth: utils.DEFAULT_CELL_MIN_WIDTH,
				RawValue: &rawValue,
				Display:  &displayValue,
				Type:     &typeValue,

				Notes:      &emptyStr,
				Valrule:    &emptyStr,
				Valrulemsg: &emptyStr,

				Color:   utils.ColorOptions["White"],
				BgColor: utils.ColorOptions["Black"],

				DecimalPoints:      utils.DEFAULT_CELL_DECIMAL_POINTS,
				ThousandsSeparator: utils.DEFAULT_CELL_THOUSANDS_SEPARATOR,
				DecimalSeparator:   utils.DEFAULT_CELL_DECIMAL_SEPARATOR,
				FinancialSign:      utils.DEFAULT_CELL_FINANCIAL_SIGN,
				DateTimeFormat:     &autotype,

				Align: 0,
				Flags: 0,

				DependsOn:  []*string{},
				Dependents: []*string{},
			}

			// Try to read formatting
			h.readCellFormatting(f, sheetName, cellCoord, c)

			cells = append(cells, c)
		}
	}

	return cells, maxRow, maxCol, nil
}

// writeSheet writes a single sheet to Excel file
func (h *ExcelFormatHandler) writeSheet(f *excelize.File, sheetName string, sheet SheetInfo) error {
	for key, cellData := range sheet.GlobalData {
		row, col := int(key[0]), int(key[1])
		
		// Get cell coordinate
		cellCoord, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			continue
		}

		// Write value or formula
		if cellData.IsFormula() && cellData.RawValue != nil {
			// Convert GoSheet formula to Excel formula
			formula := strings.TrimPrefix(*cellData.RawValue, "$=")
			formula = h.convertFormulaToExcel(formula)
			
			if err := f.SetCellFormula(sheetName, cellCoord, formula); err != nil {
				// If formula fails, write as value
				if cellData.Display != nil {
					f.SetCellValue(sheetName, cellCoord, *cellData.Display)
				}
			}
		} else if cellData.RawValue != nil {
			// Write raw value
			value := *cellData.RawValue
			
			// Try to convert to number if possible
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				f.SetCellValue(sheetName, cellCoord, num)
			} else {
				f.SetCellValue(sheetName, cellCoord, value)
			}
		}

		// Apply formatting
		h.writeCellFormatting(f, sheetName, cellCoord, cellData)
	}

	return nil
}

// readCellFormatting reads formatting from Excel cell
func (h *ExcelFormatHandler) readCellFormatting(f *excelize.File, sheetName, cellCoord string, c *cell.Cell) {
	styleID, err := f.GetCellStyle(sheetName, cellCoord)
	if err != nil {
		return
	}

	style, err := f.GetStyle(styleID)
	if err != nil {
		return
	}

	// Read font formatting
	if style.Font != nil {
		if style.Font.Bold {
			c.SetFlag(cell.FlagBold)
		}
		if style.Font.Italic {
			c.SetFlag(cell.FlagItalic)
		}
		if style.Font.Strike {
			c.SetFlag(cell.FlagStrikethrough)
		}
		if style.Font.Underline != "" {
			c.SetFlag(cell.FlagUnderline)
		}
	}

	// Read alignment
	if style.Alignment != nil {
		switch style.Alignment.Horizontal {
		case "left":
			c.Align = 1 // tview.AlignLeft
		case "center":
			c.Align = 2 // tview.AlignCenter
		case "right":
			c.Align = 3 // tview.AlignRight
		}
	}
}

// writeCellFormatting writes formatting to Excel cell
func (h *ExcelFormatHandler) writeCellFormatting(f *excelize.File, sheetName, cellCoord string, c *cell.Cell) {
	style := &excelize.Style{
		Font: &excelize.Font{
			Bold:      c.HasFlag(cell.FlagBold),
			Italic:    c.HasFlag(cell.FlagItalic),
			Strike:    c.HasFlag(cell.FlagStrikethrough),
		},
		Alignment: &excelize.Alignment{},
	}

	if c.HasFlag(cell.FlagUnderline) {
		style.Font.Underline = "single"
	}

	// Set alignment
	switch c.Align {
	case 1: // tview.AlignLeft
		style.Alignment.Horizontal = "left"
	case 2: // tview.AlignCenter
		style.Alignment.Horizontal = "center"
	case 3: // tview.AlignRight
		style.Alignment.Horizontal = "right"
	}

	styleID, err := f.NewStyle(style)
	if err != nil {
		return
	}

	f.SetCellStyle(sheetName, cellCoord, cellCoord, styleID)
}

// convertFormulaToExcel converts GoSheet formula syntax to Excel syntax
func (h *ExcelFormatHandler) convertFormulaToExcel(formula string) string {
	// Basic conversions - you may need to expand this based on your function set
	// GoSheet uses similar syntax to Excel, so most formulas should work as-is
	
	// Handle any GoSheet-specific functions that don't exist in Excel
	formula = strings.ReplaceAll(formula, "AVG(", "AVERAGE(")
	
	return formula
}
