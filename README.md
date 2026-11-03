# USB Thermal Printer

Program Go sederhana untuk mengirim data cetak langsung ke printer thermal POS melalui USB tanpa melalui print spooler sistem operasi. Program membuka perangkat USB berdasarkan Vendor ID (VID) dan Product ID (PID), mencari endpoint USB `OUT`, lalu mengirim teks serta perintah potong kertas menggunakan protokol ESC/POS.

Pelajari struktur byte, format perintah, QR Code, barcode, gambar raster, dan contoh buffer untuk ESP32-S3 melalui [dokumentasi dasar ESC/POS](docs/escpos.md).

<p align="center">
  <img src="assets/taffware-usb-thermal-printer.jpg" alt="Printer thermal Taffware POS USB 58 mm" width="500">
</p>

## Spesifikasi perangkat

Konfigurasi bawaan program dibuat untuk perangkat berikut:

| Komponen | Spesifikasi |
| --- | --- |
| Merek | Taffware |
| Jenis printer | Printer thermal POS-58 yang kompatibel dengan ESC/POS |
| Antarmuka | USB langsung melalui libusb |
| Kecepatan cetak | Hingga 90 mm/detik |
| Vendor ID | `0x0416` (Winbond Electronics Corp.) |
| Product ID | `0x5011` |
| Identitas perangkat | `POS58 Printer USB` atau printer POS-58 generik |
| Perintah keluaran | Inisialisasi printer, cetak teks, dan potong kertas |
| Kertas | Kertas thermal POS lebar 58 mm |

> **Catatan:** VID/PID `0416:5011` juga dapat digunakan oleh printer POS generik lain. Pastikan identitas USB perangkat sesuai sebelum menjalankan program. Perintah potong hanya bekerja jika printer dilengkapi auto-cutter.

## Persyaratan

### Perangkat keras

- Printer thermal Taffware POS-58 USB yang mendukung perintah ESC/POS.
- Kertas thermal POS lebar 58 mm.
- Kabel USB dan adaptor daya printer.

### Perangkat lunak

- [Go](https://go.dev/dl/) 1.25.4 atau versi lebih baru.
- [libusb 1.0](https://libusb.info/), karena paket `gousb` berkomunikasi langsung dengan perangkat USB.
- Compiler C dan `pkg-config` untuk membangun paket berbasis CGO.
- Dependency Go berikut, yang akan diunduh melalui `go mod download`:
  - [`github.com/google/gousb`](https://github.com/google/gousb)
  - [`github.com/kenshaw/escpos`](https://github.com/kenshaw/escpos)

#### Windows

Siapkan toolchain GCC, `pkg-config`, dan libusb 1.0, misalnya melalui MSYS2/MinGW. Printer juga harus memakai driver USB yang dapat diakses libusb, seperti WinUSB. Driver dapat dipilih dengan [Zadig](https://zadig.akeo.ie/).

Mengganti driver printer ke WinUSB dapat membuat printer tidak tersedia untuk aplikasi yang memakai driver/spooler Windows biasa. Kembalikan driver sebelumnya jika printer perlu digunakan dengan cara tersebut.

#### Linux

Pasang compiler, `pkg-config`, dan paket pengembangan libusb. Contoh pada Debian/Ubuntu:

```bash
sudo apt update
sudo apt install build-essential pkg-config libusb-1.0-0-dev
```

Pengguna juga harus memiliki izin untuk mengakses perangkat USB. Jika printer sedang dikuasai oleh driver kernel `usblp`, lepaskan driver tersebut atau sesuaikan konfigurasi sistem sebelum menjalankan program.

## Konfigurasi printer

VID dan PID berada di bagian atas [`main.go`](main.go):

```go
const (
	vendorID  = gousb.ID(0x0416)
	productID = gousb.ID(0x5011)
)
```

Untuk printer lain, ganti kedua nilai tersebut sesuai identitas perangkat:

- Windows: periksa melalui Device Manager, USBView, atau Zadig.
- Linux: jalankan `lsusb`.

Contoh identitas yang dicari:

```text
0416:5011
```

## Menjalankan program

1. Nyalakan printer, pasang kertas, lalu hubungkan kabel USB.
2. Unduh dependency:

   ```bash
   go mod download
   ```

3. Jalankan program:

   ```bash
   go run .
   ```

Program juga dapat dikompilasi terlebih dahulu:

```bash
go build -o usb-thermal-printer .
```

## Hasil program

Jika perangkat berhasil ditemukan, terminal menampilkan endpoint USB yang digunakan dan konfirmasi bahwa data telah dikirim. Nomor serta jenis endpoint dapat berbeda pada setiap printer.

```text
2026/11/03 10:00:00 menggunakan endpoint OUT 1 (bulk)
Data cetak berhasil dikirim ke printer.
```

Printer akan mencetak:

```text
Cetak Direct USB dari Go Berhasil!
```

Setelah teks dicetak, program mengirim perintah potong kertas. Pada printer tanpa auto-cutter, kertas perlu disobek secara manual.

## Masalah umum

- **`printer USB 0416:5011 tidak ditemukan`** — pastikan printer menyala, kabel terpasang, dan VID/PID pada kode sesuai dengan perangkat.
- **Gagal membuka perangkat atau interface** — periksa driver USB dan hak akses pengguna.
- **Program berjalan tetapi tidak mencetak** — pastikan perangkat mendukung ESC/POS dan endpoint `OUT` yang dipilih dapat menerima data cetak.
- **Perintah potong tidak bekerja** — printer mungkin tidak memiliki auto-cutter atau menggunakan variasi perintah yang berbeda.

