package file

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"gosheet/internal/services/fileop"
	"gosheet/internal/services/uitheme"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowOptions edits the settings that survive a restart.
func ShowOptions(app *tview.Application, returnTo tview.Primitive, onClose func()) {
	current := utils.LoadSettings()

	form := tview.NewForm()
	form.AddInputField("Decimal places:", strconv.Itoa(current.DecimalPoints), 4, digitsOnly, nil)
	form.AddDropDown("Thousands:", runeOptions(utils.Separators), runeIndex(utils.Separators, current.Thousands), nil)
	form.AddDropDown("Decimal mark:", runeOptions(utils.DecimalSeparators), runeIndex(utils.DecimalSeparators, current.DecimalSep), nil)
	form.AddDropDown("Currency:", runeOptions(utils.FinancialSigns), signIndex(current.Currency), nil)

	formats := make([]string, len(FileFormats))
	formatIndex := 0
	for i, f := range FileFormats {
		formats[i] = f.Extension + "  " + f.Description
		if strings.EqualFold(f.Extension, current.SaveFormat) {
			formatIndex = i
		}
	}
	form.AddDropDown("Save format:", formats, formatIndex, nil)
	form.AddInputField("Column width:", strconv.Itoa(current.ColumnWidth), 4, digitsOnly, nil)
	form.AddInputField("Recent files:", strconv.Itoa(current.RecentFiles), 4, digitsOnly, nil)
	form.AddInputField("Autosave (minutes):", strconv.Itoa(current.AutosaveMins), 4, digitsOnly, nil)

	closeBack := func() {
		if onClose != nil {
			onClose()
		}
		app.SetRoot(returnTo, true).SetFocus(returnTo)
	}

	form.AddButton("Save", func() {
		next, err := readOptions(form)
		if err != nil {
			uiWarn(app, form, err.Error())
			return
		}
		if err := utils.SaveSettings(next); err != nil {
			uiWarn(app, form, "Could not save options:\n"+err.Error())
			return
		}
		fileop.TrimRecentFiles()
		closeBack()
	})
	form.AddButton("Clear recent", func() {
		modal := tview.NewModal().
			SetText("Clear the recent file list?").
			AddButtons([]string{"Clear", "Cancel"}).
			SetDoneFunc(func(_ int, label string) {
				if label == "Clear" {
					if err := fileop.ClearRecentFiles(); err != nil {
						uiWarn(app, form, "Could not clear recent files:\n"+err.Error())
						return
					}
				}
				app.SetRoot(form, true).SetFocus(form)
			})
		uitheme.StyleModalWarn(modal, "Recent files")
		app.SetRoot(modal, true).SetFocus(modal)
	})
	form.AddButton("Cancel", closeBack)

	uitheme.StyleForm(form, "Options  ·  autosave interval takes effect after restart")
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeBack()
			return nil
		}
		return event
	})
	app.SetRoot(form, true).SetFocus(form)
}

func readOptions(form *tview.Form) (utils.Settings, error) {
	var s utils.Settings
	points, err := strconv.Atoi(strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText()))
	if err != nil || points < 0 || points > 8 {
		return s, fmt.Errorf("Decimal places must be a number from 0 to 8.")
	}
	s.DecimalPoints = points

	_, thousands := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	_, decimal := form.GetFormItem(2).(*tview.DropDown).GetCurrentOption()
	_, currency := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
	s.Thousands = optionRune(thousands)
	s.DecimalSep = optionRune(decimal)
	s.Currency = optionRune(currency)
	if s.DecimalSep == "" || s.DecimalSep == "Ø" {
		return s, fmt.Errorf("Pick a decimal mark. None is only allowed for the thousands separator.")
	}
	if s.Thousands == s.DecimalSep {
		return s, fmt.Errorf("The thousands separator and the decimal mark have to be different.")
	}
	if s.Currency == "" || s.Currency == "Ø" {
		return s, fmt.Errorf("Pick a currency sign.")
	}

	formatIdx, _ := form.GetFormItem(4).(*tview.DropDown).GetCurrentOption()
	if formatIdx < 0 || formatIdx >= len(FileFormats) {
		return s, fmt.Errorf("Pick a save format.")
	}
	s.SaveFormat = FileFormats[formatIdx].Extension

	width, err := strconv.Atoi(strings.TrimSpace(form.GetFormItem(5).(*tview.InputField).GetText()))
	if err != nil || width < 4 || width > 40 {
		return s, fmt.Errorf("Column width must be a number from 4 to 40.")
	}
	s.ColumnWidth = width

	recent, err := strconv.Atoi(strings.TrimSpace(form.GetFormItem(6).(*tview.InputField).GetText()))
	if err != nil || recent < 1 || recent > 20 {
		return s, fmt.Errorf("Recent files must be a number from 1 to 20.")
	}
	s.RecentFiles = recent

	autosave, err := strconv.Atoi(strings.TrimSpace(form.GetFormItem(7).(*tview.InputField).GetText()))
	if err != nil || autosave < 1 || autosave > 60 {
		return s, fmt.Errorf("Autosave interval must be a number from 1 to 60 minutes.")
	}
	s.AutosaveMins = autosave
	return s, nil
}

func runeOptions(list []rune) []string {
	out := make([]string, len(list))
	for i, r := range list {
		if r == 'Ø' {
			out[i] = "None"
		} else {
			out[i] = string(r)
		}
	}
	return out
}

func runeIndex(list []rune, current string) int {
	want := optionRune(current)
	if want == "" {
		want = "Ø"
	}
	r, _ := utf8.DecodeRuneInString(want)
	for i, item := range list {
		if item == r {
			return i
		}
	}
	return 0
}

func signIndex(current string) int {
	r, _ := utf8.DecodeRuneInString(current)
	for i, item := range utils.FinancialSigns {
		if item == r {
			return i
		}
	}
	return 0
}

func optionRune(label string) string {
	if label == "None" || label == "Ø" {
		return "Ø"
	}
	r, size := utf8.DecodeRuneInString(label)
	if r == utf8.RuneError || size == 0 {
		return ""
	}
	return string(r)
}

func digitsOnly(text string, ch rune) bool {
	if ch == 0 {
		return len(text) <= 2
	}
	return ch >= '0' && ch <= '9' && len(text) <= 2
}

func saveFormatIndex(name string) int {
	lower := strings.ToLower(name)
	for i, f := range FileFormats {
		if f.Extension != "" && strings.HasSuffix(lower, strings.ToLower(f.Extension)) {
			return i
		}
	}
	for i, f := range FileFormats {
		if strings.EqualFold(f.Extension, utils.DefaultSaveFormat) {
			return i
		}
	}
	return 0
}

func uiWarn(app *tview.Application, back tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) {
			app.SetRoot(back, true).SetFocus(back)
		})
	uitheme.StyleModalWarn(modal, "Options")
	app.SetRoot(modal, true).SetFocus(modal)
}
