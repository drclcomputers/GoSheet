package table

import (
	"strings"
	"testing"

	"github.com/rivo/tview"
)

func TestRenameSheetRewritesFormulas(t *testing.T) {
	wb := setupBook("Sheet1", "Prices")
	put(wb.Sheets[1], "A1", "5")
	f := put(wb.Sheets[0], "B1", `=Prices!A1+1`)
	quoted := put(wb.Sheets[0], "B2", `='Prices'!A1`)
	text := put(wb.Sheets[0], "B3", `=CONCAT("Prices!A1", Prices!A1)`)
	local := put(wb.Sheets[0], "B4", `=A1`)
	tv := tview.NewTable()
	if err := EvaluateCell(tv, f); err != nil {
		t.Fatal(err)
	}

	if err := wb.RenameSheet(1, "Costs"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(*f.RawValue, "Costs!A1") || strings.Contains(*f.RawValue, "Prices") {
		t.Fatalf("formula %s", *f.RawValue)
	}
	if !strings.Contains(*quoted.RawValue, "Costs!A1") || strings.Contains(*quoted.RawValue, "Prices") {
		t.Fatalf("quoted %s", *quoted.RawValue)
	}
	if !strings.Contains(*text.RawValue, `"Prices!A1"`) || !strings.Contains(*text.RawValue, "Costs!A1") {
		t.Fatalf("string was rewritten: %s", *text.RawValue)
	}
	if *local.RawValue != "=A1" {
		t.Fatalf("local ref changed: %s", *local.RawValue)
	}
	if numOf(t, f) != 6 {
		t.Fatalf("value after rename %s", *f.Display)
	}
}

func TestRenameQuotedSheet(t *testing.T) {
	wb := setupBook("Sheet1", "Q4 Budget")
	put(wb.Sheets[1], "A1", "3")
	f := put(wb.Sheets[0], "A1", "='Q4 Budget'!A1*2")
	if err := EvaluateCell(tview.NewTable(), f); err != nil {
		t.Fatal(err)
	}
	if err := wb.RenameSheet(1, "Q4 Actual"); err != nil {
		t.Fatal(err)
	}
	if *f.RawValue != "='Q4 Actual'!A1*2" {
		t.Fatalf("got %s", *f.RawValue)
	}
	if numOf(t, f) != 6 {
		t.Fatalf("value %s", *f.Display)
	}
}

func TestDeleteSheetBecomesRefError(t *testing.T) {
	wb := setupBook("Sheet1", "Prices")
	put(wb.Sheets[1], "A1", "5")
	put(wb.Sheets[1], "A2", "7")
	f := put(wb.Sheets[0], "B1", "=SUM(Prices!A1:A2)")
	other := put(wb.Sheets[0], "B2", "=Prices!A1+Sheet1!A1")
	put(wb.Sheets[0], "A1", "1")
	if err := EvaluateCell(tview.NewTable(), f); err != nil {
		t.Fatal(err)
	}
	if err := wb.DeleteSheet(1); err != nil {
		t.Fatal(err)
	}
	if len(wb.Sheets) != 1 {
		t.Fatalf("sheets %d", len(wb.Sheets))
	}
	if !strings.Contains(*f.RawValue, "#REF!") || strings.Contains(*f.RawValue, "Prices") {
		t.Fatalf("formula %s", *f.RawValue)
	}
	if f.Display == nil || *f.Display != "#REF!" {
		t.Fatalf("display %v", f.Display)
	}
	if other.Display == nil || *other.Display != "#REF!" {
		t.Fatalf("mixed formula %s display %v", *other.RawValue, other.Display)
	}
}
