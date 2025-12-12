// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// parsecheck.go provides parsers and input validation

package utils

import (
	"strings"
	"fmt"
	"strconv"
	"time"
)

var Separators = []rune{'Ø', '.', ',', '`', '\'', '"', '_'};
var DecimalSeparators = Separators
var FinancialSigns = []rune{'$', '€', '£', '¥', '₩', '₹', '₽', 'R', '₱', '₿', 'Ξ'}

// Checks if a given string is a number
func IsNumber(s string, financialsign rune) bool {
    s = strings.TrimSpace(s)
	s = strings.Trim(s, string(financialsign))
    if s == "" {
        return false
    }
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

// Validates value for DataTime cell type
func IsValidDateTime(s string) (isValid bool, detectedFormat string) {
	s = strings.TrimSpace(s)
	
	dateTimeFormats := []string{
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		"02-01-2006 15:04:05",
		"2006-01-02 15:04",
		"01/02/2006 15:04",
		"Jan 2, 2006 3:04 PM",
		"2 Jan 2006 15:04",
	}
	
	for _, format := range dateTimeFormats {
		if _, err := time.Parse(format, s); err == nil {
			return true, "datetime"
		}
	}
	
	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"02-01-2006",
		"01-02-2006",
		"2006/01/02",
		"Jan 2, 2006",
		"2 Jan 2006",
		"01/02/06",
	}
	
	for _, format := range dateFormats {
		if _, err := time.Parse(format, s); err == nil {
			return true, "date"
		}
	}
	
	timeFormats := []string{
		"15:04:05",
		"15:04",
		"3:04 PM",
		"3:04:05 PM",
	}
	
	for _, format := range timeFormats {
		if _, err := time.Parse(format, s); err == nil {
			return true, "time"
		}
	}
	
	return false, ""
}

// Parse datetime string and return Go time.Time
func ParseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	
	formats := []string{
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		"02-01-2006 15:04:05",
		"2006-01-02 15:04",
		"01/02/2006 15:04",
		"Jan 2, 2006 3:04 PM",
		"2 Jan 2006 15:04",
	
		"2006-01-02",
		"01/02/2006",
		"02-01-2006",
		"01-02-2006",
		"2006/01/02",
		"Jan 2, 2006",
		"2 Jan 2006",
		"01/02/06",

		"15:04:05",
		"15:04",
		"3:04 PM",
		"3:04:05 PM",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("invalid datetime format")
}

// Format time.Time based on what components it has
func FormatDateTime(t time.Time, preferredFormat string) string {
	hasDate := t.Year() != 0
	hasTime := t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0
	
	switch preferredFormat {
	case "datetime":
		if hasTime {
			return t.Format("2006-01-02 15:04:05")
		}
		return t.Format("2006-01-02")
	case "date":
		return t.Format("2006-01-02")
	case "time":
		return t.Format("15:04:05")
	default:
		if hasDate && hasTime {
			return t.Format("2006-01-02 15:04:05")
		} else if hasDate {
			return t.Format("2006-01-02")
		} else {
			return t.Format("15:04:05")
		}
	}
}

// Formats numbers
func FormatWithCommas(val float64, thousandseparator, decimalseparator rune, decimalpoints int32, financialsign rune) string {
	negative := val < 0
	if negative {
		val = -val
	}

	var result string

    parts := strings.Split(fmt.Sprintf("%.*f", decimalpoints, val), ".")
    intPart := parts[0]
	decPart := ""
    if len(parts)>1{ decPart = parts[1] }

    n := len(intPart)
    if n <= 3 { 
		if len(parts) == 1 { 
			result = intPart
			if negative {result = "-" + result}
			return result
		}
		
		result = intPart + string(decimalseparator) + decPart
		if negative {result = "-" + result}
		return result
	}

    var b strings.Builder
    pre := n % 3
    if pre > 0 {
        b.WriteString(intPart[:pre])
        if n > pre { b.WriteRune(thousandseparator) }
    }

    for i := pre; i < n; i += 3 {
        b.WriteString(intPart[i : i+3])
        if i+3 < n { b.WriteRune(thousandseparator) }
    }
	
	if len(parts) == 1 { 
		result = b.String()
		if negative {result = "-" + result}
		return result
	}

    result = b.String() + string(decimalseparator) + decPart
	if negative {result = "-" + result}
	return result
}
