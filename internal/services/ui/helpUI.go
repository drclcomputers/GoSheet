// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpUI.go contains the help modal

package ui

import (
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

// Warning Modal
func ShowWarningModal(app *tview.Application, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(returnTo, true).SetFocus(returnTo)
		})
	modal.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}
