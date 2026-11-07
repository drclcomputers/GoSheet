// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package table

import (
	"fmt"
	"strconv"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/ui/datavalidation"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"

	"github.com/rivo/tview"
)

func currentCellRaw(data map[[2]int]*cell.Cell, row, col int32) (addr, raw string) {
	addr = utils.FormatCellRef(row, col)
	if data == nil {
		return addr, ""
	}
	if c := data[[2]int{int(row), int(col)}]; c != nil && c.RawValue != nil {
		raw = *c.RawValue
	}
	return addr, raw
}

func syncFormulaBar(table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	if table == nil || activeData == nil || activeViewport == nil {
		return
	}
	vr, vc := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(vr, vc)
	if row < 1 || col < 1 {
		sheetview.SetBar("", "")
		return
	}
	addr, raw := currentCellRaw(activeData, row, col)
	sheetview.SetBar(addr, raw)
}

func startQuickEdit(app *tview.Application, table *tview.Table, seed string, keepExisting bool) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	if activeData == nil || activeViewport == nil {
		return
	}
	vr, vc := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(vr, vc)
	if row < 1 || col < 1 {
		return
	}
	text := seed
	if keepExisting {
		_, text = currentCellRaw(activeData, row, col)
		if seed != "" {
			text += seed
		}
	}
	sheetview.BeginEdit(text, func(entered string) {
		commitQuickEdit(app, table, row, col, entered)
	})
}

func commitQuickEdit(app *tview.Application, table *tview.Table, row, col int32, text string) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	if activeData == nil || activeViewport == nil {
		return
	}

	key := [2]int{int(row), int(col)}
	c, exists := activeData[key]
	if !exists || c == nil {
		c = cell.NewCell(row, col, "")
		activeData[key] = c
	}
	if !c.HasFlag(cell.FlagEditable) {
		return
	}

	old := c.Clone()
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "=") || strings.HasPrefix(text, "$=") {
		text = cell.NormalizeFormula(text)
	}

	if valid, errMsg := datavalidation.CheckValidationRule(c, text); !valid {
		if c.Valrulemsg != nil && strings.TrimSpace(*c.Valrulemsg) != "" {
			errMsg = *c.Valrulemsg
		}
		rule := ""
		if c.Valrule != nil {
			rule = *c.Valrule
		}
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Validation failed\n\n%s\n\nRule: %s", errMsg, rule)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				sheetview.Show(app)
				syncFormulaBar(table)
			})
		uitheme.StyleModalWarn(modal, "Validation")
		app.SetRoot(modal, true).SetFocus(modal)
		return
	}

	if c.RawValue == nil {
		s := ""
		c.RawValue = &s
	}
	if c.Display == nil {
		s := ""
		c.Display = &s
	}
	*c.RawValue = text

	if c.IsFormula() {
		c.SetFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)
		_ = EvaluateCell(table, c)
	} else {
		c.ClearFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)
		c.DependsOn = nil
		c.RangeDeps = nil
		applyLiteral(c, text)
	}

	activeData[key] = c
	RecordCellEdit(table, row, col, old, c)
	paintCell(table, c)
	_ = RecalculateCell(table, c)
	MarkAsModified(table)
	syncFormulaBar(table)
}

func applyLiteral(c *cell.Cell, text string) {
	if c.Type == nil {
		t := "string"
		c.Type = &t
	}
	switch strings.ToLower(*c.Type) {
	case "number", "financial":
		normalized := strings.ReplaceAll(text, string(c.ThousandsSeparator), "")
		normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
		if val, err := strconv.ParseFloat(normalized, 64); err == nil {
			*c.RawValue = fmt.Sprintf("%v", val)
			formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			if strings.EqualFold(*c.Type, "financial") {
				formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
			}
			*c.Display = formatted
			return
		}
	}
	if val, err := strconv.ParseFloat(text, 64); err == nil && !strings.ContainsAny(text, " \t") {
		*c.Type = "number"
		*c.RawValue = fmt.Sprintf("%v", val)
		*c.Display = utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
		return
	}
	*c.Type = "string"
	*c.Display = text
}
