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

func (h *ExcelFormatHandler) hasNonDefaultFormatting(f *excelize.File, sheetName, cellCoord string) bool {
	styleID, err := f.GetCellStyle(sheetName, cellCoord)
	if err != nil || styleID == 0 {
		return false
	}

	style, err := f.GetStyle(styleID)
	if err != nil {
		return false
	}

	if style.Font != nil {
		if style.Font.Bold || style.Font.Italic || style.Font.Strike || style.Font.Underline != "" {
			return true
		}
		if style.Font.Color != "" && style.Font.Color != "000000" {
			return true
		}
	}

	if len(style.Fill.Color) > 0 && style.Fill.Color[0] != "" {
		bgColor := strings.ToUpper(strings.TrimSpace(style.Fill.Color[0]))
		if len(bgColor) == 8 {
			bgColor = bgColor[2:]
		}
		if bgColor != "FFFFFF" && bgColor != "" {
			return true
		}
	}

	if style.Alignment != nil && style.Alignment.Horizontal != "" {
		return true
	}

	return false
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

	if len(sheets) > 0 {
		f.DeleteSheet("Sheet1")
	}

	for i, sheet := range sheets {
		sheetName := sheet.Name
		if sheetName == "" {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		_, err := f.NewSheet(sheetName)
		if err != nil {
			return fmt.Errorf("failed to create sheet %s: %v", sheetName, err)
		}

		if i == activeSheet {
			f.SetActiveSheet(i)
		}

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
		for colIdx := range row {
			colNum := int32(colIdx + 1)
			if colNum > maxCol {
				maxCol = colNum
			}
		}
	}

	for rowIdx := 0; rowIdx < int(maxRow); rowIdx++ {
		for colIdx := 0; colIdx < int(maxCol); colIdx++ {
			rowNum := int32(rowIdx + 1)
			colNum := int32(colIdx + 1)

			cellCoord, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			
			formula, _ := f.GetCellFormula(sheetName, cellCoord)
			cellValue, _ := f.GetCellValue(sheetName, cellCoord)
			
			hasFormatting := h.hasNonDefaultFormatting(f, sheetName, cellCoord)
			
			if cellValue == "" && formula == "" && !hasFormatting {
				continue
			}

			rawValue := cellValue
			displayValue := cellValue
			typeValue := "string"
			emptyStr := ""
			autotype := "auto"

			if formula != "" {
				formula = strings.TrimPrefix(formula, "_xludf.")
				formula = strings.TrimPrefix(formula, "_XLUDF.")
				formula = strings.TrimSpace(formula)
				
				rawValue = "$=" + h.convertExcelFormulaToGoSheet(formula)
				typeValue = "formula"
			} else {
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
				Flags: cell.FlagEditable,

				DependsOn:  []*string{},
				Dependents: []*string{},
			}

			h.readCellFormatting(f, sheetName, cellCoord, c)

			comments, _ := f.GetComments(sheetName)
			for _, comment := range comments {
				if comment.Cell == cellCoord && len(comment.Paragraph) > 0 {
					noteText := comment.Paragraph[0].Text
					*c.Notes = noteText
					break
				}
			}

			cells = append(cells, c)
		}
	}

	return cells, maxRow, maxCol, nil
}

// writeSheet writes a single sheet to Excel file
func (h *ExcelFormatHandler) writeSheet(f *excelize.File, sheetName string, sheet SheetInfo) error {
	for _, cellData := range sheet.GlobalData {
		row, col := int(cellData.Row), int(cellData.Column)

		cellCoord, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			continue
		}

		h.writeCellFormatting(f, sheetName, cellCoord, cellData)

		if cellData.IsFormula() && cellData.RawValue != nil {
			formulaStr := strings.TrimPrefix(*cellData.RawValue, "$=")
			formulaStr = strings.TrimSpace(formulaStr)
			
			if formulaStr != "" {
				excelFormula := h.convertFormulaToExcel(formulaStr)
				
				if err := f.SetCellFormula(sheetName, cellCoord, excelFormula); err != nil {
					if cellData.Display != nil && *cellData.Display != "" {
						f.SetCellValue(sheetName, cellCoord, *cellData.Display)
					}
				}
				
				goto handleMetadata
			}
		}

		if cellData.RawValue != nil && *cellData.RawValue != "" {
			value := *cellData.RawValue

			if *cellData.Type == "number" || *cellData.Type == "financial" {
				cleanValue := value
				cleanValue = strings.ReplaceAll(cleanValue, string(cellData.ThousandsSeparator), "")
				cleanValue = strings.TrimPrefix(cleanValue, string(cellData.FinancialSign))
				cleanValue = strings.TrimSpace(cleanValue)
				
				if num, err := strconv.ParseFloat(cleanValue, 64); err == nil {
					f.SetCellValue(sheetName, cellCoord, num)
				} else {
					f.SetCellValue(sheetName, cellCoord, value)
				}
			} else {
				f.SetCellValue(sheetName, cellCoord, value)
			}
		} else if cellData.Display != nil && *cellData.Display != "" {
			f.SetCellValue(sheetName, cellCoord, *cellData.Display)
		}

	handleMetadata:
		if cellData.Notes != nil && strings.TrimSpace(*cellData.Notes) != "" {
			err := f.AddComment(sheetName, excelize.Comment{
				Cell:   cellCoord,
				Author: "GoSheet",
				Paragraph: []excelize.RichTextRun{
					{Text: *cellData.Notes},
				},
			})
			if err != nil {
				return err
			}
		}

		if cellData.MinWidth > 0 {
			colName, _ := excelize.ColumnNumberToName(col)
			currentWidth, _ := f.GetColWidth(sheetName, colName)
			newWidth := float64(cellData.MinWidth)
			if newWidth > currentWidth {
				f.SetColWidth(sheetName, colName, colName, newWidth)
			}
		}
	}

	return nil
}

