# API Integrasi Stunting

Service ini menyalin data stunting dari jakantro ke database ildki. Semua
endpoint ada di bawah `/api` dan butuh API key.

```
POST http://<host>/api/kunjungan
X-API-Key: <key>
Content-Type: application/json
```

Yang di luar `/api` — `/stunting/health`, `/stunting/info`, `/stunting/docs` —
bebas diakses tanpa key.

---

## Authentication

Kirim key lewat salah satu header:

```
X-API-Key: <key>
```

```
Authorization: APIKey <key>
```

Key didapat dari pengelola service, tidak bisa di-generate sendiri.

Ada dua cara ditolak, dan bedanya menentukan apa yang harus Anda lakukan.

**401** — key tidak dikirim, salah bentuk, atau tidak terdaftar. Perbaiki key,
lalu retry.

```json
{
  "httpCode": 401,
  "errorCode": 4001,
  "errorName": "UNAUTHORIZED",
  "message": "Authorization header required"
}
```

**403** — key valid, tapi role Anda tidak diizinkan untuk method + path itu.
Retry tidak akan menolong. Minta tambahan role ke pengelola.

```json
{
  "httpCode": 403,
  "errorCode": 4002,
  "errorName": "FORBIDDEN",
  "message": "User access denied"
}
```

Role dibaca dari database tiap request, bukan dari key. Jadi kalau role Anda
diubah, efeknya terasa tanpa perlu ganti key — dengan jeda maksimal 30 detik
karena ada cache. Aturan per-endpoint punya cache sendiri, 60 detik.

---

## Dua jalur masuk

Siapa Anda menentukan endpoint mana yang dipakai. Isinya berbeda, jangan
tertukar.

| | **Jakantro (posyandu)** | **Faskes (puskesmas & RS)** |
|---|---|---|
| endpoint kirim | `POST /api/kunjungan`, `/api/kesehatan`, `/api/orangtua`, `/api/anak` | `POST /api/faskes/pemeriksaan` |
| `id` entitas | **Anda yang tentukan** | dibangkitkan sistem |
| kunci pencocokan | `id` kiriman Anda | `id_satusehat` |
| identitas pengirim | group `jakantro` di API key | group `orgid:<satusehat id>` di API key |
| isi | kunjungan posyandu dan kuisioner | kunjungan, observasi, diagnosa, layanan, rujukan |

Faskes punya pilihan kedua: kirim ke SatuSehat lewat IL, lalu data mengalir
sendiri ke sini lewat Kafka tanpa memanggil endpoint apa pun. Kedua jalur boleh
dipakai bersamaan — `id_satusehat` yang membuat datanya menyatu, bukan berganda.

Pembacaan sama untuk keduanya: [GET endpoints](?doc=baca) tidak membedakan
asal data.

---

## Yang perlu diketahui sebelum mulai

> **Penting.** Aturan `id` berbeda antara jakantro dan faskes. Salah satu
> sumber kebingungan paling sering — pastikan Anda membaca bagian yang sesuai.

**Semua `id` Anda yang tentukan.** Service tidak generate id. Kirim UUID (atau
string apa pun ≤ 36 karakter) dan itu yang jadi primary key. Kirim id yang sama
dua kali dapat `409 DUPLICATE_ID`.

Ini berlaku untuk **semua client saat ini**, tapi alasannya khusus jakantro:
jakantro bukan faskes, sehingga tidak bisa mencari pasien lewat satusehat id.
Tanpa id yang mereka tentukan sendiri, tidak ada cara mencocokkan data yang
dikirim dengan data yang sudah tersimpan.

Faskes — puskesmas dan rumah sakit — punya satusehat id dan sebetulnya tidak
memerlukan itu. Rencananya mereka nanti tidak perlu mengirim `id` untuk
orangtua, anak, kunjungan, maupun kesehatan; sistem yang membangkitkannya, dan
pencarian dilakukan lewat satusehat id.

**Jalur itu sudah tersedia** lewat [`POST /api/faskes/pemeriksaan`](?doc=faskes):
di sana faskes tidak mengirim `id` sama sekali, sistem yang membangkitkan, dan
pencocokan memakai `id_satusehat`.

Endpoint lama (`/api/kunjungan`, `/api/kesehatan`, `/api/orangtua`, `/api/anak`)
tetap mewajibkan `id`. Mengirim tanpa `id` di sana dijawab:

```json
{
  "httpCode": 400,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "kunjungan.id wajib dikirim"
}
```

