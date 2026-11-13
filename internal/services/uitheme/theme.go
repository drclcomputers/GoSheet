// Package uitheme is the shared look of the dialogs: dark panels, a teal
// border, amber titles. Screens that were restyled later should use this
// instead of the old yellow, red, and green frames.
package uitheme

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	HexInk      = "#0e111a"
	HexPanel    = "#121622"
	HexField    = "#0c0e16"
	HexLine     = "#46a0aa"
	HexAmber    = "#f0d090"
	HexLabel    = "#a8bcd6"
	HexMuted    = "#8aa0b8"
	HexDim      = "#5c6b7a"
	HexText     = "#e8ecf4"
	HexWhite    = "#ffffff"
	HexBtn      = "#2463a6"
	HexHeader   = "#20283a"
	HexCyan     = "#5ec8c8"
	HexSky      = "#8ecae6"
	HexMint     = "#7dcea0"
	HexAmberDim = "#a08246"
	HexOk       = "#5ab48c"
	HexWarn     = "#ff4a00"
)

type Swatch struct {
	Name  string
	Value string
}

var Palette = []Swatch{
	{"Default", ""},
	{"Black", "#0e111a"},
	{"DkGray", "#3a4252"},
	{"Gray", "#8a94a6"},
	{"LtGray", "#b8c0cc"},
	{"White", "#e8ecf4"},
	{"DkRed", "#8b2f3a"},
	{"Red", "#e06c75"},
	{"Rose", "#e060a0"},
	{"Pink", "#f5a3c7"},
	{"Orange", "#d19a66"},
	{"Amber", "#e8a86b"},
	{"Yellow", "#f0d090"},
	{"Cream", "#f5e6c8"},
	{"DkGreen", "#2f6b4f"},
	{"Green", "#5ab48c"},
	{"Mint", "#a3e0c0"},
	{"Teal", "#2f8b94"},
	{"Cyan", "#56b6c2"},
	{"LtCyan", "#a8d4dc"},
	{"Navy", "#2a3a6a"},
	{"Blue", "#61afef"},
	{"Sky", "#a3c8f0"},
	{"DkPurple", "#5a3a7a"},
	{"Purple", "#c678dd"},
	{"Violet", "#e0a3f0"},
	{"Olive", "#a3a83a"},
	{"Brown", "#8b6a4a"},
	{"Gold", "#d4b04a"},
}

var (
	Ink      = ParseHex(HexInk)
	Panel    = ParseHex(HexPanel)
	Field    = ParseHex(HexField)
	Line     = ParseHex(HexLine)
	Amber    = ParseHex(HexAmber)
	AmberDim = ParseHex(HexAmberDim)
	Label    = ParseHex(HexLabel)
	Muted    = ParseHex(HexMuted)
	Text     = ParseHex(HexText)
	Btn      = ParseHex(HexBtn)
	Header   = ParseHex(HexHeader)
	Sky      = ParseHex(HexSky)
	Mint     = ParseHex(HexMint)
	Warn     = ParseHex(HexWarn)
	Ok       = ParseHex(HexOk)
)

func init() {
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    Ink,
		ContrastBackgroundColor:     Btn,
		MoreContrastBackgroundColor: Amber,
		BorderColor:                 Line,
		TitleColor:                  Amber,
		GraphicsColor:               Line,
		PrimaryTextColor:            Text,
		SecondaryTextColor:          Label,
		TertiaryTextColor:           Muted,
		InverseTextColor:            Text,
		ContrastSecondaryTextColor:  Text,
	}
}

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
	b.SetTitleAlign(tview.AlignCenter)
	b.SetTitleColor(Amber)
	b.SetBorderColor(Line)
	b.SetBackgroundColor(Panel)
}

// CenteredModal places a fixed-size primitive in the middle of the terminal.
func CenteredModal(p tview.Primitive, width, height int) *tview.Flex {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

// StyleForm paints a form and its fields and buttons.
func StyleForm(f *tview.Form, title string) *tview.Form {
	Paint(f, title)
	f.SetBackgroundColor(Panel)
	f.SetLabelColor(Label)
	f.SetFieldBackgroundColor(Field)
	f.SetFieldTextColor(Text)
	f.SetButtonBackgroundColor(Btn)
	f.SetButtonTextColor(Text)
	f.SetButtonActivatedStyle(tcell.StyleDefault.Background(Amber).Foreground(Ink))
	for i := 0; i < f.GetFormItemCount(); i++ {
		if item, ok := f.GetFormItem(i).(interface {
			SetBackgroundColor(tcell.Color) *tview.Box
		}); ok {
			item.SetBackgroundColor(Panel)
		}
	}
	return f
}

// StyleButton applies the shared button colors.
func StyleButton(b *tview.Button) *tview.Button {
	b.SetBackgroundColor(Btn)
	b.SetLabelColor(Text)
	b.SetBackgroundColorActivated(Amber)
	b.SetLabelColorActivated(Ink)
	return b
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
	m.SetTitleAlign(tview.AlignCenter)
	m.SetTitleColor(titleColor)
	m.SetBorderColor(border)
	m.SetBackgroundColor(Panel)
	m.SetTextColor(Text)
	m.SetButtonBackgroundColor(Btn)
	m.SetButtonTextColor(Text)
	m.SetButtonActivatedStyle(tcell.StyleDefault.Background(Amber).Foreground(Ink))
	return m
}

// StyleList paints a list the same way as a form.
func StyleList(l *tview.List, title string) *tview.List {
	Paint(l, title)
	l.SetSelectedBackgroundColor(Btn)
	l.SetSelectedTextColor(Text)
	l.SetMainTextStyle(tcell.StyleDefault.Foreground(Text).Background(Panel))
	l.SetSecondaryTextStyle(tcell.StyleDefault.Foreground(Label).Background(Panel))
	l.SetShortcutStyle(tcell.StyleDefault.Foreground(Label).Background(Panel))
	l.SetBackgroundColor(Panel)
	return l
}

// Markup replaces palette placeholders in dynamic-color text with theme colors.
func Markup(s string) string {
	return strings.NewReplacer(
		"{key}", "["+HexCyan+"]",
		"{val}", "["+HexAmber+"]",
		"{ok}", "["+HexMint+"]",
		"{plain}", "["+HexText+"]",
		"{bright}", "["+HexWhite+"]",
		"{h1}", "["+HexAmber+"::b]",
		"{h2}", "["+HexMint+"::b]",
		"{h3}", "["+HexSky+"::b]",
		"{muted}", "["+HexMuted+"]",
		"{dim}", "["+HexDim+"]",
		"{version}", "["+HexAmber+"]",
		"{end}", "[::-]",
		"{clear}", "["+HexText+"]",
	).Replace(s)
}

// ParseHex converts a #RRGGBB color to a tcell color; invalid values use Ink.
func ParseHex(s string) tcell.Color {
	if len(s) != 7 || s[0] != '#' {
		s = HexInk
	}
	value, err := strconv.ParseUint(s[1:], 16, 24)
	if err != nil {
		value, _ = strconv.ParseUint(HexInk[1:], 16, 24)
	}
	return tcell.NewRGBColor(int32(value>>16), int32((value>>8)&0xff), int32(value&0xff))
}

// Contrast returns readable text for a foreground placed over c.
func Contrast(c tcell.Color) tcell.Color {
	r, g, b := c.RGB()
	if 0.299*float64(r)+0.587*float64(g)+0.114*float64(b) > 140 {
		return Ink
	}
	return Text
}
