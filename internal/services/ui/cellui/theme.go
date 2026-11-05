package cellui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	ink   = tcell.NewRGBColor(14, 17, 26)
	panel = tcell.NewRGBColor(18, 22, 34)
	field = tcell.NewRGBColor(12, 14, 22)
	line  = tcell.NewRGBColor(70, 160, 170)
	amber = tcell.NewRGBColor(240, 208, 144)
	label = tcell.NewRGBColor(168, 188, 214)
	textc = tcell.NewRGBColor(232, 236, 244)
	btn   = tcell.NewRGBColor(28, 110, 128)
)

func styleCellForm(f *tview.Form, title string) {
	f.SetBorder(true).
		SetTitle(" " + title + " ").
		SetTitleAlign(tview.AlignLeft).
		SetTitleColor(amber).
		SetBorderColor(line).
		SetBackgroundColor(panel)
	f.SetLabelColor(label)
	f.SetFieldBackgroundColor(field)
	f.SetFieldTextColor(textc)
	f.SetButtonBackgroundColor(btn)
	f.SetButtonTextColor(tcell.ColorWhite)
	f.SetButtonActivatedStyle(tcell.StyleDefault.Background(amber).Foreground(ink))
}

func markSection(forms []*tview.Form, active int) {
	for i, f := range forms {
		if i == active {
			f.SetBorderColor(amber)
			continue
		}
		f.SetBorderColor(line)
	}
}
