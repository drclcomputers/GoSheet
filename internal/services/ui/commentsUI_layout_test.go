package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

func TestCommentDialogFitsCompactViewport(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	app := tview.NewApplication().SetScreen(screen)
	screen.SetSize(80, 24)

	table := tview.NewTable().SetSelectable(true, true)
	table.Select(1, 1)
	viewport := &utils.Viewport{TopRow: 1, LeftCol: 1, ViewRows: 10, ViewCols: 10}
	data := map[[2]int]*cell.Cell{{1, 1}: cell.NewCell(1, 1, "")}
	ShowCommentDialog(app, table, data, viewport)
	app.ForceDraw()

	contents, width, height := screen.GetContents()
	var rendered strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rendered.WriteString(string(contents[y*width+x].Runes))
		}
		rendered.WriteByte('\n')
	}
	for _, control := range []string{"Comment", "Save", "Delete", "Cancel"} {
		if !strings.Contains(rendered.String(), control) {
			t.Errorf("%q is not visible in the compact comment dialog", control)
		}
	}
	screen.Fini()
}
