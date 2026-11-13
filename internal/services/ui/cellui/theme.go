package cellui

import (
	"gosheet/internal/services/uitheme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	ink   = uitheme.Ink
	panel = uitheme.Panel
	field = uitheme.Field
	line  = uitheme.Line
	amber = uitheme.Amber
	label = uitheme.Label
	textc = uitheme.Text
	btn   = uitheme.Btn
)

func styleCellForm(f *tview.Form, title string) {
	uitheme.StyleForm(f, title).SetTitleAlign(tview.AlignLeft)
}

func markSection(items []tview.Primitive, active int) {
	type borderSettable interface {
		SetBorderColor(tcell.Color) *tview.Box
	}
	for i, item := range items {
		color := line
		if i == active {
			color = amber
		}
		if bs, ok := item.(borderSettable); ok {
			bs.SetBorderColor(color)
		}
	}
}
