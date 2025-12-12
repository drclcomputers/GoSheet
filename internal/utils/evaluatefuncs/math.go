// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// math.go provides math functions

package evaluatefuncs

import (
	"fmt"
	"math"
)

func MathFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		// Trigonometric functions
		"SIN": func(args ...any) (any, error) {
			if err := validateArgs("SIN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SIN: %v", err)
			}
			return math.Sin(f), nil
		},
		"COS": func(args ...any) (any, error) {
			if err := validateArgs("COS", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("COS: %v", err)
			}
			return math.Cos(f), nil
		},
		"TAN": func(args ...any) (any, error) {
			if err := validateArgs("TAN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("TAN: %v", err)
			}
			return math.Tan(f), nil
		},
		"CTAN": func(args ...any) (any, error) {
			if err := validateArgs("CTAN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CTAN: %v", err)
			}
			tanVal := math.Tan(f)
			if math.Abs(tanVal) < 1e-10 {
				return math.Inf(1), fmt.Errorf("CTAN: division by zero")
			}
			return 1 / tanVal, nil
		},

		// Inverse trigonometric functions
		"ASIN": func(args ...any) (any, error) {
			if err := validateArgs("ASIN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ASIN: %v", err)
			}
			return math.Asin(f), nil
		},
		"ACOS": func(args ...any) (any, error) {
			if err := validateArgs("ACOS", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACOS: %v", err)
			}
			return math.Acos(f), nil
		},
		"ATAN": func(args ...any) (any, error) {
			if err := validateArgs("ATAN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ATAN: %v", err)
			}
			return math.Atan(f), nil
		},
		"ATAN2": func(args ...any) (any, error) {
			if err := validateArgs("ATAN2", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ATAN2: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("ATAN2: %v", err)
			}
			return math.Atan2(f1, f2), nil
		},
		"ACTAN": func(args ...any) (any, error) {
			if err := validateArgs("ACTAN", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACTAN: %v", err)
			}
			return math.Pi/2 - math.Atan(f), nil
		},

		// Additional trigonometric functions
		"SEC": func(args ...any) (any, error) {
			if err := validateArgs("SEC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SEC: %v", err)
			}
			if math.Cos(f) == 0 {
				return math.Inf(0), fmt.Errorf("division by zero")
			}
			return 1 / math.Cos(f), nil
		},
		"CSEC": func(args ...any) (any, error) {
			if err := validateArgs("CSEC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CSEC: %v", err)
			}
			if math.Sin(f) == 0 {
				return math.Inf(0), fmt.Errorf("division by zero")
			}
			return 1 / math.Sin(f), nil
		},
		"ASEC": func(args ...any) (any, error) {
			if err := validateArgs("ASEC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ASEC: %v", err)
			}
			return math.Acos(1 / f), nil
		},
		"ACSC": func(args ...any) (any, error) {
			if err := validateArgs("ACSC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACSC: %v", err)
			}
			return math.Asin(1 / f), nil
		},

		// Degrees/radians conversion
		"RAD": func(args ...any) (any, error) {
			if err := validateArgs("RAD", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("RAD: %v", err)
			}
			return f * math.Pi / 180, nil
		},
		"DEG": func(args ...any) (any, error) {
			if err := validateArgs("DEG", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("DEG: %v", err)
			}
			return f * 180 / math.Pi, nil
		},

		// Hyperbolic functions
		"SINH": func(args ...any) (any, error) {
			if err := validateArgs("SINH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SINH: %v", err)
			}
			return math.Sinh(f), nil
		},
		"COSH": func(args ...any) (any, error) {
			if err := validateArgs("COSH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("COSH: %v", err)
			}
			return math.Cosh(f), nil
		},
		"TANH": func(args ...any) (any, error) {
			if err := validateArgs("TANH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("TANH: %v", err)
			}
			return math.Tanh(f), nil
		},
		"CTANH": func(args ...any) (any, error) {
			if err := validateArgs("CTANH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CTANH: %v", err)
			}
			if math.Tanh(f) == 0 {
				return math.Inf(0), fmt.Errorf("division by zero")
			}
			return 1 / math.Tanh(f), nil
		},

		// Additional hyperbolic functions
		"SECH": func(args ...any) (any, error) {
			if err := validateArgs("SECH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SECH: %v", err)
			}
			if math.Cosh(f) == 0 {
				return math.Inf(0), fmt.Errorf("division by zero")
			}
			return 1 / math.Cosh(f), nil
		},
		"CSCH": func(args ...any) (any, error) {
			if err := validateArgs("CSCH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CSCH: %v", err)
			}
			if math.Sinh(f) == 0 {
				return math.Inf(0), fmt.Errorf("division by zero")
			}
			return 1 / math.Sinh(f), nil
		},
		"ASINH": func(args ...any) (any, error) {
			if err := validateArgs("ASINH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ASINH: %v", err)
			}
			return math.Log(f + math.Sqrt(f*f+1)), nil
		},
		"ACOSH": func(args ...any) (any, error) {
			if err := validateArgs("ACOSH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACOSH: %v", err)
			}
			return math.Log(f + math.Sqrt(f*f-1)), nil
		},
		"ATANH": func(args ...any) (any, error) {
			if err := validateArgs("ATANH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ATANH: %v", err)
			}
			return 0.5 * math.Log((1+f)/(1-f)), nil
		},
		"ASECH": func(args ...any) (any, error) {
			if err := validateArgs("ASECH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ASECH: %v", err)
			}
			return math.Log((1 + math.Sqrt(1-f*f)) / f), nil
		},
		"ACSCH": func(args ...any) (any, error) {
			if err := validateArgs("ACSCH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACSCH: %v", err)
			}
			return math.Log((1/f) + math.Sqrt(1+(1/(f*f)))), nil
		},
		"ACOTH": func(args ...any) (any, error) {
			if err := validateArgs("ACOTH", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ACOTH: %v", err)
			}
			return 0.5 * math.Log((f+1)/(f-1)), nil
		},

		// Exponential and logarithmic
		"EXP": func(args ...any) (any, error) {
			if err := validateArgs("EXP", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("EXP: %v", err)
			}
			return math.Exp(f), nil
		},
		"LOG": func(args ...any) (any, error) {
			if err := validateArgs("LOG", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("LOG: %v", err)
			}
			return math.Log(f), nil
		},
		"LOG10": func(args ...any) (any, error) {
			if err := validateArgs("LOG10", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("LOG10: %v", err)
			}
			return math.Log10(f), nil
		},
		"LOG2": func(args ...any) (any, error) {
			if err := validateArgs("LOG2", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("LOG2: %v", err)
			}
			return math.Log2(f), nil
		},

		// Power and roots
		"SQRT": func(args ...any) (any, error) {
			if err := validateArgs("SQRT", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SQRT: %v", err)
			}
			return math.Sqrt(f), nil
		},
		"CBRT": func(args ...any) (any, error) {
			if err := validateArgs("CBRT", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CBRT: %v", err)
			}
			return math.Cbrt(f), nil
		},
		"POW": func(args ...any) (any, error) {
			if err := validateArgs("POW", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("POW: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("POW: %v", err)
			}
			return math.Pow(f1, f2), nil
		},

		// Other math utilities
		"ABS": func(args ...any) (any, error) {
			if err := validateArgs("ABS", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ABS: %v", err)
			}
			return math.Abs(f), nil
		},
		"CEIL": func(args ...any) (any, error) {
			if err := validateArgs("CEIL", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CEIL: %v", err)
			}
			return math.Ceil(f), nil
		},
		"FLOOR": func(args ...any) (any, error) {
			if err := validateArgs("FLOOR", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("FLOOR: %v", err)
			}
			return math.Floor(f), nil
		},
		"ROUND": func(args ...any) (any, error) {
			if err := validateArgs("ROUND", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ROUND: %v", err)
			}
			return math.Round(f), nil
		},
		"MIN": func(args ...any) (any, error) {
			if err := validateArgs("MIN", args, 2, -1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("MIN: %v", err)
			}
			minNR := f
			for _, arg := range args {
				f, err = toFloat(arg)
				if err != nil {
					return nil, fmt.Errorf("MIN: %v", err)
				}
				minNR = math.Min(f, minNR)
			}
			return minNR, nil
		},
		"MAX": func(args ...any) (any, error) {
			if err := validateArgs("MAX", args, 2, -1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("MAX: %v", err)
			}
			maxNR := f
			for _, arg := range args {
				f, err = toFloat(arg)
				if err != nil {
					return nil, fmt.Errorf("MAX: %v", err)
				}
				maxNR = math.Max(f, maxNR)
			}
			return maxNR, nil
		},

		// Utility functions
		"SIGN": func(args ...any) (any, error) {
			if err := validateArgs("SIGN", args, 1, 1); err != nil {
				return nil, err
			}
			x, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("SIGN: %v", err)
			}
			if x > 0 {
				return 1.0, nil
			} else if x < 0 {
				return -1.0, nil
			}
			return 0.0, nil
		},
		"CLAMP": func(args ...any) (any, error) {
			if err := validateArgs("CLAMP", args, 3, 3); err != nil {
				return nil, err
			}
			x, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("CLAMP: %v", err)
			}
			min, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("CLAMP: %v", err)
			}
			max, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("CLAMP: %v", err)
			}
			if x < min {
				return min, nil
			}
			if x > max {
				return max, nil
			}
			return x, nil
		},
		"LERP": func(args ...any) (any, error) {
			if err := validateArgs("LERP", args, 3, 3); err != nil {
				return nil, err
			}
			a, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("LERP: %v", err)
			}
			b, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("LERP: %v", err)
			}
			t, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("LERP: %v", err)
			}
			return a + t*(b-a), nil
		},

		// Special mathematical functions
		"ERF": func(args ...any) (any, error) {
			if err := validateArgs("ERF", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ERF: %v", err)
			}
			return math.Erf(f), nil
		},
		"ERFC": func(args ...any) (any, error) {
			if err := validateArgs("ERFC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ERFC: %v", err)
			}
			return math.Erfc(f), nil
		},
		"GAMMA": func(args ...any) (any, error) {
			if err := validateArgs("GAMMA", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("GAMMA: %v", err)
			}
			return math.Gamma(f), nil
		},
		"J0": func(args ...any) (any, error) {
			if err := validateArgs("J0", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("J0: %v", err)
			}
			return math.J0(f), nil
		},
		"J1": func(args ...any) (any, error) {
			if err := validateArgs("J1", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("J1: %v", err)
			}
			return math.J1(f), nil
		},
		"YN": func(args ...any) (any, error) {
			if err := validateArgs("YN", args, 2, 2); err != nil {
				return nil, err
			}
			n, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("YN: %v", err)
			}
			x, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("YN: %v", err)
			}
			return math.Yn(int(n), x), nil
		},

		// Additional rounding and precision
		"TRUNC": func(args ...any) (any, error) {
			if err := validateArgs("TRUNC", args, 1, 1); err != nil {
				return nil, err
			}
			f, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("TRUNC: %v", err)
			}
			return math.Trunc(f), nil
		},
		"ROUNDTO": func(args ...any) (any, error) {
			if err := validateArgs("ROUNDTO", args, 2, 2); err != nil {
				return nil, err
			}
			value, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("ROUNDTO: %v", err)
			}
			places, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("ROUNDTO: %v", err)
			}
			scale := math.Pow(10, places)
			return math.Round(value*scale) / scale, nil
		},

		// Engineering functions
		"HYPOT": func(args ...any) (any, error) {
			if err := validateArgs("HYPOT", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("HYPOT: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("HYPOT: %v", err)
			}
			return math.Hypot(f1, f2), nil
		},
		"MOD": func(args ...any) (any, error) {
			if err := validateArgs("MOD", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("MOD: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("MOD: %v", err)
			}
			return math.Mod(f1, f2), nil
		},
		"REMAINDER": func(args ...any) (any, error) {
			if err := validateArgs("REMAINDER", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("REMAINDER: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("REMAINDER: %v", err)
			}
			return math.Remainder(f1, f2), nil
		},

		// Bit operations
		"BITAND": func(args ...any) (any, error) {
			if err := validateArgs("BITAND", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("BITAND: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("BITAND: %v", err)
			}
			return float64(int(f1) & int(f2)), nil
		},
		"BITOR": func(args ...any) (any, error) {
			if err := validateArgs("BITOR", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("BITOR: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("BITOR: %v", err)
			}
			return float64(int(f1) | int(f2)), nil
		},
		"BITXOR": func(args ...any) (any, error) {
			if err := validateArgs("BITXOR", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("BITXOR: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("BITXOR: %v", err)
			}
			return float64(int(f1) ^ int(f2)), nil
		},
		"BITSHIFTLEFT": func(args ...any) (any, error) {
			if err := validateArgs("BITSHIFTLEFT", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("BITSHIFTLEFT: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("BITSHIFTLEFT: %v", err)
			}
			return float64(int(f1) << int(f2)), nil
		},
		"BITSHIFTRIGHT": func(args ...any) (any, error) {
			if err := validateArgs("BITSHIFTRIGHT", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("BITSHIFTRIGHT: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("BITSHIFTRIGHT: %v", err)
			}
			return float64(int(f1) >> int(f2)), nil
		},

		// Additional utility functions
		"FACTORIAL": func(args ...any) (any, error) {
			if err := validateArgs("FACTORIAL", args, 1, 1); err != nil {
				return nil, err
			}
			n, _ := toFloat(args[0])
			result := 1.0
			for i := 2; i <= int(n); i++ {
				result *= float64(i)
			}
			return result, nil
		},
		"GCD": func(args ...any) (any, error) {
			if err := validateArgs("GCD", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("GCD: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("GCD: %v", err)
			}
			a, b := int(f1), int(f2)
			for b != 0 {
				a, b = b, a%b
			}
			return float64(a), nil
		},
		"LCM": func(args ...any) (any, error) {
			if err := validateArgs("LCM", args, 2, 2); err != nil {
				return nil, err
			}
			f1, err := toFloat(args[0])
			if err != nil {
				return nil, fmt.Errorf("LCM: %v", err)
			}
			f2, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("LCM: %v", err)
			}
			a, b := int(f1), int(f2)
			gcd := a
			temp := b
			for temp != 0 {
				gcd, temp = temp, gcd%temp
			}
			return float64(a / gcd * b), nil
		},

		// Constants
		"PI":  func(args ...any) (any, error) { return math.Pi, nil },
		"E":   func(args ...any) (any, error) { return math.E, nil },
		"PHI": func(args ...any) (any, error) { return (1 + math.Sqrt(5)) / 2, nil },
		"INF": func(args ...any) (any, error) { return math.Inf(1), nil },
		"NAN": func(args ...any) (any, error) { return math.NaN(), nil },
	}
}
