# Katalog Error — Modul Stunting

Setiap `errorCode` di sini punya arti tetap. Kalau Anda menulis handler error,
pakai `errorCode` sebagai patokan — bukan `message`, yang bisa berubah redaksinya.

---

## Bentuk response error

```json
{
  "httpCode": 400,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "bentuk flat hanya untuk kunjungan saja, jangan disertai key orangtua, anak, maupun kunjungan"
}
```

Di environment `development` ada dua field tambahan: `details` (isinya sama dengan
`message`) dan `stack` (trace Go, sekitar 60 baris). Keduanya hilang di
environment lain. Pesan mentah Postgres tidak pernah ikut ke response — nama
constraint dan SQLSTATE hanya masuk log.

## `httpCode` vs `errorCode`

`httpCode` menentukan sikap umum, `errorCode` lebih spesifik dan lebih stabil.
Dua error bisa sama-sama `400` tapi butuh penanganan berbeda: `1002` berarti isi
field salah, `1005` berarti data yang dirujuk belum ada.

Digit pertama `errorCode` menunjukkan apa yang perlu Anda lakukan:

| grup | arti | yang perlu dilakukan |
|---|---|---|
| `1xxx` | payload salah | perbaiki dulu, retry apa adanya percuma |
| `2xxx` | data tidak ketemu | lanjutkan sesuai alur, ini bukan kegagalan |
| `3xxx` | bentrok dengan data tersimpan | selesaikan duplikasinya |
| `4xxx` | key atau role | `4001` perbaiki key, `4002` minta role — jangan retry |
| `5xxx` | masalah di sisi service | `5002` boleh retry, `5001` laporkan |

---

## 1xxx — payload perlu diperbaiki

### `1001` BODY_INVALID · HTTP 400

> Body tidak bisa dibaca sebagai JSON

Request body bukan JSON yang sah, atau kosong sama sekali. Terjadi sebelum
apa pun diperiksa, jadi tidak ada informasi field.

**Tindakan.** Periksa serialisasi di sisi pengirim dan header `Content-Type`.
Mengulang request yang sama akan selalu gagal.

### `1002` VALIDATION_ERROR · HTTP 400

> pesan menyebut field dan syaratnya, contoh: `nik harus 16 digit angka`

Satu field tidak memenuhi aturan. `message` selalu menyebut nama field, jadi
bisa dipakai langsung untuk menunjuk kesalahan ke operator.

**Tindakan.** Perbaiki field yang disebut lalu kirim ulang.

### `1003` PAYLOAD_SHAPE_INVALID · HTTP 400

> contoh: `anak.id ("...") tidak sama dengan kunjungan.id_anak ("...")`

Kombinasi key tidak sah atau rujuk-silang id tidak konsisten. Berbeda dari
`1002`: di sini setiap field bisa saja benar, tapi susunannya yang salah.

Mencakup bentuk flat yang disertai key nested, key `kunjungan` yang hilang,
`anak.id` yang tidak sama dengan `kunjungan.id_anak`, `orangtua.id` yang tidak
sama dengan `anak.id_orangtua`, key `orangtua` tanpa key `anak`, dan orangtua
baru tanpa data anak baru.

**Tindakan.** Susun ulang payload mengikuti salah satu bentuk yang sah. Pesan
untuk kasus rujuk-silang mencantumkan kedua nilai yang berselisih.

### `1004` REQUIRED_FIELD_MISSING · HTTP 400

> `kolom wajib tidak boleh kosong: <nama_kolom>`

Database menolak karena kolom `NOT NULL` menerima nilai kosong. Umumnya berarti
ada field wajib yang lolos validasi aplikasi tapi tetap kosong sampai ke tabel.

**Tindakan.** Isi kolom yang disebut. Kalau kolom itu tidak ada di kontrak
payload, laporkan — kemungkinan ada ketidakcocokan antara validator dan skema.

### `1005` REFERENCE_NOT_FOUND · HTTP 400

> Referensi tidak ditemukan

Foreign key menunjuk baris yang tidak ada. Paling sering `kunjungan.id_anak`
atau `anak.id_orangtua` merujuk data yang belum pernah dikirim.

**Tindakan.** Kirim data induknya lebih dulu, atau gabungkan dalam satu payload
bentuk `REGISTRASI_LENGKAP` / `ANAK_BARU`.

### `1006` CONSTRAINT_VIOLATION · HTTP 400

> Data tidak memenuhi aturan database

Pelanggaran aturan database di luar kategori di atas — check constraint,
pelanggaran tipe, dan sejenisnya.

**Tindakan.** Laporkan beserta payload; kemungkinan aturan database dan kontrak
payload tidak sinkron.

### `1007` PARAM_REQUIRED · HTTP 400

> Parameter harus dikirim

Parameter path yang wajib tidak disertakan — misalnya `id` pada
`GET /api/orangtua/:id/anak`.

**Tindakan.** Sertakan parameter yang disebut pada URL.

---

## 2xxx — tidak ditemukan

### `2001` NOT_FOUND · HTTP 404

> Data orang tua tidak ditemukan

Tidak ada baris yang cocok. Ini jawaban normal, bukan kegagalan.

Muncul juga pada endpoint riwayat kalau **anaknya** yang tidak ada. Anak yang ada
tapi belum punya riwayat tetap `200`, dengan array kosong.

