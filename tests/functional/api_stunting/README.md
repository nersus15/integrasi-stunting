# Uji end-to-end API stunting

Menjalankan skenario end-to-end lewat HTTP, memeriksa
`httpCode`, `errorCode`, `errorName`, dan pesan, lalu menulis laporan HTML
berisi payload dan response penuh setiap skenario.

```bash
go test ./functional/api_stunting/
```

Aplikasi dijalankan in-process (port 2098), jadi cukup Postgres dari `.test.env`.
Bila database tidak terjangkau, test **dilewati (skip)**, bukan gagal.
Untuk menguji service yang sudah berjalan, isi `API_BASE` saat menjalankannya:

```bash
API_BASE=http://localhost:2025/api ./uji.sh api_stunting
```

Test ini **mengosongkan tabel** `orangtua`, `anak`, `kunjungan`, dan `kesehatan`
sebelum mulai, termasuk saat menembak service luar. Jangan arahkan ke database
yang berisi data sungguhan.

## Keluaran

`laporan.html` di direktori ini — dibuka langsung di browser. Berisi setiap
skenario yang bisa dibuka untuk melihat request dan response utuh, penyaring
per `httpCode` dan `errorCode`, serta keadaan akhir database.

Berkas itu hasil regenerasi tiap kali test dijalankan; tidak perlu masuk
version control.

## Pengaturan

| env | bawaan |
|---|---|
| `API_BASE` | kosong: aplikasi in-process |
| `API_HEALTH` | diturunkan dari `API_BASE` |
| `PG_DSN` | disusun dari `PG_*` di `.test.env` |
| `LAPORAN_HTML` | `laporan.html` |

## Susunan berkas

| berkas | isi |
|---|---|
| `skenario_test.go` | daftar skenario dan test utama |
| `harness_test.go` | klien HTTP, pembanding, akses database |
| `laporan_test.go` | penyusun HTML |
| `kerangka.html` | template, disisipkan lewat `go:embed` |

## Menambah skenario

Tambahkan entri di `daftarSkenario()`. Field pemeriksaan yang dibiarkan nol
tidak diperiksa, jadi skenario sukses cukup menyebut `http`.

```go
tambah(skenario{
    bagian: "H. Bentrok database",
    label:  "H5 contoh skenario baru",
    metode: "POST", path: K,
    payload:   kunj(uid("h05"), a1, "2026-09-12"),
    http:      409,
    errorCode: 3001,
    errorName: "DUPLICATE_ID",
    pesan:     "sudah dipakai",
})
```

Urutan penting: skenario awal membuat data yang dirujuk skenario berikutnya.
Sisipkan di bagian yang sesuai, jangan di awal berkas.

Bila menambah `errorCode` baru, perbarui juga `errors.md` di root.
