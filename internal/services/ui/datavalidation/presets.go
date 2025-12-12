// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// presets.go provides some presetted options for data validation

package datavalidation

import (
	"fmt"
	"regexp"
	"strings"
)

// GetValidationPresets returns all available validation presets
func GetValidationPresets() []ValidationPreset {
	return []ValidationPreset{
		{
			Name:        "Custom",
			Description: "Write your own validation expression using 'THIS' to refer to the cell value",
			Fields:      []ValidationField{},
			BuildRule:   func(params map[string]string) string { return params["custom"] },
		},
		{
			Name:        "Whole Number - Between",
			Description: "Value must be a whole number between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum:", Type: "number", Placeholder: "0"},
				{Name: "max", Label: "Maximum:", Type: "number", Placeholder: "100"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS >= %s && THIS <= %s && THIS == FLOOR(THIS)", 
					params["min"], params["max"])
			},
		},
		{
			Name:        "Whole Number - Greater Than",
			Description: "Value must be a whole number greater than a value",
			Fields: []ValidationField{
				{Name: "value", Label: "Greater than:", Type: "number", Placeholder: "0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS > %s && THIS == FLOOR(THIS)", params["value"])
			},
		},
		{
			Name:        "Whole Number - Less Than",
			Description: "Value must be a whole number less than a value",
			Fields: []ValidationField{
				{Name: "value", Label: "Less than:", Type: "number", Placeholder: "100"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS < %s && THIS == FLOOR(THIS)", params["value"])
			},
		},
		{
			Name:        "Decimal - Between",
			Description: "Value must be a decimal number between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum:", Type: "number", Placeholder: "0.0"},
				{Name: "max", Label: "Maximum:", Type: "number", Placeholder: "1.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS >= %s && THIS <= %s", params["min"], params["max"])
			},
		},
		{
			Name:        "Decimal - Greater Than",
			Description: "Value must be greater than a decimal value",
			Fields: []ValidationField{
				{Name: "value", Label: "Greater than:", Type: "number", Placeholder: "0.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS > %s", params["value"])
			},
		},
		{
			Name:        "Decimal - Less Than",
			Description: "Value must be less than a decimal value",
			Fields: []ValidationField{
				{Name: "value", Label: "Less than:", Type: "number", Placeholder: "100.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS < %s", params["value"])
			},
		},
		{
			Name:        "Text Length - Between",
			Description: "Text length must be between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum length:", Type: "number", Placeholder: "1"},
				{Name: "max", Label: "Maximum length:", Type: "number", Placeholder: "50"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("LEN(THIS) >= %s && LEN(THIS) <= %s", 
					params["min"], params["max"])
			},
		},
		{
			Name:        "Text Length - Maximum",
			Description: "Text cannot exceed a certain length",
			Fields: []ValidationField{
				{Name: "max", Label: "Maximum length:", Type: "number", Placeholder: "255"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("LEN(THIS) <= %s", params["max"])
			},
		},
		{
			Name:        "Text - Not Empty",
			Description: "Cell must contain text (cannot be empty)",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "LEN(THIS) > 0"
			},
		},
		{
			Name:        "List - Allowed Values",
			Description: "Value must be one of the specified options",
			Fields: []ValidationField{
				{Name: "list", Label: "Allowed values (comma-separated):", Type: "text", 
					Placeholder: "Yes,No,Maybe"},
			},
			BuildRule: func(params map[string]string) string {
				values := strings.Split(params["list"], ",")
				conditions := make([]string, len(values))
				for i, val := range values {
					val = strings.TrimSpace(val)
					conditions[i] = fmt.Sprintf("THIS == \"%s\"", val)
				}
				return strings.Join(conditions, " || ")
			},
		},
		{
			Name:        "Email Format",
			Description: "Value must be a valid email format",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "CONTAINS(THIS, \"@\") && CONTAINS(SUBSTR(THIS, INDEX(THIS, \"@\")), \".\")"
			},
		},
		{
			Name:        "Positive Numbers Only",
			Description: "Value must be positive (greater than 0)",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "THIS > 0"
			},
		},
		{
			Name:        "Percentage (0-100)",
			Description: "Value must be between 0 and 100",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "THIS >= 0 && THIS <= 100"
			},
		},
	}
}

// detectPresetFromRule tries to detect which preset was used to create a rule
func detectPresetFromRule(rule string) (int, map[string]string) {
	if strings.TrimSpace(rule) == "" {
		return 0, nil
	}

	presets := GetValidationPresets()
	
	for i, preset := range presets {
		if preset.Name == "Custom" {
			continue
		}

		switch preset.Name {
		case "Whole Number - Between":
			re := regexp.MustCompile(`THIS >= ([\d.]+) && THIS <= ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Whole Number - Greater Than":
			re := regexp.MustCompile(`THIS > ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Whole Number - Less Than":
			re := regexp.MustCompile(`THIS < ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Decimal - Between":
			re := regexp.MustCompile(`THIS >= ([\d.]+) && THIS <= ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Decimal - Greater Than":
			re := regexp.MustCompile(`^THIS > ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Decimal - Less Than":
			re := regexp.MustCompile(`^THIS < ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Text Length - Between":
			re := regexp.MustCompile(`LEN\(THIS\) >= ([\d]+) && LEN\(THIS\) <= ([\d]+)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Text Length - Maximum":
			re := regexp.MustCompile(`LEN\(THIS\) <= ([\d]+)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"max": matches[1]}
			}
		case "Text - Not Empty":
			if rule == "LEN(THIS) > 0" {
				return i, nil
			}
		case "List - Allowed Values":
			if strings.Contains(rule, "THIS == \"") && strings.Contains(rule, "||") {
				parts := strings.Split(rule, " || ")
				values := make([]string, 0, len(parts))
				for _, part := range parts {
					re := regexp.MustCompile(`THIS == "([^"]+)"`)
					if matches := re.FindStringSubmatch(part); matches != nil {
						values = append(values, matches[1])
					}
				}
				if len(values) > 0 {
					return i, map[string]string{"list": strings.Join(values, ",")}
				}
			}
		case "Positive Numbers Only":
			if rule == "THIS > 0" {
				return i, nil
			}
		case "Percentage (0-100)":
			if rule == "THIS >= 0 && THIS <= 100" {
				return i, nil
			}
		}
	}

	return 0, nil 
}
