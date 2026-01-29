package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hennedo/escpos"
)

// ReceiptFormat represents the complete receipt structure
type ReceiptFormat struct {
	Elements []ReceiptElement `json:"elements"`
}

// ReceiptElement represents a single element in the receipt
type ReceiptElement struct {
	Type    string                 `json:"type"` // "text", "linebreak", "table"
	Content map[string]interface{} `json:"content,omitempty"`
}

// TextElement represents a text element
type TextElement struct {
	Text      string `json:"text"`
	Font      string `json:"font"` // "A" or "B"
	Bold      bool   `json:"bold"`
	Alignment string `json:"alignment"` // "left", "center", "right"
}

// TableElement represents a table element
type TableElement struct {
	Font    string        `json:"font"`
	Columns []TableColumn `json:"columns"`
	Rows    [][]string    `json:"rows"`
}

// LoadReceiptFormat loads a receipt format from a JSON file
func LoadReceiptFormat(filename string) (*ReceiptFormat, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var format ReceiptFormat
	if err := json.Unmarshal(data, &format); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &format, nil
}

// PrintReceipt prints a receipt based on the format
func PrintReceipt(printer *escpos.Escpos, format *ReceiptFormat) error {
	for i, element := range format.Elements {
		if err := printElement(printer, element); err != nil {
			return fmt.Errorf("failed to print element %d: %w", i, err)
		}
	}
	return nil
}

func printElement(printer *escpos.Escpos, element ReceiptElement) error {
	switch element.Type {
	case "text":
		return printTextElement(printer, element.Content)
	case "linebreak":
		LineBreak(printer)
		return nil
	case "table":
		return printTableElement(printer, element.Content)
	default:
		return fmt.Errorf("unknown element type: %s", element.Type)
	}
}

func printTextElement(printer *escpos.Escpos, content map[string]interface{}) error {
	text, _ := content["text"].(string)
	fontStr, _ := content["font"].(string)
	bold, _ := content["bold"].(bool)
	alignmentStr, _ := content["alignment"].(string)

	font := parseFont(fontStr)
	alignment := parseAlignment(alignmentStr)

	PrintText(printer, text, font, bold, alignment)
	return nil
}

func printTableElement(printer *escpos.Escpos, content map[string]interface{}) error {
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

	table.Print(printer)
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
