package printer

import (
	"encoding/json"
	"fmt"
)

type ReceiptFormat struct {
	Settings FormatSettings   `json:"settings"`
	Elements []ReceiptElement `json:"elements"`
}

type FormatSettings struct {
	NoCut bool `json:"noCut"`
}

type ReceiptElement struct {
	Type    string                 `json:"type"` // "text", "linebreak", "table"
	Content map[string]interface{} `json:"content,omitempty"`
}

type TextElement struct {
	Text      string `json:"text"`
	Font      string `json:"font"` // "A" or "B"
	Bold      bool   `json:"bold"`
	Alignment string `json:"alignment"` // "left", "center", "right"
}

type TableElement struct {
	Font    string        `json:"font"`
	Columns []TableColumn `json:"columns"`
	Rows    [][]string    `json:"rows"`
}

func ReceiptFormatFromJson(jsonContent string) (*ReceiptFormat, error) {
	var format ReceiptFormat
	if err := json.Unmarshal([]byte(jsonContent), &format); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &format, nil
}

func printElement(printer *Printer, element ReceiptElement) error {
	switch element.Type {
	case "text":
		return printTextElement(printer, element.Content)
	case "linebreak":
		printer.LineBreak()
		return nil
	case "table":
		return printTableElement(printer, element.Content)
	default:
		return fmt.Errorf("unknown element type: %s", element.Type)
	}
}

func printTextElement(printer *Printer, content map[string]interface{}) error {
	text, _ := content["text"].(string)
	fontStr, _ := content["font"].(string)
	bold, _ := content["bold"].(bool)
	alignmentStr, _ := content["alignment"].(string)

	font := parseFont(fontStr)
	alignment := parseAlignment(alignmentStr)

	printer.QueueText(text, font, bold, alignment)
	return nil
}

func printTableElement(printer *Printer, content map[string]interface{}) error {
	fontStr, _ := content["font"].(string)
	font := parseFont(fontStr)

	// Parse columns
	columnsData, ok := content["columns"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid columns format")
	}

	var columns []TableColumn
	for _, col := range columnsData {
		colMap, ok := col.(map[string]interface{})
		if !ok {
			continue
		}
		header, _ := colMap["header"].(string)
		widthPct, _ := colMap["width_pct"].(float64)
		columns = append(columns, TableColumn{
			Header:   header,
			WidthPct: int(widthPct),
		})
	}

	table, err := NewTable(font, columns)
	if err != nil {
		return err
	}

	// Parse rows
	rowsData, ok := content["rows"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid rows format")
	}

	for _, row := range rowsData {
		rowArray, ok := row.([]interface{})
		if !ok {
			continue
		}
		var rowStrings []string
		for _, cell := range rowArray {
			cellStr, _ := cell.(string)
			rowStrings = append(rowStrings, cellStr)
		}
		table.AddRow(rowStrings...)
	}

	printer.QueueTable(table)
	return nil
}

func parseFont(fontStr string) Font {
	if fontStr == "B" {
		return FontB
	}
	return FontA
}

func parseAlignment(alignmentStr string) Alignment {
	switch alignmentStr {
	case "center":
		return AlignCenter
	case "right":
		return AlignRight
	default:
		return AlignLeft
	}
}
