package navigation

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

func TestFindDialogsFitCompactViewport(t *testing.T) {
	tests := []struct {
		name     string
		open     func(*tview.Application)
		controls []string
	}{
		{
			name: "find",
			open: func(app *tview.Application) {
				FindDialog(app, tview.NewTable(), nil, nil, func(*tview.Table, *utils.Viewport, map[[2]int]*cell.Cell) {})
			},
			controls: []string{"Find Next", "Find Previous", "Case sensitive", "Match whole word", "Clear", "Close"},
		},
		{
			name: "replace",
			open: func(app *tview.Application) {
				ReplaceDialog(app, tview.NewTable(), nil, nil, func(*tview.Table, *utils.Viewport, map[[2]int]*cell.Cell) {})
			},
			controls: []string{"Find:", "Replace with:", "Replace", "Replace & Find Next", "Replace All", "Close"},
		},
		{
			name: "go to cell",
			open: func(app *tview.Application) {
				GoToCellModal(app, tview.NewTable(), nil, nil, func(*tview.Table, *utils.Viewport, map[[2]int]*cell.Cell) {})
			},
			controls: []string{"Cell Address", "Go", "Cancel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := tcell.NewSimulationScreen("UTF-8")
			app := tview.NewApplication().SetScreen(screen)
			screen.SetSize(80, 24)
			tt.open(app)
			app.ForceDraw()

			contents, width, height := screen.GetContents()
			var rendered strings.Builder
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					rendered.WriteString(string(contents[y*width+x].Runes))
				}
				rendered.WriteByte('\n')
			}
			for _, control := range tt.controls {
				if !strings.Contains(rendered.String(), control) {
					t.Errorf("%q is not visible in the compact %s dialog", control, tt.name)
				}
			}
			screen.Fini()
		})
	}
}
