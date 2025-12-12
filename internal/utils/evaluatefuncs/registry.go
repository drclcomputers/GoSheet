// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// registry.go provides the main function registry that combines all function categories

package evaluatefuncs

import (
	"maps"
)

type ExprFunction func(args ...any) (any, error)

func GovalFuncs() map[string]ExprFunction {
	functions := make(map[string]ExprFunction)

	mergeFunctions(functions, MathFunctions())
	mergeFunctions(functions, StatisticalFunctions())
	mergeFunctions(functions, StringFunctions())
	mergeFunctions(functions, DateTimeFunctions())
	mergeFunctions(functions, LogicalFunctions())

	return functions
}

func mergeFunctions(target, source map[string]ExprFunction) {
	if target == nil || source == nil {
		return
	}
	maps.Copy(target, source)
}
