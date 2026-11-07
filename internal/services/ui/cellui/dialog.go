// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dialog.go provides the main edit cell dialog

package cellui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/ui/datavalidation"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown, decimalPointsInput := getDropdowns(c)

	disableFormattingFields(financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput)

	editCellDialog, leftForm := buildEditCellForm(app, table, c, oldCell, row, column,
		financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput,
		dateTimeFormatDropdown, typeIndex, alignIndex, colorIndex, bgColorIndex, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)

	app.SetRoot(editCellDialog, true).SetFocus(leftForm)
}

// Form Builder
func buildEditCellForm(app *tview.Application, table *tview.Table, c, oldCell *cell.Cell, row, column int32,
	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown *tview.DropDown,
	decimalPointsInput *tview.InputField, dateTimeFormatDropdown *tview.DropDown, typeIndex, alignIndex,
	colorIndex, bgColorIndex int, RecordCellEdit func(table *tview.Table, row, col int32, oldCell,
		newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error,
	RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell,
	globalViewport *utils.Viewport) (*tview.Flex, *tview.Form) {

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

		setFormattingEnabled(*c.Type, financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown, decimalPointsInput)
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
			c.MinWidth = int32(width)
		}
	})

	leftForm.AddInputField("Max Width", fmt.Sprintf("%d", c.MaxWidth), 5, nil, func(text string) {
		if width, err := strconv.Atoi(text); err == nil && width > 0 {
			c.MaxWidth = int32(width)
		}
	})

	styleCellForm(leftForm, "Content")

	formatForm := tview.NewForm()
	formatForm.AddFormItem(financialSignDropdown).
		AddFormItem(thousandsSeparatorDropdown).
		AddFormItem(decimalSeparatorDropdown).
		AddFormItem(decimalPointsInput).
		AddFormItem(dateTimeFormatDropdown)
	styleCellForm(formatForm, "Number format")

	// Right Column - Styling
	rightForm := tview.NewForm()
	rightForm.AddCheckbox("Bold", c.HasFlag(cell.FlagBold), func(checked bool) { c.SetFlagState(cell.FlagBold, checked) })
	rightForm.AddCheckbox("Italic", c.HasFlag(cell.FlagItalic), func(checked bool) { c.SetFlagState(cell.FlagItalic, checked) })
	rightForm.AddCheckbox("Underline", c.HasFlag(cell.FlagUnderline), func(checked bool) { c.SetFlagState(cell.FlagUnderline, checked) })
	rightForm.AddCheckbox("All Caps", c.HasFlag(cell.FlagAllCaps), func(checked bool) { c.SetFlagState(cell.FlagAllCaps, checked) })
	rightForm.AddCheckbox("Strikethrough", c.HasFlag(cell.FlagStrikethrough), func(checked bool) { c.SetFlagState(cell.FlagStrikethrough, checked) })
	rightForm.AddCheckbox("Editable", c.HasFlag(cell.FlagEditable), func(checked bool) { c.SetFlagState(cell.FlagEditable, checked) })
	//rightForm.AddCheckbox("Formula", c.Formula, func(checked bool) { c.Formula = checked })
	rightForm.AddButton("Data Validation", func() {
		datavalidation.ShowValidationRuleDialog(app, table, container, rightForm.GetFormItem(0), globalData, globalViewport)
	})
	styleCellForm(rightForm, "Style")

	// Color Form
	colorForm := tview.NewForm()

	colorForm.AddDropDown("Text", utils.ColorOptionNames, colorIndex, func(option string, index int) {
		if option == "Custom..." {
			showCustomColorPicker(app, container, c, true, colorForm)
			return
		}
		c.Color = utils.ColorOptions[option]
	})
	colorForm.AddButton("Custom text color", func() {
		showCustomColorPicker(app, container, c, true, colorForm)
	})

	colorForm.AddDropDown("Background", utils.ColorOptionNames, bgColorIndex, func(option string, index int) {
		if option == "Custom..." {
			showCustomColorPicker(app, container, c, false, colorForm)
			return
		}
		c.BgColor = utils.ColorOptions[option]
	})
	colorForm.AddButton("Custom background", func() {
		showCustomColorPicker(app, container, c, false, colorForm)
	})
	styleCellForm(colorForm, "Colors")

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
		SetText(" [#f0d090]Alt+Shift+←/→[-] section    [#f0d090]Tab[-] next field    [#f0d090]Alt+S[-] save    [#f0d090]Esc[-] cancel").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	buttonText.SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))
	buttonBar.AddItem(buttonText, 1, 0, false)

	container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(columns, 0, 1, true).
		AddItem(buttonBar, 1, 0, false)
	container.SetBackgroundColor(tcell.NewRGBColor(10, 12, 18))

	container.SetBorder(true).
		SetTitle(fmt.Sprintf(" Edit %s%d ", utils.ColumnName(column), row)).
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170))

	allForms := []*tview.Form{leftForm, formatForm, rightForm, colorForm}
	currentFormIndex := 0
	markSection(allForms, 0)
	var chords utils.AltChord

	app.SetFocus(leftForm)

	moveSection := func(dir int) {
		n := len(allForms)
		currentFormIndex = (currentFormIndex + dir + n) % n
		markSection(allForms, currentFormIndex)
		app.SetFocus(allForms[currentFormIndex])
	}

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if utils.HasAlt(event.Modifiers()) && (event.Rune() == 's' || event.Rune() == 'S') {
			SaveCellFormButtonAndKeyMap(app, table, container, c, oldCell, row, column, leftForm, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)
			return nil
		}
		dir, consumed := chords.Handle(event, func(fn func()) { app.QueueUpdateDraw(fn) }, func() {
			sheetview.Show(app)
		})
		if dir != 0 {
			moveSection(dir)
			return nil
		}
		if consumed {
			return nil
		}
		return event
	})

	return container, leftForm
}
