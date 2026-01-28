package main

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/google/gousb"
	"github.com/hennedo/escpos"
)

/*
   =====================
   Public types & enums
   =====================
*/

type Alignment int
type Font int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

const (
	FontA Font = iota
	FontB
)

/*
   =====================
   USB printer config
   =====================
*/

const (
	VendorID  gousb.ID = 0x04b8 // Epson
	ProductID gousb.ID = 0x0202 // TM-T88V
)

/*
   =====================
   Printer setup
   =====================
*/

func GetPrinter() (*escpos.Escpos, *gousb.Context) {
	ctx := gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(VendorID, ProductID)
	if err != nil {
		log.Fatal(err)
	}
	if dev == nil {
		log.Fatal("printer not found")
	}

	cfg, err := dev.Config(1)
	if err != nil {
		log.Fatal(err)
	}

	intf, err := cfg.Interface(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	ep, err := intf.OutEndpoint(1)
	if err != nil {
		log.Fatal(err)
	}

	return escpos.New(ep), ctx
}

/*
   =====================
   ESC/POS state helpers
   =====================
*/

func applyFont(p *escpos.Escpos, font Font) {
	switch font {
	case FontB:
		p.WriteRaw([]byte{0x1B, 0x4D, 0x01}) // ESC M 1
	default:
		p.WriteRaw([]byte{0x1B, 0x4D, 0x00}) // ESC M 0
	}
}

func applyDoubleWidth(p *escpos.Escpos, enabled bool) {
	if enabled {
		p.WriteRaw([]byte{0x1D, 0x21, 0x10}) // GS ! double width
	} else {
		p.WriteRaw([]byte{0x1D, 0x21, 0x00}) // GS ! normal
	}
}

func applyAlignment(p *escpos.Escpos, align Alignment) {
	switch align {
	case AlignCenter:
		p.WriteRaw([]byte{0x1B, 0x61, 0x01}) // ESC a 1
	case AlignRight:
		p.WriteRaw([]byte{0x1B, 0x61, 0x02}) // ESC a 2
	default:
		p.WriteRaw([]byte{0x1B, 0x61, 0x00}) // ESC a 0
	}
}

func resetPrinterState(p *escpos.Escpos) {
	applyFont(p, FontA)
	applyDoubleWidth(p, false)
	applyAlignment(p, AlignLeft)
}

/*
   =====================
   Layout logic
   =====================
*/

func lineWidth(font Font, doubleWidth bool) int {
	width := 48
	if font == FontB {
		width = 62
	}
	if doubleWidth {
		width /= 2
	}
	return width
}

func alignLine(line string, width int, align Alignment) string {
	lineLen := runeLen(line)
	if lineLen >= width {
		return line
	}

	padding := width - lineLen

	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + line
	case AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + line + strings.Repeat(" ", right)
	default:
		return line
	}
}

func WrapAndAlignText(
	text string,
	font Font,
	doubleWidth bool,
	align Alignment,
) []string {

	width := lineWidth(font, doubleWidth)
	words := strings.Fields(text)

	var lines []string
	var current string

	for _, word := range words {
		wordLen := runeLen(word)

		// Hard-break extremely long words
		if wordLen > width {
			if current != "" {
				lines = append(lines, alignLine(current, width, align))
				current = ""
			}

			parts := hyphenateWord(word, width)
			for i := 0; i < len(parts)-1; i++ {
				lines = append(lines, alignLine(parts[i], width, align))
			}
			current = parts[len(parts)-1]
			continue
		}

		if current == "" {
			current = word
			continue
		}

		if align == AlignRight {
			// RIGHT-ALIGNED: build line from the right edge
			if runeLen(word)+1+runeLen(current) <= width {
				current = word + " " + current
			} else {
				lines = append(lines, alignLine(current, width, align))
				current = word
			}
		} else {
			// LEFT / CENTER
			if runeLen(current)+1+wordLen <= width {
				current += " " + word
			} else {
				lines = append(lines, alignLine(current, width, align))
				current = word
			}
		}
	}

	if current != "" {
		lines = append(lines, alignLine(current, width, align))
	}

	return lines
}

func hyphenateWord(word string, width int) []string {
	var parts []string
	runes := []rune(word)

	for len(runes) > width {
		part := string(runes[:width-1]) + "-"
		parts = append(parts, part)
		runes = runes[width-1:]
	}

	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}

	return parts
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func sanitizeText(s string) string {
	replacer := strings.NewReplacer(
		"£", "\x9C", // CP858 pound
		"€", "\xD5", // CP858 euro
		"–", "-", // en dash
		"—", "-", // em dash
		"“", "\"",
		"”", "\"",
		"’", "'",
	)

	return replacer.Replace(s)
}

func applyCodePage(p *escpos.Escpos) {
	// ESC t 19 → CP858 (supports £ and €)
	p.WriteRaw([]byte{0x1B, 0x74, 0x13})
}

/*
   =====================
   High-level print API
   =====================
*/

func PrintText(
	p *escpos.Escpos,
	text string,
	font Font,
	doubleWidth bool,
	align Alignment,
) {
	ResetPrinter(p)
	applyCodePage(p)

	text = sanitizeText(text)

	applyFont(p, font)
	applyDoubleWidth(p, doubleWidth)
	applyAlignment(p, align)

	lines := WrapAndAlignText(text, font, doubleWidth, align)
	for _, line := range lines {
		p.Write(line)
		p.Write("\n")
	}

	resetPrinterState(p)
}

func ResetPrinter(p *escpos.Escpos) {
	p.WriteRaw([]byte{0x1B, 0x40}) // ESC @
}

func LineBreak(p *escpos.Escpos) {
	p.Write("\n")
}
