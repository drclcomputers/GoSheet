// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// manager.go provides unified file operations management

package fileop

import (
	"fmt"
	"os"
)

// FileManager handles all file operations with format detection
type FileManager struct {
	readers map[FileFormat]FileReader
	writers map[FileFormat]FileWriter
}

// NewFileManager creates a new file operations manager
func NewFileManager() *FileManager {
	fm := &FileManager{
		readers: make(map[FileFormat]FileReader),
		writers: make(map[FileFormat]FileWriter),
	}
	
	nativeHandler := &NativeFormatHandler{}
	fm.RegisterReader(FormatGSheet, nativeHandler)
	fm.RegisterReader(FormatJSON, nativeHandler)
	fm.RegisterWriter(FormatGSheet, nativeHandler)
	fm.RegisterWriter(FormatJSON, nativeHandler)
	
	exportHandler := &ExportFormatHandler{}
	fm.RegisterWriter(FormatCSV, exportHandler)
	fm.RegisterWriter(FormatTXT, exportHandler)
	fm.RegisterWriter(FormatHTML, exportHandler)
	
	textHandler := &TextFormatHandler{}
	fm.RegisterReader(FormatTXT, textHandler)
	
	excelHandler := &ExcelFormatHandler{}
	fm.RegisterReader(FormatXLSX, excelHandler)
	fm.RegisterWriter(FormatXLSX, excelHandler)
	
	return fm
}

// RegisterReader registers a reader for a specific format
func (fm *FileManager) RegisterReader(format FileFormat, reader FileReader) {
	fm.readers[format] = reader
}

// RegisterWriter registers a writer for a specific format
func (fm *FileManager) RegisterWriter(format FileFormat, writer FileWriter) {
	fm.writers[format] = writer
}

// Open opens a file and returns workbook data
func (fm *FileManager) Open(filename string) (*WorkbookResult, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}
	
	format, ok := DetectFormat(filename)
	if !ok {
		return nil, fmt.Errorf("unsupported file format")
	}
	
	if !format.SupportsRead() {
		return nil, fmt.Errorf("format %s does not support reading", format.Description())
	}
	
	reader, ok := fm.readers[format]
	if !ok {
		return nil, fmt.Errorf("no reader available for format %s", format.Description())
	}
	
	result, err := reader.Read(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %v", format.Description(), err)
	}
	
	result.Format = format
	return result, nil
}

// Save saves workbook data to a file
func (fm *FileManager) Save(filename string, sheets []SheetInfo, activeSheet int) error {
	format, ok := DetectFormat(filename)
	if !ok {
		return fmt.Errorf("unsupported file format")
	}
	
	if !format.SupportsWrite() {
		return fmt.Errorf("format %s does not support writing", format.Description())
	}
	
	writer, ok := fm.writers[format]
	if !ok {
		return fmt.Errorf("no writer available for format %s", format.Description())
	}
	
	if !format.SupportsMultipleSheets() && len(sheets) > 1 {
		if activeSheet >= 0 && activeSheet < len(sheets) {
			sheets = []SheetInfo{sheets[activeSheet]}
		} else {
			sheets = []SheetInfo{sheets[0]}
		}
		activeSheet = 0
	}
	
	dir := filename[:len(filename)-len(format.String())]
	if idx := len(dir) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if dir[i] == '/' || dir[i] == '\\' {
				dir = dir[:i]
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory: %v", err)
				}
				break
			}
		}
	}
	
	if err := writer.Write(filename, sheets, activeSheet); err != nil {
		return fmt.Errorf("failed to write %s file: %v", format.Description(), err)
	}
	
	return nil
}

// SaveAs saves workbook with format conversion
func (fm *FileManager) SaveAs(filename string, sheets []SheetInfo, activeSheet int, targetFormat FileFormat) error {
	if len(filename) < len(targetFormat.String()) ||
		filename[len(filename)-len(targetFormat.String()):] != targetFormat.String() {
		filename = filename + targetFormat.String()
	}
	
	return fm.Save(filename, sheets, activeSheet)
}

// CanOpen checks if a file can be opened
func (fm *FileManager) CanOpen(filename string) bool {
	format, ok := DetectFormat(filename)
	if !ok {
		return false
	}
	
	if !format.SupportsRead() {
		return false
	}
	
	_, hasReader := fm.readers[format]
	return hasReader
}

// CanSave checks if a file can be saved in the given format
func (fm *FileManager) CanSave(format FileFormat) bool {
	if !format.SupportsWrite() {
		return false
	}
	
	_, hasWriter := fm.writers[format]
	return hasWriter
}

// GetFormatInfo returns information about a file's format
func (fm *FileManager) GetFormatInfo(filename string) (FileFormat, error) {
	format, ok := DetectFormat(filename)
	if !ok {
		return FormatGSheet, fmt.Errorf("unknown file format")
	}
	return format, nil
}

// Global file manager instance
var defaultManager *FileManager

// GetDefaultManager returns the default file manager
func GetDefaultManager() *FileManager {
	if defaultManager == nil {
		defaultManager = NewFileManager()
	}
	return defaultManager
}

// OpenWorkbook opens a workbook file (replaces old OpenWorkbook)
func OpenWorkbook(filename string) (*WorkbookResult, error) {
	return GetDefaultManager().Open(filename)
}

// SaveWorkbook saves a workbook (replaces old SaveWorkbook)
func SaveWorkbook(sheets []SheetInfo, activeSheet int, filename string) error {
	return GetDefaultManager().Save(filename, sheets, activeSheet)
}

// SaveWorkbookAs saves with explicit format
func SaveWorkbookAs(sheets []SheetInfo, activeSheet int, filename string, format FileFormat) error {
	return GetDefaultManager().SaveAs(filename, sheets, activeSheet, format)
}
