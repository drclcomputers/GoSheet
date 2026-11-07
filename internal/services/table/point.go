// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// point.go inserts cell references while a formula is being typed.
// Arrows move the grid. Alt+PageUp / Alt+PageDown (and Alt+1..9) change
// sheets, and the reference is written as Prices!B2 or 'Q4 Budget'!A1.

package table

import (
	"strings"

	"gosheet/internal/services/sheetview"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pointSession struct {
	live             bool
	active           bool
	home             string
	refStart, refEnd int
}

var pointing pointSession

const pointStatus = " [#f0d090]pointing[-]   arrows pick a cell    [#f0d090]Alt+PgUp/PgDn[-] other sheet    [#f0d090]Enter[-] accept    [#f0d090]Esc[-] cancel "

func armPoint(text string) {
	home := ""
	if sh := activeSheet(); sh != nil {
		home = sh.Name
	}
	pointing = pointSession{live: true, home: home}
	if isFormulaText(text) {
		sheetview.SetStatus(pointStatus)
	}
}

func disarmPoint() {
	pointing = pointSession{}
}

func installPointing(app *tview.Application, table *tview.Table) {
	sheetview.SetEditHooks(armPoint, disarmPoint)
	sheetview.SetPointFunc(func(event *tcell.EventKey) bool {
		return handlePoint(app, table, event)
	})
}

func handlePoint(app *tview.Application, table *tview.Table, event *tcell.EventKey) bool {
	if !pointing.live || table == nil {
		return false
	}
	text, cursor := sheetview.EditText()
	if !isFormulaText(text) || inQuotes(text, cursor) {
		pointing.active = false
		return false
	}

	if sheetKey(event) {
		if !switchByKey(app, table, event) {
			return true
		}
		if pointing.active || canStartPoint(text, cursor) {
			insertPoint(table, text, cursor)
		}
		sheetview.SetStatus(pointStatus)
		return true
	}

	switch event.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		if event.Modifiers() != 0 {
			return false
		}
		if !pointing.active && !canStartPoint(text, cursor) {
			return false
		}
		nudge(table, event.Key())
		insertPoint(table, text, cursor)
		sheetview.SetStatus(pointStatus)
		return true
	default:
		pointing.active = false
		return false
	}
}

func sheetKey(event *tcell.EventKey) bool {
	if event.Modifiers()&tcell.ModAlt == 0 {
		return false
	}
	switch event.Key() {
	case tcell.KeyPgUp, tcell.KeyPgDn:
		return true
	case tcell.KeyRune:
		return event.Rune() >= '1' && event.Rune() <= '9'
	default:
		return false
	}
}

func switchByKey(app *tview.Application, table *tview.Table, event *tcell.EventKey) bool {
	if globalWorkbook == nil {
		return false
	}
	idx := globalWorkbook.ActiveSheet
	switch event.Key() {
	case tcell.KeyPgDn:
		idx++
	case tcell.KeyPgUp:
		idx--
	case tcell.KeyRune:
		idx = int(event.Rune() - '1')
	}
	if idx < 0 || idx >= len(globalWorkbook.Sheets) || idx == globalWorkbook.ActiveSheet {
		return idx == globalWorkbook.ActiveSheet
	}
	SwitchSheet(app, table, idx)
	return true
}

func insertPoint(table *tview.Table, text string, cursor int) {
	row, col, ok := selectedAbsFrom(table)
	if !ok {
		return
	}
	current := ""
	if sh := activeSheet(); sh != nil {
		current = sh.Name
	}
	ref := pointRef(pointing.home, current, row, col)
	start, end := cursor, cursor
	if pointing.active {
		start, end = pointing.refStart, pointing.refEnd
	}
	sheetview.ReplaceEdit(start, end, ref)
	pointing.active = true
	pointing.refStart = start
	pointing.refEnd = start + len(ref)
	sheetview.SetAddr(ref)
}

