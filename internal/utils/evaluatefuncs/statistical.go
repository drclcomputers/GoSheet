// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// statistical.go provides statistical functions

package evaluatefuncs

import "fmt"

func StatisticalFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"AVG": func(args ...any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("AVG requires at least 1 argument(s), got 0")
			}
			nums := NumericValues(args)
			if len(nums) == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			sum := 0.0
			for _, n := range nums {
				sum += n
			}
			return sum / float64(len(nums)), nil
		},

		"COUNT": func(args ...any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("COUNT requires at least 1 argument(s), got 0")
			}
			return float64(len(NumericValues(args))), nil
		},

		"SUM": func(args ...any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("SUM requires at least 1 argument(s), got 0")
			}
			sum := 0.0
			for _, n := range NumericValues(args) {
				sum += n
			}
			return sum, nil
		},

		"PRODUCT": func(args ...any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("PRODUCT requires at least 1 argument(s), got 0")
			}
			nums := NumericValues(args)
			if len(nums) == 0 {
				return 0.0, nil
			}
			product := 1.0
			for _, n := range nums {
				product *= n
			}
			return product, nil
		},

		"CHOOSE": func(args ...any) (any, error) {
			if err := validateArgs("CHOOSE", args, 2, -1); err != nil {
				return nil, err
			}
			index, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CHOOSE: %v", err)
			}
			idx := int(index)
			if idx < 1 || idx >= len(args) {
				return nil, fmt.Errorf("index out of range")
			}
			return args[idx], nil
		},

		"ISNUMBER": func(args ...any) (any, error) {
			if err := validateArgs("ISNUMBER", args, 1, 1); err != nil {
				return nil, err
			}
			_, ok := args[0].(float64)
			return ok, nil
		},

		"ISTEXT": func(args ...any) (any, error) {
			if err := validateArgs("ISTEXT", args, 1, 1); err != nil {
				return nil, err
			}
			_, ok := args[0].(string)
			return ok, nil
		},

		"ISBLANK": func(args ...any) (any, error) {
			if err := validateArgs("ISBLANK", args, 1, 1); err != nil {
				return nil, err
			}
			s := toString(args[0])
			return s == "", nil
		},
	}
}
