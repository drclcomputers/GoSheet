// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package file

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"
	"github.com/rivo/tview"
)

// FileFormatUI represents UI information for a file format
type FileFormatUI struct {
	Format      fileop.FileFormat
	Extension   string
	Description string
	SaveFunc    func(*tview.Table, string, map[[2]int]*cell.Cell) error
}

// GetFileFormats returns all available file formats for UI
func GetFileFormats() []FileFormatUI {
	formats := fileop.GetWritableFormats()
	result := make([]FileFormatUI, len(formats))
	
	for i, format := range formats {
		result[i] = FileFormatUI{
			Format:      format,
			Extension:   format.String(),
			Description: format.Description(),
			SaveFunc:    getSaveFunc(format),
		}
	}
	
	return result
}

// getSaveFunc returns the legacy save function for a format
func getSaveFunc(format fileop.FileFormat) func(*tview.Table, string, map[[2]int]*cell.Cell) error {
	switch format {
	case fileop.FormatGSheet:
		return fileop.SaveTable
	case fileop.FormatJSON:
		return fileop.SaveTableAsJSON
	case fileop.FormatCSV:
		return fileop.SaveTableAsCSV
	case fileop.FormatHTML:
		return fileop.SaveTableAsHTML
	case fileop.FormatTXT:
		return fileop.SaveTableAsTXT
	case fileop.FormatXLSX:
		return fileop.SaveTableAsExcel
	default:
		return nil
	}
}

// Legacy FileFormats array for compatibility with existing UI code
var FileFormats = GetFileFormats()
