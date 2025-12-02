// Copyright (c) 2025 @drclcomputers. All rights reserved.
// Licensed under the MIT license.
// See <https://opensource.org/licenses/MIT>.

// cellUI.go provides functions to display cell editing dialogs in the UI.

package ui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"


	"github.com/gdamore/tcell/v2"
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
	colorIndex, bgIndex := 10, 10 

	for i, name := range utils.ColorOptionNames {
		if c.Color == utils.ColorOptions[name] {
			colorIndex = i
		}
		if c.BgColor == utils.ColorOptions[name] {
			bgIndex = i
		}
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

func newDropdown(label string, options []string, current int32, onSelect func(string)) *tview.DropDown {
	return tview.NewDropDown().
		SetLabel(label).
		SetOptions(options, func(opt string, _ int) { onSelect(opt) }).
		SetCurrentOption(int(current))
}

// Core Functionality
func EditCellDialog(app *tview.Application, table *tview.Table, row, column int32, RecordCellEdit func(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error, RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	key := [2]int{int(row), int(column)}
	c, exists := globalData[key]
	if !exists {
    	c = cell.NewCell(row, column, "")
    	globalData[key] = c
	}	

	oldCell := c.Clone()

	if c.RawValue == nil {
		emptyStr := ""
		c.RawValue = &emptyStr
	}
	if c.Display == nil {
		emptyStr := ""
		c.Display = &emptyStr
	}
	if c.Type == nil {

		typeStr := "string"
		c.Type = &typeStr
	}
	if c.Notes == nil {
		emptyStr := ""
		c.Notes = &emptyStr
	}
	if c.Valrule == nil {
		emptyStr := ""
		c.Valrule = &emptyStr
	}

	*c.RawValue = cell.StripTviewTags(strings.TrimSpace(*c.RawValue))
	*c.Display = cell.StripTviewTags(strings.TrimSpace(*c.Display))	

	typeIndex := getTypeIndex(*c.Type)
	alignIndex := getAlignIndex(c.Align)
	colorIndex, bgColorIndex := getColorIndices(c)
	//dateTypeFormatIndex := getDateTypeFormat(c)

	financialSigns := toStringSlice(utils.FinancialSigns)
	separators := toStringSlice(utils.Separators)
	dateTimeFormats := utils.DateTimeFormats

	financialSignDropdown := newDropdown("Sign: ", financialSigns,
    findRuneIndex(financialSigns, c.FinancialSign),
    func(opt string) { 
        c.FinancialSign = []rune(opt)[0]
        if *c.Type == "financial" {
            normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
            normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
            if val, err := strconv.ParseFloat(normalized, 64); err == nil {
                formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
                *c.Display = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
            }
        }
    })

	thousandsSeparatorDropdown := newDropdown("Thousands: ", separators,
    	findRuneIndex(separators, c.ThousandsSeparator),
    	func(opt string) {
    	    c.ThousandsSeparator = []rune(opt)[0]
    	    if c.ThousandsSeparator == 'Ø' { c.ThousandsSeparator = 0 }
    	    if *c.Type == "number" || *c.Type == "financial" {
    	        normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
    	        normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
            if val, err := strconv.ParseFloat(normalized, 64); err == nil {
                formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
                if *c.Type == "financial" {
                    formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
                }
                *c.Display = formatted
            }
        }
    })

	decimalSeparatorDropdown := newDropdown("Decimal: ", separators,
    	findRuneIndex(separators, c.DecimalSeparator),
    	func(opt string) { 
    	    c.DecimalSeparator = []rune(opt)[0]
    	    if c.DecimalSeparator == 'Ø' { c.DecimalSeparator = 0 }
    	    if *c.Type == "number" || *c.Type == "financial" {
    	        normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
    	        normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
            if val, err := strconv.ParseFloat(normalized, 64); err == nil {
                formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
                if *c.Type == "financial" {
                    formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
                }
                *c.Display = formatted
            }
        }
    })

	dateTimeFormatDropdown := newDropdown("Date format", dateTimeFormats,
		int32(getDateTypeFormat(*c.DateTimeFormat)),
		func (opt string) {
	        *c.DateTimeFormat = utils.DateTimeFormats[getDateTypeFormat(opt)]
	        
	        if c.RawValue != nil && *c.RawValue != "" {
	            if t, err := utils.ParseDateTime(*c.RawValue); err == nil {
	                *c.Display = utils.FormatDateTime(t, *c.DateTimeFormat)
	            }
	        }
	    })

	decimalPointsInput := tview.NewInputField().
    	SetLabel("Decimals: ").
    	SetText(fmt.Sprintf("%d", c.DecimalPoints)).
    	SetFieldWidth(4).
    	SetAcceptanceFunc(tview.InputFieldInteger).
    	SetChangedFunc(func(text string) {
    	    if points, err := strconv.Atoi(text); err == nil && points >= 0 {
    	        c.DecimalPoints = int32(points)
    	        
    	        if *c.Type == "number" || *c.Type == "financial" {
    	            normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
    	            normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
    	            if val, err := strconv.ParseFloat(normalized, 64); err == nil {
    	                formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
    	                if *c.Type == "financial" {
    	                    formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
    	                }
    	                *c.Display = formatted
    	            }
    	        }
    	    }
    	})

	disableFormattingFields(financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput)

	editCellDialog, leftForm := buildEditCellForm(app, table, c, oldCell, row, column,
		financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput,
		dateTimeFormatDropdown, typeIndex, alignIndex, colorIndex, bgColorIndex, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)

	app.SetRoot(editCellDialog, true).SetFocus(leftForm)
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

// Form Builder
func buildEditCellForm(app *tview.Application, table *tview.Table, c, oldCell *cell.Cell, row, column int32,
	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown *tview.DropDown,
	decimalPointsInput *tview.InputField, dateTimeFormatDropdown *tview.DropDown, typeIndex, alignIndex, colorIndex, bgColorIndex int, RecordCellEdit func(table *tview.Table, row, col int32, oldCell, newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error, RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) (*tview.Flex, *tview.Form) {

	container := tview.NewFlex()

	// Left Column - Content & Type
	leftForm := tview.NewForm()
	
	rawValueStr := safeStringValue(c.RawValue)
	
	leftForm.AddInputField("Value", rawValueStr, 0, nil, func(text string) {
		//updateCellValue(app, container, c, text, leftForm)
	})
	
	leftForm.AddDropDown("Type", utils.TypeOptions, typeIndex, func(option string, _ int) {
		newType := strings.ToLower(option)
		
		if c.Type == nil {
			c.Type = new(string)
		}
		*c.Type = newType
		
		setFormattingEnabled(*c.Type, financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown,dateTimeFormatDropdown, decimalPointsInput)
	})
	
	leftForm.AddDropDown("Align", utils.AlignOptions, alignIndex, func(option string, _ int) {
		switch strings.ToLower(option) {
		case "left":
			c.Align = tview.AlignLeft
		case "center":
			c.Align = tview.AlignCenter
		case "right":
			c.Align = tview.AlignRight
		}
	})
	
	leftForm.AddInputField("Min Width", fmt.Sprintf("%d", c.MinWidth), 5, nil, func(text string) {
		if width, err := strconv.Atoi(text); err == nil && width > 0 {
			c.MinWidth = int16(width)
		}
	})
	
	leftForm.AddInputField("Max Width", fmt.Sprintf("%d", c.MaxWidth), 5, nil, func(text string) {
		if width, err := strconv.Atoi(text); err == nil && width > 0 {
			c.MaxWidth = int16(width)
		}
	})
	
	leftForm.SetBorder(true).SetTitle(" Content ").SetTitleAlign(tview.AlignLeft)

	formatForm := tview.NewForm()
	formatForm.AddFormItem(financialSignDropdown).
		AddFormItem(thousandsSeparatorDropdown).
		AddFormItem(decimalSeparatorDropdown).
		AddFormItem(decimalPointsInput).
		AddFormItem(dateTimeFormatDropdown)
	formatForm.SetBorder(true).SetTitle(" Formatting ").SetTitleAlign(tview.AlignLeft)

	// Right Column - Styling
	rightForm := tview.NewForm()
	rightForm.AddCheckbox("Bold", c.HasFlag(cell.FlagBold), func(checked bool) { c.SetFlagState(cell.FlagBold, checked) })
	rightForm.AddCheckbox("Italic", c.HasFlag(cell.FlagItalic), func(checked bool) { c.SetFlagState(cell.FlagItalic, checked) })
	rightForm.AddCheckbox("Underline", c.HasFlag(cell.FlagUnderline), func(checked bool) { c.SetFlagState(cell.FlagUnderline, checked) })
	rightForm.AddCheckbox("All Caps", c.HasFlag(cell.FlagAllCaps), func(checked bool) { c.SetFlagState(cell.FlagAllCaps, checked) })
	rightForm.AddCheckbox("Strikethrough", c.HasFlag(cell.FlagStrikethrough), func(checked bool) { c.SetFlagState(cell.FlagStrikethrough, checked) })
	rightForm.AddCheckbox("Editable", c.HasFlag(cell.FlagEditable), func(checked bool) { c.SetFlagState(cell.FlagEditable, checked) })
	//rightForm.AddCheckbox("Formula", c.Formula, func(checked bool) { c.Formula = checked })
	rightForm.AddButton("Data Validation", func() { ShowValidationRuleDialog(app, table, container, rightForm.GetFormItem(0), globalData, globalViewport) })
	rightForm.SetBorder(true).SetTitle(" Styling ").SetTitleAlign(tview.AlignLeft)

	// Color Form
	colorForm := tview.NewForm()

	colorForm.AddDropDown("Text", utils.ColorOptionNames, colorIndex, func(option string, index int) {
		if option == "Custom..." {
			showCustomColorPicker(app, container, c, true, colorForm)
			return
		}
		c.Color = utils.ColorOptions[option]
	})
	colorForm.AddButton("Custom Text Color", func() {
		showCustomColorPicker(app, container, c, true, colorForm)
	})

	colorForm.AddDropDown("Background", utils.ColorOptionNames, bgColorIndex, func(option string, index int) {
		if option == "Custom..." {
			showCustomColorPicker(app, container, c, false, colorForm)
			return
		}
		c.BgColor = utils.ColorOptions[option]
	})
	colorForm.AddButton("Custom BG Color", func() {
		showCustomColorPicker(app, container, c, false, colorForm)
	})
	colorForm.SetBorder(true).SetTitle(" Colors ").SetTitleAlign(tview.AlignLeft)	

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(leftForm, 0, 1, false).
		AddItem(formatForm, 0, 1, false)

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rightForm, 0, 3, false).
		AddItem(colorForm, 0, 2, false)

	columns := tview.NewFlex().
		AddItem(leftColumn, 0, 1, false).
		AddItem(rightColumn, 0, 1, false)

	buttonBar := tview.NewFlex().SetDirection(tview.FlexRow)
	buttonText := tview.NewTextView().
		SetText("  [yellow::b]Ctrl+→/←[::-] Switch Section   [yellow::b]Tab/Shift+Tab[::-] Next/Previous Field   [yellow::b]Alt+S[::-] Save   [yellow::b]ESC/Q[::-] Cancel").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	buttonBar.AddItem(buttonText, 1, 0, false)

	container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(columns, 0, 1, true).
		AddItem(buttonBar, 1, 0, false)

	container.SetBorder(true).
		SetTitle(fmt.Sprintf(" Edit Cell %s%d ", utils.ColumnName(column), row)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorYellow)

	allForms := []*tview.Form{leftForm, formatForm, rightForm, colorForm}
	currentFormIndex := 0

	app.SetFocus(leftForm)

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEscape || event.Rune() == 'q' || event.Rune() == 'Q':
			app.SetRoot(table, true).SetFocus(table)
			return nil
		case event.Modifiers()&tcell.ModAlt != 0 && (event.Rune() == 's' || event.Rune() == 'S'):
			SaveCellFormButtonAndKeyMap(app, table, container, c, oldCell, row, column, leftForm, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)
			return nil
		case event.Key() == tcell.KeyRight && event.Modifiers()&tcell.ModCtrl != 0:
			currentFormIndex = (currentFormIndex + 1) % len(allForms)
			app.SetFocus(allForms[currentFormIndex])
			return nil
		case event.Key() == tcell.KeyLeft && event.Modifiers()&tcell.ModCtrl != 0:
			currentFormIndex = (currentFormIndex - 1 + len(allForms)) % len(allForms)
			app.SetFocus(allForms[currentFormIndex])
			return nil
		}
		return event
	})

	return container, leftForm
}

