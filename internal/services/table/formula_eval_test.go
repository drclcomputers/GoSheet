// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package table

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"gosheet/internal/services/cell"

	"github.com/rivo/tview"
)

func setupBook(names ...string) *Workbook {
	if len(names) == 0 {
		names = []string{"Sheet1"}
	}
	wb := &Workbook{Sheets: make([]*Sheet, 0, len(names)), ActiveSheet: 0}
	for _, n := range names {
		wb.Sheets = append(wb.Sheets, NewSheet(n))
	}
	globalWorkbook = wb
	return wb
}

func put(sh *Sheet, ref, raw string) *cell.Cell {
	sheetName, row, col, err := SplitRef(ref)
	if err != nil {
		panic(err)
	}
	_ = sheetName
	c := cell.NewCell(row, col, raw)
	sh.Data[[2]int{int(row), int(col)}] = c
	return c
}

func numOf(t *testing.T, c *cell.Cell) float64 {
	t.Helper()
	if c.Display == nil {
		t.Fatalf("nil display, raw %v", c.RawValue)
	}
	s := *c.Display
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "$")
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		t.Fatalf("display %q: %v", *c.Display, err)
	}
	return f
}

func TestSumRangeIsFastAndCorrect(t *testing.T) {
	wb := setupBook("Sheet1")
	sh := wb.Sheets[0]
	const n = 2000
	var want float64
	for i := 1; i <= n; i++ {
		put(sh, "A"+strconv.Itoa(i), strconv.Itoa(i))
		want += float64(i)
	}
	f := put(sh, "B1", "$=SUM(A1:A2000)")
	start := time.Now()
	if err := EvaluateCell(tview.NewTable(), f); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("SUM of %d cells took %s", n, time.Since(start))
	}
	if got := numOf(t, f); got != want {
		t.Fatalf("sum %v want %v display %s", got, want, *f.Display)
	}
	// The range must not materialize an empty cell per address.
	if len(sh.Data) != n+1 {
		t.Fatalf("data cells = %d, want %d (range should not allocate)", len(sh.Data), n+1)
	}
}

func TestPlainEqualsAndComparison(t *testing.T) {
	wb := setupBook("Sheet1")
	sh := wb.Sheets[0]
	put(sh, "A1", "10")
	f := put(sh, "B1", `=IF(A1=10,"Y","N")`)
	if err := EvaluateCell(tview.NewTable(), f); err != nil {
		t.Fatal(err)
	}
	if f.Display == nil || *f.Display != "Y" {
		t.Fatalf("got %v", f.Display)
	}
	pow := put(sh, "C1", "=2^3")
	if err := EvaluateCell(tview.NewTable(), pow); err != nil {
		t.Fatal(err)
	}
	if numOf(t, pow) != 8 {
		t.Fatalf("power display %s", *pow.Display)
	}
}

func TestCrossSheetAndVLookup(t *testing.T) {
	wb := setupBook("Sheet1", "Prices")
	prices := wb.Sheets[1]
	put(prices, "A1", "apple")
	put(prices, "B1", "3")
	put(prices, "A2", "pear")
	put(prices, "B2", "5")

	wb.ActiveSheet = 0
	sh := wb.Sheets[0]
	ref := put(sh, "A1", "=Prices!B2")
	if err := EvaluateCell(tview.NewTable(), ref); err != nil {
		t.Fatal(err)
	}
	if numOf(t, ref) != 5 {
		t.Fatalf("cross sheet got %s", *ref.Display)
	}

	look := put(sh, "B1", `=VLOOKUP("pear", Prices!A1:B2, 2, FALSE)`)
	if err := EvaluateCell(tview.NewTable(), look); err != nil {
		t.Fatal(err)
	}
	if numOf(t, look) != 5 {
		t.Fatalf("vlookup got %s", *look.Display)
	}

	idx := put(sh, "C1", `=INDEX(Prices!A1:B2, 1, 1)`)
	if err := EvaluateCell(tview.NewTable(), idx); err != nil {
		t.Fatal(err)
	}
	if idx.Display == nil || *idx.Display != "apple" {
		t.Fatalf("index got %v", idx.Display)
	}
}

func TestEditInsideRangeRecalculates(t *testing.T) {
	wb := setupBook("Sheet1")
	sh := wb.Sheets[0]
	a1 := put(sh, "A1", "2")
	put(sh, "A2", "3")
	sum := put(sh, "B1", "=SUM(A1:A2)")
	tv := tview.NewTable()
	if err := EvaluateCell(tv, sum); err != nil {
		t.Fatal(err)
	}
	if numOf(t, sum) != 5 {
		t.Fatalf("before %s", *sum.Display)
	}
	*a1.RawValue = "10"
	*a1.Display = "10"
	a1.ClearFlag(cell.FlagFormula)
	if err := RecalculateCell(tv, a1); err != nil {
		t.Fatal(err)
	}
	if numOf(t, sum) != 13 {
		t.Fatalf("after edit %s", *sum.Display)
	}
}

func TestCircularAndSheetQualifiedName(t *testing.T) {
	wb := setupBook("Q4 Budget")
	sh := wb.Sheets[0]
	f := put(sh, "A1", "='Q4 Budget'!A1+1")
	err := EvaluateCell(tview.NewTable(), f)
	if err == nil {
		t.Fatal("expected circular error")
	}
	if f.Display == nil || *f.Display != "#CIRC!" {
		t.Fatalf("display %v err %v", f.Display, err)
	}
}
