// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// colours.go provides a colour picker dialog

package cellui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCustomColorPicker displays a color picker with a grid palette and a
// custom RGB/HEX form.
func showCustomColorPicker(app *tview.Application, returnTo *tview.Flex, c *cell.Cell, isTextColor bool, parentFocus tview.Primitive) {
	title := "Text Color"
	if !isTextColor {
		title = "Background Color"
	}

	palette := uitheme.Palette
	current := c.Color
	if !isTextColor {
		current = c.BgColor
	}

	const cols = 5
	pages := tview.NewPages()
	var pickerModal, customModal tview.Primitive

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, true)
	table.SetBackgroundColor(panel)

	selectedRow, selectedCol := 0, 0
	for i, entry := range palette {
		row, col := i/cols, i%cols
		label := entry.Name
		color := swatchColor(entry.Value, isTextColor)
		if color == current {
			label = "* " + entry.Name
			selectedRow, selectedCol = row, col
		}
		cell := tview.NewTableCell(" " + label + " ").
			SetAlign(tview.AlignCenter).
			SetExpansion(1)
		if entry.Value == "" {
			cell.SetBackgroundColor(field).SetTextColor(textc)
		} else {
			tc := color.ToTCellColor()
			cell.SetBackgroundColor(tc).SetTextColor(contrastColor(tc))
		}
		table.SetCell(row, col, cell)
	}

	customRow, customCol := len(palette)/cols, len(palette)%cols
	table.SetCell(customRow, customCol,
		tview.NewTableCell(" [ Custom... ] ").
			SetAlign(tview.AlignCenter).
			SetExpansion(1).
			SetBackgroundColor(line).
			SetTextColor(ink))
	table.Select(selectedRow, selectedCol)

	gridPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)
	gridPage.SetBackgroundColor(panel)
	pages.AddPage("grid", gridPage, true, true)

	preview := tview.NewTextView()
	preview.SetBackgroundColor(current.ToTCellColor())
	preview.SetBorder(true).SetTitle(" Preview ").SetTitleAlign(tview.AlignCenter)
	preview.SetBorderColor(line).SetTitleColor(amber)

	customForm := tview.NewForm()
	updating := false

	acceptRGB := func(text string, _ rune) bool {
		if text == "" {
			return true
		}
		if v, err := strconv.Atoi(text); err == nil {
			return v >= 0 && v <= 255
		}
		return false
	}

	redInput := tview.NewInputField().
		SetLabel("Red (0-255): ").
		SetText(fmt.Sprintf("%d", current[0])).
		SetFieldWidth(5).
		SetAcceptanceFunc(acceptRGB)

	greenInput := tview.NewInputField().
		SetLabel("Green (0-255): ").
		SetText(fmt.Sprintf("%d", current[1])).
		SetFieldWidth(5).
		SetAcceptanceFunc(acceptRGB)

	blueInput := tview.NewInputField().
		SetLabel("Blue (0-255): ").
		SetText(fmt.Sprintf("%d", current[2])).
		SetFieldWidth(5).
		SetAcceptanceFunc(acceptRGB)

	hexInput := tview.NewInputField().
		SetLabel("Hex (#RRGGBB): ").
		SetText(current.Hex()).
		SetFieldWidth(9)

	updatePreview := func() {
		if updating {
			return
		}
		updating = true
		rText, gText, bText := redInput.GetText(), greenInput.GetText(), blueInput.GetText()
		if rText == "" {
			rText = "0"
		}
		if gText == "" {
			gText = "0"
		}
		if bText == "" {
			bText = "0"
		}
		rV, _ := strconv.Atoi(rText)
		gV, _ := strconv.Atoi(gText)
		bV, _ := strconv.Atoi(bText)

		cr := utils.ColorRGB{uint8(rV), uint8(gV), uint8(bV)}
		preview.SetBackgroundColor(cr.ToTCellColor())
		hexInput.SetText(cr.Hex())
		updating = false
	}

	updateFromHex := func() {
		if updating {
			return
		}
		updating = true
		hex := hexInput.GetText()
		if len(hex) == 7 && hex[0] == '#' {
			if rV, gV, bV, ok := parseHexColor(hex); ok {
				redInput.SetText(fmt.Sprintf("%d", rV))
				greenInput.SetText(fmt.Sprintf("%d", gV))
				blueInput.SetText(fmt.Sprintf("%d", bV))
				cr := utils.ColorRGB{rV, gV, bV}
				preview.SetBackgroundColor(cr.ToTCellColor())
			}
		}
		updating = false
	}

	redInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyTab {
			updatePreview()
		}
	})
	greenInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyTab {
			updatePreview()
		}
	})
	blueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyTab {
			updatePreview()
		}
	})
	hexInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyTab {
			updateFromHex()
		}
	})

	customForm.AddFormItem(redInput).
		AddFormItem(greenInput).
		AddFormItem(blueInput).
		AddFormItem(hexInput)

	// Shared apply (used by grid selection and the custom Apply button).
	applyColor := func(color utils.ColorRGB) {
		if isTextColor {
			c.Color = color
		} else {
			c.BgColor = color
		}
		sheetview.Back(app, returnTo, parentFocus)
	}

	applyCustom := func() {
		rText, gText, bText := redInput.GetText(), greenInput.GetText(), blueInput.GetText()
		if rText == "" {
			rText = "0"
		}
		if gText == "" {
			gText = "0"
		}
		if bText == "" {
			bText = "0"
		}
		rV, _ := strconv.Atoi(rText)
		gV, _ := strconv.Atoi(gText)
		bV, _ := strconv.Atoi(bText)
		applyColor(utils.ColorRGB{uint8(rV), uint8(gV), uint8(bV)})
	}

	customForm.AddButton("Apply", applyCustom)
	customForm.AddButton("Back", func() {
		pages.SwitchToPage("grid")
		app.SetRoot(pickerModal, true).SetFocus(table)
	})

	uitheme.StyleForm(customForm, "").SetBorder(false)

	topRow := tview.NewFlex().
		AddItem(customForm, 0, 3, true).
		AddItem(preview, 0, 1, false)

	customPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topRow, 0, 1, true)
	customPage.SetBackgroundColor(ink)
	pages.AddPage("custom", customPage, true, false)

	// --- ACTIONS ---
	applyGrid := func() {
		row, col := table.GetSelection()
		if row == customRow && col == customCol {
			pages.SwitchToPage("custom")
			app.SetRoot(customModal, true).SetFocus(customForm)
			return
		}
		idx := row*cols + col
		if idx >= 0 && idx < len(palette) {
			applyColor(swatchColor(palette[idx].Value, isTextColor))
		}
	}
	table.SetSelectedFunc(func(row, col int) { applyGrid() })
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			sheetview.Back(app, returnTo, parentFocus)
			return nil
		}
		return event
	})

	customForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("grid")
			app.SetRoot(pickerModal, true).SetFocus(table)
			return nil
		}
		return event
	})

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true)
	container.SetBackgroundColor(panel)
	container.SetBorder(true).
		SetBorderColor(line).
		SetTitle(" " + title + " ").
		SetTitleColor(amber).
		SetTitleAlign(tview.AlignCenter)

	pickerModal = uitheme.CenteredModal(container, 62, 8)
	customModal = uitheme.CenteredModal(container, 62, 13)
	app.SetRoot(pickerModal, true).SetFocus(table)
}

func swatchColor(value string, isTextColor bool) utils.ColorRGB {
	if value == "" {
		if isTextColor {
			return utils.ColorRGB{255, 255, 255}
		}
		return utils.ColorRGB{0, 0, 0}
	}
	r, g, b := uitheme.ParseHex(value).RGB()
	return utils.ColorRGB{uint8(r), uint8(g), uint8(b)}
}

// contrastColor picks black or white text for readability against c.
func contrastColor(c tcell.Color) tcell.Color {
	return uitheme.Contrast(c)
}

// parseHexColor parses a hex color string (#RRGGBB) into RGB values
func parseHexColor(hex string) (r, g, b uint8, ok bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, false
	}

	var val int64
	var err error

	if val, err = strconv.ParseInt(hex[1:3], 16, 0); err != nil {
		return 0, 0, 0, false
	}
	r = uint8(val)

	if val, err = strconv.ParseInt(hex[3:5], 16, 0); err != nil {
		return 0, 0, 0, false
	}
	g = uint8(val)

	if val, err = strconv.ParseInt(hex[5:7], 16, 0); err != nil {
		return 0, 0, 0, false
	}
	b = uint8(val)

	return r, g, b, true
}
