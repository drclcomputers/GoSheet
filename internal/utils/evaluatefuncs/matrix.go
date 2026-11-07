// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package evaluatefuncs

import "strings"

// Matrix is a rectangular block of cell values passed to functions as one
// argument. Nil entries are blank cells. Ranges are not expanded into
// thousands of scalar arguments.
type Matrix struct {
	Rows  int
	Cols  int
	Cells []any
}

// At returns the value at a 0-based position.
func (m *Matrix) At(r, c int) any {
	if m == nil || r < 0 || c < 0 || r >= m.Rows || c >= m.Cols {
		return nil
	}
	return m.Cells[r*m.Cols+c]
}

// FlattenArgs expands matrices and nested slices into a single list.
func FlattenArgs(args []any) []any {
	out := make([]any, 0, len(args))
	var walk func(any)
	walk = func(a any) {
		switch v := a.(type) {
		case *Matrix:
			if v == nil {
				return
			}
			for _, c := range v.Cells {
				walk(c)
			}
		case []any:
			for _, c := range v {
				walk(c)
			}
		default:
			out = append(out, a)
		}
	}
	for _, a := range args {
		walk(a)
	}
	return out
}

// NumericValues returns numbers found in args. Text and blanks are skipped,
// matching spreadsheet aggregates (SUM/AVERAGE/MIN/MAX).
func NumericValues(args []any) []float64 {
	flat := FlattenArgs(args)
	nums := make([]float64, 0, len(flat))
	for _, a := range flat {
		if a == nil {
			continue
		}
		if s, ok := a.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		f, err := toFloat(a)
		if err != nil {
			continue
		}
		nums = append(nums, f)
	}
	return nums
}
