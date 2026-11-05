// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// colour.go provides functions for managing colours

package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var ColorOptionNames = []string{"White", "Black", "Red", "Green", "Blue", "Yellow", "Orange", "Purple", "Pink", "Gray", "Custom..."}
var ColorOptions = map[string]ColorRGB{
	"White":  {255, 255, 255},
	"Black":  {0, 0, 0},
	"Red":    {255, 0, 0},
	"Green":  {0, 255, 0},
	"Blue":   {0, 0, 255},
	"Yellow": {255, 255, 0},
	"Orange": {255, 165, 0},
	"Purple": {128, 0, 128},
	"Pink":   {255, 192, 203},
	"Gray":   {128, 128, 128},
}

// Turns custom color to tcell type
func (c ColorRGB) ToTCellColor() tcell.Color {
	return tcell.NewRGBColor(int32(c[0]), int32(c[1]), int32(c[2]))
}

// Returns the color as HEX
func (c ColorRGB) Hex() string {
	return fmt.Sprintf("#%02X%02X%02X", c[0], c[1], c[2])
}

func ParseHexColor(hex string) (ColorRGB, error) {
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.ToUpper(hex)

	if len(hex) != 6 {
		return ColorRGB{255, 255, 255}, fmt.Errorf("invalid hex color format: expected 6 characters, got %d", len(hex))
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return ColorRGB{255, 255, 255}, fmt.Errorf("invalid red component: %v", err)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return ColorRGB{255, 255, 255}, fmt.Errorf("invalid green component: %v", err)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return ColorRGB{255, 255, 255}, fmt.Errorf("invalid blue component: %v", err)
	}

	return ColorRGB{uint8(r), uint8(g), uint8(b)}, nil
}

// ColorToExcel converts ColorRGB to Excel color format (no # prefix)
func (c ColorRGB) ToExcel() string {
	return fmt.Sprintf("%02X%02X%02X", c[0], c[1], c[2])
}

// IsDefaultWhite checks if color is default white
func (c ColorRGB) IsDefaultWhite() bool {
	return c[0] == 255 && c[1] == 255 && c[2] == 255
}

// IsDefaultBlack checks if color is default black
func (c ColorRGB) IsDefaultBlack() bool {
	return c[0] == 0 && c[1] == 0 && c[2] == 0
}

// Equals checks if two colors are equal
func (c ColorRGB) Equals(other ColorRGB) bool {
	return c[0] == other[0] && c[1] == other[1] && c[2] == other[2]
}
