// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// validation implements cell validation rules

package datavalidation

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils/evaluatefuncs"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ValidateValidationRule checks if a validation rule is syntactically correct
func ValidateValidationRule(ruleText string, cellData *cell.Cell) bool {
	if strings.TrimSpace(ruleText) == "" {
		return true
	}

	upperRule := strings.ToUpper(ruleText)
	matches := validationCellRefRegex.FindAllString(upperRule, -1)

	for _, match := range matches {
		if match != "THIS" {
			return false
		}
	}

	testRule := strings.ReplaceAll(upperRule, "THIS", "5")

	functions := evaluatefuncs.GovalFuncs()

	env := make(map[string]any)
	for name, fn := range functions {
		env[name] = fn
	}

	options := []expr.Option{
		expr.AllowUndefinedVariables(),
	}
	
	for name, fn := range functions {
		options = append(options, expr.Function(name, fn))
	}

	program, err := expr.Compile(testRule, options...)
	if err != nil {
		return false 
	}

	_, err = expr.Run(program, env)
	if err != nil {
		return false 
	}	

	return true
}

// EnforceValidationOnEdit checks validation before saving a cell edit
func EnforceValidationOnEdit(app *tview.Application, returnTo tview.Primitive, cellData *cell.Cell, newValue string) bool {
	if strings.TrimSpace(newValue) == "" {
		return true
	}

	isValid, errMsg := CheckValidationRule(cellData, newValue)

	if !isValid {
		displayMsg := errMsg
		if cellData.Valrulemsg != nil && strings.TrimSpace(*cellData.Valrulemsg) != "" {
			displayMsg = *cellData.Valrulemsg
		}

		modal := tview.NewModal().
			SetText(fmt.Sprintf("Validation Failed!\n\n%s\n\nValidation Rule:\n%s", displayMsg, *cellData.Valrule)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.SetRoot(returnTo, true).SetFocus(returnTo)
			})

		modal.SetBackgroundColor(tcell.ColorDarkRed).
			SetBorderColor(tcell.ColorRed)
		modal.SetButtonBackgroundColor(tcell.ColorDarkRed).
			SetButtonTextColor(tcell.ColorWhite)

		app.SetRoot(modal, true).SetFocus(modal)
		return false
	}

	return true
}

func CheckValidationRule(cellData *cell.Cell, newValue string) (bool, string) {
	if cellData.Valrule == nil || strings.TrimSpace(*cellData.Valrule) == "" {
		return true, ""
	}

	if strings.TrimSpace(newValue) == "" {
		return true, ""
	}

	var testValue any

	cellDataTypeAux := *cellData.Type
	switch strings.ToLower(cellDataTypeAux) {
	case "number", "financial":
		normalized := strings.ReplaceAll(newValue, string(cellData.ThousandsSeparator), "")
		normalized = strings.TrimPrefix(normalized, string(cellData.FinancialSign))

		if num, err := strconv.ParseFloat(normalized, 64); err == nil {
			testValue = num
		} else {
			return false, "Value must be a number"
		}

	case "string":
		testValue = newValue

	default:
		testValue = newValue
	}

	rule := strings.TrimSpace(*cellData.Valrule)
	upperRule := strings.ToUpper(rule)

	var replacementValue string
	if _, ok := testValue.(float64); ok {
		replacementValue = fmt.Sprintf("%v", testValue)
	} else {
		strValue := fmt.Sprintf("%v", testValue)
		strValue = strings.ReplaceAll(strValue, `"`, `\"`)
		replacementValue = fmt.Sprintf(`"%s"`, strValue)
	}

	evaluableRule := strings.ReplaceAll(upperRule, "THIS", replacementValue)

	functions := evaluatefuncs.GovalFuncs()

	env := make(map[string]any)
	for name, fn := range functions {
		env[name] = fn
	}

	options := []expr.Option{
		expr.AllowUndefinedVariables(),
	}

	program, err := expr.Compile(evaluableRule, options...)
	if err != nil {
		return false, fmt.Sprintf("Invalid validation rule: %s", err.Error())
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false, "Could not compile validation rule"
	}	

	isValid, ok := result.(bool)
	if !ok {
		return false, "Validation rule must return true/false"
	}

	if !isValid {
		if cellData.Valrulemsg != nil && strings.TrimSpace(*cellData.Valrulemsg) != "" {
			return false, *cellData.Valrulemsg
		}
		return false, fmt.Sprintf("Value does not meet validation rule: %s", *cellData.Valrule)
	}

	return true, ""
}
