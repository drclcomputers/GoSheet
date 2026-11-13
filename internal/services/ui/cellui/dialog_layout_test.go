package cellui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gosheet/internal/services/cell"
)

func TestEditDialogShowsActionsInCompactViewport(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	app := tview.NewApplication().SetScreen(screen)
	screen.SetSize(90, 18)
	c := cell.NewCell(1, 1, "")
	financial, thousands, decimal, dateTime, decimalPoints := getDropdowns(c)

	dialog, focus := buildEditCellForm(
		app,
		tview.NewTable(),
		c,
		c.Clone(),
		1, 1,
		financial, thousands, decimal, decimalPoints, dateTime, 0, 0,
		func(*tview.Table, int32, int32, *cell.Cell, *cell.Cell) {},
		func(*tview.Table, *cell.Cell) error { return nil },
		func(*tview.Table, *cell.Cell) error { return nil },
		nil, nil,
	)
	app.SetRoot(dialog, true).SetFocus(focus).ForceDraw()

	cells, width, height := screen.GetContents()
	var rendered strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rendered.WriteString(string(cells[y*width+x].Runes))
		}
		rendered.WriteByte('\n')
	}
	for _, control := range []string{
		"Value", "Type", "Align", "Min Width", "Max Width",
		"Sign", "Thousands", "Decimal", "Decimals", "Date Format",
		"Bold", "Italic", "Underline", "All Caps", "Strikethrough", "Editable",
		"Data Validation", "Text Color", "Background Color",
	} {
		if !strings.Contains(rendered.String(), control) {
			t.Errorf("%q is not visible in the compact cell editor", control)
		}
	}
}
