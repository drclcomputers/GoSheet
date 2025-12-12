// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// inputcapture.go contains the implementation of the InputCaptureService

package table

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/services/ui"
	"gosheet/internal/services/ui/file"
	"gosheet/internal/services/ui/navigation"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var clipboard = [][]*cell.Cell{}

// Table input capture function. Manages everything from cell selection, to key combinations and calls other services.
func InputCaptureService(app *tview.Application, table *tview.Table, vp *utils.Viewport, data map[[2]int]*cell.Cell) *tview.Table {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.GetFocus() != table {
			return event
		}

		if event.Key() == tcell.KeyCtrlC {
		    modal := tview.NewModal().
		        SetText("Ctrl+C detected. Exiting...\nUnsaved edits will be lost.").
		        AddButtons([]string{"OK"}).
		        SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		            app.Stop()
		        })
		    app.SetRoot(modal, true).SetFocus(modal)
		    return nil
		}

		activeData := GetActiveSheetData()
		activeViewport := GetActiveViewport()
		
		if activeData == nil || activeViewport == nil {
			return event
		}

		visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())

		if event.Modifiers() == 0 {
			isSelecting = false
			clearSelectionRange()

				switch event.Key() {
			case tcell.KeyUp:
				if visualRow == 1 && activeViewport.TopRow > 1 {
					activeViewport.TopRow--
					RenderVisible(table, activeViewport, activeData)
					table.Select(1, int(visualCol))
					return nil
				}
				return event
			
			case tcell.KeyDown:
				if activeViewport.TopRow+activeViewport.ViewRows >= utils.MAX_ROWS {
					return nil
				}
				if visualRow == int32(activeViewport.ViewRows) {
					activeViewport.TopRow++
					RenderVisible(table, activeViewport, activeData)
					table.Select(int(activeViewport.ViewRows), int(visualCol))
					return nil
				}
				return event
			
			case tcell.KeyLeft:
				if visualCol == 1 && activeViewport.LeftCol > 1 {
					activeViewport.LeftCol--
					RenderVisible(table, activeViewport, activeData)
					table.Select(int(visualRow), 1)
					return nil
				}
				return event
				
			case tcell.KeyRight:
				if activeViewport.LeftCol+activeViewport.ViewCols >= utils.MAX_COLS {
					return nil
				}
				if visualCol == int32(activeViewport.ViewCols) {
					activeViewport.LeftCol++
					RenderVisible(table, activeViewport, activeData)
					table.Select(int(visualRow), int(activeViewport.ViewCols))
					return nil
				}
				return event
			}
		}

		switch {
		// Shift + Arrow Keys - Selection
		case event.Modifiers()&tcell.ModShift != 0:
			absRow, absCol := activeViewport.ToAbsolute(visualRow, visualCol)

			if !isSelecting {
				isSelecting = true
				anchorRow, anchorCol = absRow, absCol
			}

			switch event.Key() {
			case tcell.KeyRight:
				if absCol >= utils.MAX_COLS {
					return nil
				}
				if visualCol == int32(activeViewport.ViewCols) {
					activeViewport.LeftCol++
					RenderVisible(table, activeViewport, activeData)
					absCol++
					table.Select(int(visualRow), int(activeViewport.ViewCols))
				} else if visualCol < int32(table.GetColumnCount()-1) {
					visualCol++
					absRow, absCol = activeViewport.ToAbsolute(visualRow, visualCol)
					table.Select(int(visualRow), int(visualCol))
				}

			case tcell.KeyLeft:
				if visualCol == 1 && activeViewport.LeftCol > 1 {
					activeViewport.LeftCol--
					RenderVisible(table, activeViewport, activeData)
					absCol--
					table.Select(int(visualRow), 1)
				} else if visualCol > 0 {
					visualCol--
					absRow, absCol = activeViewport.ToAbsolute(visualRow, visualCol)
					table.Select(int(visualRow), int(visualCol))
				}

			case tcell.KeyDown:
				if absRow >= utils.MAX_ROWS {
					return nil
				}
				if visualRow == int32(activeViewport.ViewRows) {
					activeViewport.TopRow++
					RenderVisible(table, activeViewport, activeData)
					absRow++
					table.Select(int(activeViewport.ViewRows), int(visualCol))
				} else if visualRow < int32(table.GetRowCount()-1) {
					visualRow++
					absRow, absCol = activeViewport.ToAbsolute(visualRow, visualCol)
					table.Select(int(visualRow), int(visualCol))
				}

			case tcell.KeyUp:
				if visualRow == 1 && activeViewport.TopRow > 1 {
					activeViewport.TopRow--
					RenderVisible(table, activeViewport, activeData)
					absRow--
					table.Select(1, int(visualCol))
				} else if visualRow > 0 {
					visualRow--
					absRow, absCol = activeViewport.ToAbsolute(visualRow, visualCol)
					table.Select(int(visualRow), int(visualCol))
				}

			default:
				return event
			}

			highlightRange(table, anchorRow, anchorCol, absRow, absCol)
			return nil

		// Alt + O → Sort
		case (event.Rune() == 'o' || event.Rune() == 'O') && event.Modifiers()&tcell.ModAlt != 0:
			ShowSortDialog(app, table)
			MarkAsModified(table)
			return nil

		// Alt + G → Go to cell
		case (event.Rune() == 'g' || event.Rune() == 'G') && event.Modifiers()&tcell.ModAlt != 0:
			navigation.GoToCellModal(app, table, activeData, activeViewport, RenderVisible)
			return nil

		// Alt + F → Find
		case (event.Rune() == 'f' || event.Rune() == 'F') && event.Modifiers()&tcell.ModAlt != 0:
			navigation.FindDialog(app, table, activeData, activeViewport, RenderVisible)
			return nil

		// Alt + H → Replace
		case (event.Rune() == 'h' || event.Rune() == 'H') && event.Modifiers()&tcell.ModAlt != 0:
			navigation.ReplaceDialog(app, table, activeData, activeViewport, RenderVisible)
			MarkAsModified(table)
			return nil

		// F4 → Find Next
		case event.Key() == tcell.KeyF4:
			navigation.FindNextQuick(table, activeData, activeViewport, RenderVisible)
			return nil

		// F3 → Find Previous
		case event.Key() == tcell.KeyF3:
			navigation.FindPreviousQuick(table, activeData, activeViewport, RenderVisible)
			return nil
			
		// ALT + C → Copy
		case (event.Rune() == 'c' || event.Rune() == 'C') && event.Modifiers()&tcell.ModAlt != 0:
			copySelection(table)
			return nil

		// Alt + R → Copy cell format
		case (event.Rune() == 'r' || event.Rune() == 'R') && event.Modifiers()&tcell.ModAlt != 0:
			robCopyCellFormat(table)
			return nil

		// Alt + I → Paste cell format
		case (event.Rune() == 'i' || event.Rune() == 'I') && event.Modifiers()&tcell.ModAlt != 0:
			imitatePasteCellFormat(app, table, visualRow, visualCol)
			MarkAsModified(table)
			return nil

		// ALT + V → Paste
		case (event.Rune() == 'v' || event.Rune() == 'V') && event.Modifiers()&tcell.ModAlt != 0:
			pasteSelection(app, table, visualRow, visualCol)
			MarkAsModified(table)
			return nil

		// ALT + X → Cut
		case (event.Rune() == 'x' || event.Rune() == 'X') && event.Modifiers()&tcell.ModAlt != 0:
			cutSelection(app, table)
			MarkAsModified(table)
			return nil

		// Alt + Delete → Clear Cell/s
		case event.Key() == tcell.KeyDelete && event.Modifiers()&tcell.ModAlt != 0:
			deleteSelection(app, table)
			MarkAsModified(table)
			return nil

		// Alt + A → AutoFill selected cells 
		case (event.Rune() == 'a' || event.Rune() == 'A') && event.Modifiers()&tcell.ModAlt != 0:
    		ShowFillDialog(app, table)
    		return nil

		// Alt + Z → Undo
		case (event.Rune() == 'z' || event.Rune() == 'Z') && event.Modifiers()&tcell.ModAlt != 0:
			Undo(table)
			MarkAsModified(table)
			return nil

		// Alt + Y → Redo
		case (event.Rune() == 'y' || event.Rune() == 'Y') && event.Modifiers()&tcell.ModAlt != 0:
			Redo(table)
			MarkAsModified(table)
			return nil

		// Alt + Equal → Insert row/col
		case event.Rune() == '=' && event.Modifiers()&tcell.ModAlt != 0:
			insertRowCol(app, table)
			MarkAsModified(table)
			return nil

		// Alt + Minus → Delete row/col
		case event.Rune() == '-' && event.Modifiers()&tcell.ModAlt != 0:
			deleteRowCol(app, table)
			MarkAsModified(table)
			return nil

		// Alt + N → Edit cell comment
		case (event.Rune() == 'n' || event.Rune() == 'N') && event.Modifiers()&tcell.ModAlt != 0:
			ui.ShowCommentDialog(app, table, activeData, activeViewport)
			MarkAsModified(table)
			return nil

		// Alt + M - Open Sheet Manager
		case (event.Rune() == 'm' || event.Rune() == 'M') && event.Modifiers()&tcell.ModAlt != 0:
			ShowSheetManagerDialog(app, table)
			return nil

		// Alt + T - Open Sheet Context Menu
		case (event.Rune() == 't' || event.Rune() == 'T') && event.Modifiers()&tcell.ModAlt != 0:
			ShowSheetContextMenu(app, table)
			return nil

		// Alt + PageDown - Next Sheet
		case event.Key() == tcell.KeyPgDn && event.Modifiers()&tcell.ModAlt != 0:
			if globalWorkbook != nil && globalWorkbook.ActiveSheet < len(globalWorkbook.Sheets)-1 {
				SwitchSheet(app, table, globalWorkbook.ActiveSheet+1)
			}
			return nil

		// Alt + PageUp - Previous Sheet
		case event.Key() == tcell.KeyPgUp && event.Modifiers()&tcell.ModAlt != 0:
			if globalWorkbook != nil && globalWorkbook.ActiveSheet > 0 {
				SwitchSheet(app, table, globalWorkbook.ActiveSheet-1)
			}
			return nil

		// ALt + 1-9 - Quick switch between sheets
		case event.Rune() >= '1' && event.Rune() <= '9' && event.Modifiers()&tcell.ModAlt != 0:
    		idx := int(event.Rune() - '1')
    		if globalWorkbook != nil && idx < len(globalWorkbook.Sheets) {
    		    SwitchSheet(app, table, idx)
    		}
    		return nil

		// Alt + / → Show Help modal
		case event.Rune() == '/' && event.Modifiers()&tcell.ModAlt != 0:
			ui.ShowHelpModal(app, table)
			return nil

		// ESCAPE / Alt + S - Show Save Dialog
		case event.Key() == tcell.KeyEscape || ((event.Rune() == 's' || event.Rune() == 'S') && event.Modifiers()&tcell.ModAlt != 0):
			file.ShowUnifiedFileDialog(app, table, "save", activeData, table, SetCurrentFilename, MarkAsSaved, HasUnsavedChanges, GetCurrentFilename())
			return nil
			
		default:
			isSelecting = false
		}
		
		return event
	})
	
	return table
}	
