// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// formula.go implements spreadsheet formula evaluation

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/rivo/tview"
)

var cellRefRegex = regexp.MustCompile(`\b([A-Z]+)(\d+)\b`)

// Token types for formula parsing
type TokenType int

const (
	TokenCellRef TokenType = iota
	TokenStringLiteral
	TokenOther
)

type Token struct {
	Type     TokenType
	Value    string
	Original string
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
	matches := cellRefRegex.FindAllString(formula, -1)
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

// Checks recursively for potential circular dependencies such as A1=A2+3 and A2=A1-2
func hasCircularDependency(table *tview.Table, c *cell.Cell, visited map[string]bool) bool {
	cellRef := utils.FormatCellRef(c.Row, c.Column)
	
	if visited[cellRef] {
		return true
	}
	
	visited[cellRef] = true
	
	for _, depRef := range c.DependsOn {
		depCell, err := GetCellByRef(table, *depRef)
		if err != nil {
			continue
		}
		
		if depCell.IsFormula() {
			if hasCircularDependency(table, depCell, visited) {
				return true
			}
		}
	}
	
	//delete(visited, cellRef)
	return false
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

// Parses formula into a format usable by govaluate
func BuildEvaluableFormula(table *tview.Table, formula string, parameters map[string]any) (string, error) {
	tokens := ParseFormulaTokens(formula)
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

// Evaluates cell using govaluate and sets the result as the cell display string
func EvaluateCell(table *tview.Table, c *cell.Cell) error {
	if !c.IsFormula() {
		return nil
	}
	
	if c.RawValue == nil {
		return fmt.Errorf("cell has nil RawValue")
	}
	
	if c.HasFlag(cell.FlagEvaluated) {
		return nil 
	}	
	
	formula := c.GetFormulaExpression()
	formulaUpper := strings.ToUpper(formula) 
	
	if err := checkCircularDependencyForNewFormula(table, c, formulaUpper); err != nil {
		*c.Display = "#CIRC!"
		c.SetFlag(cell.FlagEvaluated)
		return err
	}
	
	clearOldDependencies(table, c)
	
	refs := ParseCellReferences(formulaUpper)
	c.DependsOn = refs
	
	cellRef := utils.FormatCellRef(c.Row, c.Column)
	for _, ref := range refs {
		depCell, err := GetCellByRef(table, *ref)
		if err != nil {
			*c.Display = "#REF!"
			c.SetFlag(cell.FlagEvaluated)
			return err
		}
		
		if !contains(depCell.Dependents, cellRef) {
			depCell.Dependents = append(depCell.Dependents, &cellRef)
		}
	}
	
	parameters := make(map[string]any)
	evaluableFormula, err := BuildEvaluableFormula(table, formula, parameters)
	if err != nil {
		*c.Display = "#VALUE!"
		c.SetFlag(cell.FlagEvaluated)
		return err
	}
	
	result, err := evaluateExpression(evaluableFormula, parameters)
	if err != nil {
    	errMsg := err.Error()
    	if strings.Contains(errMsg, "requires") {
    	    *c.Display = "#ARGS!"
    	} else if strings.Contains(errMsg, "division by zero") {
    	    *c.Display = "#DIV/0!"
    	} else if strings.Contains(errMsg, "invalid") {
    	    *c.Display = "#VALUE!"
	    } else {
	        *c.Display = "#ERROR!"
	    }
	c.SetFlag(cell.FlagEvaluated)
    	return err
	}	
	
	switch v := result.(type) {
	case float64:
		if c.Type == nil || *c.Type == "string" {
			*c.Type = "number"
		}
		if *c.Type == "financial" {
			formatted := utils.FormatWithCommas(v, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			*c.Display = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
		} else {
			formatted := utils.FormatWithCommas(v, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			*c.Display = formatted
		}
	case int:
		if c.Type == nil || *c.Type == "string" {
			*c.Type = "number"
		}
		floatVal := float64(v)
		if *c.Type == "financial" {
			formatted := utils.FormatWithCommas(floatVal, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			*c.Display = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
		} else {
			formatted := utils.FormatWithCommas(floatVal, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			*c.Display = formatted
		}
	case string:
		if res, err := utils.ParseDateTime(v); err == nil {
			*c.Type = "datetime"
			*c.Display = utils.FormatDateTime(res, *c.DateTimeFormat)
		} else {
			*c.Type = "string"
			*c.Display = v
		}
	case bool:
		*c.Type = "string"
		if v {
			*c.Display = "TRUE"
		} else {
			*c.Display = "FALSE"
		}
	default:
		*c.Type = "string"
		*c.Display = fmt.Sprintf("%v", result)
	}
	
	c.SetFlag(cell.FlagEvaluated)
	c.SetFlag(cell.FlagFormula)
	
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

// Uses govaluate to return a result of the formula
func evaluateExpression(formula string, parameters map[string]any) (any, error) {
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(formula, utils.GovalFuncs())
	if err != nil {
		return nil, err
	}
	
	result, err := expr.Evaluate(parameters)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// Recalculates a formula
func RecalculateCell(table *tview.Table, c *cell.Cell) error {
	c.ClearFlag(cell.FlagEvaluated)
	
	if c.IsFormula() {
		if err := EvaluateCell(table, c); err != nil {
			return err
		}
		table.SetCell(int(c.Row), int(c.Column), c.ToTViewCell())
	}
	
	for _, depRef := range c.Dependents {
		depCell, err := GetCellByRef(table, *depRef)
		if err != nil {
			continue
		}
		
		if depCell.IsFormula() {
			if err := RecalculateCell(table, depCell); err != nil {
				continue
			}
		}
	}
	
	return nil
}

// Same as the previous, but for a series of cell.
func RecalculateAllFormulas(table *tview.Table) error {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return fmt.Errorf("no active sheet")
	}

    for _, cellData := range activeData {
        if cellData.IsFormula() {
            cellData.ClearFlag(cell.FlagEvaluated)
        }
    }
    
    for _, cellData := range activeData {
        if cellData.IsFormula() && !cellData.HasFlag(cell.FlagEvaluated) {
            if err := EvaluateCell(table, cellData); err != nil {
                continue
            }
            
            if activeViewport.IsVisible(cellData.Row, cellData.Column) {
                visualR, visualC := activeViewport.ToRelative(cellData.Row, cellData.Column)
                table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
            }
        }
    }
    
    return nil
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