**Tindakan.** Lanjutkan sesuai alur — biasanya artinya data induk perlu dikirim
lebih dulu.

---

## 3xxx — bentrok dengan data tersimpan

### `3001` DUPLICATE_ID · HTTP 409

> `id kunjungan sudah dipakai` (pesan menyebut entitasnya)

Primary key sudah ada. Paling sering karena kunjungan yang sama dikirim dua kali.

**Tindakan.** Kalau ini pengiriman ulang yang tidak disengaja, abaikan — datanya
sudah tersimpan. Kalau memang data baru, terbitkan id baru.

### `3002` DUPLICATE_NIK_ORANGTUA · HTTP 409

> nik orangtua sudah terdaftar

NIK orangtua sudah dipakai baris lain. Dalam alur normal ini **tidak akan
muncul**, karena service mencari lebih dulu dan memakai orangtua yang sudah ada.
Sifatnya penjaga terakhir bila dua request dengan NIK sama tiba bersamaan.

**Tindakan.** Coba ulang sekali; kalau tetap muncul, laporkan.

### `3003` DUPLICATE_NO_KK · HTTP 409

> no_kk sudah terdaftar

Nomor KK sudah dipakai keluarga lain. Tidak seperti NIK, `no_kk` tidak dipakai
untuk mencari, jadi ini muncul dalam alur normal.

**Tindakan.** Periksa apakah keluarga itu sudah terdaftar dengan NIK berbeda.
Salah satu datanya perlu dikoreksi di sumber.

### `3004` DUPLICATE_NIK_ANAK · HTTP 409

> nik anak sudah terdaftar

NIK anak sudah dipakai. Seperti `3002`, dalam alur normal tidak muncul karena
service mencari lebih dulu — kecuali pada kondisi balapan.

**Tindakan.** Coba ulang sekali; kalau tetap muncul, laporkan.

### `3005` DUPLICATE_DATA · HTTP 409

> Data sudah ada

Pelanggaran constraint unik yang belum punya pemetaan khusus. Muncul juga bila
ditemukan lebih dari satu anak dengan orangtua dan urutan yang sama.

**Tindakan.** Laporkan beserta payload supaya bisa dipetakan lebih spesifik.

### `3006` ANAK_TERDAFTAR_DI_ORANGTUA_LAIN · HTTP 409

> nik anak `<nik>` sudah terdaftar pada orangtua yang berbeda

NIK anak sudah ada di database, tetapi di bawah `id_orangtua` yang berbeda dari
yang dikirim. Payloadnya sah — dua sumber data yang berselisih tentang anak ini
milik siapa.

**Tindakan.** Jangan paksa kirim ulang. Tentukan dulu mana yang benar: NIK anak
salah, atau `id_orangtua` salah. Seluruh transaksi dibatalkan, tidak ada data
setengah tersimpan.

---

## 4xxx — key atau role

Grup ini muncul di lapisan auth, **sebelum** request sampai ke handler. Jadi bisa
kena di endpoint mana pun, termasuk yang payloadnya sudah benar.

Bedakan keduanya baik-baik: sama-sama penolakan, tapi tindakannya berlawanan.

### `4001` UNAUTHORIZED · HTTP 401

> `Authorization header required` · `User not found: <key>` · `Invalid or expired token`

Service tidak mengenali Anda. Key tidak dikirim, salah bentuk, tidak terdaftar,
atau sudah expired.

Header yang diterima:

```
X-API-Key: <key>
Authorization: APIKey <key>
```

**Tindakan.** Periksa key, lalu retry. Kalau key-nya memang salah, retry berapa
kali pun hasilnya sama.

### `4002` FORBIDDEN · HTTP 403

> `User access denied`

Key Anda valid dan Anda dikenali — tapi role Anda tidak diizinkan untuk
kombinasi method + path itu.

**Tindakan.** Jangan retry. Ini bukan soal key, tapi soal role. Minta tambahan
role ke pengelola, atau pastikan endpoint yang Anda panggil memang yang dimaksud.

Endpoint yang tidak punya aturan khusus terbuka untuk semua role, jadi `4002`
hanya muncul di endpoint yang memang dibatasi.

Perubahan role berlaku tanpa ganti key, dengan jeda maksimal 30 detik karena
ada cache. Aturan per-endpoint punya cache sendiri, 60 detik — jadi kalau
pengelola baru menambahkan aturan, efeknya bisa telat sampai satu menit.

---

## 5xxx — masalah di sisi layanan

### `5001` INTERNAL_ERROR · HTTP 500

> Terjadi kesalahan

Ada yang rusak di sisi service. Sebab aslinya cuma masuk log, tidak ikut ke
response.

Termasuk kalau lapisan auth sendiri yang bermasalah. Bedanya dengan `4002`:
`4002` berarti Anda memang tidak berhak, `5001` berarti service tidak bisa
menentukan berhak atau tidak.

**Tindakan.** Laporkan beserta waktu kejadian dan payload. Retry biasanya tidak
menolong.

### `5002` SERVICE_UNAVAILABLE · HTTP 503

> Layanan sedang sibuk, silakan coba lagi

Batas waktu terlampaui — context deadline atau `statement_timeout` Postgres.
Transaksi dibatalkan, tidak ada data setengah tersimpan.

**Tindakan.** Boleh dicoba ulang setelah jeda. Satu-satunya kode yang memang
dirancang untuk diulang.
