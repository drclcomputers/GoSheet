// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// setvalue.go provides the functions which edit the cell properties according to the options in the edit cell dialog

package cellui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/ui/datavalidation"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

func updateCellValue(app *tview.Application, container *tview.Flex, c *cell.Cell, text string, leftForm *tview.Form) {
	text = strings.TrimSpace(text)
	
	if c.RawValue == nil {
		c.RawValue = new(string)
	}
	if c.Display == nil {
		c.Display = new(string)
	}
	if c.Type == nil {
		typeStr := "string"
		c.Type = &typeStr
	}
	
	if strings.HasPrefix(text, "$=") {
		*c.RawValue = text
		c.SetFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)
		*c.Display = text 
		return
	}
	
	if c.HasFlag(cell.FlagFormula) && !strings.HasPrefix(text, "$=") {
		c.ClearFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)
		c.DependsOn = nil
	}
	
	switch strings.ToLower(*c.Type) {
	case "string":
		*c.RawValue = text
		*c.Display = text
	case "number", "financial":
		normalized := strings.ReplaceAll(text, string(c.ThousandsSeparator), "")
		normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
		if val, err := strconv.ParseFloat(normalized, 64); err == nil {
			*c.RawValue = fmt.Sprintf("%v", val)
			formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
			if *c.Type == "financial" {
				formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
			}
			*c.Display = formatted
		}
	case "datetime", "date", "time":
		if isValid, _ := utils.IsValidDateTime(text); isValid {
			*c.RawValue = text
			*c.Display = text
		} else {
			ShowTypeErrorModal(app, container, c, leftForm)
		}
	}
}



func SaveCellFormButtonAndKeyMap(app *tview.Application, table *tview.Table, container *tview.Flex, c, oldCell *cell.Cell, row, column int32, leftForm *tview.Form, RecordCellEdit func(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error, RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	valueField := leftForm.GetFormItem(0).(*tview.InputField)
	currentValue := strings.TrimSpace(valueField.GetText())

	updateCellValue(app, container, c, currentValue, leftForm)

	if c.RawValue == nil {
		c.RawValue = new(string)
	}
	if c.Display == nil {
		c.Display = new(string)
	}
	if c.Type == nil {
		typeStr := "string"
		c.Type = &typeStr
	}
	if c.Valrule == nil {
		emptyStr := ""
		c.Valrule = &emptyStr
	}

	if !datavalidation.EnforceValidationOnEdit(app, container, c, currentValue) {
		return
	}

	rawValueCopy := currentValue
	c.RawValue = &rawValueCopy

	if c.Display == c.RawValue {
		displayCopy := currentValue
		c.Display = &displayCopy
	}

	if c.IsFormula() {
		c.SetFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)

		if err := EvaluateCell(table, c); err != nil {
			if c.Display != nil && strings.HasPrefix(*c.Display, "#") {
				ShowFormulaErrorModal(app, container, *c.Display, err.Error(), leftForm)
			} else {
				ShowTypeErrorModal(app, container, c, leftForm)
			}
		}
	} else {
		c.ClearFlag(cell.FlagFormula)
		c.ClearFlag(cell.FlagEvaluated)

		switch strings.ToLower(*c.Type) {
		case "number", "financial":
			normalized := strings.ReplaceAll(currentValue, string(c.ThousandsSeparator), "")
			normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
			if val, err := strconv.ParseFloat(normalized, 64); err == nil {
				formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
				if *c.Type == "financial" {
					formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
				}
				*c.Display = formatted
			} else {
				ShowTypeErrorModal(app, container, c, leftForm)
				return
			}
		case "datetime":
			if isValid, _ := utils.IsValidDateTime(currentValue); !isValid {
				ShowTypeErrorModal(app, container, c, leftForm)
				return
			}
			*c.Display = currentValue
		default:
			*c.Display = currentValue
		}
	}

	key := [2]int{int(row), int(column)}
	globalData[key] = c

	RecordCellEdit(table, row, column, oldCell, c)

	if globalViewport.IsVisible(row, column) {
		visualR, visualC := globalViewport.ToRelative(row, column)
		table.SetCell(int(visualR), int(visualC), c.ToTViewCell())
	}

	RecalculateCell(table, c)

	app.SetRoot(table, true).SetFocus(table)
}
