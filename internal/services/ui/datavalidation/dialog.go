// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dialog.go provides the main dialog for data validation

package datavalidation

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowValidationRuleDialog displays the enhanced validation rule editor with presets
func ShowValidationRuleDialog(app *tview.Application, table *tview.Table, returnTo tview.Primitive, focus tview.Primitive, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := globalViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 || col == 0 {
		return
	}

	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		cellData = cell.NewCell(row, col, "")
		globalData[key] = cellData
	}

	if cellData.Valrule == nil {
		emptyStr := ""
		cellData.Valrule = &emptyStr
	}
	if cellData.Valrulemsg == nil {
		emptyStr := ""
		cellData.Valrulemsg = &emptyStr
	}

	presets := GetValidationPresets()
	presetNames := make([]string, len(presets))
	for i, preset := range presets {
		presetNames[i] = preset.Name
	}

	detectedPresetIdx, detectedParams := detectPresetFromRule(*cellData.Valrule)

	container := tview.NewFlex().SetDirection(tview.FlexRow)

	presetDropdown := tview.NewDropDown().
		SetLabel("Validation Type: ").
		SetOptions(presetNames, nil).
		SetCurrentOption(detectedPresetIdx)
	presetDropdown.SetBorder(true).
		SetTitle(" 1. Select Type ").
		SetBorderColor(tcell.ColorLightBlue)

	dynamicForm := tview.NewForm()
	dynamicForm.SetBorder(true).
		SetTitle(" 2. Configure ").
		SetBorderColor(tcell.ColorLightBlue)

	customRuleArea := tview.NewTextArea().
		SetPlaceholder("Enter custom validation rule using 'THIS'...\nExample: THIS > 0 && THIS < 100")
	customRuleArea.SetText(*cellData.Valrule, true)
	customRuleArea.SetBorder(true).
		SetTitle(" Custom Rule (Advanced) ").
		SetBorderColor(tcell.ColorYellow)

	customMessageInput := tview.NewInputField().
		SetLabel("Custom Error Message (optional): ").
		SetText(*cellData.Valrulemsg).
		SetFieldWidth(60).
		SetPlaceholder("Leave empty for default message")
	customMessageInput.SetBorder(true).
		SetTitle(" 3. Error Message ").
		SetBorderColor(tcell.ColorPurple)

	previewText := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	previewText.SetBorder(true).
		SetTitle(" Preview ").
		SetBorderColor(tcell.ColorGreen)

	buttonForm := tview.NewForm()
	buttonForm.AddButton("Apply", func() {
		currentPresetIdx, _ := presetDropdown.GetCurrentOption()
		preset := presets[currentPresetIdx]

		var finalRule string
		if preset.Name == "Custom" {
			finalRule = strings.TrimSpace(customRuleArea.GetText())
		} else {
			params := make(map[string]string)
			for i := 0; i < dynamicForm.GetFormItemCount(); i++ {
				item := dynamicForm.GetFormItem(i)
				if inputField, ok := item.(*tview.InputField); ok {
					if i < len(preset.Fields) {
						fieldName := preset.Fields[i].Name
						params[fieldName] = inputField.GetText()
					}
				}
			}
			
			allFilled := true
			for _, field := range preset.Fields {
				if params[field.Name] == "" {
					allFilled = false
					break
				}
			}
			
			if !allFilled {
				showValidationErrorModal(app, container, container, "Please fill in all required fields before applying.")
				return
			}
			
			finalRule = preset.BuildRule(params)
		}

		customMsg := strings.TrimSpace(customMessageInput.GetText())
		cellData.Valrulemsg = &customMsg

		saveRule(app, table, cellData, finalRule, row, col, buttonForm, returnTo, focus, container, globalData, globalViewport)
	})

	buttonForm.AddButton("Delete", func() {
		deleteRule(cellData, table, row, col, globalData, globalViewport)
		app.SetRoot(returnTo, true).SetFocus(focus)
	})

	buttonForm.AddButton("Cancel", func() {
		app.SetRoot(returnTo, true).SetFocus(focus)
	})

	var currentFormItemsInDynamic []tview.FormItem

	updateDynamicForm := func(presetIdx int) {
		dynamicForm.Clear(true)
		currentFormItemsInDynamic = nil
		preset := presets[presetIdx]

		previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[gray]Empty cells are always allowed, validation only applies when entering a value.[-]", 
			preset.Name, preset.Description))

		if preset.Name == "Custom" {
			//if container.GetItemCount() == 5 {
			//	container.RemoveItem(dynamicForm)
			//}
			//if container.GetItemCount() == 4 {
			//	items := make([]tview.Primitive, 0)
			//	for i := 0; i < container.GetItemCount(); i++ {
			//		item := container.GetItem(i)
			//		if item == customMessageInput {
			//			items = append(items, customRuleArea)
			//		}
			//		items = append(items, item)
			//	}
			container.Clear()
				container.AddItem(presetDropdown, 0, 1, false).
				AddItem(previewText, 0, 2, false).
				AddItem(customRuleArea, 0, 5, true).
				AddItem(customMessageInput, 0, 1, false).
				AddItem(buttonForm, 0, 1, false)
			//}
		} else {
			if container.GetItemCount() == 5 {
				container.Clear()
				container.AddItem(presetDropdown, 0, 1, false).
					AddItem(previewText, 0, 2, false).
					AddItem(dynamicForm, 0, 3, false).
					AddItem(customMessageInput, 3, 1, false).
					AddItem(buttonForm, 0, 1, false)
			}

			for i, field := range preset.Fields {
				inputField := tview.NewInputField().
					SetLabel(field.Label).
					SetPlaceholder(field.Placeholder).
					SetFieldWidth(30)
				
				if detectedParams != nil && detectedPresetIdx == presetIdx {
					if val, ok := detectedParams[field.Name]; ok {
						inputField.SetText(val)
					}
				}
				
				dynamicForm.AddFormItem(inputField)
				currentFormItemsInDynamic = append(currentFormItemsInDynamic, inputField)
				
				idx := i
				inputField.SetChangedFunc(func(text string) {
					params := make(map[string]string)
					for j := 0; j < len(preset.Fields); j++ {
						fieldItem := dynamicForm.GetFormItem(j)
						if fi, ok := fieldItem.(*tview.InputField); ok {
							params[preset.Fields[j].Name] = fi.GetText()
						}
					}
					
					allFilled := true
					for _, f := range preset.Fields {
						if params[f.Name] == "" {
							allFilled = false
							break
						}
					}
					
					if allFilled {
						rule := preset.BuildRule(params)
						previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[green]Generated Rule:[white]\n%s\n\n[gray]Empty cells are always allowed.[-]", 
							preset.Name, preset.Description, rule))
					} else {
						previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[gray]Fill in all fields to see the generated rule.\nEmpty cells are always allowed.[-]", 
							preset.Name, preset.Description))
					}
					_ = idx 
				})
			}
			
			if detectedParams != nil && detectedPresetIdx == presetIdx && len(preset.Fields) > 0 {
				rule := preset.BuildRule(detectedParams)
				previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[green]Current Rule:[white]\n%s\n\n[gray]Empty cells are always allowed.[-]", 
					preset.Name, preset.Description, rule))
			}
		}
	}

	presetDropdown.SetSelectedFunc(func(text string, index int) {
		updateDynamicForm(index)
	})

	container.
		AddItem(presetDropdown, 0, 1, false).
		AddItem(previewText, 0, 2, false).
		AddItem(dynamicForm, 0, 3, false).
		AddItem(customMessageInput, 0, 1, false).
		AddItem(buttonForm, 0, 1, false)

	updateDynamicForm(detectedPresetIdx)


	container.SetBorder(true).
		SetTitle(fmt.Sprintf(" Data Validation - %s%d  •  Ctrl+←/→ to navigate  •  Esc to cancel ", utils.ColumnName(col), row)).
		SetBorderColor(tcell.ColorYellow)

	getFocusablePrimitives := func() []tview.Primitive {
		focusable := []tview.Primitive{}
		
		focusable = append(focusable, presetDropdown)
		
		for i := 0; i < container.GetItemCount(); i++ {
			item := container.GetItem(i)
			if item == customRuleArea {
				focusable = append(focusable, customRuleArea)
				break
			} else if item == dynamicForm && dynamicForm.GetFormItemCount() > 0 {
				focusable = append(focusable, dynamicForm)
				break
			}
		}
		
		focusable = append(focusable, customMessageInput)
		focusable = append(focusable, buttonForm)

		return focusable
	}

	currentPrim := 0

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(returnTo, true).SetFocus(focus)
			return nil
		}
		if event.Modifiers()&tcell.ModCtrl != 0 {
			focusables := getFocusablePrimitives()
			
			if event.Key() == tcell.KeyRight {
				currentPrim++
				currentPrim %= len(focusables)
				app.SetFocus(focusables[currentPrim])
				return nil
			} else if event.Key() == tcell.KeyLeft {
				currentPrim--
				if currentPrim < 0 {
					currentPrim = len(focusables) - 1
				}
				app.SetFocus(focusables[currentPrim])
				return nil
			}
		}
		return event
	})

	app.SetRoot(container, true).SetFocus(presetDropdown)
}

