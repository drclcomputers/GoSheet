package file

import "testing"

func TestIsSupportedSpreadsheetFile(t *testing.T) {
	tests := []struct {
		extension string
		want      bool
	}{
		{extension: ".gsheet", want: true},
		{extension: ".csv", want: true},
		{extension: ".pdf", want: false},
	}

	for _, test := range tests {
		if got := isSupportedSpreadsheetFile(test.extension); got != test.want {
			t.Errorf("isSupportedSpreadsheetFile(%q) = %t, want %t", test.extension, got, test.want)
		}
	}
}
