// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// frame.go is the formula bar and status line around the grid.

package sheetview

import (
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
	addr.SetBackgroundColor(tcell.ColorDarkBlue)
	addr.SetTextColor(tcell.ColorWhite)

	formula := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(false)
	formula.SetBackgroundColor(tcell.ColorBlack)
	formula.SetTextColor(tcell.ColorWhite)

	editor := tview.NewTextArea().SetWrap(false)
	editor.SetBackgroundColor(tcell.ColorBlack)
	editor.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow))

	pages := tview.NewPages()
	pages.AddPage("view", formula, true, true)
	pages.AddPage("edit", editor, true, false)

	status := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	status.SetBackgroundColor(tcell.ColorDarkSlateGray)
	status.SetTextColor(tcell.ColorWhite)
	status.SetText(" type to replace  ·  F2 edit  ·  Enter format  ·  Del clear  ·  Alt+/ help ")

	bar := tview.NewFlex().SetDirection(tview.FlexColumn)
	bar.AddItem(addr, 18, 0, false)
	bar.AddItem(pages, 0, 1, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(bar, 1, 0, false)
	root.AddItem(grid, 0, 1, true)
	root.AddItem(status, 1, 0, false)

	current = &frame{
		app: app, root: root, grid: grid,
		addr: addr, pages: pages, formula: formula, editor: editor, status: status,
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
	app.SetRoot(current.root, true).SetFocus(current.grid)
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
		current.formula.SetText(" [gray]empty")
		return
	}
	current.formula.SetText(" " + tview.Escape(formula))
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
func SetStatus(text string) {
	if current == nil {
		return
	}
	current.status.SetText(text)
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
		// Drop focus before hiding the editor page. Pages.SwitchToPage
		// refocuses itself when a child is focused, and that callback is nil
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
