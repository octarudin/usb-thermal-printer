package main

import (
	"fmt"
	"io"
	"log"

	"github.com/google/gousb"
	"github.com/kenshaw/escpos"
)

const (
	vendorID  = gousb.ID(0x0416) // Lihat menggunakan Zadig atau lsusb.
	productID = gousb.ID(0x5011) // Lihat menggunakan Zadig atau lsusb.
)

// usbWriter membuat endpoint USB OUT memenuhi io.ReadWriter yang diminta
// package escpos. Printer hanya membutuhkan Write untuk proses cetak.
type usbWriter struct {
	out *gousb.OutEndpoint
}

func (w *usbWriter) Write(data []byte) (int, error) {
	return w.out.Write(data)
}

func (w *usbWriter) Read([]byte) (int, error) {
	return 0, io.EOF
}

func main() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	device, err := ctx.OpenDeviceWithVIDPID(vendorID, productID)
	if err != nil {
		log.Fatalf("gagal membuka printer USB %s:%s: %v", vendorID, productID, err)
	}
	if device == nil {
		log.Fatalf("printer USB %s:%s tidak ditemukan", vendorID, productID)
	}
	defer device.Close()

	intf, release, err := device.DefaultInterface()
	if err != nil {
		log.Fatalf("gagal membuka interface printer: %v", err)
	}
	defer release()

	var endpoint *gousb.OutEndpoint
	for _, descriptor := range intf.Setting.Endpoints {
		if descriptor.Direction != gousb.EndpointDirectionOut {
			continue
		}

		endpoint, err = intf.OutEndpoint(descriptor.Number)
		if err != nil {
			log.Fatalf("gagal membuka endpoint OUT %d: %v", descriptor.Number, err)
		}
		log.Printf("menggunakan endpoint OUT %d (%s)", descriptor.Number, descriptor.TransferType)
		break
	}
	if endpoint == nil {
		log.Fatal("printer tidak memiliki endpoint USB OUT")
	}

	p := escpos.New(&usbWriter{out: endpoint})
	p.Init()
	if _, err := p.Write("Cetak Direct USB dari Go Berhasil!\n\n\n"); err != nil {
		log.Fatalf("gagal mengirim data cetak: %v", err)
	}
	p.Cut()

	fmt.Println("Data cetak berhasil dikirim ke printer.")
}
