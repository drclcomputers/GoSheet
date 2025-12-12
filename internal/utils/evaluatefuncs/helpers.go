// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// helpers.go provides helper functions

package evaluatefuncs

import (
	"fmt"
	"strings"
	"time"
)

// Helper function to convert any type to string
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case int:
		return fmt.Sprintf("%d", val)
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Helper function to convert any type to float64
func toFloat(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func validateArgs(funcName string, args []any, minArgs, maxArgs int) error {
	if len(args) < minArgs {
		if minArgs == maxArgs {
			return fmt.Errorf("%s requires exactly %d argument(s), got %d", funcName, minArgs, len(args))
		}
		return fmt.Errorf("%s requires at least %d argument(s), got %d", funcName, minArgs, len(args))
	}
	if maxArgs != -1 && len(args) > maxArgs {
		return fmt.Errorf("%s accepts at most %d argument(s), got %d", funcName, maxArgs, len(args))
	}
	return nil
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
