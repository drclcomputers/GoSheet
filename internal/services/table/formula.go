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
	"gosheet/internal/utils/evaluatefuncs"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/rivo/tview"
)

var cellRefRegex = regexp.MustCompile(`\b([A-Z]+)(\d+)\b`)
var rangeRegex = regexp.MustCompile(`\b([A-Z]+)(\d+):([A-Z]+)(\d+)\b`)

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
	
	delete(visited, cellRef)
	return false
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

	expandedFormula, err := ExpandRangesInFormula(formulaUpper)
	if err != nil {
		*c.Display = "#REF!"
		c.SetFlag(cell.FlagEvaluated)
		return err
	}

	if err := checkCircularDependencyForNewFormula(table, c, expandedFormula); err != nil {
		*c.Display = "#CIRC!"
		c.SetFlag(cell.FlagEvaluated)
		return err
	}

	clearOldDependencies(table, c)

	refs := ParseCellReferences(expandedFormula)
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

func EvaluateAllFormulasOnLoad(table *tview.Table) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	for _, sheet := range globalWorkbook.Sheets {
		if sheet == nil {
			continue
		}

		for _, cellData := range sheet.Data {
			if cellData.IsFormula() {
				cellData.ClearFlag(cell.FlagEvaluated)
			}
		}

		for _, cellData := range sheet.Data {
			if cellData.IsFormula() && !cellData.HasFlag(cell.FlagEvaluated) {
				if err := EvaluateCell(table, cellData); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

// Uses govaluate to return a result of the formula
func evaluateExpression(formula string, env map[string]any) (any, error) {
	functions := evaluatefuncs.GovalFuncs()
	
	options := []expr.Option{
		expr.Env(env),
		expr.AllowUndefinedVariables(),
	}
	
	for name, fn := range functions {
		options = append(options, expr.Function(name, fn))
	}

	program, err := expr.Compile(formula, options...)
	if err != nil {
		return nil, fmt.Errorf("compile error: %v", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
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
