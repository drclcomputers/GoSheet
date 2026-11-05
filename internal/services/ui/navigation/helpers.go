// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpers.go provides some helper functions

package navigation

import (
	"gosheet/internal/services/cell"
	"strings"
)

var lastSearchTerm string
var lastSearchRow, lastSearchCol int32
var lastCaseSensitive bool
var lastMatchWholeWord bool

func SplitAlphaNumeric(s string) (letters, numbers string) {
	for i, ch := range s {
		if ch >= '0' && ch <= '9' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

func matchCellInData(globalData map[[2]int]*cell.Cell, row, col int32, searchTerm string, caseSensitive, matchWholeWord bool) bool {
	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		return false
	}

	cellValue := cellData.Display
	if cellValue == nil {
		cellValue = cellData.RawValue
	}
	if cellValue == nil {
		return false
	}

	if matchWholeWord {
		for word := range strings.FieldsSeq(*cellValue) {
			if caseSensitive {
				if word == searchTerm {
					return true
				}
			} else {
				if strings.EqualFold(word, searchTerm) {
					return true
				}
			}
		}
		return false
	}

	searchValue := *cellValue
	if !caseSensitive {
		searchValue = strings.ToLower(searchValue)
		searchTerm = strings.ToLower(searchTerm)
	}

	return strings.Contains(searchValue, searchTerm)
}