func nudge(table *tview.Table, key tcell.Key) {
	data := GetActiveSheetData()
	vp := GetActiveViewport()
	if table == nil || data == nil || vp == nil {
		return
	}
	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	if visualRow < 1 {
		visualRow = 1
	}
	if visualCol < 1 {
		visualCol = 1
	}
	switch key {
	case tcell.KeyUp:
		if visualRow <= 1 && vp.TopRow > 1 {
			vp.TopRow--
			RenderVisible(table, vp, data)
			table.Select(1, int(visualCol))
			return
		}
		if visualRow > 1 {
			table.Select(int(visualRow-1), int(visualCol))
		}
	case tcell.KeyDown:
		if vp.TopRow+vp.ViewRows >= utils.MAX_ROWS {
			return
		}
		if visualRow >= int32(vp.ViewRows) {
			vp.TopRow++
			RenderVisible(table, vp, data)
			table.Select(int(vp.ViewRows), int(visualCol))
			return
		}
		table.Select(int(visualRow+1), int(visualCol))
	case tcell.KeyLeft:
		if visualCol <= 1 && vp.LeftCol > 1 {
			vp.LeftCol--
			RenderVisible(table, vp, data)
			table.Select(int(visualRow), 1)
			return
		}
		if visualCol > 1 {
			table.Select(int(visualRow), int(visualCol-1))
		}
	case tcell.KeyRight:
		if vp.LeftCol+vp.ViewCols >= utils.MAX_COLS {
			return
		}
		if visualCol >= int32(vp.ViewCols) {
			vp.LeftCol++
			RenderVisible(table, vp, data)
			table.Select(int(visualRow), int(vp.ViewCols))
			return
		}
		table.Select(int(visualRow), int(visualCol+1))
	}
}

func selectedAbsFrom(table *tview.Table) (int32, int32, bool) {
	vp := GetActiveViewport()
	if table == nil || vp == nil {
		return 0, 0, false
	}
	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := vp.ToAbsolute(visualRow, visualCol)
	if row < 1 || col < 1 {
		return 0, 0, false
	}
	return row, col, true
}

func isFormulaText(text string) bool {
	s := strings.TrimLeft(text, " \t")
	return strings.HasPrefix(s, "=") || strings.HasPrefix(s, "$=")
}

func canStartPoint(text string, cursor int) bool {
	if cursor < 0 || cursor > len(text) || !isFormulaText(text) || inQuotes(text, cursor) {
		return false
	}
	i := cursor
	for i > 0 && (text[i-1] == ' ' || text[i-1] == '\t') {
		i--
	}
	if strings.HasPrefix(text, "$=") && i <= 2 {
		return i >= 2
	}
	trimmed := strings.TrimLeft(text, " \t")
	prefixAt := strings.Index(text, trimmed)
	if strings.HasPrefix(trimmed, "=") && !strings.HasPrefix(trimmed, "$=") && i <= prefixAt+1 {
		return i >= prefixAt+1
	}
	if i == 0 {
		return false
	}
	switch text[i-1] {
	case '+', '-', '*', '/', '^', '&', '(', ',', ':', '=', '<', '>', '!':
		return true
	default:
		return false
	}
}

func inQuotes(text string, cursor int) bool {
	single, double := false, false
	limit := cursor
	if limit > len(text) {
		limit = len(text)
	}
	for i := 0; i < limit; i++ {
		switch text[i] {
		case '"':
			if single {
				continue
			}
			if double && i+1 < limit && text[i+1] == '"' {
				i++
				continue
			}
			double = !double
		case '\'':
			if double {
				continue
			}
			if single && i+1 < limit && text[i+1] == '\'' {
				i++
				continue
			}
			single = !single
		}
	}
	return single || double
}

// pointRef is the token written into a formula for a pointed cell.
// The home sheet is omitted so =B2 stays local; any other sheet is qualified.
func pointRef(home, current string, row, col int32) string {
	cell := utils.FormatCellRef(row, col)
	if current == "" || strings.EqualFold(strings.TrimSpace(home), strings.TrimSpace(current)) {
		return cell
	}
	return quoteSheet(current) + "!" + cell
}
