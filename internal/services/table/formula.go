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
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
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
	cellRef := qualifiedCell(c)

	if visited[cellRef] {
		return true
	}

	visited[cellRef] = true
	defer delete(visited, cellRef)

	for _, depRef := range c.DependsOn {
		if depRef == nil {
			continue
		}
		depCell, err := GetCellByRef(table, *depRef)
		if err != nil || depCell == nil {
			continue
		}
		if depCell == c {
			return true
		}
		if depCell.IsFormula() {
			if hasCircularDependency(table, depCell, visited) {
				return true
			}
		}
	}

	selfSheet := sheetNameOf(c)
	for _, rg := range c.RangeDeps {
		sh, r1, c1, r2, c2, ok := parseRangeDep(rg)
		if !ok {
			continue
		}
		if strings.EqualFold(sh, selfSheet) && c.Row >= r1 && c.Row <= r2 && c.Column >= c1 && c.Column <= c2 {
			return true
		}
		sheet := sheetByName(sh)
		if sheet == nil {
			continue
		}
		for coord, depCell := range sheet.Data {
			if depCell == nil || !depCell.IsFormula() {
				continue
			}
			row, col := int32(coord[0]), int32(coord[1])
			if row < r1 || row > r2 || col < c1 || col > c2 {
				continue
			}
			if hasCircularDependency(table, depCell, visited) {
				return true
			}
		}
	}

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

	restoreSheet := activateSheetOf(c)
	defer restoreSheet()

	fail := func(display string, err error) error {
		if c.Display == nil {
			s := ""
			c.Display = &s
		}
		*c.Display = display
		c.SetFlag(cell.FlagEvaluated)
		c.SetFlag(cell.FlagFormula)
		return err
	}

	pieces, err := parseFormula(c.GetFormulaExpression())
	if err != nil {
		return fail("#REF!", err)
	}

	cellDeps, rangeDeps := depsFromPieces(pieces)

	prevDeps, prevRanges := c.DependsOn, c.RangeDeps
	c.DependsOn = ptrSlice(cellDeps)
	c.RangeDeps = rangeDeps
	if hasCircularDependency(table, c, map[string]bool{}) {
		c.DependsOn, c.RangeDeps = prevDeps, prevRanges
		return fail("#CIRC!", fmt.Errorf("circular dependency detected"))
	}
	c.DependsOn, c.RangeDeps = prevDeps, prevRanges

	clearOldDependencies(table, c)
	c.DependsOn = ptrSlice(cellDeps)
	c.RangeDeps = append([]string(nil), rangeDeps...)

	owner := qualifiedCell(c)
	for _, ref := range cellDeps {
		depCell, err := GetCellByRef(table, ref)
		if err != nil {
			return fail("#REF!", err)
		}
		if !contains(depCell.Dependents, owner) {
			ownerCopy := owner
			depCell.Dependents = append(depCell.Dependents, &ownerCopy)
		}
	}

	var evalErr error
	eval := func(dep *cell.Cell) error {
		if dep == nil || !dep.IsFormula() || dep.HasFlag(cell.FlagEvaluated) {
			return nil
		}
		return EvaluateCell(table, dep)
	}
	exprText, params, err := bindPieces(pieces,
		func(a addr) (any, error) {
			dep, err := GetCellByRef(table, qualify(a))
			if err != nil {
				return nil, err
			}
			return cellValueForFormula(dep, eval)
		},
		func(a, b addr) (*evaluatefuncs.Matrix, error) {
			return readRange(a, b, eval)
		},
	)
	if err != nil {
		msg := "#REF!"
		if strings.Contains(err.Error(), "too large") {
			msg = "#REF!"
		}
		return fail(msg, err)
	}

	result, err := evaluateExpression(exprText, params)
	if err != nil {
		evalErr = err
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "#N/A"):
			return fail("#N/A", err)
		case strings.Contains(errMsg, "requires"), strings.Contains(errMsg, "argument"):
			return fail("#ARGS!", err)
		case strings.Contains(errMsg, "division by zero"):
			return fail("#DIV/0!", err)
		case strings.Contains(errMsg, "not found"), strings.Contains(errMsg, "sheet"):
			return fail("#REF!", err)
		case strings.Contains(errMsg, "unknown"), strings.Contains(errMsg, "unexpected"):
			return fail("#NAME?", err)
		default:
			return fail("#ERROR!", err)
		}
	}
	_ = evalErr

	if c.Display == nil {
		s := ""
		c.Display = &s
	}
	if c.Type == nil {
		t := "string"
		c.Type = &t
	}

	switch v := result.(type) {
	case float64:
		if *c.Type == "string" {
			*c.Type = "number"
		}
		formatted := utils.FormatWithCommas(v, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
		if *c.Type == "financial" {
			formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
		}
		*c.Display = formatted
	case int:
		if *c.Type == "string" {
			*c.Type = "number"
		}
		formatted := utils.FormatWithCommas(float64(v), c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
		if *c.Type == "financial" {
			formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
		}
		*c.Display = formatted
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
		if result == nil {
			*c.Display = ""
		} else {
			*c.Type = "string"
			*c.Display = fmt.Sprintf("%v", result)
		}
	}

	c.SetFlag(cell.FlagEvaluated)
	c.SetFlag(cell.FlagFormula)
	return nil
}

func EvaluateAllFormulasOnLoad(table *tview.Table) error {
	if globalWorkbook == nil {
		return fmt.Errorf("no workbook loaded")
	}

	saved := globalWorkbook.ActiveSheet
	defer func() { globalWorkbook.ActiveSheet = saved }()

	for _, sheet := range globalWorkbook.Sheets {
		if sheet == nil {
			continue
		}
		for _, cellData := range sheet.Data {
			if cellData != nil && cellData.IsFormula() {
				cellData.ClearFlag(cell.FlagEvaluated)
			}
		}
	}

	for i, sheet := range globalWorkbook.Sheets {
		if sheet == nil {
			continue
		}
		globalWorkbook.ActiveSheet = i
		for _, cellData := range sheet.Data {
			if cellData != nil && cellData.IsFormula() && !cellData.HasFlag(cell.FlagEvaluated) {
				_ = EvaluateCell(table, cellData)
			}
		}
	}

	return nil
}

var (
	programCache = map[string]*vm.Program{}
	programMu    sync.Mutex
	baseOptions  []expr.Option
	optionsOnce  sync.Once
)

func formulaOptions() []expr.Option {
	optionsOnce.Do(func() {
		baseOptions = []expr.Option{
			expr.AllowUndefinedVariables(),
		}
		for name, fn := range evaluatefuncs.GovalFuncs() {
			baseOptions = append(baseOptions, expr.Function(name, fn))
		}
	})
	return baseOptions
}

func evaluateExpression(formula string, env map[string]any) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("runtime error: %v", r)
			result = nil
		}
	}()

	programMu.Lock()
	program := programCache[formula]
	programMu.Unlock()

	if program == nil {
		compiled, compileErr := expr.Compile(formula, formulaOptions()...)
		if compileErr != nil {
			return nil, fmt.Errorf("compile error: %v", compileErr)
		}
		program = compiled
		programMu.Lock()
		programCache[formula] = program
		programMu.Unlock()
	}

	result, err = expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}
	return result, nil
}

