// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dropdowns.go provides various dropdowns in the edit cell dialog

package cellui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

func newDropdown(label string, options []string, current int32, onSelect func(string)) *tview.DropDown {
	return tview.NewDropDown().
		SetLabel(label).
		SetOptions(options, func(opt string, _ int) { onSelect(opt) }).
		SetCurrentOption(int(current))
}

func getDropdowns(c *cell.Cell) (*tview.DropDown, *tview.DropDown, *tview.DropDown, *tview.DropDown, *tview.InputField) {
	financialSigns := toStringSlice(utils.FinancialSigns)
	separators := toStringSlice(utils.Separators)
	dateTimeFormats := utils.DateTimeFormats	

	financialSignDropdown := newDropdown("Sign: ", financialSigns,
		findRuneIndex(financialSigns, c.FinancialSign),
		func(opt string) {
			c.FinancialSign = []rune(opt)[0]
			if *c.Type == "financial" {
				normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
				normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
				if val, err := strconv.ParseFloat(normalized, 64); err == nil {
					formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
					*c.Display = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
				}
			}
		})

	thousandsSeparatorDropdown := newDropdown("Thousands: ", separators,
		findRuneIndex(separators, c.ThousandsSeparator),
		func(opt string) {
			c.ThousandsSeparator = []rune(opt)[0]
			if c.ThousandsSeparator == 'Ø' {
				c.ThousandsSeparator = 0
			}
			if *c.Type == "number" || *c.Type == "financial" {
				normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
				normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
				if val, err := strconv.ParseFloat(normalized, 64); err == nil {
					formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
					if *c.Type == "financial" {
						formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
					}
					*c.Display = formatted
				}
			}
		})

	decimalSeparatorDropdown := newDropdown("Decimal: ", separators,
		findRuneIndex(separators, c.DecimalSeparator),
		func(opt string) {
			c.DecimalSeparator = []rune(opt)[0]
			if c.DecimalSeparator == 'Ø' {
				c.DecimalSeparator = 0
			}
			if *c.Type == "number" || *c.Type == "financial" {
				normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
				normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
				if val, err := strconv.ParseFloat(normalized, 64); err == nil {
					formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
					if *c.Type == "financial" {
						formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
					}
					*c.Display = formatted
				}
			}
		})

	dateTimeFormatDropdown := newDropdown("Date Format: ", dateTimeFormats,
		int32(getDateTypeFormat(*c.DateTimeFormat)),
		func(opt string) {
			*c.DateTimeFormat = utils.DateTimeFormats[getDateTypeFormat(opt)]

			if c.RawValue != nil && *c.RawValue != "" {
				if t, err := utils.ParseDateTime(*c.RawValue); err == nil {
					*c.Display = utils.FormatDateTime(t, *c.DateTimeFormat)
				}
			}
		})

	decimalPointsInput := tview.NewInputField().
		SetLabel("Decimals: ").
		SetText(fmt.Sprintf("%d", c.DecimalPoints)).
		SetFieldWidth(4).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetChangedFunc(func(text string) {
			if points, err := strconv.Atoi(text); err == nil && points >= 0 {
				c.DecimalPoints = int32(points)

				if *c.Type == "number" || *c.Type == "financial" {
					normalized := strings.ReplaceAll(*c.RawValue, string(c.ThousandsSeparator), "")
					normalized = strings.TrimPrefix(normalized, string(c.FinancialSign))
					if val, err := strconv.ParseFloat(normalized, 64); err == nil {
						formatted := utils.FormatWithCommas(val, c.ThousandsSeparator, c.DecimalSeparator, c.DecimalPoints, c.FinancialSign)
						if *c.Type == "financial" {
							formatted = fmt.Sprintf("%c%s", c.FinancialSign, formatted)
						}
						*c.Display = formatted
					}
				}
			}
		})

	return financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown, decimalPointsInput
}
