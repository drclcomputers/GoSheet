// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// excel_handler_helpers.go provides helper functions for exporting/importing xlsx files

package fileop

import "strings"

// convertFormulaToExcel converts GoSheet formula syntax to Excel syntax
func (h *ExcelFormatHandler) convertFormulaToExcel(formula string) string {
	formula = strings.ToUpper(formula)

	formula = strings.TrimSpace(formula)
	if !strings.HasPrefix(formula, "=") {
		formula = "=" + formula
	}

	simpleReplacements := map[string]string{
		"AVG":           "AVERAGE",
		"CEIL":          "CEILING",
		"RAD":           "RADIANS",
		"DEG":           "DEGREES",
		"DATEDIFF":      "DATEDIF",
		"BITSHIFTLEFT":  "BITLSHIFT",
		"BITSHIFTRIGHT": "BITRSHIFT",
	}

	for gosheet, excel := range simpleReplacements {
		formula = replaceFunction(formula, gosheet, excel)
	}

	formula = replaceFunctionWithWrapper(formula, "CTAN", "1/TAN")
	formula = replaceFunctionWithWrapper(formula, "SEC", "1/COS")
	formula = replaceFunctionWithWrapper(formula, "CSEC", "1/SIN")
	formula = replaceFunctionWithWrapper(formula, "CTANH", "1/TANH")
	formula = replaceFunctionWithWrapper(formula, "SECH", "1/COSH")
	formula = replaceFunctionWithWrapper(formula, "CSCH", "1/SINH")

	formula = replaceFunctionWithInverse(formula, "ACTAN", "ATAN", "1/")
	formula = replaceFunctionWithInverse(formula, "ASEC", "ACOS", "1/")
	formula = replaceFunctionWithInverse(formula, "ACSC", "ASIN", "1/")
	formula = replaceFunctionWithInverse(formula, "ASECH", "ACOSH", "1/")
	formula = replaceFunctionWithInverse(formula, "ACSCH", "ASINH", "1/")
	formula = replaceFunctionWithInverse(formula, "ACOTH", "ATANH", "1/")

	formula = replaceBesselJ0(formula)
	formula = replaceBesselJ1(formula)
	formula = replaceBesselYN(formula)

	return formula
}

// convertExcelFormulaToGoSheet converts Excel formula syntax to GoSheet syntax
func (h *ExcelFormatHandler) convertExcelFormulaToGoSheet(formula string) string {
	formula = strings.TrimPrefix(formula, "=")
	formula = strings.ToUpper(formula)
	formula = strings.TrimPrefix(formula, "_xludf.")
	formula = strings.TrimPrefix(formula, "_XLUDF.")
	formula = strings.TrimPrefix(formula, "_xlfn.")
	formula = strings.TrimPrefix(formula, "_XLFN.")
	formula = strings.TrimSpace(formula)

	simpleReplacements := map[string]string{
		"AVERAGE":   "AVG",
		"CEILING":   "CEIL",
		"RADIANS":   "RAD",
		"DEGREES":   "DEG",
		"DATEDIF":   "DATEDIFF",
		"BITLSHIFT": "BITSHIFTLEFT",
		"BITRSHIFT": "BITSHIFTRIGHT",
	}

	for excel, gosheet := range simpleReplacements {
		formula = replaceFunction(formula, excel, gosheet)
	}

	formula = replaceInverseWrapper(formula, "1/TAN", "CTAN")
	formula = replaceInverseWrapper(formula, "1/COS", "SEC")
	formula = replaceInverseWrapper(formula, "1/SIN", "CSEC")
	formula = replaceInverseWrapper(formula, "1/TANH", "CTANH")
	formula = replaceInverseWrapper(formula, "1/COSH", "SECH")
	formula = replaceInverseWrapper(formula, "1/SINH", "CSCH")

	formula = replaceInverseArgument(formula, "ATAN", "1/", "ACTAN")
	formula = replaceInverseArgument(formula, "ACOS", "1/", "ASEC")
	formula = replaceInverseArgument(formula, "ASIN", "1/", "ACSC")
	formula = replaceInverseArgument(formula, "ACOSH", "1/", "ASECH")
	formula = replaceInverseArgument(formula, "ASINH", "1/", "ACSCH")
	formula = replaceInverseArgument(formula, "ATANH", "1/", "ACOTH")

	formula = replaceExcelBesselJ0(formula)
	formula = replaceExcelBesselJ1(formula)
	formula = replaceExcelBesselYN(formula)

	return formula
}