// Recalculates a formula
func RecalculateCell(table *tview.Table, c *cell.Cell) error {
	if c == nil {
		return nil
	}
	key := qualifiedCell(c)
	if recalcStack[key] {
		return nil
	}
	if recalcStack == nil {
		recalcStack = map[string]bool{}
	}
	recalcStack[key] = true
	defer delete(recalcStack, key)

	c.ClearFlag(cell.FlagEvaluated)

	if c.IsFormula() {
		if err := EvaluateCell(table, c); err != nil {
			paintCell(table, c)
			return err
		}
		paintCell(table, c)
	}

	for _, depRef := range c.Dependents {
		if depRef == nil {
			continue
		}
		depCell, err := GetCellByRef(table, *depRef)
		if err != nil || depCell == nil {
			continue
		}
		if depCell.IsFormula() {
			_ = RecalculateCell(table, depCell)
		}
	}

	recalcRangesTouching(table, sheetNameOf(c), c.Row, c.Column)
	return nil
}

var recalcStack map[string]bool

func paintCell(table *tview.Table, c *cell.Cell) {
	if table == nil || c == nil || globalWorkbook == nil {
		return
	}
	active := globalWorkbook.GetActiveSheet()
	if active == nil || active.Data[[2]int{int(c.Row), int(c.Column)}] != c {
		return
	}
	vp := active.Viewport
	if vp != nil && vp.IsVisible(c.Row, c.Column) {
		visualR, visualC := vp.ToRelative(c.Row, c.Column)
		table.SetCell(int(visualR), int(visualC), c.ToTViewCell())
	}
}

func recalcRangesTouching(table *tview.Table, sheet string, row, col int32) {
	if globalWorkbook == nil || sheet == "" {
		return
	}
	var pending []*cell.Cell
	for _, sh := range globalWorkbook.Sheets {
		if sh == nil {
			continue
		}
		for _, formulaCell := range sh.Data {
			if formulaCell == nil || !formulaCell.IsFormula() {
				continue
			}
			for _, rg := range formulaCell.RangeDeps {
				if rangeCovers(rg, sheet, row, col) {
					pending = append(pending, formulaCell)
					break
				}
			}
		}
	}
	for _, formulaCell := range pending {
		_ = RecalculateCell(table, formulaCell)
	}
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
