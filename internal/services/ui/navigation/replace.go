// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// replace.go provides the replace functions

package navigation

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strings"

	"github.com/rivo/tview"
)

func replaceInCell(app *tview.Application, table *tview.Table, row, col int32, findTerm, replaceTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) bool {
	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		return false
	}

	if !cellData.HasFlag(cell.FlagEditable) {
		ShowWarningModal(app, table, "Cell is not editable")
		return false
	}

	if !matchCellInData(globalData, row, col, findTerm, caseSensitive, matchWholeWord) {
		return false
	}

	cellValue := cellData.RawValue
	if cellValue == nil {
		return false
	}

	if matchWholeWord {
		aux := replaceWholeWord(*cellValue, findTerm, replaceTerm, caseSensitive)
		cellData.RawValue = &aux
	} else {
		if caseSensitive {
			aux := strings.ReplaceAll(*cellValue, findTerm, replaceTerm)
			cellData.RawValue = &aux
		} else {
			aux := replaceInsensitive(*cellValue, findTerm, replaceTerm)
			cellData.RawValue = &aux
		}
	}

	cellData.Display = cellData.RawValue
	globalData[key] = cellData

	if globalViewport.IsVisible(row, col) {
		visualR, visualC := globalViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}

	return true
}

func countMatches(globalData map[[2]int]*cell.Cell, searchTerm string, caseSensitive, matchWholeWord bool) int32 {
	count := int32(0)

	for _, cellData := range globalData {
		if matchCellInData(globalData, cellData.Row, cellData.Column, searchTerm, caseSensitive, matchWholeWord) {
			count++
		}
	}

	return count
}

func replaceAll(table *tview.Table, findTerm, replaceTerm string, caseSensitive, matchWholeWord bool, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport, RenderVisible func(table *tview.Table, globalViewport *utils.Viewport, globalData map[[2]int]*cell.Cell)) int {
	count := 0

	for key, cellData := range globalData {
		if !cellData.HasFlag(cell.FlagEditable) {
			continue
		}

		if matchCellInData(globalData, cellData.Row, cellData.Column, findTerm, caseSensitive, matchWholeWord) {
			if cellData.RawValue != nil {
				if matchWholeWord {
					*cellData.RawValue = replaceWholeWord(*cellData.RawValue, findTerm, replaceTerm, caseSensitive)
				} else {
					if caseSensitive {
						*cellData.RawValue = strings.ReplaceAll(*cellData.RawValue, findTerm, replaceTerm)
					} else {
						*cellData.RawValue = replaceInsensitive(*cellData.RawValue, findTerm, replaceTerm)
					}
				}
				cellData.Display = cellData.RawValue
				globalData[key] = cellData
				count++
			}
		}
	}

	RenderVisible(table, globalViewport, globalData)

	return count
}

func replaceInsensitive(s, old, new string) string {
	lowerOld := strings.ToLower(old)

	result := ""
	i := 0
	for i < len(s) {
		if i+len(old) <= len(s) && strings.ToLower(s[i:i+len(old)]) == lowerOld {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

func replaceWholeWord(s, old, new string, caseSensitive bool) string {
	words := strings.Fields(s)
	result := make([]string, len(words))

	for i, word := range words {
		if caseSensitive {
			if word == old {
				result[i] = new
			} else {
				result[i] = word
			}
		} else {
			if strings.EqualFold(word, old) {
				result[i] = new
			} else {
				result[i] = word
			}
		}
	}

	return strings.Join(result, " ")
}
