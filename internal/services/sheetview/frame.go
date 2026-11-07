// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// frame.go is the formula bar and status line around the grid.

package sheetview

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type frame struct {
	app         *tview.Application
	root        *tview.Flex
	grid        tview.Primitive
	addr        *tview.TextView
	pages       *tview.Pages
	formula     *tview.TextView
	editor      *tview.TextArea
	status      *tview.TextView
	editing     bool
	addrText    string
	formulaText string
	idle        string
}

var (
	current    *frame
	pointFn    func(event *tcell.EventKey) bool
	editStart  func(text string)
	editFinish func()
)

// Mount builds the chrome around the grid. Call once, then Show.
func Mount(app *tview.Application, grid tview.Primitive) tview.Primitive {
	addr := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	addr.SetBackgroundColor(tcell.NewRGBColor(58, 46, 22))
	addr.SetTextColor(tcell.NewRGBColor(240, 208, 144))

	formula := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(false)
	formula.SetBackgroundColor(tcell.NewRGBColor(18, 22, 34))
	formula.SetTextColor(tcell.NewRGBColor(232, 236, 244))

	editor := tview.NewTextArea().SetWrap(false)
	editor.SetBackgroundColor(tcell.NewRGBColor(18, 22, 34))
	editor.SetTextStyle(tcell.StyleDefault.
		Background(tcell.NewRGBColor(18, 22, 34)).
		Foreground(tcell.NewRGBColor(240, 208, 144)))

	pages := tview.NewPages()
	pages.AddPage("view", formula, true, true)
	pages.AddPage("edit", editor, true, false)

	idle := " [#7dcea0]Sheet[-]    [#8aa0b8]type[-] edit    [#f0d090]F2[-] formula    [#f0d090]Enter[-] format    [#f0d090]Del[-] clear    [#f0d090]?[-] help "
	status := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	status.SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))
	status.SetTextColor(tcell.NewRGBColor(210, 218, 230))
	status.SetText(idle)

	bar := tview.NewFlex().SetDirection(tview.FlexColumn)
	bar.AddItem(addr, 18, 0, false)
	bar.AddItem(pages, 0, 1, false)

	bar.SetBackgroundColor(tcell.NewRGBColor(18, 22, 34))

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.SetBackgroundColor(tcell.NewRGBColor(14, 17, 26))
	root.AddItem(bar, 1, 0, false)
	root.AddItem(grid, 0, 1, true)
	root.AddItem(status, 1, 0, false)

	current = &frame{
		app: app, root: root, grid: grid,
		addr: addr, pages: pages, formula: formula, editor: editor, status: status,
		idle: idle,
	}
	return root
}

// Root is the full screen primitive (bar + grid + status).
func Root() tview.Primitive {
	if current == nil {
		return nil
	}
	return current.root
}

// Show returns focus to the sheet, keeping the formula bar visible.
func Show(app *tview.Application) {
	if current == nil || app == nil {
		return
	}
	current.editing = false
	current.pages.SwitchToPage("view")
	if current.idle != "" {
		current.status.SetText(current.idle)
	}
	app.SetRoot(current.root, true).SetFocus(current.grid)
}

// Back closes a dialog. If the screen under it is the grid, the formula bar
// and the shortcut line are restored. Putting the table back by itself is
// what made that line vanish.
func Back(app *tview.Application, returnTo, focus tview.Primitive) {
	if app == nil {
		return
	}
	if current != nil && (returnTo == nil || returnTo == current.grid || returnTo == current.root) {
		Show(app)
		return
	}
	if focus == nil {
		focus = returnTo
	}
	app.SetRoot(returnTo, true).SetFocus(focus)
}

// SetBar updates the address and the raw contents of the selected cell.
func SetBar(addr, formula string) {
	if current == nil || current.editing {
		return
	}
	current.addrText = addr
	current.formulaText = formula
	current.addr.SetText(addr)
	if formula == "" {
		current.formula.SetText(" [#5ec8c8]fx[-]  [#5c6b7a]empty[-]")
		return
	}
	current.formula.SetText(" [#5ec8c8]fx[-]  " + tview.Escape(formula))
}

// SetAddr updates the name box. Safe while a formula is being edited.
func SetAddr(addr string) {
	if current == nil {
		return
	}
	current.addr.SetText(addr)
}

// Editing reports whether the formula bar currently has the keyboard.
func Editing() bool {
	return current != nil && current.editing
}

// SetEditHooks are called when formula editing starts and when it ends.
func SetEditHooks(start func(text string), finish func()) {
	editStart = start
	editFinish = finish
}

// SetPointFunc handles keys while the formula bar is focused.
// Returning true consumes the key so it can point at a cell instead of moving the cursor.
func SetPointFunc(fn func(event *tcell.EventKey) bool) {
	pointFn = fn
}

// EditText returns the formula bar contents and the cursor as a byte index.
// The index matches TextArea for ASCII formulas; that is everything arrows insert.
func EditText() (string, int) {
	if current == nil || current.editor == nil {
		return "", 0
	}
	text := current.editor.GetText()
	_, _, _, col := current.editor.GetCursor()
	if col < 0 {
		col = 0
	}
	if col > len(text) {
		col = len(text)
	}
	return text, col
}

// ReplaceEdit replaces text[start:end] in the formula bar. Indexes are bytes.
func ReplaceEdit(start, end int, insert string) {
	if current == nil || current.editor == nil {
		return
	}
	text := current.editor.GetText()
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start > end {
		start = end
	}
	current.editor.Replace(start, end, insert)
}

// SetStatus replaces the shortcut hint line.
// While a formula is being typed the hint is temporary; the previous one
// comes back when editing ends or a dialog closes.
func SetStatus(text string) {
	if current == nil {
		return
	}
	if !current.editing {
		current.idle = text
	}
	current.status.SetText(text)
}

// Hint is the one-line shortcut bar for the sheet.
func Hint(sheet string) string {
	if strings.TrimSpace(sheet) == "" {
		sheet = "Sheet"
	}
	return " [#7dcea0]" + sheet + "[-]    [#8aa0b8]type[-] edit    [#f0d090]F2[-] formula    [#f0d090]Enter[-] format    [#f0d090]Del[-] clear    [#f0d090]?[-] help "
}

// BeginEdit opens the formula bar for typing. commit is called only on Enter.
func BeginEdit(text string, commit func(string)) {
	if current == nil || current.app == nil {
		return
	}
	current.editing = true
	current.editor.SetText(text, true)
	current.pages.SwitchToPage("edit")
	finish := func(accept bool) {
		if current == nil || !current.editing {
			return
		}
		current.editing = false
		entered := current.editor.GetText()
		if editFinish != nil {
			editFinish()
		}
		if current.idle != "" {
			current.status.SetText(current.idle)
		}
		// Drop focus before hiding the editor page. Pages.SwitchToPage
		// refocuses itself when a child is focused, and that callback is nil
		// if focus was placed on the field directly — which panics (exit 255).
		current.app.SetFocus(current.grid)
		current.pages.SwitchToPage("view")
		if accept && commit != nil {
			commit(entered)
			return
		}
		SetBar(current.addrText, current.formulaText)
	}
	current.editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			finish(true)
			return nil
		case tcell.KeyEscape:
			finish(false)
			return nil
		}
		if pointFn != nil && pointFn(event) {
			return nil
		}
		return event
	})
	if editStart != nil {
		editStart(text)
	}
	current.app.SetFocus(current.editor)
}
