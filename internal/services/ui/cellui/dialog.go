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
	"gosheet/internal/services/uitheme"
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

	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown, decimalPointsInput := getDropdowns(c)

	disableFormattingFields(financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput)

	editCellDialog, leftForm := buildEditCellForm(app, table, c, oldCell, row, column,
		financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, decimalPointsInput,
		dateTimeFormatDropdown, typeIndex, alignIndex, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)

	app.SetRoot(editCellDialog, true).SetFocus(leftForm)
}

// Form Builder
func buildEditCellForm(app *tview.Application, table *tview.Table, c, oldCell *cell.Cell, row, column int32,
	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown *tview.DropDown,
	decimalPointsInput *tview.InputField, dateTimeFormatDropdown *tview.DropDown, typeIndex, alignIndex int,
	RecordCellEdit func(table *tview.Table, row, col int32, oldCell,
		newCell *cell.Cell), EvaluateCell func(table *tview.Table, c *cell.Cell) error,
	RecalculateCell func(table *tview.Table, c *cell.Cell) error, globalData map[[2]int]*cell.Cell,
	globalViewport *utils.Viewport) (*tview.Flex, *tview.Form) {

	container := tview.NewFlex()
	var editModal *tview.Flex

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

	// Style toggles and editing actions.
	rightForm := tview.NewForm()
	rightForm.AddCheckbox("Bold", c.HasFlag(cell.FlagBold), func(checked bool) { c.SetFlagState(cell.FlagBold, checked) })
	rightForm.AddCheckbox("Italic", c.HasFlag(cell.FlagItalic), func(checked bool) { c.SetFlagState(cell.FlagItalic, checked) })
	rightForm.AddCheckbox("Underline", c.HasFlag(cell.FlagUnderline), func(checked bool) { c.SetFlagState(cell.FlagUnderline, checked) })
	rightForm.AddCheckbox("All Caps", c.HasFlag(cell.FlagAllCaps), func(checked bool) { c.SetFlagState(cell.FlagAllCaps, checked) })
	rightForm.AddCheckbox("Strikethrough", c.HasFlag(cell.FlagStrikethrough), func(checked bool) { c.SetFlagState(cell.FlagStrikethrough, checked) })
	rightForm.AddCheckbox("Editable", c.HasFlag(cell.FlagEditable), func(checked bool) { c.SetFlagState(cell.FlagEditable, checked) })
	styleCellForm(rightForm, "Style")
	rightForm.SetItemPadding(0)

	var validationButton, textColorButton, backgroundButton *tview.Button
	validationButton = uitheme.StyleButton(tview.NewButton("Data Validation").SetSelectedFunc(func() {
		datavalidation.ShowValidationRuleDialog(app, table, editModal, rightForm, globalData, globalViewport)
	}))
	textColorButton = uitheme.StyleButton(tview.NewButton("Text Color…").SetSelectedFunc(func() {
		showCustomColorPicker(app, editModal, c, true, textColorButton)
	}))
	backgroundButton = uitheme.StyleButton(tview.NewButton("Background Color…").SetSelectedFunc(func() {
		showCustomColorPicker(app, editModal, c, false, backgroundButton)
	}))
	actionButtons := []*tview.Button{validationButton, textColorButton, backgroundButton}
	for i, button := range actionButtons {
		i := i
		button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab, tcell.KeyDown:
				app.SetFocus(actionButtons[(i+1)%len(actionButtons)])
				return nil
			case tcell.KeyBacktab, tcell.KeyUp:
				app.SetFocus(actionButtons[(i-1+len(actionButtons))%len(actionButtons)])
				return nil
			}
			return event
		})
	}
	actionPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(validationButton, 1, 0, true).
		AddItem(textColorButton, 1, 0, false).
		AddItem(backgroundButton, 1, 0, false)
	uitheme.Paint(actionPanel, "Actions")

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rightForm, 0, 1, false).
		AddItem(actionPanel, 5, 0, false)

	columns := tview.NewFlex().
		AddItem(leftForm, 0, 1, false).
		AddItem(formatForm, 0, 1, false).
		AddItem(rightColumn, 0, 1, false)

	buttonBar := tview.NewFlex().SetDirection(tview.FlexRow)
	buttonText := tview.NewTextView().
		SetText(uitheme.Markup(" {val}Alt+Shift+←/→{end} section    {val}Tab{end} next field    {val}Alt+S{end} save    {val}Esc{end} cancel")).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	buttonText.SetBackgroundColor(uitheme.Ink)
	buttonBar.AddItem(buttonText, 1, 0, false)

	container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(columns, 0, 1, true).
		AddItem(buttonBar, 1, 0, false)
	uitheme.Paint(container, fmt.Sprintf("Edit %s%d", utils.ColumnName(column), row))
	container.SetTitleAlign(tview.AlignCenter).SetBackgroundColor(uitheme.Ink)
	editModal = uitheme.CenteredModal(container, 90, 18)

	// Mixed *tview.Form and *tview.Flex — both expose SetBorderColor.
	allSections := []tview.Primitive{leftForm, formatForm, rightForm, actionPanel}
	currentFormIndex := 0
	markSection(allSections, 0)
	var chords utils.AltChord

	app.SetFocus(leftForm)

	moveSection := func(dir int) {
		n := len(allSections)
		currentFormIndex = (currentFormIndex + dir + n) % n
		markSection(allSections, currentFormIndex)
		app.SetFocus(allSections[currentFormIndex])
	}

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if utils.HasAlt(event.Modifiers()) && (event.Rune() == 's' || event.Rune() == 'S') {
			SaveCellFormButtonAndKeyMap(app, table, editModal, c, oldCell, row, column, leftForm, RecordCellEdit, EvaluateCell, RecalculateCell, globalData, globalViewport)
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

	return editModal, leftForm
}
