// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// util.go provides utility functions and constants for the spreadsheet application

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var (
	VER = "2.8.7"
	FILEVER = "2.0"

	DEFAULT_SHEET_NAME = "Sheet1"
	DEFAULT_FILE_NAME = "Workbook1"

	MAX_ROWS int32 = 1073741824 //2^30
	MAX_COLS int32 = 1048576 //2^20 - BGQCV
	
	DEFAULT_CELL_MIN_WIDTH int32 = 10
	DEFAULT_CELL_MAX_WIDTH int32 = 40
	
	DEFAULT_CELL_DECIMAL_POINTS int32 = 2
	DEFAULT_CELL_THOUSANDS_SEPARATOR = ','
	DEFAULT_CELL_DECIMAL_SEPARATOR = '.'
	DEFAULT_CELL_FINANCIAL_SIGN = '$'
	
	DEFAULT_VIEWPORT_COLS int32
	DEFAULT_VIEWPORT_ROWS int32

	DEFAULT_RECENT_FILES_NUMBER int = 10
)

type ColorRGB [3]uint8

// According to terminal dimensions, modifies the viewport
func UpdateNrCellsOnScrn(){
	width, height := GetTermDimension()
	DEFAULT_VIEWPORT_COLS = width/DEFAULT_CELL_MIN_WIDTH-2
	DEFAULT_VIEWPORT_ROWS = height-3
}

// Returns terminal dimensions
func GetTermDimension() (int32, int32){
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width = 80
		height = 24
	}
	return int32(width), int32(height)
}

// ColumnName converts a column number to its corresponding column name (A, B, ..., AA, AB, ...)
func ColumnName(col int32) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name
}

// ColumnNumber converts a column name to its corresponding column number
func ColumnNumber(name string) int {
    var col int = 0
    for i := 0; i < len(name); i++ {
        if name[i] < 'A' || name[i] > 'Z' {
            return 0 
        }
        col = col*26 + int(name[i]-'A'+1)
    }
    return col
}

// ParseCellRef converts "A1" to (row: 1, col: 1)
func ParseCellRef(ref string) (row int32, col int32) {
	ref = strings.TrimSpace(strings.ToUpper(ref))
	
	var letters string
	var numbers string
	
	for _, ch := range ref {
		if ch >= 'A' && ch <= 'Z' {
			letters += string(ch)
		} else if ch >= '0' && ch <= '9' {
			numbers += string(ch)
		}
	}
	
	col = int32(ColumnNumber(letters))
	rowAux, _ := strconv.Atoi(numbers)
	row = int32(rowAux)
	
	return row, col 
}

// FormatCellRef converts (row: 1, col: 1) to "A1"
func FormatCellRef(row, col int32) string {
	return fmt.Sprintf("%s%d", ColumnName(col), row)
}

// MinMax returns the min and max of 2 numbers. 
func MinMax(a, b int32) (int32, int32) {
	if a > b {
		return b, a
	}
	return a, b
}

func ConvertToInt(a, b int32) (int, int) {
	return int(a), int(b)
} 

func ConvertToInt32(a, b int) (int32, int32) {
	return int32(a), int32(b)
}

// Logs and other strings to be printed after program completion.
var TOBEPRINTED []string

var TypeOptions = []string{"String", "Number", "Financial", "DateTime"} // Data types
var AlignOptions = []string{"Left", "Center", "Right"} // Alignment options
var DateTimeFormats = []string{"auto", "date", "time", "datetime"} // Date formats

// Format path to look better
func PrettyPath(full string, mode string) string {
    home, _ := os.UserHomeDir()

    full = filepath.ToSlash(full)
    home = filepath.ToSlash(home)

    rel := full
    if strings.HasPrefix(full, home) {
        rel = strings.TrimPrefix(full, home+"/")
    }

    parts := strings.Split(rel, "/")
	
    if mode == "recentfiles" && len(parts) > 1 {
        parts = parts[:len(parts)-1]
    }

    return strings.Join(parts, " > ")
}
