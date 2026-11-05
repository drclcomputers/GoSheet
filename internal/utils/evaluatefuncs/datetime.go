// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// datetime.go provides functions related to datetime manipulation

package evaluatefuncs

import (
	"fmt"
	"time"
)

func DateTimeFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"NOW": func(args ...any) (any, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},

		"TODAY": func(args ...any) (any, error) {
			return time.Now().Format("2006-01-02"), nil
		},

		"DATE": func(args ...any) (any, error) {
			if err := validateArgs("DATE", args, 3, 3); err != nil {
				return nil, err
			}
			year, err1 := toFloat(args[0])
			month, err2 := toFloat(args[1])
			day, err3 := toFloat(args[2])
			if err1 != nil || err2 != nil || err3 != nil {
				return nil, fmt.Errorf("invalid date arguments")
			}
			date := time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC)
			return date.Format("2006-01-02"), nil
		},

		"TIME": func(args ...any) (any, error) {
			if err := validateArgs("TIME", args, 2, 3); err != nil {
				return nil, err
			}
			hour, err1 := toFloat(args[0])
			minute, err2 := toFloat(args[1])
			second := 0.0
			var err3 error
			if len(args) > 2 {
				second, err3 = toFloat(args[2])
			}
			if err1 != nil || err2 != nil || err3 != nil {
				return nil, fmt.Errorf("invalid time arguments")
			}
			t := time.Date(0, 1, 1, int(hour), int(minute), int(second), 0, time.UTC)
			return t.Format("15:04:05"), nil
		},

		"YEAR": func(args ...any) (any, error) {
			if err := validateArgs("YEAR", args, 1, 1); err != nil {
				return nil, err
			}
			dateStr := toString(args[0])
			t, err := ParseDateTime(dateStr)
			if err != nil {
				return nil, fmt.Errorf("invalid date format")
			}
			return float64(t.Year()), nil
		},

		"MONTH": func(args ...any) (any, error) {
			if err := validateArgs("MONTH", args, 1, 1); err != nil {
				return nil, err
			}
			dateStr := toString(args[0])
			t, err := ParseDateTime(dateStr)
			if err != nil {
				return nil, fmt.Errorf("invalid date format")
			}
			return float64(t.Month()), nil
		},

		"DAY": func(args ...any) (any, error) {
			if err := validateArgs("DAY", args, 1, 1); err != nil {
				return nil, err
			}
			dateStr := toString(args[0])
			t, err := ParseDateTime(dateStr)
			if err != nil {
				return nil, fmt.Errorf("invalid date format")
			}
			return float64(t.Day()), nil
		},

		"HOUR": func(args ...any) (any, error) {
			if err := validateArgs("HOUR", args, 1, 1); err != nil {
				return nil, err
			}
			timeStr := toString(args[0])
			t, err := ParseDateTime(timeStr)
			if err != nil {
				return nil, fmt.Errorf("invalid time format")
			}
			return float64(t.Hour()), nil
		},

		"MINUTE": func(args ...any) (any, error) {
			if err := validateArgs("MINUTE", args, 1, 1); err != nil {
				return nil, err
			}
			timeStr := toString(args[0])
			t, err := ParseDateTime(timeStr)
			if err != nil {
				return nil, fmt.Errorf("invalid time format")
			}
			return float64(t.Minute()), nil
		},

		"SECOND": func(args ...any) (any, error) {
			if err := validateArgs("SECOND", args, 1, 1); err != nil {
				return nil, err
			}
			timeStr := toString(args[0])
			t, err := ParseDateTime(timeStr)
			if err != nil {
				return nil, fmt.Errorf("invalid time format")
			}
			return float64(t.Second()), nil
		},

		"WEEKDAY": func(args ...any) (any, error) {
			if err := validateArgs("WEEKDAY", args, 1, 1); err != nil {
				return nil, err
			}
			dateStr := toString(args[0])
			t, err := ParseDateTime(dateStr)
			if err != nil {
				return nil, fmt.Errorf("invalid date format")
			}
			return float64(t.Weekday()) + 1, nil
		},

		"DATEDIFF": func(args ...any) (any, error) {
			if err := validateArgs("DATEDIFF", args, 2, 2); err != nil {
				return nil, err
			}
			date1Str := toString(args[0])
			date2Str := toString(args[1])
			t1, err := ParseDateTime(date1Str)
			if err != nil {
				return nil, fmt.Errorf("DATEDIFF: %v", err)
			}
			t2, err := ParseDateTime(date2Str)
			if err != nil {
				return nil, fmt.Errorf("DATEDIFF: %v", err)
			}
			days := t2.Sub(t1).Hours() / 24
			return days, nil
		},

		"DATEADD": func(args ...any) (any, error) {
			if err := validateArgs("DATEADD", args, 2, 2); err != nil {
				return nil, err
			}
			dateStr := toString(args[0])
			days, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("DATEADD: %v", err)
			}
			t, err := ParseDateTime(dateStr)
			if err != nil {
				return nil, fmt.Errorf("DATEADD: %v", err)
			}
			newDate := t.AddDate(0, 0, int(days))
			return newDate.Format("2006-01-02"), nil
		},
	}
}
