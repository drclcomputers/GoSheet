// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// conversions.go provides functions to convert between custom Cell structs and tview.TableCell structs.

package cell

import (
	"gosheet/internal/utils"
	"regexp"
	"strings"

	"github.com/rivo/tview"
)

// ApplyTextEffects applies formatting effects
func (c *Cell) ApplyTextEffects(text string) string {
	if text == "" {
		return text
	}

	result := text

	if c.HasFlag(FlagStrikethrough) {
		result = "~" + result + "~"
	}

	var tags []string

	if c.HasFlag(FlagBold) {
		tags = append(tags, "b")
	}
	if c.HasFlag(FlagItalic) {
		tags = append(tags, "i")
	}
	if c.HasFlag(FlagUnderline) {
		tags = append(tags, "u")
	}

	if len(tags) > 0 {
		tagString := strings.Join(tags, "")
		result = "[::" + tagString + "]" + result + "[::-]"
	}

	return result
}

// ToTViewCell converts a custom Cell to a tview.TableCell
func (c *Cell) ToTViewCell() *tview.TableCell {
	var textValue string
	if c.Display != nil {
		textValue = *c.Display
	} else if c.RawValue != nil {
		textValue = *c.RawValue
	} else {
		textValue = ""
	}

	if c.HasFlag(FlagAllCaps) {
		textValue = strings.ToUpper(textValue)
	}

	textValue = c.ApplyTextEffects(textValue)
	textValue = c.SetMinCellWidth(textValue)

	w := c.MaxWidth
	if w <= 0 {
		w = utils.DEFAULT_CELL_MAX_WIDTH
	}

	tvCell := tview.NewTableCell(textValue).
		SetAlign(int(c.Align)).
		SetTextColor(c.Color.ToTCellColor()).
		SetBackgroundColor(c.BgColor.ToTCellColor()).
		SetExpansion(0).
		SetMaxWidth(int(w))

	tvCell.SetReference(c)
	c.SetTableCell(tvCell)

	return tvCell
}

// stripTviewTags removes tview formatting tags from a string using regexp
var tviewTagRegex = regexp.MustCompile(`\[(?:[^]]+)\]`)

func StripTviewTags(s string) string {
	return tviewTagRegex.ReplaceAllString(s, "")
}

// ToCustomCell converts a tview.TableCell to a custom Cell
func ToCustomCell(tv *tview.TableCell, row, col int32) *Cell {
	if tv == nil {
		return NewCell(row, col, "")
	}

	text := StripTviewTags(tv.Text)
	align := tv.Align

	rawValueCopy := text
	displayCopy := text
	t := "string"
	autotype := "auto"
	
	c := &Cell{
		RawValue:           &rawValueCopy,
		Display:            &displayCopy,
		Type:               &t,
		Row:                row,
		Column:             col,
		MaxWidth:           utils.DEFAULT_CELL_MAX_WIDTH,
		MinWidth:           utils.DEFAULT_CELL_MIN_WIDTH,
		Align:              int8(align),
		Flags:              FlagEditable,
		DecimalPoints:      utils.DEFAULT_CELL_DECIMAL_POINTS,
		ThousandsSeparator: utils.DEFAULT_CELL_DECIMAL_SEPARATOR,
		DecimalSeparator:   utils.DEFAULT_CELL_DECIMAL_SEPARATOR,
		FinancialSign:      utils.DEFAULT_CELL_FINANCIAL_SIGN,
		DateTimeFormat:     &autotype,
		Color:              utils.ColorRGB{255, 255, 255},
		BgColor:            utils.ColorRGB{0, 0, 0},
		Notes:              nil,
		Valrule:            nil,
		Valrulemsg:         nil,
		DependsOn:          nil,
		Dependents:         nil,
		tvCell:             tv,
	}

	if ref := tv.GetReference(); ref != nil {
		if old, ok := ref.(*Cell); ok {
			if old.RawValue != nil {
				rawCopy := *old.RawValue
				c.RawValue = &rawCopy
			}
			if old.Display != nil {
				dispCopy := *old.Display
				c.Display = &dispCopy
			}
			if old.Type != nil {
				typeCopy := *old.Type
				c.Type = &typeCopy
			}
			if old.Notes != nil {
				notesCopy := *old.Notes
				c.Notes = &notesCopy
			}
			if old.Valrule != nil {
				valruleCopy := *old.Valrule
				c.Valrule = &valruleCopy
			}
			if old.Valrulemsg != nil {
				valrulemsgCopy := *old.Valrulemsg
				c.Valrulemsg = &valrulemsgCopy
			}
			if old.DateTimeFormat != nil {
				dateTimeCopy := *old.DateTimeFormat
				c.DateTimeFormat = &dateTimeCopy
			}
			
			c.Color = old.Color
			c.BgColor = old.BgColor
			c.Flags = old.Flags
			c.Align = old.Align
			c.DecimalPoints = old.DecimalPoints
			c.ThousandsSeparator = old.ThousandsSeparator
			c.DecimalSeparator = old.DecimalSeparator
			c.FinancialSign = old.FinancialSign
			
			if old.DependsOn != nil {
				c.DependsOn = make([]*string, len(old.DependsOn))
				for i, dep := range old.DependsOn {
					if dep != nil {
						depCopy := *dep
						c.DependsOn[i] = &depCopy
					}
				}
			}
			if old.Dependents != nil {
				c.Dependents = make([]*string, len(old.Dependents))
				for i, dep := range old.Dependents {
					if dep != nil {
						depCopy := *dep
						c.Dependents[i] = &depCopy
					}
				}
			}
		}
	}

	return c
}
