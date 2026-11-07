// Copyright (c) 2026 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// logical.go provides logic functions

package evaluatefuncs

import "fmt"

func LogicalFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"IF": func(args ...any) (any, error) {
			if err := validateArgs("IF", args, 3, 3); err != nil {
				return nil, err
			}
			condition, err := isTruthy(args[0])
			if err != nil {
				return nil, fmt.Errorf("IF: %v", err)
			}
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
				ok, err := isTruthy(args[i])
				if err != nil {
					return nil, fmt.Errorf("IFS: %v", err)
				}
				if ok {
					return args[i+1], nil
				}
			}
			if len(args)%2 == 1 {
				return args[len(args)-1], nil
			}
			return nil, fmt.Errorf("IFS: no condition matched")
		},

		"AND": func(args ...any) (any, error) {
			if err := validateArgs("AND", args, 1, -1); err != nil {
				return nil, err
			}
			flat := FlattenArgs(args)
			if len(flat) == 0 {
				return nil, fmt.Errorf("AND requires at least 1 argument(s), got 0")
			}
			for _, arg := range flat {
				ok, err := isTruthy(arg)
				if err != nil {
					return nil, fmt.Errorf("AND: %v", err)
				}
				if !ok {
					return false, nil
				}
			}
			return true, nil
		},

		"OR": func(args ...any) (any, error) {
			if err := validateArgs("OR", args, 1, -1); err != nil {
				return nil, err
			}
			flat := FlattenArgs(args)
			if len(flat) == 0 {
				return nil, fmt.Errorf("OR requires at least 1 argument(s), got 0")
			}
			for _, arg := range flat {
				ok, err := isTruthy(arg)
				if err != nil {
					return nil, fmt.Errorf("OR: %v", err)
				}
				if ok {
					return true, nil
				}
			}
			return false, nil
		},

		"NOT": func(args ...any) (any, error) {
			if err := validateArgs("NOT", args, 1, 1); err != nil {
				return nil, err
			}
			ok, err := isTruthy(args[0])
			if err != nil {
				return nil, fmt.Errorf("NOT: %v", err)
			}
			return !ok, nil
		},

		"XOR": func(args ...any) (any, error) {
			if err := validateArgs("XOR", args, 1, -1); err != nil {
				return nil, err
			}
			flat := FlattenArgs(args)
			count := 0
			for _, arg := range flat {
				ok, err := isTruthy(arg)
				if err != nil {
					return nil, fmt.Errorf("XOR: %v", err)
				}
				if ok {
					count++
				}
			}
			return count%2 == 1, nil
		},
	}
}
