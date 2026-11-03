# Dasar Format Data ESC/POS untuk ESP32-S3

Dokumen ini menjelaskan cara menyusun data ESC/POS yang nantinya dikirim oleh ESP32-S3 ke endpoint USB Bulk `OUT` printer Taffware POS-58. Fokusnya adalah format payload printer, bukan implementasi enumerasi atau driver USB Host.

> **Penting:** ESC/POS berasal dari Epson, tetapi printer POS generik biasanya hanya mengimplementasikan sebagian perintah atau memiliki variasi firmware. Uji perintah secara bertahap pada printer yang digunakan.

## 1. Dua lapisan yang berbeda

Komunikasi dengan printer USB terdiri dari dua lapisan:

```text
ESP32-S3 USB Host
        │
        │ USB control transfer: enumerasi dan claim interface
        │ USB bulk transfer: membawa byte ESC/POS
        ▼
Printer POS-58
```

- **USB Host** bertugas menemukan perangkat `0416:5011`, membaca descriptor, memilih interface, dan membuka endpoint Bulk `OUT`.
- **ESC/POS** adalah isi byte yang dikirim melalui endpoint tersebut.

ESC/POS tidak mempunyai header untuk satu struk penuh. Printer membaca aliran byte secara berurutan. Sebagian byte dianggap sebagai teks, sedangkan byte kontrol seperti `ESC` dan `GS` mengawali perintah.

## 2. Cara membaca notasi byte

Byte kontrol yang sering digunakan:

| Nama | Hex | Desimal | Arti umum |
| --- | ---: | ---: | --- |
| `NUL` | `00` | 0 | Null/parameter tertentu |
| `HT` | `09` | 9 | Tab horizontal |
| `LF` | `0A` | 10 | Cetak buffer baris dan pindah baris |
| `CR` | `0D` | 13 | Carriage return; perilakunya dapat bergantung firmware |
| `DLE` | `10` | 16 | Awalan perintah real-time |
| `ESC` | `1B` | 27 | Awalan kelompok perintah ESC |
| `FS` | `1C` | 28 | Awalan kelompok perintah FS |
| `GS` | `1D` | 29 | Awalan kelompok perintah GS |

Contoh notasi:

```text
ASCII     ESC @
Hex       1B 40
Desimal   27 64
```

Tulisan `ESC @` bukan string lima karakter `"ESC @"`. Data sebenarnya hanya dua byte: `0x1B` dan `0x40`.

Dalam C/C++:

```c
const uint8_t initialize[] = {0x1B, 0x40};
printer_write(initialize, sizeof(initialize));
```

`printer_write()` pada contoh-contoh di dokumen ini adalah abstraksi untuk mengirim buffer ke endpoint Bulk `OUT`. Implementasinya akan dibuat di lapisan USB Host.

## 3. Bentuk umum perintah

### 3.1 Perintah tetap

Tidak mempunyai parameter:

```text
ESC @
1B 40
```

### 3.2 Perintah dengan parameter

Parameter berada setelah kode perintah:

```text
ESC a n
1B 61 n
```

Nilai `n` adalah satu byte, bukan karakter angka kecuali dokumentasi perintah menyatakan demikian.

```c
const uint8_t align_center[] = {0x1B, 0x61, 0x01};
```

### 3.3 Perintah dengan panjang data

Perintah gambar dan QR membawa ukuran payload:

```text
prefix pL pH parameter... data...
```

Nilai panjang umumnya disimpan little-endian:

```text
panjang = pL + (pH × 256)
pL      = panjang & 0xFF
pH      = (panjang >> 8) & 0xFF
```

## 4. Inisialisasi printer

Gunakan `ESC @` pada awal pekerjaan cetak:

| Format | Data |
| --- | --- |
| ASCII | `ESC @` |
| Hex | `1B 40` |
| C | `{0x1B, 0x40}` |

Perintah ini membersihkan print buffer dan mengembalikan mode cetak ke keadaan awal printer. Data permanen seperti logo pada NV memory tidak ikut dihapus.

