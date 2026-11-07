// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package evaluatefuncs

import (
	"fmt"
	"strings"
)

func LookupFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"VLOOKUP":    vlookup,
		"HLOOKUP":    hlookup,
		"INDEX":      indexFn,
		"MATCH":      matchFn,
		"COUNTBLANK": countBlank,
		"COUNTA":     countA,
	}
}

func vlookup(args ...any) (any, error) {
	if err := validateArgs("VLOOKUP", args, 3, 4); err != nil {
		return nil, err
	}
	m, ok := args[1].(*Matrix)
	if !ok || m == nil || m.Cols == 0 {
		return nil, fmt.Errorf("VLOOKUP range must be a cell range")
	}
	col, err := toFloat(args[2])
	if err != nil {
		return nil, fmt.Errorf("VLOOKUP: %v", err)
	}
	idx := int(col)
	if idx < 1 || idx > m.Cols {
		return nil, fmt.Errorf("VLOOKUP: column index out of range")
	}
	exact := true
	if len(args) == 4 {
		exact, err = exactMatchFlag(args[3])
		if err != nil {
			return nil, fmt.Errorf("VLOOKUP: %v", err)
		}
	}
	row := findInColumn(m, 0, args[0], exact)
	if row < 0 {
		return nil, fmt.Errorf("#N/A")
	}
	return m.At(row, idx-1), nil
}

func hlookup(args ...any) (any, error) {
	if err := validateArgs("HLOOKUP", args, 3, 4); err != nil {
		return nil, err
	}
	m, ok := args[1].(*Matrix)
	if !ok || m == nil || m.Rows == 0 {
		return nil, fmt.Errorf("HLOOKUP range must be a cell range")
	}
	rowN, err := toFloat(args[2])
	if err != nil {
		return nil, fmt.Errorf("HLOOKUP: %v", err)
	}
	idx := int(rowN)
	if idx < 1 || idx > m.Rows {
		return nil, fmt.Errorf("HLOOKUP: row index out of range")
	}
	exact := true
	if len(args) == 4 {
		exact, err = exactMatchFlag(args[3])
		if err != nil {
			return nil, fmt.Errorf("HLOOKUP: %v", err)
		}
	}
	col := findInRow(m, 0, args[0], exact)
	if col < 0 {
		return nil, fmt.Errorf("#N/A")
	}
	return m.At(idx-1, col), nil
}

func indexFn(args ...any) (any, error) {
	if err := validateArgs("INDEX", args, 2, 3); err != nil {
		return nil, err
	}
	m, ok := args[0].(*Matrix)
	if !ok || m == nil {
		return nil, fmt.Errorf("INDEX range must be a cell range")
	}
	rowN, err := toFloat(args[1])
	if err != nil {
		return nil, fmt.Errorf("INDEX: %v", err)
	}
	row := int(rowN)
	col := 1
	if len(args) == 3 {
		colN, err := toFloat(args[2])
		if err != nil {
			return nil, fmt.Errorf("INDEX: %v", err)
		}
		col = int(colN)
	} else if m.Cols == 1 {
		col = 1
	} else if m.Rows == 1 {
		col = row
		row = 1
	} else {
		return nil, fmt.Errorf("INDEX: column index required")
	}
	if row < 1 || col < 1 || row > m.Rows || col > m.Cols {
		return nil, fmt.Errorf("INDEX: index out of range")
	}
	return m.At(row-1, col-1), nil
}

func matchFn(args ...any) (any, error) {
	if err := validateArgs("MATCH", args, 2, 3); err != nil {
		return nil, err
	}
	m, ok := args[1].(*Matrix)
	if !ok || m == nil {
		return nil, fmt.Errorf("MATCH lookup range must be a cell range")
	}
	if len(args) == 3 {
		flag, err := toFloat(args[2])
		if err != nil {
			return nil, fmt.Errorf("MATCH: %v", err)
		}
		if int(flag) != 0 {
			return nil, fmt.Errorf("MATCH: only exact match (0) is supported")
		}
	}
	if m.Rows > 1 && m.Cols > 1 {
		return nil, fmt.Errorf("MATCH range must be a single row or column")
	}
	if m.Cols == 1 {
		row := findInColumn(m, 0, args[0], true)
		if row < 0 {
			return nil, fmt.Errorf("#N/A")
		}
		return float64(row + 1), nil
	}
	col := findInRow(m, 0, args[0], true)
	if col < 0 {
		return nil, fmt.Errorf("#N/A")
	}
	return float64(col + 1), nil
}

func countA(args ...any) (any, error) {
	if err := validateArgs("COUNTA", args, 1, -1); err != nil {
		return nil, err
	}
	n := 0
	for _, a := range FlattenArgs(args) {
		if a == nil {
			continue
		}
		if s, ok := a.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		n++
	}
	return float64(n), nil
}

func countBlank(args ...any) (any, error) {
	if err := validateArgs("COUNTBLANK", args, 1, -1); err != nil {
		return nil, err
	}
	n := 0
	for _, a := range FlattenArgs(args) {
		if a == nil {
			n++
			continue
		}
		if s, ok := a.(string); ok && strings.TrimSpace(s) == "" {
			n++
		}
	}
	return float64(n), nil
}

// exactMatchFlag reads Excel's range-lookup argument.
// TRUE means approximate, FALSE means exact. Returned value is "exact?".
func exactMatchFlag(v any) (exact bool, err error) {
	switch t := v.(type) {
	case bool:
		return !t, nil
	case float64:
		return t == 0, nil
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		switch s {
		case "true", "1", "yes":
			return false, nil
		case "false", "0", "no":
			return true, nil
		}
	}
	return false, fmt.Errorf("expected TRUE or FALSE")
}

func findInColumn(m *Matrix, col int, key any, exact bool) int {
	if !exact {
		return findApproxColumn(m, col, key)
	}
	for r := 0; r < m.Rows; r++ {
		if valuesEqual(m.At(r, col), key) {
			return r
		}
	}
	return -1
}

func findInRow(m *Matrix, row int, key any, exact bool) int {
	if !exact {
		return -1
	}
	for c := 0; c < m.Cols; c++ {
		if valuesEqual(m.At(row, c), key) {
			return c
		}
	}
	return -1
}

func findApproxColumn(m *Matrix, col int, key any) int {
	best := -1
	keyNum, keyErr := toFloat(key)
	for r := 0; r < m.Rows; r++ {
		cell := m.At(r, col)
		if cell == nil {
			continue
		}
		if keyErr == nil {
			n, err := toFloat(cell)
			if err != nil {
				continue
			}
			if n == keyNum {
				return r
			}
			if n < keyNum {
				best = r
				continue
			}
			break
		}
		cmp := strings.Compare(strings.ToLower(toString(cell)), strings.ToLower(toString(key)))
		if cmp == 0 {
			return r
		}
		if cmp < 0 {
			best = r
			continue
		}
		break
	}
	return best
}

func valuesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	af, aerr := toFloat(a)
	bf, berr := toFloat(b)
	if aerr == nil && berr == nil {
		return af == bf
	}
	return strings.EqualFold(strings.TrimSpace(toString(a)), strings.TrimSpace(toString(b)))
}
