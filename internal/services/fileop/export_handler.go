// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// export_handler.go handles CSV, TXT, and HTML export formats

package fileop

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
)

type ExportFormatHandler struct{}

// SupportsFormat checks whether this handler supports the format
func (h *ExportFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatCSV || format == FormatTXT || format == FormatHTML
}

// Write exports workbook to CSV, TXT, or HTML
func (h *ExportFormatHandler) Write(filename string, sheets []SheetInfo, activeSheet int) error {
	if activeSheet < 0 || activeSheet >= len(sheets) {
		activeSheet = 0
	}

	sheet := sheets[activeSheet]

	format, _ := DetectFormat(filename)
	switch format {
	case FormatCSV:
		return h.writeCSV(filename, sheet)
	case FormatTXT:
		return h.writeTXT(filename, sheet)
	case FormatHTML:
		return h.writeHTML(filename, sheet)
	default:
		return fmt.Errorf("unsupported export format")
	}
}

// writeCSV exports to CSV format
func (h *ExportFormatHandler) writeCSV(filename string, sheet SheetInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var maxRow, maxCol int32
	for key := range sheet.GlobalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	for row := int32(1); row <= maxRow; row++ {
		record := make([]string, maxCol)

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			if cellData, exists := sheet.GlobalData[key]; exists && cellData.RawValue != nil {
				if *cellData.RawValue != "" {
					record[col-1] = *cellData.RawValue
				} else if cellData.Display != nil {
					record[col-1] = *cellData.Display
				} else {
					record[col-1] = ""
				}
			} else {
				record[col-1] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// writeTXT exports to tab-delimited text format
func (h *ExportFormatHandler) writeTXT(filename string, sheet SheetInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var maxRow, maxCol int32
	for key := range sheet.GlobalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	for row := int32(1); row <= maxRow; row++ {
		var values []string

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			if cellData, exists := sheet.GlobalData[key]; exists && cellData.RawValue != nil {
				if *cellData.RawValue != "" {
					values = append(values, *cellData.RawValue)
				} else if cellData.Display != nil {
					values = append(values, *cellData.Display)
				} else {
					values = append(values, "")
				}
			} else {
				values = append(values, "")
			}
		}

		line := strings.Join(values, "\t") + "\n"
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// writeHTML exports to HTML table format
func (h *ExportFormatHandler) writeHTML(filename string, sheet SheetInfo) error {
	var maxRow, maxCol int32
	for key := range sheet.GlobalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>GoSheet Export - `)
	html.WriteString(sheet.Name)
	html.WriteString(`</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; }
		th { background-color: #4CAF50; color: white; font-weight: bold; }
		tr:hover { background-color: #f5f5f5; }
		.formula-cell { border-left: 3px solid #2196F3; }
	</style>
</head>
<body>
	<h1>`)
	html.WriteString(sheet.Name)
	html.WriteString(`</h1>
	<table>
		<thead>
			<tr>
				<th>#</th>
`)

	for col := int32(1); col <= maxCol; col++ {
		html.WriteString(fmt.Sprintf("<th>%s</th>\n", utils.ColumnName(col)))
	}

	html.WriteString("</tr>\n</thead>\n<tbody>\n")

	for row := int32(1); row <= maxRow; row++ {
		html.WriteString("<tr>\n")
		html.WriteString(fmt.Sprintf("<td style=\"background-color: #4CAF50; color: white; font-weight: bold;\"><b>%d</b></td>\n", row))

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			cellData, exists := sheet.GlobalData[key]

			if !exists || cellData == nil || cellData.Display == nil {
				html.WriteString("<td></td>\n")
				continue
			}

			content := *cellData.Display

			if cellData.HasFlag(cell.FlagAllCaps) {
				content = strings.ToUpper(content)
			}

			style := h.buildCellStyle(cellData)

			class := ""
			tooltip := ""
			if cellData.HasFlag(cell.FlagFormula) {
				class = " class=\"formula-cell\""
				if cellData.RawValue != nil {
					tooltip = fmt.Sprintf(" title=\"Formula: %s\"", htmlEscape(*cellData.RawValue))
				}
			}

			if style != "" {
				html.WriteString(
					fmt.Sprintf("<td%s style=\"%s\"%s>%s</td>\n",
						class, style, tooltip, htmlEscape(content)),
				)
			} else {
				html.WriteString(
					fmt.Sprintf("<td%s%s>%s</td>\n",
						class, tooltip, htmlEscape(content)),
				)
			}
		}

		html.WriteString("</tr>\n")
	}

	html.WriteString("</tbody>\n</table>\n</body>\n</html>")

	_, err = file.WriteString(html.String())
	return err
}

// buildCellStyle builds CSS style string for a cell
func (h *ExportFormatHandler) buildCellStyle(cellData *cell.Cell) string {
	var styles []string

	if cellData.Color != utils.ColorOptions["White"] {
		styles = append(styles, "color: "+cellData.Color.Hex())
	}

	if cellData.BgColor != utils.ColorOptions["Black"] {
		styles = append(styles, "background-color: "+cellData.BgColor.Hex())
	}

	switch cellData.Align {
	case 1:
		styles = append(styles, "text-align: left")
	case 2:
		styles = append(styles, "text-align: center")
	case 3:
		styles = append(styles, "text-align: right")
	}

	if cellData.HasFlag(cell.FlagBold) {
		styles = append(styles, "font-weight: bold")
	}

	if cellData.HasFlag(cell.FlagItalic) {
		styles = append(styles, "font-style: italic")
	}

	if cellData.HasFlag(cell.FlagUnderline) && cellData.HasFlag(cell.FlagStrikethrough) {
		styles = append(styles, "text-decoration: underline line-through")
	} else if cellData.HasFlag(cell.FlagUnderline) {
		styles = append(styles, "text-decoration: underline")
	} else if cellData.HasFlag(cell.FlagStrikethrough) {
		styles = append(styles, "text-decoration: line-through")
	}

	styles = append(styles, "padding: 8px")

	return strings.Join(styles, "; ")
}

// htmlEscape escapes HTML special characters
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