```c
const uint8_t cmd_init[] = {0x1B, 0x40};
printer_write(cmd_init, sizeof(cmd_init));
```

Jangan menjalankan `ESC @` di tengah struk kecuali memang ingin membatalkan pengaturan aktif seperti alignment, bold, ukuran teks, dan line spacing.

## 5. Mencetak teks dan pindah baris

Karakter biasa dikirim langsung setelah inisialisasi:

```c
static const uint8_t text[] = "TOKO CONTOH\nJakarta\n";
printer_write(text, sizeof(text) - 1);
```

Byte yang terkirim:

```text
54 4F 4B 4F 20 43 4F 4E 54 4F 48 0A
4A 61 6B 61 72 74 61 0A
```

`LF` (`0x0A`) mencetak data yang masih berada pada buffer baris lalu memindahkan posisi ke baris berikutnya.

### Encoding karakter

Printer thermal generik biasanya tidak menafsirkan teks sebagai UTF-8 secara otomatis. Karakter ASCII dasar aman digunakan, tetapi simbol atau huruf non-ASCII bergantung pada code page printer.

Pemilihan code page menggunakan:

```text
ESC t n
1B 74 n
```

Nilai `n` berbeda antarfirmware. Cetak self-test printer untuk melihat tabel code page yang tersedia. Jangan mengasumsikan satu nilai code page berlaku pada semua printer.

## 6. Alignment

Format:

```text
ESC a n
1B 61 n
```

| `n` | Posisi |
| ---: | --- |
| `0` | Kiri |
| `1` | Tengah |
| `2` | Kanan |

```c
const uint8_t left[]   = {0x1B, 0x61, 0x00};
const uint8_t center[] = {0x1B, 0x61, 0x01};
const uint8_t right[]  = {0x1B, 0x61, 0x02};
```

Kirim perintah alignment saat printer berada di awal baris. Alignment tetap aktif sampai diubah lagi atau printer diinisialisasi.

## 7. Bold atau emphasized

Format:

```text
ESC E n
1B 45 n
```

| `n` | Mode |
| ---: | --- |
| `0` | Nonaktif |
| `1` | Aktif |

```c
const uint8_t bold_on[]  = {0x1B, 0x45, 0x01};
const uint8_t bold_off[] = {0x1B, 0x45, 0x00};
```

## 8. Underline

Format:

```text
ESC - n
1B 2D n
```

| `n` | Mode |
| ---: | --- |
| `0` | Nonaktif |
| `1` | Garis bawah 1 dot |
| `2` | Garis bawah 2 dot |

```c
const uint8_t underline_on[]  = {0x1B, 0x2D, 0x01};
const uint8_t underline_off[] = {0x1B, 0x2D, 0x00};
```

## 9. Ukuran teks

Format:

```text
GS ! n
1D 21 n
```

Bit 4–6 mengatur pengali lebar dan bit 0–2 mengatur pengali tinggi:

```text
n = ((pengali_lebar - 1) << 4) | (pengali_tinggi - 1)
```

Contoh umum:

| Ukuran | `n` hex | Data lengkap |
| --- | ---: | --- |
| Lebar 1×, tinggi 1× | `00` | `1D 21 00` |
| Lebar 2×, tinggi 1× | `10` | `1D 21 10` |
| Lebar 1×, tinggi 2× | `01` | `1D 21 01` |
| Lebar 2×, tinggi 2× | `11` | `1D 21 11` |
| Lebar 3×, tinggi 3× | `22` | `1D 21 22` |

```c
const uint8_t size_2x[]     = {0x1D, 0x21, 0x11};
const uint8_t size_normal[] = {0x1D, 0x21, 0x00};
```

Walaupun spesifikasi ESC/POS mendefinisikan pengali sampai 8×, printer generik mungkin mendukung rentang yang lebih kecil.

## 10. Pemilihan font

Format:

```text
ESC M n
1B 4D n
```

| `n` | Font umum |
| ---: | --- |
| `0` | Font A, lebih besar |
| `1` | Font B, lebih kecil |

