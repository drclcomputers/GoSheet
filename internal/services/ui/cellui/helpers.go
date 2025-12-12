// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpers.go provides helper functions for the cellui package

package cellui

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strings"

	"github.com/rivo/tview"
)

// Utility Functions
func getTypeIndex(cellType string) int {
	switch strings.TrimSpace(strings.ToLower(cellType)) {
	case "number": return 1
	case "financial": return 2
	case "datetime": return 3
	default: return 0
	}
}

func getAlignIndex(align int8) int {
	switch align {
	case tview.AlignCenter: return 1
	case tview.AlignRight: return 2
	default: return 0
	}
}

func getColorIndices(c *cell.Cell) (int, int) {
	colorIndex, bgIndex := -1, -1

	for i, name := range utils.ColorOptionNames {
		if name == "Custom..." {
			continue
		}
		if colorIndex == -1 && c.Color == utils.ColorOptions[name] {
			colorIndex = i
		}
		if bgIndex == -1 && c.BgColor == utils.ColorOptions[name] {
			bgIndex = i
		}
		if colorIndex != -1 && bgIndex != -1 {
			break
		}
	}

	if colorIndex == -1 {
		colorIndex = 0
	}
	if bgIndex == -1 {
		bgIndex = 0
	}

	return colorIndex, bgIndex
}

func getDateTypeFormat(opt string) int {
	switch opt {
	case "date": return 1
	case "time": return 2
	case "datetime": return 3
	default: return 0
	}
}

func findRuneIndex(slice []string, target rune) int32 {
	for i, s := range slice {
		if []rune(s)[0] == target {
			return int32(i)
		}
	}
	return 0
}

func safeStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func disableFormattingFields(items ...tview.Primitive) {
	for _, item := range items {
		if d, ok := item.(interface{ SetDisabled(bool) *tview.DropDown }); ok { d.SetDisabled(true) }
	}
}

func toStringSlice(runes []rune) []string {
	out := make([]string, len(runes))
	for i, r := range runes { out[i] = string(r) }
	return out
}


