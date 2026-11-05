package fileop

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gosheet/internal/services/cell"

	"github.com/xuri/excelize/v2"
)

func TestExcelKeepsRealSheet1AndQuotedNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.xlsx")

	number := cell.NewCell(1, 1, "10")
	numType := "number"
	number.Type = &numType

	formula := cell.NewCell(1, 2, "$=AVG('Q4 Budget'!A1:A2)")
	shown := "7.5"
	formula.Display = &shown

	label := cell.NewCell(1, 1, "pear")

	sheets := []SheetInfo{
		{Name: "Sheet1", GlobalData: map[[2]int]*cell.Cell{{1, 1}: number, {1, 2}: formula}},
		{Name: "Q4 Budget", GlobalData: map[[2]int]*cell.Cell{{1, 1}: label}},
	}
	if err := (&ExcelFormatHandler{}).Write(path, sheets, 1); err != nil {
		t.Fatal(err)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if got, want := f.GetSheetList(), []string{"Sheet1", "Q4 Budget"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sheets = %v", got)
	}
	if got := f.GetActiveSheetIndex(); got != 1 {
		t.Fatalf("active sheet = %d", got)
	}
	got, err := f.GetCellFormula("Sheet1", "B1")
	if err != nil {
		t.Fatal(err)
	}
	got = strings.TrimPrefix(got, "=")
	if got != "AVERAGE('Q4 Budget'!A1:A2)" {
		t.Fatalf("formula = %q", got)
	}

	only := filepath.Join(dir, "prices.xlsx")
	if err := (&ExcelFormatHandler{}).Write(only, []SheetInfo{{Name: "Prices"}}, 0); err != nil {
		t.Fatal(err)
	}
	pf, err := excelize.OpenFile(only)
	if err != nil {
		t.Fatal(err)
	}
	defer pf.Close()
	if got := pf.GetSheetList(); !reflect.DeepEqual(got, []string{"Prices"}) {
		t.Fatalf("sheets = %v, leftover Sheet1", got)
	}
}

func TestFormulaQuotesSurviveConversion(t *testing.T) {
	h := &ExcelFormatHandler{}
	if got := h.convertFormulaToExcel(`avg('Q4 Budget'!A1:A2)`); got != `=AVERAGE('Q4 Budget'!A1:A2)` {
		t.Fatalf("export = %s", got)
	}
	if got := h.convertExcelFormulaToGoSheet(`=AVERAGE('Q4 Budget'!A1:A2)`); got != `AVG('Q4 Budget'!A1:A2)` {
		t.Fatalf("import = %s", got)
	}
	if got := h.convertFormulaToExcel(`concat("Avg(keep)", A1)`); !strings.Contains(got, `"Avg(keep)"`) {
		t.Fatalf("string literal folded: %s", got)
	}
}

func TestCSVAndTXTExportFormulaResult(t *testing.T) {
	dir := t.TempDir()
	formula := cell.NewCell(1, 1, "$=SUM(A1:A3)")
	shown := "60"
	formula.Display = &shown
	plain := cell.NewCell(1, 2, "pear")
	sheets := []SheetInfo{{
		Name: "Sheet1",
		GlobalData: map[[2]int]*cell.Cell{
			{1, 1}: formula,
			{1, 2}: plain,
		},
	}}

	csvPath := filepath.Join(dir, "out.csv")
	if err := (&CSVFormatHandler{}).Write(csvPath, sheets, 0); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "60,pear" {
		t.Fatalf("csv = %q", got)
	}

	txtPath := filepath.Join(dir, "out.txt")
	if err := (&TextFormatHandler{}).Write(txtPath, sheets, 0); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(txtPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "60\tpear" {
		t.Fatalf("txt = %q", got)
	}
}

func TestODSIsNotOffered(t *testing.T) {
	for _, format := range GetWritableFormats() {
		if format == FormatODS {
			t.Fatal("ODS is still listed as writable")
		}
	}
	for _, format := range GetReadableFormats() {
		if format == FormatODS {
			t.Fatal("ODS is still listed as readable")
		}
	}
	if err := SaveWorkbookAs(nil, 0, filepath.Join(t.TempDir(), "book.ods"), FormatODS); err == nil {
		t.Fatal("expected ODS save to fail")
	}
}