Pada printer 58 mm dengan area cetak 384 dot, Font A biasanya menghasilkan sekitar 32 karakter per baris. Nilai sebenarnya harus dikonfirmasi melalui self-test printer.

## 11. Line spacing dan paper feed

### Kembalikan line spacing default

```text
ESC 2
1B 32
```

### Atur line spacing

```text
ESC 3 n
1B 33 n
```

`n` menentukan jarak dalam unit vertikal printer.

### Cetak dan feed sejumlah baris

```text
ESC d n
1B 64 n
```

Contoh feed tiga baris:

```c
const uint8_t feed_3_lines[] = {0x1B, 0x64, 0x03};
```

### Cetak dan feed sejumlah dot/unit

```text
ESC J n
1B 4A n
```

Untuk penggunaan umum, `ESC d n` lebih mudah karena mengikuti line spacing yang aktif.

## 12. Reverse atau putih-di-atas-hitam

Format yang umum:

```text
GS B n
1D 42 n
```

| `n` | Mode |
| ---: | --- |
| `0` | Nonaktif |
| `1` | Aktif |

Dukungan perintah ini perlu diuji pada printer Taffware.

## 13. Barcode satu dimensi

Pengaturan yang umum digunakan:

```text
GS H n    Posisi teks HRI
1D 48 n

GS h n    Tinggi barcode
1D 68 n

GS w n    Lebar modul barcode
1D 77 n
```

Posisi HRI:

| `n` | Posisi teks barcode |
| ---: | --- |
| `0` | Tidak dicetak |
| `1` | Di atas barcode |
| `2` | Di bawah barcode |
| `3` | Di atas dan bawah |

Format modern `GS k` menyertakan panjang data:

```text
GS k m n d1...dn
1D 6B m n data...
```

Contoh CODE39 berisi `ABC123`:

```c
const uint8_t barcode_setup[] = {
    0x1D, 0x48, 0x02, // HRI di bawah
    0x1D, 0x68, 0x50, // tinggi 80 dot
    0x1D, 0x77, 0x02  // lebar modul 2
};

const uint8_t barcode[] = {
    0x1D, 0x6B, 0x45, 0x06, // CODE39, panjang 6
    'A', 'B', 'C', '1', '2', '3'
};
```

Nilai tipe barcode dan batas parameter dapat berbeda pada firmware printer generik. Mulai dengan CODE39 karena struktur datanya sederhana.

## 14. QR Code

QR Code ESC/POS dibentuk melalui beberapa perintah `GS ( k`. Urutannya adalah:

1. Pilih QR model.
2. Atur ukuran module.
3. Atur error correction.
4. Simpan payload.
5. Cetak payload yang tersimpan.

### 14.1 Pilih QR Model 2

```text
1D 28 6B 04 00 31 41 32 00
```

`0x32` berarti ASCII `2`, yaitu QR Model 2.

### 14.2 Atur ukuran module

```text
1D 28 6B 03 00 31 43 n
```

Nilai `n` adalah ukuran satu kotak/module QR dalam dot. Nilai `4` sampai `6` biasanya cocok untuk kertas 58 mm, tergantung panjang data.

### 14.3 Atur error correction

```text
1D 28 6B 03 00 31 45 n
```

| Level | `n` hex | Perkiraan pemulihan |
| --- | ---: | ---: |
| L | `30` | 7% |
| M | `31` | 15% |
| Q | `32` | 25% |
| H | `33` | 30% |

### 14.4 Simpan data QR

```text
1D 28 6B pL pH 31 50 30 data...
```

Panjang yang ditulis pada `pL` dan `pH` adalah:

```text
p = jumlah_byte_data + 3
pL = p & 0xFF
pH = (p >> 8) & 0xFF
```

Tiga byte tambahan berasal dari `31 50 30`.

### 14.5 Cetak QR

```text
1D 28 6B 03 00 31 51 30
```

### Contoh fungsi QR untuk ESP32-S3

