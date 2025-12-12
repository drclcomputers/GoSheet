// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// formatters.go provides various formatters in the edit cell dialog

package cellui

import "github.com/rivo/tview"

func setFormattingEnabled(cellType string,
	financialSignDropdown, thousandsSeparatorDropdown, decimalSeparatorDropdown, dateTimeFormatDropdown *tview.DropDown,
	decimalPointsInput *tview.InputField) {

	enable := func(p tview.Primitive, enabled bool) {
		switch v := p.(type) {
		case *tview.DropDown:
			v.SetDisabled(!enabled)
		case *tview.InputField:
			v.SetDisabled(!enabled)
		}
	}

	switch cellType {
	case "financial":
		enable(financialSignDropdown, true)
		enable(thousandsSeparatorDropdown, true)
		enable(decimalSeparatorDropdown, true)
		enable(decimalPointsInput, true)
		enable(dateTimeFormatDropdown, false)
	case "number":
		enable(financialSignDropdown, false) 
		enable(thousandsSeparatorDropdown, true)
		enable(decimalSeparatorDropdown, true)
		enable(decimalPointsInput, true)
		enable(dateTimeFormatDropdown, false)
	case "datetime":
		enable(financialSignDropdown, false)
		enable(thousandsSeparatorDropdown, false)
		enable(decimalSeparatorDropdown, false)
		enable(decimalPointsInput, false)
		enable(dateTimeFormatDropdown, true)
	default:
		enable(financialSignDropdown, false)
		enable(thousandsSeparatorDropdown, false)
		enable(decimalSeparatorDropdown, false)
		enable(decimalPointsInput, false)
		enable(dateTimeFormatDropdown, false)
	}
}

