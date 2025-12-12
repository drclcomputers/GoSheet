// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// native_handler.go handles .gsheet and .json formats

package fileop

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

// NativeFormatHandler handles .gsheet (compressed) and .json formats
type NativeFormatHandler struct{}

// SupportsFormat returns whether this handler supports the format
func (h *NativeFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatGSheet || format == FormatJSON
}

// Read reads a .gsheet or .json file
func (h *NativeFormatHandler) Read(filename string) (*WorkbookResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	isGSheet := strings.HasSuffix(filename, ".gsheet")
	
	if isGSheet {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress: %v", err)
		}
		defer gz.Close()
		reader = gz
	}

	var wbData WorkbookData
	if err := json.NewDecoder(reader).Decode(&wbData); err != nil {
		return nil, fmt.Errorf("failed to decode: %v", err)
	}

	if wbData.Version == "" && len(wbData.Sheets) == 0 {
		file.Close()
		return h.readLegacyFormat(filename)
	}

	result := &WorkbookResult{
		Sheets:      make([]SheetResult, 0, len(wbData.Sheets)),
		ActiveSheet: wbData.ActiveSheet,
		Version:     wbData.Version,
	}

	for _, sheetData := range wbData.Sheets {
		cells := h.processCellData(sheetData.Cells)
		result.Sheets = append(result.Sheets, SheetResult{
			Name:  sheetData.Name,
			Cells: cells,
			Rows:  sheetData.Rows,
			Cols:  sheetData.Cols,
		})
	}

	return result, nil
}

// Write writes workbook to .gsheet or .json format
func (h *NativeFormatHandler) Write(filename string, sheets []SheetInfo, activeSheet int) error {
	isGSheet := strings.HasSuffix(filename, ".gsheet")
	
	wbData := WorkbookData{
		Version:     utils.FILEVER,
		ActiveSheet: activeSheet,
		Sheets:      make([]SheetData, 0, len(sheets)),
	}

	for _, sheet := range sheets {
		sheetData := SheetData{
			Name:  sheet.Name,
			Rows:  sheet.Rows,
			Cols:  sheet.Cols,
			Cells: make(map[string]*CellData),
		}

		for _, c := range sheet.GlobalData {
			cName := fmt.Sprintf("%s%d", utils.ColumnName(int32(c.Column)), c.Row)
			cleanRawValue := cell.StripTviewTags(strings.TrimSpace(*c.RawValue))

			sheetData.Cells[cName] = &CellData{
				Cell:     c,
				RawValue: cleanRawValue,
			}
		}

		wbData.Sheets = append(wbData.Sheets, sheetData)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonBytes, err := json.MarshalIndent(wbData, "", "  ")
	if err != nil {
		return err
	}

	if isGSheet {
		gz := gzip.NewWriter(f)
		defer gz.Close()
		
		_, err = gz.Write(jsonBytes)
		return err
	}

	_, err = f.Write(jsonBytes)
	return err
}

// processCellData converts CellData map to cell slice
func (h *NativeFormatHandler) processCellData(cellDataMap map[string]*CellData) []*cell.Cell {
	var cells []*cell.Cell

	for _, c := range cellDataMap {
		c.Cell.RawValue = &c.RawValue

		if c.Cell.Display == nil {
			displayValue := c.RawValue
			c.Cell.Display = &displayValue
		}

		if c.Cell.Type == nil {
			typeValue := "string"
			c.Cell.Type = &typeValue
		}

		if c.Cell.Notes == nil {
			emptyStr := ""
			c.Cell.Notes = &emptyStr
		}

		if c.Cell.Valrule == nil {
			emptyStr := ""
			c.Cell.Valrule = &emptyStr
		}

		if c.Cell.Valrulemsg == nil {
			emptyStr := ""
			c.Cell.Valrulemsg = &emptyStr
		}

		if c.Cell.DependsOn == nil {
			c.Cell.DependsOn = []*string{}
		}

		if c.Cell.Dependents == nil {
			c.Cell.Dependents = []*string{}
		}

		cells = append(cells, c.Cell)
	}

	return cells
}

// readLegacyFormat handles v1.0 single-sheet format
func (h *NativeFormatHandler) readLegacyFormat(filename string) (*WorkbookResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filename, ".gsheet") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	var legacyData struct {
		Rows  int32                `json:"rows"`
		Cols  int32                `json:"cols"`
		Cells map[string]*CellData `json:"cells"`
	}

	if err := json.NewDecoder(reader).Decode(&legacyData); err != nil {
		return nil, fmt.Errorf("failed to decode legacy format: %v", err)
	}

	cells := h.processCellData(legacyData.Cells)

	return &WorkbookResult{
		Sheets: []SheetResult{
			{
				Name:  "Sheet1",
				Cells: cells,
				Rows:  legacyData.Rows,
				Cols:  legacyData.Cols,
			},
		},
		ActiveSheet: 0,
		Version:     "1.0",
	}, nil
}
