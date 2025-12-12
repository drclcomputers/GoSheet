// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// struct.go provides the definition of the Cell struct used in the spreadsheet application.

package cell

import (
	"gosheet/internal/utils"

	"github.com/rivo/tview"
)

// Format style bits
const (
    FlagItalic = 1 << iota
    FlagBold
    FlagUnderline
    FlagAllCaps
    FlagStrikethrough
    FlagEditable
	FlagFormula
	FlagEvaluated
)

// Custom Cell definition
type Cell struct {
    RawValue      *string  
    Display       *string 
    Type          *string  

    Row           int32     
	Column        int32
    MaxWidth      int32
	MinWidth      int32
  
    Align         int8
	Flags 		  uint8	

	DecimalPoints      int32
    ThousandsSeparator rune
    DecimalSeparator   rune
    FinancialSign      rune
	DateTimeFormat     *string

	Color    utils.ColorRGB
    BgColor  utils.ColorRGB 

	Notes      *string
	Valrule    *string
	Valrulemsg *string
	              
    DependsOn     []*string          
    Dependents    []*string

	tvCell        *tview.TableCell
}