**Semua kolom id bertipe `varchar`,** jadi nilainya kembali apa adanya —
sepanjang yang Anda kirim, tanpa tambahan apa pun. Bandingkan langsung, tidak
perlu `TRIM()`.

Yang masih bertipe `character` dan **dipadding spasi sampai panjang penuh**:

| kolom | panjang |
|---|---|
| `source_data` | 36 |
| `rt`, `rw` | 3 |
| `jenis_kelamin` | 1 |

Padding itu perilaku Postgres untuk tipe `character`, bukan bug. Untuk `rt`,
`rw`, dan `jenis_kelamin` tidak terasa karena nilainya memang selalu sepanjang
itu. Tapi `source_data` terasa:

```text
"source_data": "jakantro                            "
```

Kalau Anda mencocokkan `source_data` dengan string tertentu, potong dulu spasi
ekornya.

**Tanggal selalu `YYYY-MM-DD`.** Format lain ditolak `400`. Timestamp
(`created_at`, `updated_at`, `deleted_at`) dikirim balik sebagai RFC 3339 UTC.

**Field yang tidak dikirim jadi `null`,** kecuali `source_data` yang punya
default di database.

### Field wajib

Yang tidak disebut di sini boleh dikosongkan atau dihilangkan.

**orangtua** — `id`, `id_posyandu`, `no_kk`, `nik`, `nama_ayah`, `nama_ibu`,
`telepon`, `rt`, `rw`, `alamat`, `kia`.
`no_kk` dan `nik` harus 16 digit angka. `kia` hanya `0` atau `1`.
`usia_hamil` tidak boleh negatif, `kia_bayi_kecil` hanya `0` atau `1`.

**anak** — `id`, `id_orangtua`, `nama`, `tanggal_lahir`, `jenis_kelamin`,
`anak_ke`, `imd`, `bb_lahir`, `tb_lahir`, `lk_lahir`, `source_data`.
`jenis_kelamin` hanya `L` atau `P`. `anak_ke`, `bb_lahir`, `tb_lahir`, dan
`lk_lahir` harus lebih dari 0. `imd` hanya `0` atau `1`.
`nik` opsional — tapi kalau dikirim harus 16 digit angka.

**kunjungan** — `id`, `id_anak`, `tanggal_pengukuran`.
Field ukuran (`berat_badan`, `tinggi_badan`, `lingkar_kepala`, `lingkar_lengan`,
`lingkar_dada`) opsional, tapi kalau dikirim harus lebih dari 0. Field `asi_*`,
`vit_*`, `pitting_edema`, dan `kelas_ibu_balita` hanya menerima `0` atau `1`.

**kesehatan** — `id`, `id_anak`, `tanggal_pemantauan`.
Semua field `tbc_*`, `layanan_*`, dan `penyuluhan_*` opsional dan hanya menerima
`0` atau `1`.

---

## Laporan pengujian

Contoh di halaman ini sengaja dibatasi supaya terbaca. Kalau butuh lebih banyak
kasus — terutama kombinasi yang ditolak — ada laporan pengujian berisi **136
skenario** terhadap seluruh endpoint, lengkap dengan payload yang dikirim, status
code, dan response utuh untuk masing-masing.

[Lihat laporan pengujian](?doc=laporan)

Isinya dikelompokkan per topik: bentuk payload yang diterima, NIK anak opsional,
kombinasi yang ditolak, validasi tiap field, bentrok database, endpoint GET,
kesehatan, summary, autentikasi, dan dokumentasi.

Laporan itu dihasilkan dari test yang ditembakkan ke service sungguhan, jadi
isinya bukan contoh yang diketik tangan — dan versi yang Anda lihat di sini ikut
tertanam di biner service, jadi selalu sesuai dengan versi yang sedang berjalan.

Untuk membangkitkan ulang:

```bash
cd tests
KUNCI_WRITE=<key> KUNCI_READ=<key> go test ./functional/api_stunting/
```

---

## Error

Semua error punya bentuk yang sama:

```json
{
  "httpCode": 400,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "anak.id (\"a1\") tidak sama dengan kunjungan.id_anak (\"a2\")"
}
```

`httpCode` menentukan sikap umum, `errorCode` lebih spesifik dan stabil. Dua
error bisa sama-sama `400` tapi butuh penanganan berbeda — `1002` berarti isi
field salah, `1005` berarti data yang dirujuk belum ada.

Di environment `development` ada tambahan `details` dan `stack`. Keduanya hilang
di environment lain.

Daftar lengkap kode beserta artinya ada di [Katalog Error](?doc=errors).