```c
void printer_qr(const uint8_t *data, size_t data_len)
{
    const uint8_t model[] = {
        0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00
    };
    const uint8_t module_size[] = {
        0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, 0x05
    };
    const uint8_t error_level[] = {
        0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, 0x31
    };
    const uint8_t print[] = {
        0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30
    };

    uint16_t parameter_len = (uint16_t)data_len + 3;
    uint8_t store_header[] = {
        0x1D, 0x28, 0x6B,
        parameter_len & 0xFF,
        (parameter_len >> 8) & 0xFF,
        0x31, 0x50, 0x30
    };

    printer_write(model, sizeof(model));
    printer_write(module_size, sizeof(module_size));
    printer_write(error_level, sizeof(error_level));
    printer_write(store_header, sizeof(store_header));
    printer_write(data, data_len);
    printer_write(print, sizeof(print));
}
```

Jika `printer_write()` bersifat asynchronous, tunggu callback transfer selesai sebelum menggunakan kembali atau mengubah buffer. Pastikan urutan transfer tetap terjaga.

## 15. Gambar raster

Perintah yang banyak didukung printer POS generik adalah `GS v 0`:

```text
GS v 0 m xL xH yL yH d1...dk
1D 76 30 m xL xH yL yH data...
```

Parameter:

```text
lebar_byte = xL + (xH × 256)
tinggi_dot = yL + (yH × 256)
jumlah_data = lebar_byte × tinggi_dot
```

| `m` | Skala horizontal | Skala vertikal |
| ---: | ---: | ---: |
| `0` | 1× | 1× |
| `1` | 2× | 1× |
| `2` | 1× | 2× |
| `3` | 2× | 2× |

Setiap bit mewakili satu titik:

- Bit `1`: titik dicetak hitam.
- Bit `0`: titik tidak dicetak.
- Bit paling signifikan mewakili piksel paling kiri dalam satu byte.

Untuk area cetak 384 dot:

```text
384 / 8 = 48 byte per baris
```

Gambar tinggi 100 dot memerlukan:

```text
48 × 100 = 4.800 byte payload
```

Contoh header gambar 384 × 100 dot, skala normal:

```text
1D 76 30 00 30 00 64 00
```

Setelah header tersebut harus dikirim tepat 4.800 byte data raster.

Walaupun Epson menandai `GS v 0` sebagai perintah lama, perintah ini sering lebih kompatibel dengan printer POS generik. Implementasi baru dapat mencoba fungsi graphics `GS ( L`, tetapi dukungannya tetap harus diuji pada printer Taffware.

## 16. Paper cut

Format umum:

```text
GS V m
1D 56 m
```

| `m` | Aksi umum |
| ---: | --- |
| `0` atau `48` | Full cut |
| `1` atau `49` | Partial cut |

Varian dengan feed:

```text
GS V 66 n
1D 56 42 n
```

Printer Taffware POS-58 pada proyek ini tampak menggunakan tear bar manual. Jika tidak memiliki mekanisme auto-cutter, jangan mengandalkan perintah ini; printer dapat mengabaikannya.

## 17. Membaca status printer

Permintaan status real-time yang umum:

```text
DLE EOT n
10 04 n
```

| `n` | Kelompok status umum |
| ---: | --- |
| `1` | Status printer |
| `2` | Penyebab offline |
| `3` | Penyebab error |
| `4` | Status sensor kertas |

Printer mengembalikan byte status melalui endpoint USB Bulk `IN`. Karena program Go awal hanya mencari endpoint `OUT`, implementasi ESP32-S3 harus membuka endpoint `IN` juga jika ingin membaca status.

Makna bit respons dapat berbeda antarprinter kompatibel. Gunakan dokumentasi atau eksperimen perangkat sebelum menjadikan status sebagai dasar keputusan penting.

## 18. Perintah periferal opsional

### Membuka cash drawer

```text
ESC p m t1 t2
1B 70 m t1 t2
```

Perintah ini menghasilkan pulsa pada konektor cash drawer. Jangan menggunakannya jika hardware printer tidak menyediakan konektor dan spesifikasi kelistrikan yang sesuai.

### Buzzer

