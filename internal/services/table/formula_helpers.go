// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// formula_helpers.go provide auxiliary functions for the formula engine

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

// Parses formula into a format usable by govaluate
func BuildEvaluableFormula(table *tview.Table, formula string, parameters map[string]any) (string, error) {
	expandedFormula, err := ExpandRangesInFormula(formula)
	if err != nil {
		return "", fmt.Errorf("range expansion error: %v", err)
	}

	tokens := ParseFormulaTokens(expandedFormula)
	var result strings.Builder

	for _, token := range tokens {
		switch token.Type {
		case TokenStringLiteral:
			paramName := fmt.Sprintf("STR_LITERAL_%d", len(parameters))
			parameters[paramName] = token.Value
			result.WriteString(paramName)

		case TokenCellRef:
			paramName := "CELL_" + token.Value

			c, err := GetCellByRef(table, token.Value)
			if err != nil {
				return "", err
			}

			if c.IsFormula() && !c.HasFlag(cell.FlagEvaluated) {
				if err := EvaluateCell(table, c); err != nil {
					return "", err
				}
			}

			val := strings.TrimSpace(*c.Display)
			val = strings.ReplaceAll(val, string(c.ThousandsSeparator), "")
			val = strings.TrimPrefix(val, string(c.FinancialSign))

			if num, err := strconv.ParseFloat(val, 64); err == nil {
				parameters[paramName] = num
			} else {
				parameters[paramName] = strings.TrimSpace(*c.Display)
			}

			result.WriteString(paramName)

		default:
			result.WriteString(token.Value)
		}
	}

	return result.String(), nil
}

