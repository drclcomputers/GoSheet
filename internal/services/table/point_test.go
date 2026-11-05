package table

import "testing"

func TestPointRef(t *testing.T) {
	if got := pointRef("Sheet1", "Sheet1", 2, 2); got != "B2" {
		t.Fatalf("same sheet: %s", got)
	}
	if got := pointRef("Sheet1", "sheet1", 2, 2); got != "B2" {
		t.Fatalf("case fold: %s", got)
	}
	if got := pointRef("Sheet1", "Prices", 2, 2); got != "Prices!B2" {
		t.Fatalf("other sheet: %s", got)
	}
	if got := pointRef("Sheet1", "Q4 Budget", 1, 1); got != "'Q4 Budget'!A1" {
		t.Fatalf("quoted: %s", got)
	}
}

func TestCanStartPoint(t *testing.T) {
	cases := []struct {
		text   string
		cursor int
		want   bool
	}{
		{"=", 1, true},
		{"$=", 2, true},
		{"=SUM(", len("=SUM("), true},
		{"=SUM(A1", len("=SUM(A1"), false},
		{"=A1+", len("=A1+"), true},
		{"=A1:B2", len("=A1:"), true},
		{`=CONCAT("=A1+", "x")`, len(`=CONCAT("=A1+`), false},
		{"hello", 5, false},
		{"=Prices!", len("=Prices!"), true},
	}
	for _, tc := range cases {
		if got := canStartPoint(tc.text, tc.cursor); got != tc.want {
			t.Fatalf("canStartPoint(%q, %d) = %v", tc.text, tc.cursor, got)
		}
	}
}
