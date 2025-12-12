// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// autofill.go provides functions for auto-filling data based on patterns

package table

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/services/ui"
	"gosheet/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FillDirection represents the direction to fill
type FillDirection int

const (
	FillDown FillDirection = iota
	FillRight
	FillUp
	FillLeft
)

// ShowFillDialog displays options for filling data from selection
func ShowFillDialog(app *tview.Application, table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	r1, c1, r2, c2 := getSelectionRange(table)
	
	if (r1 == r2 && c1 == c2) {
		ui.ShowWarningModal(app, table, "Please select at least 2 cells to create a fill pattern.")
		return
	}

	form := tview.NewForm()

	directions := []string{}
	directionMap := map[string]FillDirection{}
	
	if r2 < utils.MAX_ROWS {
		directions = append(directions, "Down")
		directionMap["Down"] = FillDown
	}
	if c2 < utils.MAX_COLS {
		directions = append(directions, "Right")
		directionMap["Right"] = FillRight
	}
	if r1 > 1 {
		directions = append(directions, "Up")
		directionMap["Up"] = FillUp
	}
	if c1 > 1 {
		directions = append(directions, "Left")
		directionMap["Left"] = FillLeft
	}

	if len(directions) == 0 {
		ui.ShowWarningModal(app, table, "No space available to fill in any direction.")
		return
	}

	var selectedDirection FillDirection
	form.AddDropDown("Direction:", directions, 0, func(option string, index int) {
		selectedDirection = directionMap[option]
	})

	countInput := tview.NewInputField().
		SetLabel("Number of cells:").
		SetText("10").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)
	form.AddFormItem(countInput)

	fillTypes := []string{
		"Linear Series (1, 2, 3, 4...)",
		"Growth Series (1, 2, 4, 8...)",
		"Date Series",
		"Text Pattern",
		"Copy Values",
	}
	
	var fillType int
	form.AddDropDown("Fill Type:", fillTypes, 0, func(option string, index int) {
		fillType = index
	})

	form.AddButton("Fill", func() {
		countStr := strings.TrimSpace(countInput.GetText())
		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			ui.ShowWarningModal(app, form, "Please enter a valid positive number.")
			return
		}

		performFill(table, r1, c1, r2, c2, selectedDirection, count, fillType)
		app.SetRoot(table, true).SetFocus(table)
		clearSelectionRange()
	})

	form.AddButton("x Cancel", func() {
		app.SetRoot(table, true).SetFocus(table)
	})

	form.SetBorder(true).
		SetTitle(" Fill Data from Selection ").
		SetBorderColor(tcell.ColorBlue).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// performFill executes the fill operation
func performFill(table *tview.Table, r1, c1, r2, c2 int32, direction FillDirection, count int, fillType int) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	var sourceCells []*cell.Cell
	if direction == FillDown || direction == FillUp {
		for r := r1; r <= r2; r++ {
			key := [2]int{int(r), int(c1)}
			if cellData, exists := activeData[key]; exists {
				sourceCells = append(sourceCells, cellData)
			} else {
				sourceCells = append(sourceCells, cell.NewCell(r, c1, ""))
			}
		}
	} else {
		for c := c1; c <= c2; c++ {
			key := [2]int{int(r1), int(c)}
			if cellData, exists := activeData[key]; exists {
				sourceCells = append(sourceCells, cellData)
			} else {
				sourceCells = append(sourceCells, cell.NewCell(r1, c, ""))
			}
		}
	}

	pattern := detectPattern(sourceCells, fillType)

	var fillR1, fillC1, fillR2, fillC2 int32
	switch direction {
	case FillDown:
		fillR1, fillC1 = r2+1, c1
		fillR2, fillC2 = r2+int32(count), c2
	case FillRight:
		fillR1, fillC1 = r1, c2+1
		fillR2, fillC2 = r2, c2+int32(count)
	case FillUp:
		fillR1, fillC1 = r1-int32(count), c1
		fillR2, fillC2 = r1-1, c2
	case FillLeft:
		fillR1, fillC1 = r1, c1-int32(count)
		fillR2, fillC2 = r2, c1-1
	}

	oldCells := captureCellRange(fillR1, fillC1, fillR2, fillC2)

	fillIndex := len(sourceCells)
	if direction == FillDown || direction == FillRight {
		for i := range count {
			if direction == FillDown {
				r := r2 + 1 + int32(i)
				for colOffset := int32(0); colOffset <= c2-c1; colOffset++ {
					c := c1 + colOffset
					value := pattern.GetNext(fillIndex)
					createFilledCell(table, r, c, value)
					fillIndex++
				}
			} else {
				c := c2 + 1 + int32(i)
				for rowOffset := int32(0); rowOffset <= r2-r1; rowOffset++ {
					r := r1 + rowOffset
					value := pattern.GetNext(fillIndex)
					createFilledCell(table, r, c, value)
					fillIndex++
				}
			}
		}
	} else {
		for i := count - 1; i >= 0; i-- {
			if direction == FillUp {
				r := r1 - int32(count) + int32(i)
				for colOffset := int32(0); colOffset <= c2-c1; colOffset++ {
					c := c1 + colOffset
					value := pattern.GetNext(fillIndex)
					createFilledCell(table, r, c, value)
					fillIndex++
				}
			} else {
				c := c1 - int32(count) + int32(i)
				for rowOffset := int32(0); rowOffset <= r2-r1; rowOffset++ {
					r := r1 + rowOffset
					value := pattern.GetNext(fillIndex)
					createFilledCell(table, r, c, value)
					fillIndex++
				}
			}
		}
	}

	newCells := captureCellRange(fillR1, fillC1, fillR2, fillC2)
	RecordMultiCellAction(ActionPasteCells, fillR1, fillC1, fillR2, fillC2, oldCells, newCells)
}

