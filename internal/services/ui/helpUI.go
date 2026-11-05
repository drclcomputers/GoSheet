// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpUI.go contains the help modal

package ui

import (
	"gosheet/internal/services/sheetview"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Help Modal
func ShowHelpModal(app *tview.Application, table *tview.Table) {
	helpText := `[#f0d090::b]EVERY DAY[-]
  Arrows          Move
  Shift+Arrows    Select a block
  Type            Replace the cell
  F2              Edit in the formula bar
  Enter           Format the cell
  Delete          Clear
  ?               This help
  Esc             Save

[#7dcea0::b]FORMULAS[-]
  =               Start a formula. Arrows insert the highlighted cell
  Alt+PgUp/PgDn   Point at another sheet, for example Prices!B2
  Alt+1 … 9       Jump to a sheet

[#8ecae6::b]LESS OFTEN[-]  (Alt + the letter)
  S save     G go to cell     F find     H replace
  C copy     X cut     V paste     Z undo     Y redo
  A fill     O sort     N comment
  = insert row/column     - delete row/column
  M sheets     T sheet menu

[#8aa0b8]Alt+Shift+←/→ moves between panels. Plain Alt+arrow is swallowed by the terminal.[-]
Everything else is Alt, or a plain key. F3 / F4 repeat a search.[-]

[#5c6b7a]Esc closes this.[-]`

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetScrollable(true).
		SetWrap(true)

	textView.SetBorder(true).
		SetTitle(" GoSheet  ·  ? or Esc closes ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.NewRGBColor(240, 208, 144)).
		SetBorderColor(tcell.NewRGBColor(70, 160, 170)).
		SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			sheetview.Show(app)
			return nil
		case tcell.KeyEnter:
			sheetview.Show(app)
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
			sheetview.Back(app, returnTo, returnTo)
		})
	modal.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}
