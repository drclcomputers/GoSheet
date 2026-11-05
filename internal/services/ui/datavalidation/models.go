// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// models.go provides the structure for data validation

package datavalidation

import "regexp"

var validationCellRefRegex = regexp.MustCompile(`\b([A-Z]+)(\d+)\b`)

// ValidationPreset represents a predefined validation type
type ValidationPreset struct {
	Name        string
	Description string
	BuildRule   func(params map[string]string) string
	Fields      []ValidationField
}

type ValidationField struct {
	Name        string
	Label       string
	Type        string
	Placeholder string
}
