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
			if err := validateArgs("AVG", args, 2, -1); err != nil {
				return nil, err
			}
			sum := 0.0
			for _, arg := range args {
				f, _ := toFloat(arg)
				sum += f
			}
			return sum / float64(len(args)), nil
		},

		"COUNT": func(args ...any) (any, error) {
			if err := validateArgs("COUNT", args, 1, -1); err != nil {
				return nil, err
			}
			count := 0
			for _, arg := range args {
				if _, ok := arg.(float64); ok {
					count++
				}
			}
			return float64(count), nil
		},

		"SUM": func(args ...any) (any, error) {
			if err := validateArgs("SUM", args, 2, -1); err != nil {
				return nil, err
			}
			sum := 0.0
			for _, arg := range args {
				f, _ := toFloat(arg)
				sum += f
			}
			return sum, nil
		},

		"PRODUCT": func(args ...any) (any, error) {
			if err := validateArgs("PRODUCT", args, 2, -1); err != nil {
				return nil, err
			}
			product := 1.0
			for _, arg := range args {
				f, _ := toFloat(arg)
				product *= f
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
