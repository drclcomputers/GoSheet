// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// text.go provides text manipulation

package evaluatefuncs

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func StringFunctions() map[string]ExprFunction {
	return map[string]ExprFunction{
		"LEFT": func(args ...any) (any, error) {
			if err := validateArgs("LEFT", args, 2, 2); err != nil {
				return nil, err
			}
			text := toString(args[0])
			length, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("LEFT: %v", err)
			}
			n := int(length)
			runes := []rune(text)
			if n > len(runes) {
				return text, nil
			}
			return string(runes[:n]), nil
		},

		"RIGHT": func(args ...any) (any, error) {
			if err := validateArgs("RIGHT", args, 2, 2); err != nil {
				return nil, err
			}
			text := toString(args[0])
			length, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("RIGHT: %v", err)
			}
			n := int(length)
			runes := []rune(text)
			if n > len(runes) {
				return text, nil
			}
			return string(runes[len(runes)-n:]), nil
		},

		"MID": func(args ...any) (any, error) {
			if err := validateArgs("MID", args, 3, 3); err != nil {
				return nil, err
			}
			text := toString(args[0])
			start, err := toFloat(args[1])
			if err != nil {
				return nil, fmt.Errorf("MID: %v", err)
			}
			length, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("MID: %v", err)
			}
			startIdx := int(start) - 1
			n := int(length)
			runes := []rune(text)
			if startIdx < 0 {
				startIdx = 0
			}
			if startIdx+n > len(runes) {
				return string(runes[startIdx:]), nil
			}
			return string(runes[startIdx : startIdx+n]), nil
		},

		"UPPER": func(args ...any) (any, error) {
			if err := validateArgs("UPPER", args, 1, 1); err != nil {
				return nil, err
			}
			return strings.ToUpper(toString(args[0])), nil
		},

		"LOWER": func(args ...any) (any, error) {
			if err := validateArgs("LOWER", args, 1, 1); err != nil {
				return nil, err
			}
			return strings.ToLower(toString(args[0])), nil
		},

		"PROPER": func(args ...any) (any, error) {
			if err := validateArgs("PROPER", args, 1, 1); err != nil {
				return nil, err
			}
			caser := cases.Title(language.English)
			return caser.String(strings.ToLower(toString(args[0]))), nil
		},

		"TRIM": func(args ...any) (any, error) {
			if err := validateArgs("TRIM", args, 1, 1); err != nil {
				return nil, err
			}
			return strings.TrimSpace(toString(args[0])), nil
		},

		"FIND": func(args ...any) (any, error) {
			if err := validateArgs("FIND", args, 2, 3); err != nil {
				return nil, err
			}
			findText := toString(args[0])
			withinText := toString(args[1])
			startPos := 1
			if len(args) > 2 {
				sp, err := toFloat(args[2])
				if err != nil {
					return nil, fmt.Errorf("FIND: %v", err)
				}
				startPos = int(sp)
			}
			if startPos < 1 {
				return nil, fmt.Errorf("start position must be >= 1")
			}
			if startPos > len(withinText) {
				return -1.0, nil
			}
			pos := strings.Index(withinText[startPos-1:], findText)
			if pos == -1 {
				return -1.0, nil
			}
			return float64(pos + startPos), nil
		},

		"SUBSTITUTE": func(args ...any) (any, error) {
			if err := validateArgs("SUBSTITUTE", args, 3, 4); err != nil {
				return nil, err
			}
			text := toString(args[0])
			oldText := toString(args[1])
			newText := toString(args[2])
			instanceNum := -1
			if len(args) > 3 {
				in, err := toFloat(args[3])
				if err != nil {
					return nil, fmt.Errorf("SUBSTITUTE: %v", err)
				}
				instanceNum = int(in)
			}

			if instanceNum == -1 {
				return strings.ReplaceAll(text, oldText, newText), nil
			}

			parts := strings.Split(text, oldText)
			if instanceNum >= len(parts) {
				return text, nil
			}
			result := strings.Join(parts[:instanceNum], oldText) + newText + strings.Join(parts[instanceNum:], oldText)
			return result, nil
		},

		"LEN": func(args ...any) (any, error) {
			if err := validateArgs("LEN", args, 1, 1); err != nil {
				return nil, err
			}
			return float64(len(toString(args[0]))), nil
		},

		"CONCAT": func(args ...any) (any, error) {
			if err := validateArgs("CONCAT", args, 1, -1); err != nil {
				return nil, err
			}
			result := ""
			for _, arg := range args {
				result += toString(arg)
			}
			return result, nil
		},
	}
}
