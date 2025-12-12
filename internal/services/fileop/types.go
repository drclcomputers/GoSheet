// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// types.go provides unified data structures for file operations

package fileop

import (
	"gosheet/internal/services/cell"
)

// FileFormat represents a supported file format
type FileFormat int

const (
	FormatGSheet FileFormat = iota
	FormatJSON
	FormatCSV
	FormatTXT
	FormatHTML
	FormatXLSX
	FormatPDF
)

// String returns the file extension for the format
func (f FileFormat) String() string {
	switch f {
	case FormatGSheet:
		return ".gsheet"
	case FormatJSON:
		return ".json"
	case FormatCSV:
		return ".csv"
	case FormatTXT:
		return ".txt"
	case FormatHTML:
		return ".html"
	case FormatXLSX:
		return ".xlsx"
	case FormatPDF:
		return ".pdf"
	default:
		return ""
	}
}

// Description returns human-readable format description
func (f FileFormat) Description() string {
	switch f {
	case FormatGSheet:
		return "GSheet (Native Format)"
	case FormatJSON:
		return "JSON Spreadsheet"
	case FormatCSV:
		return "Comma-Separated Values"
	case FormatTXT:
		return "Tab-Delimited Text"
	case FormatHTML:
		return "HTML Table"
	case FormatXLSX:
		return "Excel Spreadsheet"
	case FormatPDF:
		return "PDF Document"
	default:
		return "Unknown Format"
	}
}

// SupportsRead checks whether format supports reading
func (f FileFormat) SupportsRead() bool {
	switch f {
	case FormatGSheet, FormatJSON, FormatTXT, FormatXLSX, FormatCSV:
		return true
	default:
		return false
	}
}

// SupportsWrite checks whether format supports writing
func (f FileFormat) SupportsWrite() bool {
	return true
}

// PreservesFormatting checks whether format preserves cell formatting
func (f FileFormat) PreservesFormatting() bool {
	switch f {
	case FormatGSheet, FormatJSON, FormatHTML, FormatXLSX, FormatPDF:
		return true
	default:
		return false
	}
}

// PreservesFormulas checks whether format preserves formulas
func (f FileFormat) PreservesFormulas() bool {
	switch f {
	case FormatGSheet, FormatJSON, FormatXLSX:
		return true
	default:
		return false
	}
}

// SupportsMultipleSheets checks whether format supports multiple sheets
func (f FileFormat) SupportsMultipleSheets() bool {
	switch f {
	case FormatGSheet, FormatJSON, FormatXLSX:
		return true
	default:
		return false
	}
}

// DetectFormat detects file format from filename
func DetectFormat(filename string) (FileFormat, bool) {
	formats := []FileFormat{
		FormatGSheet,
		FormatJSON,
		FormatCSV,
		FormatTXT,
		FormatHTML,
		FormatXLSX,
		FormatPDF,
	}
	
	for _, format := range formats {
		if len(filename) >= len(format.String()) &&
			filename[len(filename)-len(format.String()):] == format.String() {
			return format, true
		}
	}
	
	return FormatGSheet, false
}

// WorkbookData represents the complete workbook structure (v2.0 format)
type WorkbookData struct {
	Version     string      `json:"version"`
	ActiveSheet int         `json:"active_sheet"`
	Sheets      []SheetData `json:"sheets"`
}

// SheetData represents a single sheet's data
type SheetData struct {
	Name  string               `json:"name"`
	Rows  int32                `json:"rows"`
	Cols  int32                `json:"cols"`
	Cells map[string]*CellData `json:"cells"`
}

// CellData represents serializable cell data
type CellData struct {
	Cell     *cell.Cell `json:"cell"`
	RawValue string     `json:"raw_value"`
}

// SheetInfo contains runtime sheet information for saving
type SheetInfo struct {
	Name       string
	Rows       int32
	Cols       int32
	GlobalData map[[2]int]*cell.Cell
}

// WorkbookResult contains loaded workbook data
type WorkbookResult struct {
	Sheets      []SheetResult
	ActiveSheet int
	Version     string
	Format      FileFormat
}

// SheetResult contains loaded sheet data
type SheetResult struct {
	Name  string
	Cells []*cell.Cell
	Rows  int32
	Cols  int32
}

// FileReader interface for reading different formats
type FileReader interface {
	Read(filename string) (*WorkbookResult, error)
	SupportsFormat(format FileFormat) bool
}

// FileWriter interface for writing different formats
type FileWriter interface {
	Write(filename string, sheets []SheetInfo, activeSheet int) error
	SupportsFormat(format FileFormat) bool
}

// GetAllFormats returns all available file formats
func GetAllFormats() []FileFormat {
	return []FileFormat{
		FormatGSheet,
		FormatJSON,
		FormatCSV,
		FormatTXT,
		FormatHTML,
		FormatXLSX,
		FormatPDF,
	}
}

// GetReadableFormats returns formats that support reading
func GetReadableFormats() []FileFormat {
	formats := GetAllFormats()
	readable := make([]FileFormat, 0)
	for _, f := range formats {
		if f.SupportsRead() {
			readable = append(readable, f)
		}
	}
	return readable
}

// GetWritableFormats returns formats that support writing
func GetWritableFormats() []FileFormat {
	formats := GetAllFormats()
	writable := make([]FileFormat, 0)
	for _, f := range formats {
		if f.SupportsWrite() {
			writable = append(writable, f)
		}
	}
	return writable
}
