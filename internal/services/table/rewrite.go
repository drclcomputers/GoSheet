// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// rewrite.go parses formulas into sheet-aware references and compact ranges.
// Ranges stay as one matrix argument so SUM(A1:A10000) does not allocate
// 10,000 expression nodes or empty cells.

package table

import (
	"fmt"
	"strconv"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"gosheet/internal/utils/evaluatefuncs"
)

const maxRangeCells = 250000

type addr struct {
	sheet string
	row   int32
	col   int32
}

type pieceKind int

const (
	pieceRaw pieceKind = iota
	pieceCell
	pieceRange
	pieceBroken
)

type piece struct {
	kind pieceKind
	raw  string
	a    addr
	b    addr
}

func parseFormula(formula string) ([]piece, error) {
	var pieces []piece
	s := formula
	i := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "#REF!") {
			pieces = append(pieces, piece{kind: pieceBroken})
			i += len("#REF!")
			continue
		}
		ch := s[i]
		if ch == '"' {
			j := i + 1
			var b strings.Builder
			for j < len(s) {
				if s[j] == '\\' && j+1 < len(s) {
					b.WriteByte(s[j+1])
					j += 2
					continue
				}
				if s[j] == '"' {
					if j+1 < len(s) && s[j+1] == '"' {
						b.WriteByte('"')
						j += 2
						continue
					}
					break
				}
				b.WriteByte(s[j])
				j++
			}
			if j >= len(s) {
				return nil, fmt.Errorf("unterminated string")
			}
			pieces = append(pieces, piece{kind: pieceRaw, raw: strconv.Quote(b.String())})
			i = j + 1
			continue
		}

		if ch == '<' && i+1 < len(s) && s[i+1] == '>' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "!="})
			i += 2
			continue
		}
		if ch == '<' && i+1 < len(s) && s[i+1] == '=' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "<="})
			i += 2
			continue
		}
		if ch == '>' && i+1 < len(s) && s[i+1] == '=' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: ">="})
			i += 2
			continue
		}
		if ch == '!' && i+1 < len(s) && s[i+1] == '=' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "!="})
			i += 2
			continue
		}
		if ch == '=' && i+1 < len(s) && s[i+1] == '=' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "=="})
			i += 2
			continue
		}
		if ch == '=' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "=="})
			i++
			continue
		}
		if ch == '^' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "**"})
			i++
			continue
		}
		if ch == '%' {
			pieces = append(pieces, piece{kind: pieceRaw, raw: "/100"})
			i++
			continue
		}

		if ch == '\'' || ch == '$' || isIdentStart(ch) {
			p, next, ok, err := parseRefToken(s, i)
			if err != nil {
				return nil, err
			}
			if ok {
				pieces = append(pieces, p)
				i = next
				continue
			}
		}

		if isIdentStart(ch) {
			word, next := readWord(s, i)
			if next < len(s) && s[next] == '(' {
				pieces = append(pieces, piece{kind: pieceRaw, raw: strings.ToUpper(word)})
				i = next
				continue
			}
			up := strings.ToUpper(word)
			switch up {
			case "TRUE":
				pieces = append(pieces, piece{kind: pieceRaw, raw: "true"})
			case "FALSE":
				pieces = append(pieces, piece{kind: pieceRaw, raw: "false"})
			default:
				pieces = append(pieces, piece{kind: pieceRaw, raw: up})
			}
			i = next
			continue
		}

		pieces = append(pieces, piece{kind: pieceRaw, raw: string(ch)})
		i++
	}
	return pieces, nil
}

