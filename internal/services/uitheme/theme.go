// Package uitheme is the shared look of the dialogs: dark panels, a teal
// border, amber titles. Screens that were restyled later should use this
// instead of the old yellow, red, and green frames.
package uitheme

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	Ink   = tcell.NewRGBColor(14, 17, 26)
	Panel = tcell.NewRGBColor(18, 22, 34)
	Field = tcell.NewRGBColor(12, 14, 22)
	Line  = tcell.NewRGBColor(70, 160, 170)
	Amber = tcell.NewRGBColor(240, 208, 144)
	Label = tcell.NewRGBColor(168, 188, 214)
	Text  = tcell.NewRGBColor(232, 236, 244)
	Btn   = tcell.NewRGBColor(28, 110, 128)
	Warn  = tcell.NewRGBColor(196, 122, 92)
	Ok    = tcell.NewRGBColor(90, 180, 140)
)

type boxed interface {
	SetBorder(bool) *tview.Box
	SetTitle(string) *tview.Box
	SetTitleAlign(int) *tview.Box
	SetTitleColor(tcell.Color) *tview.Box
	SetBorderColor(tcell.Color) *tview.Box
	SetBackgroundColor(tcell.Color) *tview.Box
}

// Paint gives a panel the standard border and title.
func Paint(b boxed, title string) {
	b.SetBorder(true)
	if t := strings.TrimSpace(title); t != "" {
		b.SetTitle(" " + t + " ")
	}
	b.SetTitleAlign(tview.AlignLeft)
	b.SetTitleColor(Amber)
	b.SetBorderColor(Line)
	b.SetBackgroundColor(Panel)
}

// StyleForm paints a form and its fields and buttons.
func StyleForm(f *tview.Form, title string) *tview.Form {
	Paint(f, title)
	f.SetLabelColor(Label)
	f.SetFieldBackgroundColor(Field)
	f.SetFieldTextColor(Text)
	f.SetButtonBackgroundColor(Btn)
	f.SetButtonTextColor(tcell.ColorWhite)
	f.SetButtonActivatedStyle(tcell.StyleDefault.Background(Amber).Foreground(Ink))
	return f
}

// StyleModal paints a question or message. The background stays dark.
// A warning uses a rust border, a success a green one. Neither fills the box.
func StyleModal(m *tview.Modal, title string) *tview.Modal {
	return tone(m, title, Line, Amber)
}

func StyleModalWarn(m *tview.Modal, title string) *tview.Modal {
	return tone(m, title, Warn, Warn)
}

func StyleModalOK(m *tview.Modal, title string) *tview.Modal {
	return tone(m, title, Ok, Ok)
}

func tone(m *tview.Modal, title string, border, titleColor tcell.Color) *tview.Modal {
	m.SetBorder(true)
	if t := strings.TrimSpace(title); t != "" {
		m.SetTitle(" " + t + " ")
	}
	m.SetTitleAlign(tview.AlignLeft)
	m.SetTitleColor(titleColor)
	m.SetBorderColor(border)
	m.SetBackgroundColor(Panel)
	m.SetTextColor(Text)
	m.SetButtonBackgroundColor(Btn)
	m.SetButtonTextColor(tcell.ColorWhite)
	m.SetButtonActivatedStyle(tcell.StyleDefault.Background(Amber).Foreground(Ink))
	return m
}

// StyleList paints a list the same way as a form.
func StyleList(l *tview.List, title string) *tview.List {
	Paint(l, title)
	l.SetSelectedBackgroundColor(Btn)
	l.SetSelectedTextColor(tcell.ColorWhite)
	l.SetMainTextColor(Text)
	l.SetSecondaryTextColor(Label)
	return l
}
