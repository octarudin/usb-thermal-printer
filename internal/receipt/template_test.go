package receipt

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestBuildTextReceipt(t *testing.T) {
	data, err := Build(Template{Blocks: []Block{
		{Type: "text", Text: "TOKO", Align: "center", Bold: true, Width: 2, Height: 2},
		{Type: "separator", Character: "-"},
		{Type: "total", Label: "TOTAL", Value: "Rp10.000"},
		{Type: "feed", Lines: 3},
	}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !bytes.HasPrefix(data, []byte{0x1B, 0x40}) {
		t.Fatalf("hasil tidak diawali ESC @: % X", data[:2])
	}
	if !bytes.Contains(data, []byte("TOKO\n")) {
		t.Fatal("hasil tidak memuat teks")
	}
	if !bytes.Contains(data, []byte{0x1B, 0x64, 0x03}) {
		t.Fatal("hasil tidak memuat feed tiga baris")
	}
}

func TestBuildQRLength(t *testing.T) {
	data, err := Build(Template{Blocks: []Block{{
		Type: "qr", QRData: "ABC", QRSize: 5, QRLevel: 49,
	}}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	want := []byte{0x1D, 0x28, 0x6B, 0x06, 0x00, 0x31, 0x50, 0x30, 'A', 'B', 'C'}
	if !bytes.Contains(data, want) {
		t.Fatalf("header/payload QR salah; ingin % X", want)
	}
}

func TestBuildRasterImage(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 8, 1))
	for x := 0; x < 4; x++ {
		img.SetGray(x, 0, color.Gray{Y: 0})
	}
	for x := 4; x < 8; x++ {
		img.SetGray(x, 0, color.Gray{Y: 255})
	}

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())

	data, err := Build(Template{Blocks: []Block{{Type: "image", Image: dataURL}}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	want := []byte{0x1D, 0x76, 0x30, 0x00, 0x01, 0x00, 0x01, 0x00, 0xF0}
	if !bytes.Contains(data, want) {
		t.Fatalf("raster salah; ingin % X, dapat % X", want, data)
	}
}

func TestColumnLineHasFixedWidth(t *testing.T) {
	line := columnLine("Produk sangat panjang sekali", "Rp10.000", TextColumns)
	if len(line) != TextColumns {
		t.Fatalf("panjang baris = %d, ingin %d: %q", len(line), TextColumns, line)
	}
	if !strings.HasSuffix(line, "Rp10.000") {
		t.Fatalf("harga tidak rata kanan: %q", line)
	}
}
