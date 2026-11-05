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
	editor      *tview.InputField
	status      *tview.TextView
	editing     bool
	addrText    string
	formulaText string
}

var current *frame

// Mount builds the chrome around the grid. Call once, then Show.
func Mount(app *tview.Application, grid tview.Primitive) tview.Primitive {
	addr := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	addr.SetBackgroundColor(tcell.ColorDarkBlue)
	addr.SetTextColor(tcell.ColorWhite)

	formula := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(false)
	formula.SetBackgroundColor(tcell.ColorBlack)
	formula.SetTextColor(tcell.ColorWhite)

	editor := tview.NewInputField()
	editor.SetFieldBackgroundColor(tcell.ColorBlack)
	editor.SetFieldTextColor(tcell.ColorYellow)
	editor.SetLabel("")

	pages := tview.NewPages()
	pages.AddPage("view", formula, true, true)
	pages.AddPage("edit", editor, true, false)

	status := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	status.SetBackgroundColor(tcell.ColorDarkSlateGray)
	status.SetTextColor(tcell.ColorWhite)
	status.SetText(" type to replace  ·  F2 edit  ·  Enter format  ·  Del clear  ·  Alt+/ help ")

	bar := tview.NewFlex().SetDirection(tview.FlexColumn)
	bar.AddItem(addr, 12, 0, false)
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
	current.editor.SetText(text)
	current.pages.SwitchToPage("edit")
	current.editor.SetDoneFunc(func(key tcell.Key) {
		if current == nil || !current.editing {
			return
		}
		current.editing = false
		entered := current.editor.GetText()
		// Drop focus before hiding the editor page. Pages.SwitchToPage
		// refocuses itself when a child is focused, and that callback is nil
		current.app.SetFocus(current.grid)
		current.pages.SwitchToPage("view")
		if key == tcell.KeyEnter && commit != nil {
			commit(entered)
			return
		}
		SetBar(current.addrText, current.formulaText)
	})
	current.app.SetFocus(current.editor)
}