// replaceFunction replaces a function name while preserving its arguments
func replaceFunction(formula, oldFunc, newFunc string) string {
	result := formula
	searchStr := oldFunc + "("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+len(oldFunc):] // Use temporary marker
				continue
			}
		}
		
		result = result[:idx] + newFunc + result[idx+len(oldFunc):]
	}
	
	result = strings.ReplaceAll(result, "§", oldFunc)
	
	return result
}

// replaceFunctionWithWrapper wraps a function: CTAN(x) -> 1/TAN(x)
func replaceFunctionWithWrapper(formula, oldFunc, wrapper string) string {
	result := formula
	searchStr := oldFunc + "("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+len(oldFunc):]
				continue
			}
		}
		
		args, endIdx := extractFunctionArgs(result, idx+len(searchStr)-1)
		if endIdx == -1 {
			break
		}
		
		replacement := "(" + wrapper + "(" + args + "))"
		result = result[:idx] + replacement + result[endIdx+1:]
	}
	
	result = strings.ReplaceAll(result, "§", oldFunc)
	return result
}

// replaceFunctionWithInverse wraps arguments: ACTAN(x) -> ATAN(1/x)
func replaceFunctionWithInverse(formula, oldFunc, newFunc, wrapper string) string {
	result := formula
	searchStr := oldFunc + "("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+len(oldFunc):]
				continue
			}
		}
		
		args, endIdx := extractFunctionArgs(result, idx+len(searchStr)-1)
		if endIdx == -1 {
			break
		}
		
		replacement := newFunc + "(" + wrapper + "(" + args + "))"
		result = result[:idx] + replacement + result[endIdx+1:]
	}
	
	result = strings.ReplaceAll(result, "§", oldFunc)
	return result
}

// replaceInverseWrapper: 1/TAN(x) -> CTAN(x)
func replaceInverseWrapper(formula, pattern, newFunc string) string {
	result := formula
	searchStr := pattern + "("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		args, endIdx := extractFunctionArgs(result, idx+len(searchStr)-1)
		if endIdx == -1 {
			break
		}
		
		startIdx := idx
		if idx > 0 && result[idx-1] == '(' {
			startIdx = idx - 1
			if endIdx+1 < len(result) && result[endIdx+1] == ')' {
				endIdx = endIdx + 1
			}
		}
		
		replacement := newFunc + "(" + args + ")"
		result = result[:startIdx] + replacement + result[endIdx+1:]
	}
	
	return result
}

// replaceInverseArgument: ATAN(1/x) -> ACTAN(x)
func replaceInverseArgument(formula, funcName, pattern, newFunc string) string {
	result := formula
	searchStr := funcName + "(" + pattern
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		argStart := idx + len(searchStr)
		
		args, endIdx := extractFunctionArgs(result, argStart-1)
		if endIdx == -1 {
			break
		}
		
		funcStart := idx
		
		replacement := newFunc + "(" + args + ")"
		result = result[:funcStart] + replacement + result[endIdx+1:]
	}
	
	return result
}

// replaceBesselJ0: J0(x) -> BESSELJ(x,0)
func replaceBesselJ0(formula string) string {
	result := formula
	searchStr := "J0("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+2:]
				continue
			}
		}
		
		args, endIdx := extractFunctionArgs(result, idx+2)
		if endIdx == -1 {
			break
		}
		
		replacement := "BESSELJ(" + args + ",0)"
		result = result[:idx] + replacement + result[endIdx+1:]
	}
	
	result = strings.ReplaceAll(result, "§", "J0")
	return result
}