// readCellFormatting reads formatting from Excel cell INCLUDING COLORS
func (h *ExcelFormatHandler) readCellFormatting(f *excelize.File, sheetName, cellCoord string, c *cell.Cell) {
	styleID, err := f.GetCellStyle(sheetName, cellCoord)
	if err != nil {
		return
	}

	style, err := f.GetStyle(styleID)
	if err != nil {
		return
	}

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

		if style.Font.Color != "" {
			if color, err := parseExcelColor(style.Font.Color); err == nil {
				c.Color = color
			}
		}
	}

	if style.Alignment != nil {
		switch style.Alignment.Horizontal {
		case "left":
			c.Align = 1
		case "center":
			c.Align = 2
		case "right":
			c.Align = 3
		}
	}

	if len(style.Fill.Color) > 0 && style.Fill.Color[0] != "" {
		if color, err := parseExcelColor(style.Fill.Color[0]); err == nil {
			c.BgColor = color
		}
	}
}

// writeCellFormatting writes formatting to Excel cell INCLUDING COLORS
func (h *ExcelFormatHandler) writeCellFormatting(f *excelize.File, sheetName, cellCoord string, c *cell.Cell) {
	style := &excelize.Style{
		Font:      &excelize.Font{},
		Alignment: &excelize.Alignment{},
	}

	style.Font.Bold = c.HasFlag(cell.FlagBold)
	style.Font.Italic = c.HasFlag(cell.FlagItalic)
	style.Font.Strike = c.HasFlag(cell.FlagStrikethrough)

	if c.HasFlag(cell.FlagUnderline) {
		style.Font.Underline = "single"
	}

	if !c.Color.IsDefaultWhite() {
		style.Font.Color = strings.TrimPrefix(c.Color.Hex(), "#")
	}

	if !c.BgColor.IsDefaultBlack() {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{strings.TrimPrefix(c.BgColor.Hex(), "#")},
		}
	}

	switch c.Align {
	case 1:
		style.Alignment.Horizontal = "left"
	case 2:
		style.Alignment.Horizontal = "center"
	case 3:
		style.Alignment.Horizontal = "right"
	}

	styleID, err := f.NewStyle(style)
	if err != nil {
		return
	}

	f.SetCellStyle(sheetName, cellCoord, cellCoord, styleID)
}

// parseExcelColor converts Excel color format to ColorRGB
func parseExcelColor(excelColor string) (utils.ColorRGB, error) {
	excelColor = strings.TrimSpace(excelColor)
	excelColor = strings.ToUpper(excelColor)
	
	if len(excelColor) == 8 {
		excelColor = excelColor[2:]
	}
	
	if len(excelColor) == 6 {
		r, err1 := strconv.ParseUint(excelColor[0:2], 16, 8)
		g, err2 := strconv.ParseUint(excelColor[2:4], 16, 8)
		b, err3 := strconv.ParseUint(excelColor[4:6], 16, 8)
		
		if err1 == nil && err2 == nil && err3 == nil {
			return utils.ColorRGB{uint8(r), uint8(g), uint8(b)}, nil
		}
	}
	
	if color, ok := utils.ColorOptions[excelColor]; ok {
		return color, nil
	}
	
	return utils.ColorRGB{255, 255, 255}, fmt.Errorf("invalid color format: %s", excelColor)
}