func parseRefToken(s string, i int) (piece, int, bool, error) {
	start := i
	sheet := ""
	explicit := false

	if s[i] == '\'' {
		j := i + 1
		var name strings.Builder
		for j < len(s) {
			if s[j] == '\'' {
				if j+1 < len(s) && s[j+1] == '\'' {
					name.WriteByte('\'')
					j += 2
					continue
				}
				break
			}
			name.WriteByte(s[j])
			j++
		}
		if j >= len(s) || s[j] != '\'' || j+1 >= len(s) || s[j+1] != '!' {
			return piece{}, start, false, fmt.Errorf("invalid sheet name")
		}
		sheet = name.String()
		explicit = true
		i = j + 2
	} else if isIdentStart(s[i]) {
		word, j := readWord(s, i)
		if j < len(s) && s[j] == '!' {
			sheet = word
			explicit = true
			i = j + 1
		}
	}

	a, next, ok := parseA1(s, i)
	if !ok {
		if explicit {
			return piece{}, start, false, fmt.Errorf("invalid reference")
		}
		return piece{}, start, false, nil
	}
	// LOG10( is a function, not column LOG row 10.
	if !explicit && next < len(s) && s[next] == '(' {
		return piece{}, start, false, nil
	}
	a.sheet = sheet

	k := next
	for k < len(s) && (s[k] == ' ' || s[k] == '\t') {
		k++
	}
	if k < len(s) && s[k] == ':' {
		k++
		for k < len(s) && (s[k] == ' ' || s[k] == '\t') {
			k++
		}
		b, bend, bok, err := parseMaybeSheetAddr(s, k)
		if err != nil {
			return piece{}, start, false, err
		}
		if !bok {
			return piece{}, start, false, fmt.Errorf("invalid range")
		}
		if b.sheet == "" {
			b.sheet = a.sheet
		}
		return piece{kind: pieceRange, a: a, b: b}, bend, true, nil
	}
	return piece{kind: pieceCell, a: a}, next, true, nil
}

func parseMaybeSheetAddr(s string, i int) (addr, int, bool, error) {
	if i >= len(s) {
		return addr{}, i, false, nil
	}
	p, next, ok, err := parseRefToken(s, i)
	if err != nil || !ok || p.kind != pieceCell {
		if err != nil {
			return addr{}, i, false, err
		}
		return addr{}, i, false, nil
	}
	return p.a, next, true, nil
}

func parseA1(s string, i int) (addr, int, bool) {
	if i < len(s) && s[i] == '$' {
		i++
	}
	if i >= len(s) || !isLetter(s[i]) {
		return addr{}, i, false
	}
	colStart := i
	for i < len(s) && isLetter(s[i]) {
		i++
	}
	letters := strings.ToUpper(s[colStart:i])
	if i < len(s) && s[i] == '$' {
		i++
	}
	if i >= len(s) || !isDigit(s[i]) {
		return addr{}, i, false
	}
	rowStart := i
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	row, err := strconv.Atoi(s[rowStart:i])
	if err != nil || row < 1 {
		return addr{}, i, false
	}
	col := utils.ColumnNumber(letters)
	if col < 1 {
		return addr{}, i, false
	}
	return addr{row: int32(row), col: int32(col)}, i, true
}

func readWord(s string, i int) (string, int) {
	j := i
	if j < len(s) && isIdentStart(s[j]) {
		j++
	}
	for j < len(s) && isIdentCont(s[j]) {
		j++
	}
	return s[i:j], j
}