// createFilledCell creates a new cell with the given value
func createFilledCell(table *tview.Table, r, c int32, value string) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	newCell := cell.GetOrCreateCell(table, r, c, activeData)
	*newCell.RawValue = value
	*newCell.Display = value

	if activeViewport.IsVisible(r, c) {
		visualR, visualC := activeViewport.ToRelative(r, c)
		table.SetCell(int(visualR), int(visualC), newCell.ToTViewCell())
	}
}

// Pattern interface for different fill types
type Pattern interface {
	GetNext(index int) string
}

// LinearPattern for 1, 2, 3...
type LinearPattern struct {
	start  float64
	step   float64
	prefix string
	suffix string
}

func (p *LinearPattern) GetNext(index int) string {
	val := p.start + float64(index)*p.step
	var numStr string
	if val == float64(int64(val)) {
		numStr = fmt.Sprintf("%d", int64(val))
	} else {
		numStr = fmt.Sprintf("%.2f", val)
	}
	return p.prefix + numStr + p.suffix
}

// GrowthPattern for 2, 4, 8...
type GrowthPattern struct {
	start  float64
	factor float64
}

func (p *GrowthPattern) GetNext(index int) string {
	val := p.start
	for range index {
		val *= p.factor
	}
	if val == float64(int64(val)) {
		return fmt.Sprintf("%d", int64(val))
	}
	return fmt.Sprintf("%.2f", val)
}

// DatePattern for date series
type DatePattern struct {
	start time.Time
	days  int
}

func (p *DatePattern) GetNext(index int) string {
	date := p.start.AddDate(0, 0, index*p.days)
	return date.Format("2006-01-02")
}

// TextPattern for repeating text
type TextPattern struct {
	values []string
}

func (p *TextPattern) GetNext(index int) string {
	return p.values[index%len(p.values)]
}

// CopyPattern for copying values
type CopyPattern struct {
	values []string
}

func (p *CopyPattern) GetNext(index int) string {
	return p.values[index%len(p.values)]
}

// detectPattern analyzes source cells and creates appropriate pattern
func detectPattern(cells []*cell.Cell, fillType int) Pattern {
	if len(cells) == 0 {
		return &TextPattern{values: []string{""}}
	}

	values := make([]string, len(cells))
	for i, c := range cells {
		if c.RawValue != nil {
			values[i] = strings.TrimSpace(*c.RawValue)
		} else {
			values[i] = ""
		}
	}

	switch fillType {
	case 0:
		return detectLinearPattern(values)
	case 1:
		return detectGrowthPattern(values)
	case 2:
		return detectDatePattern(values)
	case 3:
		return &TextPattern{values: values}
	case 4:
		return &CopyPattern{values: values}
	default:
		return &CopyPattern{values: values}
	}
}

// detectLinearPattern finds arithmetic progression
func detectLinearPattern(values []string) Pattern {
	type numWithContext struct {
		num    float64
		prefix string
		suffix string
	}
	
	numsWithContext := []numWithContext{}
	
	for _, v := range values {
		v = strings.TrimSpace(v)
		
		var prefix, suffix, numStr string
		var foundNum bool
		
		numStart, numEnd := -1, -1
		inNumber := false
		hasDecimal := false
		
		for i, ch := range v {
			if ch >= '0' && ch <= '9' {
				if !inNumber {
					numStart = i
					inNumber = true
				}
				numEnd = i + 1
			} else if ch == '.' && inNumber && !hasDecimal {
				hasDecimal = true
				numEnd = i + 1
			} else if inNumber {
				break
			}
		}
		
		if numStart >= 0 && numEnd > numStart {
			prefix = v[:numStart]
			numStr = v[numStart:numEnd]
			if numEnd < len(v) {
				suffix = v[numEnd:]
			}
			
			if num, err := strconv.ParseFloat(numStr, 64); err == nil {
				numsWithContext = append(numsWithContext, numWithContext{
					num:    num,
					prefix: prefix,
					suffix: suffix,
				})
				foundNum = true
			}
		}
		
		if !foundNum {
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				numsWithContext = append(numsWithContext, numWithContext{
					num:    num,
					prefix: "",
					suffix: "",
				})
			}
		}
	}

	if len(numsWithContext) < 2 {
		return &LinearPattern{start: 1, step: 1, prefix: "", suffix: ""}
	}

	prefix := numsWithContext[0].prefix
	suffix := numsWithContext[0].suffix
	
	step := numsWithContext[1].num - numsWithContext[0].num
	
	return &LinearPattern{
		start:  numsWithContext[0].num,
		step:   step,
		prefix: prefix,
		suffix: suffix,
	}
}

// detectGrowthPattern finds geometric progression
func detectGrowthPattern(values []string) Pattern {
	nums := []float64{}
	for _, v := range values {
		if num, err := strconv.ParseFloat(v, 64); err == nil && num != 0 {
			nums = append(nums, num)
		}
	}

	if len(nums) < 2 {
		return &GrowthPattern{start: 1, factor: 2}
	}

	factor := nums[1] / nums[0]
	return &GrowthPattern{start: nums[0], factor: factor}
}

// detectDatePattern finds date progression
func detectDatePattern(values []string) Pattern {
	dates := []time.Time{}
	formats := []string{"2006-01-02", "01/02/2006", "02-01-2006", "2006/01/02"}

	for _, v := range values {
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				dates = append(dates, t)
				break
			}
		}
	}

	if len(dates) < 2 {
		return &DatePattern{start: time.Now(), days: 1}
	}

	days := int(dates[1].Sub(dates[0]).Hours() / 24)
	if days == 0 {
		days = 1
	}
	return &DatePattern{start: dates[0], days: days}
}
