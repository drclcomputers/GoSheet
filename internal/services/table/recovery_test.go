package table

import (
	"testing"

	"gosheet/internal/services/cell"
)

func TestCaptureRecoverySnapshotDetachesCellData(t *testing.T) {
	wb := NewWorkbook()
	value := "before"
	message := "invalid value"
	wb.HasChanges = true
	wb.Sheets[0].Data[[2]int{1, 1}] = &cell.Cell{
		RawValue:   &value,
		Valrulemsg: &message,
		Row:        1,
		Column:     1,
	}
	globalWorkbook = wb
	t.Cleanup(func() { globalWorkbook = nil })

	sheets, activeSheet, dirty := CaptureRecoverySnapshot()
	if !dirty || activeSheet != 0 {
		t.Fatalf("snapshot state: dirty=%v active=%d", dirty, activeSheet)
	}

	cellSnapshot := sheets[0].GlobalData[[2]int{1, 1}]
	*wb.Sheets[0].Data[[2]int{1, 1}].RawValue = "after"
	*wb.Sheets[0].Data[[2]int{1, 1}].Valrulemsg = "changed"
	if *cellSnapshot.RawValue != "before" || *cellSnapshot.Valrulemsg != "invalid value" {
		t.Fatalf("snapshot shared cell pointers: value=%q message=%q", *cellSnapshot.RawValue, *cellSnapshot.Valrulemsg)
	}
}
