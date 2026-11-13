// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpUI.go contains the help modal

package ui

import (
	"gosheet/internal/services/sheetview"
	"gosheet/internal/services/uitheme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Help Modal
func ShowHelpModal(app *tview.Application, table *tview.Table) {
	helpText := uitheme.Markup(`{h1}EVERY DAY{end}
  Arrows          Move
  Shift+Arrows    Select a block
  Type            Replace the cell
  F2              Edit in the formula bar
  Enter           Format the cell
  Delete          Clear
  ?               This help
  Esc             Save

{h2}FORMULAS{end}
  =               Start a formula. Arrows insert the highlighted cell
  Alt+PgUp/PgDn   Point at another sheet, for example Prices!B2
  Alt+1 … 9       Jump to a sheet

{h3}LESS OFTEN{end}  (Alt + the letter)
  S save     G go to cell     F find     H replace
  C copy     X cut     V paste     Z undo     Y redo
  A fill     O sort     N comment
  = insert row/column     - delete row/column
  M sheets     T sheet menu

{muted}Alt+Shift+←/→ moves between panels. Plain Alt+arrow is swallowed by the terminal.{end}
Everything else is Alt, or a plain key. F3 / F4 repeat a search.

{dim}Esc closes this.{end}`)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetScrollable(true).
		SetWrap(true)

	uitheme.Paint(textView, "GoSheet  ·  ? or Esc closes")
	textView.SetTitleAlign(tview.AlignCenter).SetBackgroundColor(uitheme.Ink)

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
	uitheme.StyleModal(modal, "Info")
	app.SetRoot(modal, true).SetFocus(modal)
}
