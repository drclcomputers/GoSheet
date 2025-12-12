// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// cell.go provides functions to create and manage individual cells in a spreadsheet application.

package cell

import (
	"gosheet/internal/utils"
	"strings"

	"github.com/rivo/tview"
)

// NewCell creates and returns a new Cell instance
func NewCell(r, c int32, text string) *Cell {
	t := "string"
	autotype := "auto"
	return &Cell{
		RawValue: &text,
		Display:  &text,
		Type:     &t,

		Row:       r,
		Column:    c,
		MaxWidth:  utils.DEFAULT_CELL_MAX_WIDTH,
		MinWidth:  utils.DEFAULT_CELL_MIN_WIDTH,
		
		Align: tview.AlignLeft,
		Flags: FlagEditable,

		DecimalPoints:      utils.DEFAULT_CELL_DECIMAL_POINTS,
		ThousandsSeparator: utils.DEFAULT_CELL_THOUSANDS_SEPARATOR,
		DecimalSeparator:   utils.DEFAULT_CELL_DECIMAL_SEPARATOR,
		FinancialSign:      utils.DEFAULT_CELL_FINANCIAL_SIGN,
		DateTimeFormat:     &autotype,

		Color:   utils.ColorRGB{255, 255, 255},
		BgColor: utils.ColorRGB{0, 0, 0},

		Notes: nil,
		Valrule: nil,
		Valrulemsg: nil,

		DependsOn:  nil,
		Dependents: nil,
	}
}

// Checks if cells exists, otherwise creates it
func GetOrCreateCell(table *tview.Table, absRow, absCol int32, data map[[2]int]*Cell) *Cell {
    key := [2]int{int(absRow), int(absCol)}
    
    if cellData, exists := data[key]; exists {
        return cellData
    }
    
    newCell := NewCell(absRow, absCol, "")
    data[key] = newCell
    
    return newCell
}

// SetFlag turns a flag ON
func (c *Cell) SetFlag(flag uint8) {
	c.Flags |= flag
}

// ClearFlag turns a flag OFF
func (c *Cell) ClearFlag(flag uint8) {
	c.Flags &^= flag
}

// HasFlag checks if a flag is set
func (c *Cell) HasFlag(flag uint8) bool {
	return c.Flags&flag != 0
}

// ToggleFlag toggles bit
func (c *Cell) ToggleFlag(flag uint8) {
	c.Flags ^= flag
}

// SetFlagState sets the flag to a given bool
func (c *Cell) SetFlagState(flag uint8, enabled bool) {
	if enabled {
		c.Flags |= flag
	} else {
		c.Flags &^= flag
	}
}

// SetTableCell associates a tview.TableCell with this Cell
func (c *Cell) SetTableCell(tv *tview.TableCell) {
	c.tvCell = tv
}

// GetTableCell retrieves the associated tview.TableCell
func (c *Cell) GetTableCell() *tview.TableCell {
	return c.tvCell
}

// SetFormulaTag sets the cell's RawValue and updates its Type if it starts with "$="
func (c *Cell) SetFormulaTag(v string) *Cell {
    if strings.HasPrefix(v, "$=") {
		c.RawValue = &v
	}
	return c
}

// Checks if a cell contains a formula
func (c *Cell) IsFormula() bool {
    return strings.HasPrefix(strings.TrimSpace(*c.RawValue), "$=") 
}

// Returns the formula expression
func (c *Cell) GetFormulaExpression() string {
    if c.IsFormula() {
		aux := *c.RawValue
		return strings.TrimSpace(aux[2:])
    }
    return ""
}

// SetAlign sets the alignment of the cell based on the provided string
func (c *Cell) SetAlign(align string) *Cell {
	switch align {
		case "left": c.Align = tview.AlignLeft
		case "center": c.Align = tview.AlignCenter
		case "right": c.Align = tview.AlignRight
		default: c.Align = tview.AlignLeft
	}
	return c
}

// SetMinCellWidth ensures the cell's Display meets the minimum width by padding with spaces
func (c *Cell) SetMinCellWidth(text string) string {
	w := c.MinWidth
	if w <= 0 {
		w = 10
	}

	visibleLen := int32(len(StripTviewTags(text)))
	if visibleLen < w {
		padding := int(w - visibleLen)
		switch c.Align {
		case tview.AlignLeft:
			text = text + strings.Repeat(" ", padding)
		case tview.AlignRight:
			text = strings.Repeat(" ", padding) + text
		case tview.AlignCenter:
			leftPad := padding / 2
			rightPad := padding - leftPad
			text = strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
		default:
			text = text + strings.Repeat(" ", padding)
		}
	}
	return text
}

// Clones a cell and returns the clone
func (c *Cell) Clone() *Cell {
    if c == nil {
        return nil
    }
    
    clone := *c
    
    if c.RawValue != nil {
        rawCopy := *c.RawValue
        clone.RawValue = &rawCopy
    }
    if c.Display != nil {
        displayCopy := *c.Display
        clone.Display = &displayCopy
    }
    if c.Type != nil {
        typeCopy := *c.Type
        clone.Type = &typeCopy
    }
    if c.Notes != nil {
        notesCopy := *c.Notes
        clone.Notes = &notesCopy
    }
    if c.Valrule != nil {
        valruleCopy := *c.Valrule
        clone.Valrule = &valruleCopy
    }
	if c.DateTimeFormat != nil {
		dateTimeFormatCopy := *c.DateTimeFormat
		clone.DateTimeFormat = &dateTimeFormatCopy
	}
    
    clone.DependsOn = make([]*string, len(c.DependsOn))
    for i, dep := range c.DependsOn {
        if dep != nil {
            depCopy := *dep
            clone.DependsOn[i] = &depCopy
        }
    }
    
    clone.Dependents = make([]*string, len(c.Dependents))
    for i, dep := range c.Dependents {
        if dep != nil {
            depCopy := *dep
            clone.Dependents[i] = &depCopy
        }
    }
    
    return &clone
}