// replaceBesselJ1: J1(x) -> BESSELJ(x,1)
func replaceBesselJ1(formula string) string {
	result := formula
	searchStr := "J1("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+2:]
				continue
			}
		}
		
		args, endIdx := extractFunctionArgs(result, idx+2)
		if endIdx == -1 {
			break
		}
		
		replacement := "BESSELJ(" + args + ",1)"
		result = result[:idx] + replacement + result[endIdx+1:]
	}
	
	result = strings.ReplaceAll(result, "§", "J1")
	return result
}

// replaceBesselYN: YN(n,x) -> BESSELY(x,n) - swaps arguments!
func replaceBesselYN(formula string) string {
	result := formula
	searchStr := "YN("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		if idx > 0 {
			prev := result[idx-1]
			if (prev >= 'A' && prev <= 'Z') || (prev >= '0' && prev <= '9') || prev == '_' {
				result = result[:idx] + "§" + result[idx+2:]
				continue
			}
		}
		
		args, endIdx := extractFunctionArgs(result, idx+2)
		if endIdx == -1 {
			break
		}
		
		parts := splitFunctionArgs(args)
		if len(parts) == 2 {
			replacement := "BESSELY(" + parts[1] + "," + parts[0] + ")"
			result = result[:idx] + replacement + result[endIdx+1:]
		} else {
			break
		}
	}
	
	result = strings.ReplaceAll(result, "§", "YN")
	return result
}

// Reverse Bessel conversions for import
func replaceExcelBesselJ0(formula string) string {
	result := formula
	searchStr := "BESSELJ("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		args, endIdx := extractFunctionArgs(result, idx+7)
		if endIdx == -1 {
			break
		}
		
		parts := splitFunctionArgs(args)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == "0" {
			replacement := "J0(" + parts[0] + ")"
			result = result[:idx] + replacement + result[endIdx+1:]
		} else {
			result = result[:idx] + "§BESSELJ§" + result[idx+8:]
		}
	}
	
	result = strings.ReplaceAll(result, "§BESSELJ§", "BESSELJ")
	return result
}

func replaceExcelBesselJ1(formula string) string {
	result := formula
	searchStr := "BESSELJ("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		args, endIdx := extractFunctionArgs(result, idx+7)
		if endIdx == -1 {
			break
		}
		
		parts := splitFunctionArgs(args)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == "1" {
			replacement := "J1(" + parts[0] + ")"
			result = result[:idx] + replacement + result[endIdx+1:]
		} else {
			result = result[:idx] + "§BESSELJ§" + result[idx+8:]
		}
	}
	
	result = strings.ReplaceAll(result, "§BESSELJ§", "BESSELJ")
	return result
}

func replaceExcelBesselYN(formula string) string {
	result := formula
	searchStr := "BESSELY("
	
	for {
		idx := strings.Index(result, searchStr)
		if idx == -1 {
			break
		}
		
		args, endIdx := extractFunctionArgs(result, idx+7)
		if endIdx == -1 {
			break
		}
		
		parts := splitFunctionArgs(args)
		if len(parts) == 2 {
			replacement := "YN(" + parts[1] + "," + parts[0] + ")"
			result = result[:idx] + replacement + result[endIdx+1:]
		} else {
			break
		}
	}
	
	return result
}

// extractFunctionArgs extracts arguments from a function call
func extractFunctionArgs(formula string, startIdx int) (string, int) {
	if startIdx >= len(formula) || formula[startIdx] != '(' {
		return "", -1
	}
	
	depth := 1
	i := startIdx + 1
	
	for i < len(formula) && depth > 0 {
		switch formula[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		i++
	}
	
	if depth != 0 {
		return "", -1
	}
	
	return formula[startIdx+1 : i-1], i - 1
}

// splitFunctionArgs splits comma-separated arguments respecting nested parentheses
func splitFunctionArgs(args string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	
	for _, ch := range args {
		if ch == '(' {
			depth++
			current.WriteRune(ch)
		} else if ch == ')' {
			depth--
			current.WriteRune(ch)
		} else if ch == ',' && depth == 0 {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}
	
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	
	return parts
}
