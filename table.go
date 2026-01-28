package main

import (
	"fmt"
	"strings"

	"github.com/hennedo/escpos"
)

type TableColumn struct {
	Header   string
	WidthPct int // 0 means "auto"
}

type Table struct {
	columns    []TableColumn
	rows       [][]string
	font       Font
	widthChars int
}

func NewTable(
	font Font,
	cols []TableColumn,
) (*Table, error) {

	if len(cols) < 1 || len(cols) > 4 {
		return nil, fmt.Errorf("tables must have between 1 and 4 columns")
	}

	width := lineWidth(font, false)

	// Count unspecified widths
	totalPct := 0
	autoCols := 0
	for _, c := range cols {
		if c.WidthPct > 0 {
			totalPct += c.WidthPct
		} else {
			autoCols++
		}
	}

	if totalPct > 100 {
		return nil, fmt.Errorf("column widths exceed 100%%")
	}

	// Distribute remaining width equally
	if autoCols > 0 {
		remaining := 100 - totalPct
		each := remaining / autoCols
		for i := range cols {
			if cols[i].WidthPct == 0 {
				cols[i].WidthPct = each
			}
		}
	}

	return &Table{
		columns:    cols,
		font:       font,
		widthChars: width,
	}, nil
}

func (t *Table) AddRow(values ...string) error {
	if len(values) != len(t.columns) {
		return fmt.Errorf(
			"expected %d columns, got %d",
			len(t.columns),
			len(values),
		)
	}

	t.rows = append(t.rows, values)
	return nil
}

func (t *Table) colWidths() []int {
	// Subtract space for inner borders
	totalWidth := t.widthChars - (len(t.columns) * 2) - 1

	widths := make([]int, len(t.columns))
	used := 0

	for i, col := range t.columns {
		w := (totalWidth * col.WidthPct) / 100
		widths[i] = w
		used += w
	}

	// Fix rounding loss
	diff := totalWidth - used
	if diff > 0 {
		widths[len(widths)-1] += diff
	}

	return widths
}

func wrapCell(text string, width int) []string {
	words := strings.Fields(text)
	var lines []string
	var current string

	for _, w := range words {
		if runeLen(w) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			for runeLen(w) > width {
				lines = append(lines, w[:width-1]+"-")
				w = w[width-1:]
			}
			current = w
			continue
		}

		if current == "" {
			current = w
		} else if runeLen(current)+1+runeLen(w) <= width {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func (t *Table) renderLines() []string {
	colWidths := t.colWidths()
	var output []string

	renderRow := func(cells []string) {
		wrapped := make([][]string, len(cells))
		maxLines := 0

		for i, cell := range cells {
			wrapped[i] = wrapCell(cell, colWidths[i]-1)
			if len(wrapped[i]) > maxLines {
				maxLines = len(wrapped[i])
			}
		}

		for line := 0; line < maxLines; line++ {
			var row strings.Builder
			for i := range wrapped {
				row.WriteString("|")
				if line < len(wrapped[i]) {
					row.WriteString(padRight(wrapped[i][line], colWidths[i]-1))
				} else {
					row.WriteString(strings.Repeat(" ", colWidths[i]-1))
				}
			}
			row.WriteString("|")
			output = append(output, row.String())
		}
	}

	// Header
	headers := make([]string, len(t.columns))
	for i, c := range t.columns {
		headers[i] = c.Header
	}
	renderRow(headers)

	times := 1
	for _, i := range colWidths {
		times += i
	}

	bar := strings.Repeat("-", times)
	output = append(output, bar)

	// Rows
	for _, row := range t.rows {
		renderRow(row)
	}

	return output
}

func padRight(s string, width int) string {
	l := runeLen(s)
	if l >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-l)
}

func (t *Table) Print(p *escpos.Escpos) {
	ResetPrinter(p)
	applyFont(p, t.font)
	applyDoubleWidth(p, false)
	applyAlignment(p, AlignLeft)

	lines := t.renderLines()
	for _, l := range lines {
		p.Write(sanitizeText(l))
		p.Write("\n")
	}

	resetPrinterState(p)
}