Perintah buzzer sangat bergantung model. Jangan mengasumsikan satu format buzzer dapat digunakan pada seluruh printer ESC/POS generik.

## 19. Contoh satu struk sederhana

```c
void print_test_receipt(void)
{
    const uint8_t init[]       = {0x1B, 0x40};
    const uint8_t center[]     = {0x1B, 0x61, 0x01};
    const uint8_t left[]       = {0x1B, 0x61, 0x00};
    const uint8_t bold_on[]    = {0x1B, 0x45, 0x01};
    const uint8_t bold_off[]   = {0x1B, 0x45, 0x00};
    const uint8_t double_xy[]  = {0x1D, 0x21, 0x11};
    const uint8_t normal[]     = {0x1D, 0x21, 0x00};
    const uint8_t feed[]       = {0x1B, 0x64, 0x03};

    static const uint8_t title[] = "TOKO CONTOH\n";
    static const uint8_t body[] =
        "Item A          Rp10.000\n"
        "Item B          Rp15.000\n"
        "------------------------\n"
        "TOTAL           Rp25.000\n";
    static const uint8_t footer[] = "Terima kasih\n";

    printer_write(init, sizeof(init));
    printer_write(center, sizeof(center));
    printer_write(bold_on, sizeof(bold_on));
    printer_write(double_xy, sizeof(double_xy));
    printer_write(title, sizeof(title) - 1);

    printer_write(normal, sizeof(normal));
    printer_write(bold_off, sizeof(bold_off));
    printer_write(left, sizeof(left));
    printer_write(body, sizeof(body) - 1);

    printer_write(center, sizeof(center));
    printer_write(footer, sizeof(footer) - 1);
    printer_write(feed, sizeof(feed));
}
```

Aliran byte konseptualnya:

```text
[initialize]
[center][bold on][size 2×2][title][LF]
[normal][bold off][left][detail items]
[center][footer][LF]
[feed 3 lines]
```

## 20. Aturan pengiriman melalui USB Host

1. Claim interface printer sebelum mengirim data.
2. Gunakan endpoint dengan arah `OUT`, biasanya bertipe Bulk.
3. Kirim byte mentah; jangan mengubahnya menjadi representasi teks hex seperti string `"1B 40"`.
4. Jaga urutan transfer.
5. Jangan mengubah atau membebaskan buffer sebelum callback transfer asynchronous selesai.
6. Stack USB akan membagi buffer besar menjadi paket USB sesuai ukuran endpoint. Aplikasi tidak perlu menambahkan header USB sendiri.
7. Untuk gambar besar, kirim header lalu seluruh payload sesuai panjang yang dinyatakan. Kekurangan satu byte dapat membuat data berikutnya dianggap bagian gambar.
8. Jangan menambahkan byte padding ke payload ESC/POS kecuali memang merupakan bagian data perintah.

## 21. Urutan pengujian yang disarankan

Uji fitur satu per satu:

1. Enumerasi USB dan cocokkan VID/PID.
2. Kirim `ESC @`.
3. Cetak `TEST\n`.
4. Uji alignment.
5. Uji bold dan ukuran teks.
6. Uji feed kertas.
7. Uji QR Code dengan payload pendek.
8. Uji gambar kecil hitam-putih.
9. Uji endpoint `IN` dan pembacaan status, jika diperlukan.
10. Uji perintah tambahan hanya jika hardware mendukungnya.

Catat perintah yang berhasil karena label “ESC/POS compatible” tidak menjamin seluruh command set Epson tersedia.

## 22. Referensi utama

- [Epson ESC/POS commands in code order](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/commands.html)
- [Epson: Initialize printer (`ESC @`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/esc_atsign.html)
- [Epson: Character size (`GS !`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/gs_exclamation.html)
- [Epson: Emphasized mode (`ESC E`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/esc_ce.html)
- [Epson: QR Code (`GS ( k`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/gs_lparen_lk.html)
- [Epson: Raster bit image (`GS v 0`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/gs_lv_0.html)
- [Epson: Paper cut (`GS V`)](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/gs_cv.html)

