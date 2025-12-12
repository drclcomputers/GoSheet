// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// models.go provides the structure for managing sheets

package sheetmanager

// SheetManagerCallbacks defines callbacks for sheet operations
type SheetManagerCallbacks struct {
	GetSheets          func() []SheetInfo
	GetActiveIndex     func() int
	GetWorkbookInfo    func() WorkbookInfo
	AddSheet           func(name string) error
	RenameSheet        func(index int, name string) error
	DeleteSheet        func(index int) error
	DuplicateSheet     func(index int) error
	MoveSheet          func(fromIndex, toIndex int) error
	SwitchToSheet      func(index int) error
	UpdateTabBar       func()
	UpdateTableTitle   func()
	MarkAsModified     func()
	RenderActiveSheet  func()
}

// SheetInfo contains information about a single sheet
type SheetInfo struct {
	Name      string
	CellCount int
	IsActive  bool
}

// WorkbookInfo contains information about the workbook
type WorkbookInfo struct {
	TotalSheets int
	ActiveSheet string
	TotalCells  int
	FileName    string
	HasChanges  bool
}


