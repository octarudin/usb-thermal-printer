package printer

import (
	"fmt"
	"sync"

	"github.com/google/gousb"
)

const (
	DefaultVendorID  = gousb.ID(0x0416)
	DefaultProductID = gousb.ID(0x5011)
)

type Result struct {
	Endpoint     int    `json:"endpoint"`
	TransferType string `json:"transferType"`
	BytesWritten int    `json:"bytesWritten"`
}

var printMu sync.Mutex

func Send(vendorID, productID gousb.ID, data []byte) (Result, error) {
	printMu.Lock()
	defer printMu.Unlock()

	ctx := gousb.NewContext()
	defer ctx.Close()

	device, err := ctx.OpenDeviceWithVIDPID(vendorID, productID)
	if err != nil {
		return Result{}, fmt.Errorf("gagal membuka printer USB %s:%s: %w", vendorID, productID, err)
	}
	if device == nil {
		return Result{}, fmt.Errorf("printer USB %s:%s tidak ditemukan", vendorID, productID)
	}
	defer device.Close()

	intf, release, err := device.DefaultInterface()
	if err != nil {
		return Result{}, fmt.Errorf("gagal membuka interface printer: %w", err)
	}
	defer release()

	for _, descriptor := range intf.Setting.Endpoints {
		if descriptor.Direction != gousb.EndpointDirectionOut || descriptor.TransferType != gousb.TransferTypeBulk {
			continue
		}

		endpoint, openErr := intf.OutEndpoint(descriptor.Number)
		if openErr != nil {
			return Result{}, fmt.Errorf("gagal membuka endpoint OUT %d: %w", descriptor.Number, openErr)
		}

		written, writeErr := endpoint.Write(data)
		if writeErr != nil {
			return Result{}, fmt.Errorf("gagal mengirim data ke endpoint OUT %d: %w", descriptor.Number, writeErr)
		}
		if written != len(data) {
			return Result{}, fmt.Errorf("data tidak terkirim lengkap: %d dari %d byte", written, len(data))
		}

		return Result{
			Endpoint:     descriptor.Number,
			TransferType: descriptor.TransferType.String(),
			BytesWritten: written,
		}, nil
	}

	return Result{}, fmt.Errorf("printer tidak memiliki endpoint USB Bulk OUT")
}
