// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// otherUI.go contains other UI functions.

package ui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Help Modal
func ShowHelpModal(app *tview.Application, table *tview.Table) {
	helpText := `[yellow]NAVIGATION:[white]
  Arrow Keys           Navigate cells
  Shift + Arrows       Select range
  Alt + G              Go to cell
  Escape               Save dialog
  
  Note: In menus, such as the start menu, you can use Ctrl+←/→ to navigate around.

[yellow]EDITING:[white]
  Enter                Edit cell
  Alt + S              Save Menu
  Alt + Delete         Clear selection
  Alt + N              Edit cell comment
  Alt + A              AutoFill

[yellow]Sheet Management:[white]
  Alt+M           Open Sheet Manager
  Alt+T           Quick Sheet Menu
  Alt+PageUp      Previous Sheet
  Alt+PageDown    Next Sheet
  
[yellow]In Sheet Manager:[white]
  Alt+N           New Sheet
  Alt+R           Rename Sheet
  Alt+D           Delete Sheet
  Alt+M           Move/Reorder
  Alt+C           Duplicate Sheet
  Alt+S           Switch to Sheet

[yellow]CLIPBOARD:[white]
  Alt + C              Copy
  Alt + V              Paste
  Alt + X              Cut
  Alt + R              Copy format
  Alt + I              Paste format
  Alt + Z              Undo last action
  Alt + Y              Redo last action

[yellow]SEARCH & REPLACE:[white]
  Alt + F              Find dialog
  Alt + H              Replace dialog
  F3                   Find previous
  F4                   Find next

[yellow]ROWS & COLUMNS:[white]
  Alt + Minus (-)      Delete row/column
  Alt + Equal (=)      Insert row/column

[yellow]SORTING:[white]
  Alt + O              Sort dialog

[yellow]HELP:[white]
  Alt + /              Show this help`

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetScrollable(true).
		SetWrap(false)
	
	textView.SetBorder(true).
		SetTitle(" Keyboard Shortcuts - Use arrow keys to scroll • Press ESC to close ").
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlack)
	
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.SetRoot(table, true).SetFocus(table)
			return nil
		case tcell.KeyEnter:
			app.SetRoot(table, true).SetFocus(table)
			return nil
		}
		return event
	})

	app.SetRoot(textView, true).SetFocus(textView)
}

// showCustomColorPicker displays a custom RGB color picker
func showCustomColorPicker(app *tview.Application, returnTo *tview.Flex, c *cell.Cell, isTextColor bool, parentForm *tview.Form) {
	pickerForm := tview.NewForm()
	
	var currentColor utils.ColorRGB
	if isTextColor {
		currentColor = c.Color
	} else {
		currentColor = c.BgColor
	}
	
	preview := tview.NewTextView()
	preview.SetBackgroundColor(currentColor.ToTCellColor())
	preview.SetBorder(true).SetTitle(" Preview ")

	updating := false
	
	redInput := tview.NewInputField().
		SetLabel("Red (0-255): ").
		SetText(fmt.Sprintf("%d", currentColor[0])).
		SetFieldWidth(5).
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			if text == "" {
				return true
			}
			if val, err := strconv.Atoi(text); err == nil {
				return val >= 0 && val <= 255
			}
			return false
		})
	
	greenInput := tview.NewInputField().
		SetLabel("Green (0-255): ").
		SetText(fmt.Sprintf("%d", currentColor[1])).
		SetFieldWidth(5).
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			if text == "" {
				return true
			}
			if val, err := strconv.Atoi(text); err == nil {
				return val >= 0 && val <= 255
			}
			return false
		})
	
	blueInput := tview.NewInputField().
		SetLabel("Blue (0-255): ").
		SetText(fmt.Sprintf("%d", currentColor[2])).
		SetFieldWidth(5).
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			if text == "" {
				return true
			}
			if val, err := strconv.Atoi(text); err == nil {
				return val >= 0 && val <= 255
			}
			return false
		})
	
	hexInput := tview.NewInputField().
		SetLabel("Hex (#RRGGBB): ").
		SetText(currentColor.Hex()).
		SetFieldWidth(9)
	
	updatePreview := func() {
		if updating {
			return
		}
		updating = true
		
		rText := redInput.GetText()
		gText := greenInput.GetText()
		bText := blueInput.GetText()
		
		if rText == "" {
			rText = "0"
		}
		if gText == "" {
			gText = "0"
		}
		if bText == "" {
			bText = "0"
		}
		
		r, _ := strconv.Atoi(rText)
		g, _ := strconv.Atoi(gText)
		b, _ := strconv.Atoi(bText)
		
		if r < 0 {
			r = 0
		}
		if r > 255 {
			r = 255
		}
		if g < 0 {
			g = 0
		}
		if g > 255 {
			g = 255
		}
		if b < 0 {
			b = 0
		}
		if b > 255 {
			b = 255
		}
		
		color := utils.ColorRGB{uint8(r), uint8(g), uint8(b)}
		preview.SetBackgroundColor(color.ToTCellColor())
		hexInput.SetText(color.Hex())
		
		updating = false
	}
	
	updateFromHex := func() {
		if updating {
			return
		}
		updating = true
		
		hex := hexInput.GetText()
		if len(hex) == 7 && hex[0] == '#' {
			if r, g, b, ok := parseHexColor(hex); ok {
				redInput.SetText(fmt.Sprintf("%d", r))
				greenInput.SetText(fmt.Sprintf("%d", g))
				blueInput.SetText(fmt.Sprintf("%d", b))
				
				color := utils.ColorRGB{r, g, b}
				preview.SetBackgroundColor(color.ToTCellColor())
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
	
	pickerForm.AddFormItem(redInput).
		AddFormItem(greenInput).
		AddFormItem(blueInput).
		AddFormItem(hexInput)
	
	pickerForm.AddButton("Apply", func() {
		rText := redInput.GetText()
		gText := greenInput.GetText()
		bText := blueInput.GetText()
		
		if rText == "" {
			rText = "0"
		}
		if gText == "" {
			gText = "0"
		}
		if bText == "" {
			bText = "0"
		}
		
		r, _ := strconv.Atoi(rText)
		g, _ := strconv.Atoi(gText)
		b, _ := strconv.Atoi(bText)
		
		color := utils.ColorRGB{uint8(r), uint8(g), uint8(b)}
		
		if isTextColor {
			c.Color = color
		} else {
			c.BgColor = color
		}
		
		app.SetRoot(returnTo, true).SetFocus(parentForm)
	})
	
	pickerForm.AddButton("Cancel", func() {
		app.SetRoot(returnTo, true).SetFocus(parentForm)
	})
	
	title := " Custom Text Color "
	if !isTextColor {
		title = " Custom Background Color "
	}
	pickerForm.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)
	
	topRow := tview.NewFlex().
		AddItem(pickerForm, 0, 3, true).
		AddItem(preview, 0, 1, false) 
	
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topRow, 0, 1, true)

	container.SetBorder(true).SetBorderColor(tcell.ColorYellow)
	
	app.SetRoot(container, true).SetFocus(pickerForm)
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
