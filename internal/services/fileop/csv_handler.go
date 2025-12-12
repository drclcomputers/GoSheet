package fileop

import (
	"encoding/csv"
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"io"
	"os"
	"strings"
)

type CSVFormatHandler struct{}

// SupportsFormat returns whether this handler supports the format
func (h *CSVFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatCSV
}

// writeCSV exports to CSV format
func (h *CSVFormatHandler) Write(filename string, sheets []SheetInfo, activeSheet int) error {
	sheet := sheets[activeSheet]

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var maxRow, maxCol int32
	for key := range sheet.GlobalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	for row := int32(1); row <= maxRow; row++ {
		record := make([]string, maxCol)

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			if cellData, exists := sheet.GlobalData[key]; exists && cellData.RawValue != nil {
				if *cellData.RawValue != "" {
					record[col-1] = *cellData.RawValue
				} else if cellData.Display != nil {
					record[col-1] = *cellData.Display
				} else {
					record[col-1] = ""
				}
			} else {
				record[col-1] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Read imports a CSV file into a single sheet
func (h *CSVFormatHandler) Read(filename string) (*WorkbookResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 

	var cells []*cell.Cell
	var maxCol int32
	rowNum := int32(1)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if int32(len(record)) > maxCol {
			maxCol = int32(len(record))
		}

		for colIdx, value := range record {
			value = strings.TrimSpace(value)
			if value == "" {
				continue 
			}

			colNum := int32(colIdx + 1)
			emptyStr := ""
			autotype := "auto"
			cellType := "string"

			if utils.IsNumber(value, utils.DEFAULT_CELL_FINANCIAL_SIGN) {
				cellType = "number"
			}

			c := &cell.Cell{
				Row:      rowNum,
				Column:   colNum,

				MaxWidth: utils.DEFAULT_CELL_MAX_WIDTH,
				MinWidth: utils.DEFAULT_CELL_MIN_WIDTH,

				RawValue: &value,
				Display:  &value,
				Type:     &cellType,

				Notes:      &emptyStr,
				Valrule:    &emptyStr,
				Valrulemsg: &emptyStr,

				Color:              utils.ColorOptions["White"],
				BgColor:            utils.ColorOptions["Black"],

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

			cells = append(cells, c)
		}

		rowNum++
	}

	if len(cells) == 0 {
		maxCol = 10 // reasonable default
	}

	return &WorkbookResult{
		Sheets: []SheetResult{
			{
				Name:  "Sheet1",
				Cells: cells,
				Rows:  rowNum - 1,
				Cols:  maxCol,
			},
		},
		ActiveSheet: 0,
		Version:     utils.FILEVER,
		Format:      FormatCSV,
	}, nil
}