func setFormattingEnabled(cellType string,
	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown *tview.DropDown,
	decimalPointsInput *tview.InputField) {

	enable := func(p tview.Primitive, enabled bool) {
		switch v := p.(type) {
		case *tview.DropDown:
			v.SetDisabled(!enabled)
		case *tview.InputField:
			v.SetDisabled(!enabled)
		}
	}

	switch cellType {
	case "financial":
		enable(financialSignDropdown, true)
		enable(thousandsSeparatorDropdown, true)
		enable(decimalSeparatorDropdown, true)
		enable(decimalPointsInput, true)
		enable(dateTimeFormatDropdown, false)
	case "number":
		enable(financialSignDropdown, false) 
		enable(thousandsSeparatorDropdown, true)
		enable(decimalSeparatorDropdown, true)
		enable(decimalPointsInput, true)
		enable(dateTimeFormatDropdown, false)
	case "datetime":
		enable(financialSignDropdown, false)
		enable(thousandsSeparatorDropdown, false)
		enable(decimalSeparatorDropdown, false)
		enable(decimalPointsInput, false)
		enable(dateTimeFormatDropdown, true)
	default:
		enable(financialSignDropdown, false)
		enable(thousandsSeparatorDropdown, false)
		enable(decimalSeparatorDropdown, false)
		enable(decimalPointsInput, false)
		enable(dateTimeFormatDropdown, false)
	}
}

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
			// You could add a DateTimeFormat field to cell.Cell
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

	if !EnforceValidationOnEdit(app, container, c, currentValue) {
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
