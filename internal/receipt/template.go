package receipt

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
	"unicode/utf8"
)

const (
	PaperWidthDots = 384
	TextColumns    = 32
)

type Template struct {
	Name   string  `json:"name"`
	Blocks []Block `json:"blocks"`
}

type Block struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Align     string `json:"align,omitempty"`
	Bold      bool   `json:"bold,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Character string `json:"character,omitempty"`
	Label     string `json:"label,omitempty"`
	Value     string `json:"value,omitempty"`
	Image     string `json:"image,omitempty"`
	QRData    string `json:"qrData,omitempty"`
	QRSize    int    `json:"qrSize,omitempty"`
	QRLevel   int    `json:"qrLevel,omitempty"`
	Lines     int    `json:"lines,omitempty"`
	Items     []Item `json:"items,omitempty"`
}

type Item struct {
	Name  string `json:"name"`
	Qty   string `json:"qty,omitempty"`
	Price string `json:"price"`
}

func Build(template Template) ([]byte, error) {
	if len(template.Blocks) == 0 {
		return nil, errors.New("template tidak memiliki blok")
	}

	var output bytes.Buffer
	output.Write([]byte{0x1B, 0x40})

	for index, block := range template.Blocks {
		if err := writeBlock(&output, block); err != nil {
			return nil, fmt.Errorf("blok %d (%s): %w", index+1, block.Type, err)
		}
	}

	return output.Bytes(), nil
}

func writeBlock(output *bytes.Buffer, block Block) error {
	switch block.Type {
	case "text":
		writeText(output, block)
	case "separator":
		character := firstASCII(block.Character, '-')
		output.Write(bytes.Repeat([]byte{character}, TextColumns))
		output.WriteByte('\n')
	case "items":
		for _, item := range block.Items {
			output.WriteString(columnLine(item.Name, item.Price, TextColumns))
			output.WriteByte('\n')
			if strings.TrimSpace(item.Qty) != "" {
				output.WriteString("  ")
				output.WriteString(ascii(item.Qty))
				output.WriteByte('\n')
			}
		}
	case "total":
		setAlign(output, block.Align)
		output.Write([]byte{0x1B, 0x45, 0x01})
		output.WriteString(columnLine(block.Label, block.Value, TextColumns))
		output.WriteByte('\n')
		output.Write([]byte{0x1B, 0x45, 0x00})
	case "image":
		if strings.TrimSpace(block.Image) == "" {
			return errors.New("gambar belum dipilih")
		}
		raster, err := imageRaster(block.Image)
		if err != nil {
			return err
		}
		setAlign(output, normalizedAlign(block.Align))
		output.Write(raster)
		output.WriteByte('\n')
	case "qr":
		if strings.TrimSpace(block.QRData) == "" {
			return errors.New("data QR masih kosong")
		}
		setAlign(output, normalizedAlign(block.Align))
		writeQR(output, []byte(ascii(block.QRData)), block.QRSize, block.QRLevel)
		output.WriteByte('\n')
	case "feed":
		lines := clamp(block.Lines, 1, 20, 3)
		output.Write([]byte{0x1B, 0x64, byte(lines)})
	default:
		return fmt.Errorf("tipe blok %q tidak dikenal", block.Type)
	}

	return nil
}

func writeText(output *bytes.Buffer, block Block) {
	setAlign(output, normalizedAlign(block.Align))
	if block.Bold {
		output.Write([]byte{0x1B, 0x45, 0x01})
	}

	width := clamp(block.Width, 1, 8, 1)
	height := clamp(block.Height, 1, 8, 1)
	size := byte(((width - 1) << 4) | (height - 1))
	output.Write([]byte{0x1D, 0x21, size})
	output.WriteString(ascii(block.Text))
	if !strings.HasSuffix(block.Text, "\n") {
		output.WriteByte('\n')
	}
	output.Write([]byte{0x1D, 0x21, 0x00})
	if block.Bold {
		output.Write([]byte{0x1B, 0x45, 0x00})
	}
}

func setAlign(output *bytes.Buffer, align string) {
	value := byte(0)
	switch align {
	case "center":
		value = 1
	case "right":
		value = 2
	}
	output.Write([]byte{0x1B, 0x61, value})
}

func normalizedAlign(align string) string {
	switch align {
	case "center", "right":
		return align
	default:
		return "left"
	}
}

func writeQR(output *bytes.Buffer, data []byte, moduleSize, errorLevel int) {
	moduleSize = clamp(moduleSize, 1, 16, 5)
	if errorLevel < 48 || errorLevel > 51 {
		errorLevel = 49
	}

	output.Write([]byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00})
	output.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, byte(moduleSize)})
	output.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, byte(errorLevel)})

	parameterLength := len(data) + 3
	output.Write([]byte{
		0x1D, 0x28, 0x6B,
		byte(parameterLength & 0xFF),
		byte((parameterLength >> 8) & 0xFF),
		0x31, 0x50, 0x30,
	})
	output.Write(data)
	output.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30})
}

func imageRaster(dataURL string) ([]byte, error) {
	comma := strings.IndexByte(dataURL, ',')
	if comma < 0 {
		return nil, errors.New("format data gambar tidak valid")
	}

	encoded := dataURL[comma+1:]
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca base64 gambar: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("format gambar tidak didukung: %w", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, errors.New("dimensi gambar tidak valid")
	}

	width := bounds.Dx()
	if width > PaperWidthDots {
		width = PaperWidthDots
	}
	height := int(math.Round(float64(bounds.Dy()) * float64(width) / float64(bounds.Dx())))
	if height < 1 {
		height = 1
	}
	if height > 2048 {
		return nil, errors.New("gambar terlalu tinggi; maksimum 2048 dot setelah diskalakan")
	}

	bytesPerRow := (width + 7) / 8
	payload := make([]byte, bytesPerRow*height)
	for y := 0; y < height; y++ {
		sourceY := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < width; x++ {
			sourceX := bounds.Min.X + x*bounds.Dx()/width
			r, g, b, a := img.At(sourceX, sourceY).RGBA()
			if a < 0x8000 {
				continue
			}
			luminance := (299*r + 587*g + 114*b) / 1000
			if luminance < 0x8000 {
				payload[y*bytesPerRow+x/8] |= 0x80 >> (x % 8)
			}
		}
	}

	result := make([]byte, 0, 8+len(payload))
	result = append(result,
		0x1D, 0x76, 0x30, 0x00,
		byte(bytesPerRow&0xFF), byte((bytesPerRow>>8)&0xFF),
		byte(height&0xFF), byte((height>>8)&0xFF),
	)
	result = append(result, payload...)
	return result, nil
}

func columnLine(left, right string, width int) string {
	left = ascii(strings.TrimSpace(left))
	right = ascii(strings.TrimSpace(right))
	if len(right) >= width {
		return truncateASCII(right, width)
	}

	leftLimit := width - len(right) - 1
	left = truncateASCII(left, leftLimit)
	spaces := width - len(left) - len(right)
	return left + strings.Repeat(" ", spaces) + right
}

func ascii(value string) string {
	var output strings.Builder
	for len(value) > 0 {
		r, size := utf8.DecodeRuneInString(value)
		value = value[size:]
		if r == '\n' || r == '\t' || (r >= 0x20 && r <= 0x7E) {
			output.WriteRune(r)
		} else {
			output.WriteByte('?')
		}
	}
	return output.String()
}

func truncateASCII(value string, length int) string {
	if length <= 0 {
		return ""
	}
	if len(value) <= length {
		return value
	}
	return value[:length]
}

func firstASCII(value string, fallback byte) byte {
	value = ascii(value)
	if value == "" || value[0] == '\n' || value[0] == '\t' {
		return fallback
	}
	return value[0]
}

func clamp(value, minimum, maximum, fallback int) int {
	if value == 0 {
		return fallback
	}
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
