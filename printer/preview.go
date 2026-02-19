package printer

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
)

const (
	fontASize = 16
	fontBSize = 11
)

type PreviewCanvas struct {
	ctx        *gg.Context
	y          float64
	lineHeight float64
	widthPx    float64
	fontPath   string
	charW      float64
}

func (c *PreviewCanvas) SetFont(font Font) error {
	var size float64
	switch font {
	case FontB:
		size = fontBSize
	default:
		size = fontASize
	}

	if err := c.ctx.LoadFontFace(c.fontPath, size); err != nil {
		return err
	}

	c.charW, _ = c.ctx.MeasureString("M")
	c.lineHeight = size * 1.4
	return nil
}

func defaultMonospaceFontPath() (string, error) {
	home, _ := os.UserHomeDir()

	paths := []string{
		filepath.Join(home, "Library/Fonts/DejaVuSansMono.ttf"),
		filepath.Join(home, "Library/Fonts/DejaVu Sans Mono.ttf"),

		"/Library/Fonts/DejaVuSansMono.ttf",
		"/Library/Fonts/DejaVu Sans Mono.ttf",
		"/System/Library/Fonts/Supplemental/DejaVu Sans Mono.ttf",

		"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
		"/usr/share/fonts/TTF/DejaVuSansMono.ttf",

		"C:\\Windows\\Fonts\\DejaVuSansMono.ttf",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("DejaVuSansMono.ttf not found (checked user and system font directories)")
}

func NewPreviewCanvas(
	fontSize float64,
	widthChars int,
) (*PreviewCanvas, error) {

	fontPath, err := defaultMonospaceFontPath()
	if err != nil {
		return nil, err
	}

	tmp := gg.NewContext(10, 10)
	if err := tmp.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, err
	}

	charW, _ := tmp.MeasureString("M")
	widthPx := charW * float64(widthChars+2)

	ctx := gg.NewContext(int(widthPx+20), 5000)
	ctx.SetRGB(1, 1, 1)
	ctx.Clear()

	if err := ctx.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, err
	}

	ctx.SetRGB(0, 0, 0)

	return &PreviewCanvas{
		ctx:        ctx,
		y:          20,
		lineHeight: fontSize * 1.4,
		widthPx:    widthPx,
		fontPath:   fontPath,
	}, nil
}

func (c *PreviewCanvas) DrawLine(text string, align Alignment) {
	w, _ := c.ctx.MeasureString(text)

	x := c.charW
	switch align {
	case AlignCenter:
		x = (c.widthPx - w) / 2
	case AlignRight:
		x = c.widthPx - w - c.charW
	}

	c.ctx.DrawString(text, x, c.y)
	c.y += c.lineHeight
}

func RenderReceiptPreviewCanvas(
	format *ReceiptFormat,
	outputPath string,
) error {

	canvas, err := NewPreviewCanvas(fontASize, 46)
	if err != nil {
		return err
	}

	for _, element := range format.Elements {
		switch element.Type {

		case "text":
			text := element.Content["text"].(string)
			font := parseFont(element.Content["font"].(string))
			align := parseAlignment(element.Content["alignment"].(string))
			bold := element.Content["bold"].(bool)

			canvas.SetFont(font)

			lines := wrapAndAlignText(text, font, bold, align)
			for _, l := range lines {
				canvas.DrawLine(l, align)
			}

		case "linebreak":
			canvas.y += canvas.lineHeight

		case "table":
			table, err := buildTableFromElement(element)
			if err != nil {
				return err
			}

			canvas.SetFont(table.font)

			for _, line := range table.renderLines() {
				canvas.DrawLine(line, AlignLeft)
			}
		}
	}

	return canvas.SaveTrimmedPNG(outputPath)
}

func buildTableFromElement(e ReceiptElement) (*Table, error) {
	font := parseFont(e.Content["font"].(string))

	var cols []TableColumn
	for _, c := range e.Content["columns"].([]interface{}) {
		col := c.(map[string]interface{})

		widthPct := 0
		if raw, ok := col["width_pct"]; ok && raw != nil {
			if f, ok := raw.(float64); ok {
				widthPct = int(f)
			}
		}

		cols = append(cols, TableColumn{
			Header:   col["header"].(string),
			WidthPct: widthPct,
		})
	}

	table, err := NewTable(font, cols)
	if err != nil {
		return nil, err
	}

	for _, r := range e.Content["rows"].([]interface{}) {
		var row []string
		for _, cell := range r.([]interface{}) {
			row = append(row, cell.(string))
		}
		table.AddRow(row...)
	}

	return table, nil
}

func (c *PreviewCanvas) SaveTrimmedPNG(path string) error {
	height := int(c.y + c.lineHeight)
	img := c.ctx.Image()

	cropped := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(0, 0, img.Bounds().Dx(), height))

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, cropped)
}
