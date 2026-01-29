package printer

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/google/gousb"
	"github.com/hennedo/escpos"
)

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

const (
	VendorID  gousb.ID = 0x04b8 // Epson
	ProductID gousb.ID = 0x0202 // TM-T88V
)

type Printer struct {
	Printer *escpos.Escpos
	Context *gousb.Context
}

func (p *Printer) Initialise() {
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

	p.Printer = escpos.New(ep)
	p.Context = ctx

	p.HardResetPrinter()
	p.ApplyCodePage()
}

func (p *Printer) ApplyFont(font Font) {
	switch font {
	case FontB:
		p.Printer.WriteRaw([]byte{0x1B, 0x4D, 0x01}) // ESC M 1
	default:
		p.Printer.WriteRaw([]byte{0x1B, 0x4D, 0x00}) // ESC M 0
	}
}

func (p *Printer) ApplyDoubleWidth(enabled bool) {
	if enabled {
		p.Printer.WriteRaw([]byte{0x1D, 0x21, 0x10}) // GS ! double width
	} else {
		p.Printer.WriteRaw([]byte{0x1D, 0x21, 0x00}) // GS ! normal
	}
}

func (p *Printer) ApplyAlignment(align Alignment) {
	switch align {
	case AlignCenter:
		p.Printer.WriteRaw([]byte{0x1B, 0x61, 0x01}) // ESC a 1
	case AlignRight:
		p.Printer.WriteRaw([]byte{0x1B, 0x61, 0x02}) // ESC a 2
	default:
		p.Printer.WriteRaw([]byte{0x1B, 0x61, 0x00}) // ESC a 0
	}
}

func (p *Printer) ApplyDefaultTextSettings() {
	p.ApplyFont(FontA)
	p.ApplyDoubleWidth(false)
	p.ApplyAlignment(AlignLeft)
}

func lineWidth(font Font, doubleWidth bool) int {
	width := 46
	if font == FontB {
		width = 60
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

func wrapAndAlignText(text string, font Font, doubleWidth bool, align Alignment) []string {

	width := lineWidth(font, doubleWidth)
	words := strings.Fields(text)

	var lines []string
	var current string

	for _, word := range words {
		wordLen := runeLen(word)

		// Hard-break long words
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
		"£", "\x9C",
		"€", "\xD5",
		"–", "-",
		"—", "-",
		"“", "\"",
		"”", "\"",
		"’", "'",
	)

	return replacer.Replace(s)
}

func (p *Printer) ApplyCodePage() {
	p.Printer.WriteRaw([]byte{0x1B, 0x74, 0x13})
}

func (p *Printer) QueueText(text string, font Font, doubleWidth bool, align Alignment) {
	text = sanitizeText(text)
	p.ApplyFont(font)
	p.ApplyDoubleWidth(doubleWidth)
	p.ApplyAlignment(align)

	lines := wrapAndAlignText(text, font, doubleWidth, align)
	for _, line := range lines {
		p.Printer.Write(line)
		p.Printer.Write("\n")
	}

	p.ApplyDefaultTextSettings()
}

func (p *Printer) QueueReceiptFormat(format *ReceiptFormat) error {
	for i, element := range format.Elements {
		if err := printElement(p, element); err != nil {
			return fmt.Errorf("failed to print element %d: %w", i, err)
		}
	}
	return nil
}

func (p *Printer) QueueTable(t *Table) {
	p.ApplyFont(t.font)
	p.ApplyDoubleWidth(false)
	p.ApplyAlignment(AlignLeft)

	lines := t.renderLines()
	for _, l := range lines {
		p.Printer.Write(sanitizeText(l))
		p.Printer.Write("\n")
	}

	p.ApplyDefaultTextSettings()
}

func (p *Printer) Execute() {
	p.Printer.PrintAndCut()
}

func (p *Printer) ExecuteWithoutCut() {
	p.Printer.Print()
}

func (p *Printer) Cut() {
	p.Printer.WriteRaw([]byte{
		0x1D, 0x56, 0x42, 0x00, // GS V 66 0 (FULL CUT)
	})
	p.Printer.Print() // WriteRaw writes to a buffer, Print flushes the buffer
}

func (p *Printer) HardResetPrinter() {
	p.Printer.WriteRaw([]byte{0x1B, 0x40})
}

func (p *Printer) LineBreak() {
	p.Printer.Write("\n")
}