// Turns a formula into tokens
func ParseFormulaTokens(formula string) []Token {
	var tokens []Token
	i := 0
	
	for i < len(formula) {
		ch := formula[i]
		
		if ch == '"' {
			j := i + 1
			escaped := false
			for j < len(formula) {
				if formula[j] == '\\' && !escaped {
					escaped = true
					j++
					continue
				}
				if formula[j] == '"' && !escaped {
					break
				}
				escaped = false
				j++
			}
			
			if j < len(formula) {
				content := formula[i+1:j]
				content = strings.ReplaceAll(content, `\"`, `"`)
				content = strings.ReplaceAll(content, `\\`, `\`)
				
				tokens = append(tokens, Token{
					Type:     TokenStringLiteral,
					Value:    content,
					Original: formula[i:j+1],
				})
				i = j + 1
				continue
			}
		}
		
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			j := i
			for j < len(formula) && ((formula[j] >= 'A' && formula[j] <= 'Z') || (formula[j] >= 'a' && formula[j] <= 'z')) {
				j++
			}
			digitStart := j
			for j < len(formula) && formula[j] >= '0' && formula[j] <= '9' {
				j++
			}
			
			if digitStart > i && j > digitStart {
				cellRef := formula[i:j]
				tokens = append(tokens, Token{
					Type:     TokenCellRef,
					Value:    strings.ToUpper(cellRef), 
					Original: cellRef,
				})
				i = j
				continue
			}
		}
		
		tokens = append(tokens, Token{
			Type:     TokenOther,
			Value:    strings.ToUpper(string(ch)), 
			Original: string(ch),
		})
		i++
	}
	
	return tokens
}

// Returns an array of all of the cells used in the formula
func ParseCellReferences(formula string) []*string {
	expandedFormula, err := ExpandRangesInFormula(formula)
	if err != nil {
		expandedFormula = formula
	}

	matches := cellRefRegex.FindAllString(expandedFormula, -1)
	seen := make(map[string]bool)
	var refs []*string

	for _, match := range matches {
		if !seen[match] {
			matchCopy := match
			refs = append(refs, &matchCopy)
			seen[match] = true
		}
	}

	return refs
}

// Initiates the circular dependency search process 
func checkCircularDependencyForNewFormula(table *tview.Table, c *cell.Cell, newFormula string) error {
	newRefs := ParseCellReferences(strings.ToUpper(newFormula))
	
	oldDependsOn := c.DependsOn
	c.DependsOn = newRefs
	
	visited := make(map[string]bool)
	if hasCircularDependency(table, c, visited) {
		c.DependsOn = oldDependsOn
		return fmt.Errorf("circular dependency detected")
	}
	
	c.DependsOn = oldDependsOn
	return nil
}

// Clears dependencies recursively
func clearOldDependencies(table *tview.Table, c *cell.Cell) {
	cellRef := utils.FormatCellRef(c.Row, c.Column)
	
	for _, oldRef := range c.DependsOn {
		oldDepCell, err := GetCellByRef(table, *oldRef)
		if err != nil {
			continue
		}
		
		oldDepCell.Dependents = removeFromSlice(oldDepCell.Dependents, cellRef)
	}
}

// ExpandRange converts a range like "A1:B4" into individual cell references
func ExpandRange(rangeStr string) ([]string, error) {
	matches := rangeRegex.FindStringSubmatch(strings.ToUpper(rangeStr))
	if matches == nil {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	startCol := matches[1]
	startRow := matches[2]
	endCol := matches[3]
	endRow := matches[4]

	startRowNum, err := strconv.Atoi(startRow)
	if err != nil {
		return nil, fmt.Errorf("invalid start row: %s", startRow)
	}

	endRowNum, err := strconv.Atoi(endRow)
	if err != nil {
		return nil, fmt.Errorf("invalid end row: %s", endRow)
	}

	startColNum := utils.ColumnNumber(startCol)
	endColNum := utils.ColumnNumber(endCol)

	if startRowNum > endRowNum {
		startRowNum, endRowNum = endRowNum, startRowNum
	}
	if startColNum > endColNum {
		startColNum, endColNum = endColNum, startColNum
	}

	var cells []string
	for r := startRowNum; r <= endRowNum; r++ {
		for c := startColNum; c <= endColNum; c++ {
			cellRef := fmt.Sprintf("%s%d", utils.ColumnName(int32(c)), r)
			cells = append(cells, cellRef)
		}
	}

	return cells, nil
}

// ExpandRangesInFormula replaces all ranges with comma-separated cell lists
func ExpandRangesInFormula(formula string) (string, error) {
	matches := rangeRegex.FindAllStringSubmatchIndex(formula, -1)
	if matches == nil {
		return formula, nil
	}

	result := formula
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		rangeStart := match[0]
		rangeEnd := match[1]
		rangeStr := formula[rangeStart:rangeEnd]

		cells, err := ExpandRange(rangeStr)
		if err != nil {
			return "", err
		}

		cellList := strings.Join(cells, ", ")
		result = result[:rangeStart] + cellList + result[rangeEnd:]
	}

	return result, nil
}

// Returns the cell based on its address
func GetCellByRef(table *tview.Table, ref string) (*cell.Cell, error) {
	activeData := GetActiveSheetData()
    if activeData == nil {
        return nil, fmt.Errorf("no active sheet")
    }

	row, col := utils.ParseCellRef(ref)
	
	key := [2]int{int(row), int(col)}
	
	if cellData, exists := activeData[key]; exists {
		return cellData, nil
	}
	
	newCell := cell.NewCell(row, col, "")
	activeData[key] = newCell
	
	return newCell, nil
}

// Returns cell's value
func GetCellValue(table *tview.Table, ref string) (float64, error) {
	c, err := GetCellByRef(table, ref)
	if err != nil {
		return 0, err
	}
	
	if c.Display == nil || c.RawValue == nil {
		return 0, nil
	}
	
	if c.IsFormula() && !c.HasFlag(cell.FlagEvaluated) {
		if err := EvaluateCell(table, c); err != nil {
			return 0, err
		}
	}
	
	val := strings.TrimSpace(*c.Display)
	val = strings.ReplaceAll(val, string(c.ThousandsSeparator), "")
	val = strings.TrimPrefix(val, string(c.FinancialSign))
	
	if num, err := strconv.ParseFloat(val, 64); err == nil {
		return num, nil
	}
	
	if !c.IsFormula() {
		rawVal := strings.TrimSpace(*c.RawValue)
		if num, err := strconv.ParseFloat(rawVal, 64); err == nil {
			return num, nil
		}
	}
	
	return 0, fmt.Errorf("cell %s does not contain a numeric value", ref)
}


// Helper function: checks if an item exists in a slice
func contains(slice []*string, item string) bool {
	for _, ptr := range slice {
		if ptr != nil && *ptr == item {
			return true
		}
	}
	return false
}

// Helper function: removes an item from a slice
func removeFromSlice(slice []*string, item string) []*string {
	result := make([]*string, 0, len(slice))
	for _, s := range slice {
		if *s != item {
			result = append(result, s)
		}
	}
	return result
}
