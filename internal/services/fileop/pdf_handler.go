// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// pdf_handler.go handles .pdf export (write-only)

package fileop

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/jung-kurt/gofpdf"
)

// PDFFormatHandler handles .pdf export (write-only)
type PDFFormatHandler struct{}

func (h *PDFFormatHandler) SupportsFormat(format FileFormat) bool {
	return format == FormatPDF
}

func (h *PDFFormatHandler) Write(filename string, sheets []SheetInfo, activeSheet int) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetCreator("GoSheet", true)
	pdf.SetAuthor("GoSheet User", true)
	pdf.SetTitle("Exported Spreadsheet", true)
	pdf.SetMargins(15, 15, 15)

	pdf.AddPage()

	for i, sheet := range sheets {
		if i > 0 {
			pdf.AddPage()
		}

		if err := h.writeSheetToPDF(pdf, sheet, i+1, len(sheets)); err != nil {
			return err
		}
	}

	return pdf.OutputFileAndClose(filename)
}

func (h *PDFFormatHandler) writeSheetToPDF(pdf *gofpdf.Fpdf, sheet SheetInfo, sheetNum, totalSheets int) error {
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, fmt.Sprintf("Sheet %d/%d: %s", sheetNum, totalSheets, sheet.Name), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	var maxRow, maxCol int32
	for k := range sheet.GlobalData {
		if int32(k[0]) > maxRow {
			maxRow = int32(k[0])
		}
		if int32(k[1]) > maxCol {
			maxCol = int32(k[1])
		}
	}

	if maxCol == 0 || maxRow == 0 {
		pdf.SetFont("Helvetica", "I", 12)
		pdf.Cell(0, 10, "(Empty sheet)")
		return nil
	}

	const (
		margin    = 15.0
		headerH   = 8.0
		rowH      = 7.0
		maxColW   = 40.0
		minColW   = 12.0
		pageWidth = 297.0 - 2*margin
	)

	colCount := int(maxCol)
	colWidths := make([]float64, colCount)
	baseWidth := pageWidth / float64(colCount)

	for i := range colWidths {
		colWidths[i] = baseWidth
		if colWidths[i] < minColW {
			colWidths[i] = minColW
		}
		if colWidths[i] > maxColW {
			colWidths[i] = maxColW
		}
	}

	for _, c := range sheet.GlobalData {
		if c.MinWidth > 0 {
			needed := float64(c.MinWidth) * 1.8
			idx := int(c.Column - 1)
			if idx >= 0 && idx < len(colWidths) {
				if needed > colWidths[idx] {
					colWidths[idx] = needed
					if colWidths[idx] > maxColW {
						colWidths[idx] = maxColW
					}
				}
			}
		}
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	h.drawHeaderRow(pdf, colWidths, headerH, maxCol)

	pdf.SetFont("Courier", "", 9)

	for row := int32(1); row <= maxRow; row++ {
		if pdf.GetY() > 185 {
			pdf.AddPage()
			h.drawHeaderRow(pdf, colWidths, headerH, maxCol)
			pdf.SetFont("Courier", "", 9)
		}

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			cellData, exists := sheet.GlobalData[key]

			text := ""
			align := "L"
			style := ""
			fill := false

			if exists && cellData != nil {
				if cellData.Display != nil {
					text = tr(*cellData.Display)
					text = cell.StripTviewTags(text)

					if len([]rune(text)) > 40 {
						runes := []rune(text)
						text = string(runes[:37]) + "..."
					}
				}

				switch cellData.Align {
				case 1:
					align = "L"
				case 2:
					align = "C"
				case 3:
					align = "R"
				}

				if cellData.HasFlag(cell.FlagBold) && cellData.HasFlag(cell.FlagItalic) {
					style = "BI"
				} else if cellData.HasFlag(cell.FlagBold) {
					style = "B"
				} else if cellData.HasFlag(cell.FlagItalic) {
					style = "I"
				}

				if !cellData.BgColor.IsDefaultBlack() && !cellData.BgColor.IsDefaultWhite() {
					r, g, b := cellData.BgColor[0], cellData.BgColor[1], cellData.BgColor[2]
					pdf.SetFillColor(int(r), int(g), int(b))
					fill = true
				}
			}

			if style != "" {
				pdf.SetFont("Courier", style, 9)
			}

			colIdx := int(col - 1)
			if colIdx >= 0 && colIdx < len(colWidths) {
				pdf.CellFormat(colWidths[colIdx], rowH, text, "1", 0, align, fill, 0, "")
			}

			if fill {
				pdf.SetFillColor(255, 255, 255)
			}
			if style != "" {
				pdf.SetFont("Courier", "", 9)
			}
		}
		pdf.Ln(-1)
	}

	return nil
}

// drawHeaderRow draws the column header row
func (h *PDFFormatHandler) drawHeaderRow(pdf *gofpdf.Fpdf, colWidths []float64, height float64, maxCol int32) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(220, 220, 220)

	for i := range maxCol {
		colName := utils.ColumnName(i + 1)

		colIdx := int(i)
		if colIdx >= 0 && colIdx < len(colWidths) {
			pdf.CellFormat(colWidths[colIdx], height, colName, "1", 0, "C", true, 0, "")
		}
	}
	pdf.Ln(-1)

	pdf.SetFillColor(255, 255, 255)
}
