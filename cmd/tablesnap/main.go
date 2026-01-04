package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

//go:embed fonts/Inter-Regular.ttf
var interFont []byte

type Theme struct {
	Background color.Color
	Text       color.Color
	Header     color.Color
	HeaderBg   color.Color
	Border     color.Color
	AltRow     color.Color
}

var darkTheme = Theme{
	Background: color.RGBA{0x1a, 0x1a, 0x1a, 0xff},
	Text:       color.RGBA{0xe0, 0xe0, 0xe0, 0xff},
	Header:     color.RGBA{0x4f, 0xc3, 0xf7, 0xff},
	HeaderBg:   color.RGBA{0x26, 0x26, 0x26, 0xff},
	Border:     color.RGBA{0x3a, 0x3a, 0x3a, 0xff},
	AltRow:     color.RGBA{0x22, 0x22, 0x22, 0xff},
}

var lightTheme = Theme{
	Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
	Text:       color.RGBA{0x33, 0x33, 0x33, 0xff},
	Header:     color.RGBA{0x1a, 0x73, 0xe8, 0xff},
	HeaderBg:   color.RGBA{0xf5, 0xf5, 0xf5, 0xff},
	Border:     color.RGBA{0xdd, 0xdd, 0xdd, 0xff},
	AltRow:     color.RGBA{0xfa, 0xfa, 0xfa, 0xff},
}

const maxScanTokenSize = 1024 * 1024

// emojiReplacements maps common emoji to text equivalents
// since most system fonts don't include emoji glyphs
var emojiReplacements = map[string]string{
	"✅": "✓",
	"❌": "✗",
	"⭕": "○",
	"❎": "✗",
	"☑️": "✓",
	"✔️": "✓",
	"✖️": "✗",
	"⚠️": "⚠",
	"🔴": "●",
	"🟢": "●",
	"🟡": "●",
	"⬜": "□",
	"⬛": "■",
	"🔲": "□",
	"🔳": "□",
}

func replaceEmoji(input string) string {
	result := input
	for emoji, replacement := range emojiReplacements {
		result = strings.ReplaceAll(result, emoji, replacement)
	}
	return result
}

func splitRow(line string) []string {
	parts := strings.Split(line, "|")
	if len(parts) == 0 {
		return nil
	}
	if strings.TrimSpace(parts[0]) == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return nil
	}
	cells := make([]string, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	return cells
}

func isSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		if len(cell) < 3 {
			return false
		}
		for _, r := range cell {
			if r != '-' {
				return false
			}
		}
	}
	return true
}

func normalizeRows(rows [][]string) {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	for i, row := range rows {
		if len(row) < maxCols {
			padded := make([]string, maxCols)
			copy(padded, row)
			rows[i] = padded
		}
	}
}

func parseTable(input string) ([][]string, error) {
	var rows [][]string
	scanner := bufio.NewScanner(strings.NewReader(input))
	// Allow large rows beyond the default 64K scanner limit.
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.Contains(line, "|") {
			continue
		}

		cells := splitRow(line)
		if len(cells) == 0 || isSeparatorRow(cells) {
			continue
		}
		rows = append(rows, cells)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no table data found")
	}
	normalizeRows(rows)
	return rows, scanner.Err()
}

func loadFont(dc *gg.Context, size float64) error {
	// Use embedded Inter font for cross-platform support
	font, err := truetype.Parse(interFont)
	if err != nil {
		return fmt.Errorf("failed to parse embedded font: %w", err)
	}
	face := truetype.NewFace(font, &truetype.Options{Size: size})
	dc.SetFontFace(face)
	return nil
}

func measureTable(dc *gg.Context, rows [][]string, padding float64) ([]float64, float64) {
	colWidths := make([]float64, len(rows[0]))
	_, fontHeight := dc.MeasureString("Mg")
	rowHeight := fontHeight + padding*2
	
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(colWidths) {
				continue
			}
			w, _ := dc.MeasureString(cell)
			if w+padding*2 > colWidths[i] {
				colWidths[i] = w + padding*2
			}
		}
	}
	return colWidths, rowHeight
}

func renderTable(rows [][]string, theme Theme, fontSize, padding float64) (*gg.Context, error) {
	// Create temp context for measuring
	tmpDc := gg.NewContext(1, 1)
	if err := loadFont(tmpDc, fontSize); err != nil {
		return nil, err
	}
	
	colWidths, rowHeight := measureTable(tmpDc, rows, padding)
	
	// Calculate total size
	totalWidth := padding * 2
	for _, w := range colWidths {
		totalWidth += w
	}
	totalHeight := padding*2 + float64(len(rows))*rowHeight
	
	// Create actual context
	dc := gg.NewContext(int(totalWidth), int(totalHeight))
	if err := loadFont(dc, fontSize); err != nil {
		return nil, err
	}
	
	// Background
	dc.SetColor(theme.Background)
	dc.Clear()
	
	// Draw table
	y := padding
	for rowIdx, row := range rows {
		x := padding
		isHeader := rowIdx == 0
		isAltRow := rowIdx%2 == 0 && rowIdx > 0
		
		// Row background
		if isHeader {
			dc.SetColor(theme.HeaderBg)
			dc.DrawRectangle(padding, y, totalWidth-padding*2, rowHeight)
			dc.Fill()
		} else if isAltRow {
			dc.SetColor(theme.AltRow)
			dc.DrawRectangle(padding, y, totalWidth-padding*2, rowHeight)
			dc.Fill()
		}
		
		// Draw cells
		for i, cell := range row {
			if i >= len(colWidths) {
				continue
			}
			
			// Cell border
			dc.SetColor(theme.Border)
			dc.SetLineWidth(1)
			dc.DrawRectangle(x, y, colWidths[i], rowHeight)
			dc.Stroke()
			
			// Cell text
			if isHeader {
				dc.SetColor(theme.Header)
			} else {
				dc.SetColor(theme.Text)
			}
			_, fh := dc.MeasureString("Mg")
			dc.DrawString(cell, x+padding, y+fh+padding/2)
			
			x += colWidths[i]
		}
		y += rowHeight
	}
	
	return dc, nil
}

func main() {
	inputFile := flag.String("i", "", "Input file (default: stdin)")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	themeName := flag.String("theme", "dark", "Theme: dark or light")
	fontSize := flag.Float64("font-size", 14, "Font size")
	padding := flag.Float64("padding", 10, "Cell padding")
	flag.Parse()
	
	// Read input
	var input string
	if *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		input = string(data)
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		input = string(data)
	}
	
	// Replace emoji with text equivalents
	input = replaceEmoji(input)

	// Parse table
	rows, err := parseTable(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing table: %v\n", err)
		os.Exit(1)
	}
	
	// Select theme
	theme := darkTheme
	if *themeName == "light" {
		theme = lightTheme
	}
	
	// Render
	dc, err := renderTable(rows, theme, *fontSize, *padding)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}
	
	// Output
	if *outputFile != "" {
		if err := dc.SavePNG(*outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving PNG: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := png.Encode(os.Stdout, dc.Image()); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding PNG: %v\n", err)
			os.Exit(1)
		}
	}
}
