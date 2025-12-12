// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// logical.go provides logic functions

package evaluatefuncs

func LogicalFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"IF": func(args ...any) (any, error) {
			if err := validateArgs("IF", args, 3, 3); err != nil {
				return nil, err
			}
			condition := args[0].(bool)
			if condition {
				return args[1], nil
			}
			return args[2], nil
		},

		"IFS": func(args ...any) (any, error) {
			if err := validateArgs("IFS", args, 2, -1); err != nil {
				return nil, err
			}
			for i := 0; i < len(args)-1; i += 2 {
				if args[i].(bool) {
					return args[i+1], nil
				}
			}
			return args[len(args)-1], nil
		},

		"AND": func(args ...any) (any, error) {
			if err := validateArgs("AND", args, 2, -1); err != nil {
				return nil, err
			}
			for _, arg := range args {
				if !arg.(bool) {
					return false, nil
				}
			}
			return true, nil
		},

		"OR": func(args ...any) (any, error) {
			if err := validateArgs("OR", args, 2, -1); err != nil {
				return nil, err
			}
			for _, arg := range args {
				if arg.(bool) {
					return true, nil
				}
			}
			return false, nil
		},

		"NOT": func(args ...any) (any, error) {
			if err := validateArgs("NOT", args, 1, 1); err != nil {
				return nil, err
			}
			return !args[0].(bool), nil
		},

		"XOR": func(args ...any) (any, error) {
			if err := validateArgs("XOR", args, 2, -1); err != nil {
				return nil, err
			}
			count := 0
			for _, arg := range args {
				if arg.(bool) {
					count++
				}
			}
			return count%2 == 1, nil
		},
	}
}
