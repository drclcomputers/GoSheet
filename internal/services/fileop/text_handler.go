// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// text_handler.go handles importing tab-delimited text files

package fileop

import (
	"bufio"
	"os"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

// TextFormatHandler handles importing .txt files
type TextFormatHandler struct{}

// SupportsFormat returns whether this handler supports the format
func (h *TextFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatTXT
}

// Read reads a tab-delimited text file
func (h *TextFormatHandler) Read(filename string) (*WorkbookResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cells []*cell.Cell
	scanner := bufio.NewScanner(file)
	var maxCol int32
	row := int32(1)

	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, "\t")

		if int32(len(values)) > maxCol {
			maxCol = int32(len(values))
		}

		for col, value := range values {
			if value == "" {
				continue
			}

			cellValue := value
			displayValue := value
			typeValue := "string"
			emptyStr := ""
			autotype := "auto"

			c := &cell.Cell{
				Row:      row,
				Column:   int32(col + 1),
				MaxWidth: utils.DEFAULT_CELL_MAX_WIDTH,
				MinWidth: utils.DEFAULT_CELL_MIN_WIDTH,
				RawValue: &cellValue,
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
			cells = append(cells, c)
		}
		row++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	rows := row - 1

	return &WorkbookResult{
		Sheets: []SheetResult{
			{
				Name:  "Sheet1",
				Cells: cells,
				Rows:  rows,
				Cols:  maxCol,
			},
		},
		ActiveSheet: 0,
		Version:     utils.FILEVER,
	}, nil
}
