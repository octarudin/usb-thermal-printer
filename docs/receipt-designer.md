# Taffware Receipt Designer

Taffware Receipt Designer adalah GUI lokal untuk menyusun dan menguji struk ESC/POS pada printer thermal Taffware POS-58. Backend dibuat dengan Go, sedangkan antarmuka berjalan di browser lokal tanpa mengirim data ke internet.

## Fitur MVP

- Editor berbasis blok dengan drag-and-drop.
- Preview kertas 58 mm dengan area cetak 384 dot.
- Blok teks, separator, daftar barang, total, logo, QR Code, dan paper feed.
- Alignment, bold, serta skala lebar dan tinggi teks.
- Konversi PNG, JPEG, atau GIF menjadi raster hitam-putih ESC/POS.
- Simpan dan buka template JSON.
- Test print langsung melalui USB `0416:5011`.
- Validasi template dan pesan error printer.

## Menjalankan dari source

Pastikan persyaratan Go, libusb, dan driver WinUSB pada README utama sudah terpenuhi.

```bash
go run ./cmd/receipt-designer
```

GUI terbuka otomatis pada:

```text
http://127.0.0.1:8787
```

Gunakan opsi berikut jika browser tidak perlu dibuka otomatis:

```bash
go run ./cmd/receipt-designer -no-browser
```

Alamat lokal juga dapat diganti:

```bash
go run ./cmd/receipt-designer -addr 127.0.0.1:9000
```

## Membuat executable Windows

```bash
go build -o receipt-designer.exe ./cmd/receipt-designer
```

Jalankan:

```powershell
.\receipt-designer.exe
```

Seluruh HTML, CSS, dan JavaScript sudah ditanam ke executable menggunakan `go:embed`, sehingga tidak perlu menyalin folder static saat mendistribusikan aplikasi.

## Alur penggunaan

1. Jalankan aplikasi.
2. Tambahkan komponen melalui panel **Tambah blok**.
3. Pilih blok pada panel **Urutan struk**.
4. Ubah isi melalui panel **Properti blok**.
5. Seret kartu untuk mengatur urutan.
6. Periksa hasil pada **Pratinjau**.
7. Simpan desain dengan **Simpan JSON**.
8. Nyalakan dan hubungkan printer.
9. Klik **Test print**.

## Format template

Template disimpan sebagai JSON agar dapat digunakan kembali atau diproses oleh aplikasi lain.

```json
{
  "name": "Struk Taffware",
  "blocks": [
    {
      "id": "title",
      "type": "text",
      "text": "TOKO CONTOH",
      "align": "center",
      "bold": true,
      "width": 2,
      "height": 2
    },
    {
      "id": "separator",
      "type": "separator",
      "character": "-"
    },
    {
      "id": "total",
      "type": "total",
      "label": "TOTAL",
      "value": "Rp28.000"
    },
    {
      "id": "feed",
      "type": "feed",
      "lines": 3
    }
  ]
}
```

ID blok digunakan oleh GUI. Generator ESC/POS menggunakan urutan elemen pada array `blocks`.

## Endpoint lokal

GUI menggunakan API yang hanya mendengarkan pada loopback `127.0.0.1`.

| Endpoint | Method | Fungsi |
| --- | --- | --- |
| `/api/info` | `GET` | Informasi printer dan lebar kertas |
| `/api/render` | `POST` | Memvalidasi template dan menghitung ukuran ESC/POS |
| `/api/print` | `POST` | Mengubah template lalu mengirimkannya ke printer |

Request `/api/render` dan `/api/print` menggunakan template JSON sebagai body.

## Pengolahan logo

Logo yang dipilih melalui GUI akan:

1. Dibaca sebagai data URL base64.
2. Diperkecil secara proporsional hingga maksimum 384 dot.
3. Diubah menjadi hitam-putih menggunakan ambang luminance.
4. Dikemas sebagai perintah raster `GS v 0`.

Ukuran file pada GUI dibatasi 3 MB dan tinggi hasil raster maksimum 2048 dot. Gunakan logo dengan kontras tinggi agar hasil thermal lebih jelas.

## Batasan versi MVP

- Preview menggunakan font browser, sehingga tidak identik sepenuhnya dengan font internal printer.
- Teks dikonversi ke ASCII; karakter yang tidak tersedia diganti dengan `?`.
- QR pada preview adalah placeholder visual. QR sebenarnya dibuat oleh firmware printer ketika dicetak.
- Belum tersedia variabel dinamis, undo/redo, barcode satu dimensi, atau editor posisi bebas.
- Aplikasi belum membaca status melalui endpoint USB `IN`.
- Tidak ada perintah auto-cut karena printer Taffware POS-58 pada proyek ini menggunakan tear bar manual.

## Struktur source

```text
cmd/receipt-designer/
├── main.go             Server lokal dan API
└── static/
    ├── index.html      Struktur GUI
    ├── app.css         Tampilan dan preview kertas
    └── app.js          State editor dan interaksi

internal/
├── printer/usb.go      Pengiriman USB Bulk OUT
└── receipt/
    ├── template.go     Generator template ke ESC/POS
    └── template_test.go
```