func isLetter(b byte) bool { return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') }
func isDigit(b byte) bool  { return b >= '0' && b <= '9' }
func isIdentStart(b byte) bool {
	return isLetter(b) || b == '_'
}
func isIdentCont(b byte) bool {
	return isLetter(b) || isDigit(b) || b == '_'
}

func quoteSheet(name string) string {
	if name == "" {
		return name
	}
	for i := 0; i < len(name); i++ {
		needsQuote := !isIdentCont(name[i]) || (i == 0 && !isIdentStart(name[i]))
		if needsQuote {
			return "'" + strings.ReplaceAll(name, "'", "''") + "'"
		}
	}
	return name
}

func unquoteSheet(name string) string {
	name = strings.TrimSpace(name)
	if len(name) >= 2 && name[0] == '\'' && name[len(name)-1] == '\'' {
		return strings.ReplaceAll(name[1:len(name)-1], "''", "'")
	}
	return name
}

func qualify(a addr) string {
	name := a.sheet
	if name == "" {
		if sh := activeSheet(); sh != nil {
			name = sh.Name
		}
	} else if sh := sheetByName(name); sh != nil {
		name = sh.Name
	}
	return quoteSheet(name) + "!" + utils.FormatCellRef(a.row, a.col)
}

func qualifyRange(a, b addr) string {
	aa, bb := normalizeRange(a, b)
	name := aa.sheet
	if name == "" {
		if sh := activeSheet(); sh != nil {
			name = sh.Name
		}
	} else if sh := sheetByName(name); sh != nil {
		name = sh.Name
	}
	return fmt.Sprintf("%s!%s:%s", quoteSheet(name), utils.FormatCellRef(aa.row, aa.col), utils.FormatCellRef(bb.row, bb.col))
}

func normalizeRange(a, b addr) (addr, addr) {
	if a.row > b.row {
		a.row, b.row = b.row, a.row
	}
	if a.col > b.col {
		a.col, b.col = b.col, a.col
	}
	if b.sheet == "" {
		b.sheet = a.sheet
	}
	if a.sheet == "" {
		a.sheet = b.sheet
	}
	return a, b
}

func depsFromPieces(pieces []piece) (cells []string, ranges []string) {
	seenC := map[string]bool{}
	seenR := map[string]bool{}
	for _, p := range pieces {
		switch p.kind {
		case pieceCell:
			q := qualify(p.a)
			if !seenC[q] {
				seenC[q] = true
				cells = append(cells, q)
			}
		case pieceRange:
			q := qualifyRange(p.a, p.b)
			if !seenR[q] {
				seenR[q] = true
				ranges = append(ranges, q)
			}
		}
	}
	return cells, ranges
}

func bindPieces(pieces []piece, tableCellValue func(addr) (any, error), tableRange func(addr, addr) (*evaluatefuncs.Matrix, error)) (string, map[string]any, error) {
	params := map[string]any{}
	var b strings.Builder
	for _, p := range pieces {
		switch p.kind {
		case pieceRaw:
			b.WriteString(p.raw)
		case pieceCell:
			val, err := tableCellValue(p.a)
			if err != nil {
				return "", nil, err
			}
			name := fmt.Sprintf("C%d", len(params))
			params[name] = val
			b.WriteString(name)
		case pieceRange:
			m, err := tableRange(p.a, p.b)
			if err != nil {
				return "", nil, err
			}
			name := fmt.Sprintf("R%d", len(params))
			params[name] = m
			b.WriteString(name)
		case pieceBroken:
			return "", nil, fmt.Errorf("sheet not found")
		}
	}
	return b.String(), params, nil
}

func activeSheet() *Sheet {
	if globalWorkbook == nil {
		return nil
	}
	return globalWorkbook.GetActiveSheet()
}

func sheetByName(name string) *Sheet {
	if globalWorkbook == nil {
		return nil
	}
	if strings.TrimSpace(name) == "" {
		return globalWorkbook.GetActiveSheet()
	}
	for _, sh := range globalWorkbook.Sheets {
		if sh != nil && strings.EqualFold(sh.Name, name) {
			return sh
		}
	}
	return nil
}

func dataForSheet(name string) map[[2]int]*cell.Cell {
	sh := sheetByName(name)
	if sh == nil {
		return nil
	}
	return sh.Data
}

// SplitRef parses "A1", "$B$2", "Sheet2!A1" or "'My Sheet'!A1".
func SplitRef(ref string) (sheet string, row, col int32, err error) {
	ref = strings.TrimSpace(ref)
	p, next, ok, perr := parseRefToken(ref, 0)
	if perr != nil {
		return "", 0, 0, perr
	}
	if !ok || p.kind != pieceCell || next != len(ref) {
		// Also accept a bare A1 that parseRefToken might have rejected? It accepts A1.
		return "", 0, 0, fmt.Errorf("invalid reference %q", ref)
	}
	return p.a.sheet, p.a.row, p.a.col, nil
}

func sheetNameOf(c *cell.Cell) string {
	if c == nil || globalWorkbook == nil {
		return ""
	}
	key := [2]int{int(c.Row), int(c.Column)}
	for _, sh := range globalWorkbook.Sheets {
		if sh != nil && sh.Data[key] == c {
			return sh.Name
		}
	}
	if sh := globalWorkbook.GetActiveSheet(); sh != nil {
		return sh.Name
	}
	return ""
}

func qualifiedCell(c *cell.Cell) string {
	return quoteSheet(sheetNameOf(c)) + "!" + utils.FormatCellRef(c.Row, c.Column)
}

func activateSheetOf(c *cell.Cell) func() {
	if globalWorkbook == nil || c == nil {
		return func() {}
	}
	key := [2]int{int(c.Row), int(c.Column)}
	prev := globalWorkbook.ActiveSheet
	for i, sh := range globalWorkbook.Sheets {
		if sh != nil && sh.Data[key] == c {
			globalWorkbook.ActiveSheet = i
			break
		}
	}
	return func() { globalWorkbook.ActiveSheet = prev }
}

func ptrSlice(vals []string) []*string {
	out := make([]*string, len(vals))
	for i, v := range vals {
		copyV := v
		out[i] = &copyV
	}
	return out
}

func cellValueForFormula(c *cell.Cell, eval func(*cell.Cell) error) (any, error) {
	if c == nil || c.RawValue == nil || strings.TrimSpace(*c.RawValue) == "" {
		return 0.0, nil
	}
	if c.IsFormula() && !c.HasFlag(cell.FlagEvaluated) {
		if err := eval(c); err != nil && (c.Display == nil || !strings.HasPrefix(*c.Display, "#")) {
			return nil, err
		}
	}
	if c.Display == nil {
		return 0.0, nil
	}
	val := strings.TrimSpace(*c.Display)
	if val == "" {
		return 0.0, nil
	}
	if strings.HasPrefix(val, "#") {
		return nil, fmt.Errorf("%s", val)
	}
	cleaned := val
	if c.FinancialSign != 0 {
		cleaned = strings.TrimPrefix(cleaned, string(c.FinancialSign))
	}
	if c.ThousandsSeparator != 0 {
		cleaned = strings.ReplaceAll(cleaned, string(c.ThousandsSeparator), "")
	}
	if c.DecimalSeparator != 0 && c.DecimalSeparator != '.' {
		cleaned = strings.ReplaceAll(cleaned, string(c.DecimalSeparator), ".")
	}
	if num, err := strconv.ParseFloat(strings.TrimSpace(cleaned), 64); err == nil {
		return num, nil
	}
	if strings.EqualFold(val, "TRUE") {
		return true, nil
	}
	if strings.EqualFold(val, "FALSE") {
		return false, nil
	}
	return val, nil
}

func readRange(a, b addr, eval func(*cell.Cell) error) (*evaluatefuncs.Matrix, error) {
	a, b = normalizeRange(a, b)
	data := dataForSheet(a.sheet)
	if data == nil {
		return nil, fmt.Errorf("sheet not found")
	}
	rows := int(b.row-a.row) + 1
	cols := int(b.col-a.col) + 1
	if rows < 1 || cols < 1 || rows*cols > maxRangeCells {
		return nil, fmt.Errorf("range too large")
	}
	m := &evaluatefuncs.Matrix{
		Rows:  rows,
		Cols:  cols,
		Cells: make([]any, rows*cols),
	}
	for r := a.row; r <= b.row; r++ {
		for c := a.col; c <= b.col; c++ {
			var v any
			if cellData := data[[2]int{int(r), int(c)}]; cellData != nil {
				got, err := cellValueForFormula(cellData, eval)
				if err != nil {
					return nil, err
				}
				v = got
			}
			m.Cells[(int(r-a.row))*cols+int(c-a.col)] = v
		}
	}
	return m, nil
}

func parseRangeDep(dep string) (sheet string, r1, c1, r2, c2 int32, ok bool) {
	dep = strings.TrimSpace(dep)
	bang := strings.LastIndex(dep, "!")
	if bang < 0 {
		return "", 0, 0, 0, 0, false
	}
	sheet = unquoteSheet(dep[:bang])
	body := dep[bang+1:]
	parts := strings.Split(body, ":")
	if len(parts) != 2 {
		return "", 0, 0, 0, 0, false
	}
	p1, n1, ok1 := parseA1(parts[0], 0)
	p2, n2, ok2 := parseA1(parts[1], 0)
	if !ok1 || !ok2 || n1 != len(parts[0]) || n2 != len(parts[1]) {
		return "", 0, 0, 0, 0, false
	}
	if p1.row > p2.row {
		p1.row, p2.row = p2.row, p1.row
	}
	if p1.col > p2.col {
		p1.col, p2.col = p2.col, p1.col
	}
	return sheet, p1.row, p1.col, p2.row, p2.col, true
}

func rangeCovers(dep, sheet string, row, col int32) bool {
	sh, r1, c1, r2, c2, ok := parseRangeDep(dep)
	if !ok {
		return false
	}
	if !strings.EqualFold(sh, sheet) {
		return false
	}
	return row >= r1 && row <= r2 && col >= c1 && col <= c2
}
func sheetPrefix(s string, i int) (name string, afterBang int, ok bool) {
	if i >= len(s) {
		return "", i, false
	}
	if s[i] == '\'' {
		j := i + 1
		var b strings.Builder
		for j < len(s) {
			if s[j] == '\'' {
				if j+1 < len(s) && s[j+1] == '\'' {
					b.WriteByte('\'')
					j += 2
					continue
				}
				break
			}
			b.WriteByte(s[j])
			j++
		}
		if j >= len(s) || s[j] != '\'' || j+1 >= len(s) || s[j+1] != '!' {
			return "", i, false
		}
		return b.String(), j + 2, true
	}
	if !isIdentStart(s[i]) {
		return "", i, false
	}
	word, j := readWord(s, i)
	if j >= len(s) || s[j] != '!' {
		return "", i, false
	}
	return word, j + 1, true
}

func appendStringLiteral(s string, i int, b *strings.Builder) int {
	b.WriteByte(s[i])
	j := i + 1
	for j < len(s) {
		if s[j] == '\\' && j+1 < len(s) {
			b.WriteByte(s[j])
			b.WriteByte(s[j+1])
			j += 2
			continue
		}
		b.WriteByte(s[j])
		if s[j] == '"' {
			if j+1 < len(s) && s[j+1] == '"' {
				b.WriteByte('"')
				j += 2
				continue
			}
			return j + 1
		}
		j++
	}
	return j
}

// rewriteSheetRefs renames or deletes explicit sheet references in a formula.
// newName empty means the sheet was deleted and the reference becomes #REF!.
// Text inside quotes is left alone. A1 with no sheet name is left alone.
func rewriteSheetRefs(formula, oldName, newName string) (string, bool) {
	if oldName == "" || strings.EqualFold(oldName, newName) && newName != "" && oldName == newName {
		return formula, false
	}
	var b strings.Builder
	changed := false
	s := formula
	i := 0
	for i < len(s) {
		if s[i] == '"' {
			i = appendStringLiteral(s, i, &b)
			continue
		}
		name, after, ok := sheetPrefix(s, i)
		if ok && strings.EqualFold(name, oldName) {
			if newName == "" {
				if p, end, pok, err := parseRefToken(s, i); err == nil && pok && (p.kind == pieceCell || p.kind == pieceRange) {
					b.WriteString("#REF!")
					i = end
					changed = true
					continue
				}
			} else {
				b.WriteString(quoteSheet(newName))
				b.WriteByte('!')
				i = after
				changed = true
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	if !changed {
		return formula, false
	}
	return b.String(), true
}

func (wb *Workbook) rewriteFormulas(oldName, newName string, skip *Sheet) {
	if wb == nil {
		return
	}
	for _, sh := range wb.Sheets {
		if sh == nil || sh == skip {
			continue
		}
		for _, c := range sh.Data {
			if c == nil || c.RawValue == nil || !c.IsFormula() {
				continue
			}
			next, ok := rewriteSheetRefs(*c.RawValue, oldName, newName)
			if !ok {
				continue
			}
			s := next
			c.RawValue = &s
			c.ClearFlag(cell.FlagEvaluated)
		}
	}
}

func (wb *Workbook) scrubDeadDependents() {
	if wb == nil {
		return
	}
	for _, sh := range wb.Sheets {
		if sh == nil {
			continue
		}
		for _, c := range sh.Data {
			if c == nil || len(c.Dependents) == 0 {
				continue
			}
			kept := c.Dependents[:0]
			for _, d := range c.Dependents {
				if d == nil {
					continue
				}
				sheet, _, _, err := SplitRef(*d)
				if err != nil {
					continue
				}
				if sheet != "" && sheetByName(sheet) == nil {
					continue
				}
				kept = append(kept, d)
			}
			c.Dependents = kept
		}
	}
}

func (wb *Workbook) recalculateAll() {
	if wb == nil {
		return
	}
	for _, sh := range wb.Sheets {
		if sh == nil {
			continue
		}
		for _, c := range sh.Data {
			if c != nil && c.IsFormula() {
				c.ClearFlag(cell.FlagEvaluated)
			}
		}
	}
	for _, sh := range wb.Sheets {
		if sh == nil {
			continue
		}
		for _, c := range sh.Data {
			if c != nil && c.IsFormula() && !c.HasFlag(cell.FlagEvaluated) {
				_ = EvaluateCell(nil, c)
			}
		}
	}
}
