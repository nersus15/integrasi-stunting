# Katalog Error — Modul Stunting

Setiap `errorCode` di sini punya arti tetap. Kalau Anda menulis handler error,
pakai `errorCode` sebagai patokan — bukan `message`, yang bisa berubah redaksinya.

---

## Bentuk response error

```json
{
  "httpCode": 422,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "bentuk flat hanya untuk kunjungan saja, jangan disertai key orangtua, anak, maupun kunjungan"
}
```

Di environment `development` ada dua field tambahan: `details` (isinya sama
dengan `message`) dan `stack` (trace Go, sekitar 60 baris). Keduanya hilang di
environment lain. Pesan mentah Postgres tidak pernah ikut ke response — nama
constraint dan SQLSTATE hanya masuk log.

## Cara membacanya

**Percabangkan handler Anda pada `errorCode`, bukan `httpCode`.** Satu `httpCode`
bisa dipakai beberapa sebab yang tindakannya berbeda: `422` dipakai `1002`
(isi field salah) sampai `6001` (payload benar, sengaja tidak disimpan), dan
`403` dipakai `4002` (role tidak boleh) maupun `1008` (data milik pihak lain).

Di keluarga `1xxx` hanya `1001` yang `400`, karena hanya di situ body gagal
dibaca sebagai JSON. Sisanya `422`: JSON sudah terbaca, isinya yang belum
memenuhi aturan. Jadi `400` menunjuk ke bug serialisasi, `422` menunjuk ke data
yang perlu diperbaiki operator.

Digit pertama `errorCode` menunjukkan apa yang perlu Anda lakukan:

| grup | arti | yang perlu dilakukan |
|---|---|---|
| `1xxx` | payload salah, atau kewenangan kurang | perbaiki dulu, retry apa adanya percuma |
| `2xxx` | data tidak ketemu | lanjutkan sesuai alur, ini bukan kegagalan |
| `3xxx` | bentrok dengan data tersimpan | selesaikan duplikasinya |
| `4xxx` | key atau role | `4001` perbaiki key, `4002` minta role — jangan retry |
| `5xxx` | masalah di sisi service | `5002` boleh retry, `5001` laporkan |
| `6xxx` | payload sah tapi sengaja tidak disimpan | bukan kesalahan Anda, jangan retry |

## Tabel acuan lengkap

Seluruh kode yang bisa dikembalikan service. Kolom `message` adalah pesan
bawaan; banyak kode menggantinya dengan pesan yang lebih spesifik, dan pesan
itulah yang perlu Anda tampilkan ke operator.

| HTTP | code | name | message bawaan | arti | tindak lanjut |
|---|---|---|---|---|---|
| 400 | 1001 | `BODY_INVALID` | Body tidak bisa dibaca sebagai JSON | Body bukan JSON yang sah, atau kosong | Perbaiki serialisasi dan `Content-Type`. Retry percuma |
| 422 | 1002 | `VALIDATION_ERROR` | Data tidak memenuhi aturan | Isi sebuah field melanggar aturan | Perbaiki field yang disebut di `message`, kirim ulang |
| 422 | 1003 | `PAYLOAD_SHAPE_INVALID` | Bentuk payload tidak dikenali | Susunan key tidak sah atau id rujuk-silang tidak konsisten | Susun ulang mengikuti salah satu bentuk yang sah |
| 422 | 1004 | `REQUIRED_FIELD_MISSING` | Kolom wajib tidak boleh kosong | Kolom `NOT NULL` menerima nilai kosong | Isi kolom yang disebut. Kalau di luar kontrak, laporkan |
| 422 | 1005 | `REFERENCE_NOT_FOUND` | Referensi tidak ditemukan | Foreign key menunjuk baris yang tidak ada | Kirim data induknya lebih dulu |
| 422 | 1006 | `CONSTRAINT_VIOLATION` | Data tidak memenuhi aturan database | Melanggar check constraint (umumnya nilai enum), atau nilai di luar batas kolom | Perbaiki nilainya sesuai daftar atau batas yang sah |
| 422 | 1007 | `PARAM_REQUIRED` | Parameter harus dikirim | Path parameter wajib terbaca kosong | Sertakan parameter pada URL |
| 403 | 1008 | `FORBIDDEN` | User tidak memiliki akses | Key sah, tapi Anda tidak berwenang atas posyandu atau faskes itu | Jangan retry. Periksa keterkaitan posyandu–faskes |
| 404 | 2001 | `NOT_FOUND` | Data tidak ditemukan | Tidak ada baris yang cocok | Lanjutkan sesuai alur, ini bukan kegagalan |
| 409 | 3001 | `DUPLICATE_ID` | Id sudah dipakai | Primary key sudah ada | Abaikan kalau kiriman ulang; kalau data baru, terbitkan id baru |
| 409 | 3002 | `DUPLICATE_NIK_ORANGTUA` | nik orangtua sudah terdaftar | NIK orangtua dipakai baris lain | Coba ulang sekali; kalau tetap, laporkan |
| 409 | 3003 | `DUPLICATE_NO_KK` | no_kk sudah terdaftar | Nomor KK dipakai keluarga lain | Periksa apakah keluarga itu terdaftar dengan NIK berbeda |
| 409 | 3004 | `DUPLICATE_NIK_ANAK` | nik anak sudah terdaftar | NIK anak dipakai baris lain | Coba ulang sekali; kalau tetap, laporkan |
| 409 | 3005 | `DUPLICATE_DATA` | Data sudah ada | Pelanggaran unik yang belum dipetakan khusus | Laporkan beserta payload |
| 409 | 3006 | `ANAK_TERDAFTAR_DI_ORANGTUA_LAIN` | nik anak sudah terdaftar pada orangtua yang berbeda | Dua sumber berselisih soal anak ini milik siapa | Jangan retry. Tentukan dulu mana yang benar |
| 401 | 4001 | `UNAUTHORIZED` | Authorization header required | Key tidak dikirim, salah bentuk, tidak terdaftar, atau kedaluwarsa | Perbaiki key, lalu retry |
| 403 | 4002 | `FORBIDDEN` | User access denied | Key dikenali, tapi role tidak diizinkan untuk method + path itu | Jangan retry. Minta tambahan role |
| 500 | 5001 | `INTERNAL_ERROR` | Terjadi kesalahan | Ada yang rusak di sisi service | Laporkan beserta waktu dan payload |
| 503 | 5002 | `SERVICE_UNAVAILABLE` | Layanan sedang sibuk, silakan coba lagi | Batas waktu terlampaui, transaksi dibatalkan | Boleh diulang setelah jeda |
| 422 | 6001 | `TIDAK_DISIMPAN` | Data tidak memenuhi kriteria pemantauan stunting | Payload benar, tapi memang tidak disimpan | Perlakukan sebagai "diterima, tidak relevan". Jangan retry |

Dua kode sama-sama `403` `FORBIDDEN` tapi berbeda sebab: `4002` berarti role
Anda tidak boleh memanggil endpoint itu sama sekali, `1008` berarti endpointnya
boleh tapi data yang Anda minta milik posyandu atau faskes lain.

---

## 1xxx — payload perlu diperbaiki

### `1001` BODY_INVALID · HTTP 400

> Body tidak bisa dibaca sebagai JSON

Request body bukan JSON yang sah, atau kosong sama sekali. Terjadi sebelum
apa pun diperiksa, jadi tidak ada informasi field.

**Tindakan.** Periksa serialisasi di sisi pengirim dan header `Content-Type`.
Mengulang request yang sama akan selalu gagal.

### `1002` VALIDATION_ERROR · HTTP 422

> pesan menyebut field dan syaratnya, contoh: `nik harus 16 digit angka`

Satu field tidak memenuhi aturan. `message` selalu menyebut nama field, jadi
bisa dipakai langsung untuk menunjuk kesalahan ke operator.

Muncul juga pada endpoint GET yang kriteria pencariannya kurang lengkap,
misalnya mencari anak tanpa menyebut nik maupun id orangtua.

**Tindakan.** Perbaiki field yang disebut lalu kirim ulang.

### `1003` PAYLOAD_SHAPE_INVALID · HTTP 422

> contoh: `anak.id ("...") tidak sama dengan kunjungan.id_anak ("...")`

Kombinasi key tidak sah atau rujuk-silang id tidak konsisten. Berbeda dari
`1002`: di sini setiap field bisa saja benar, tapi susunannya yang salah.

Mencakup bentuk flat yang disertai key nested, key `kunjungan` yang hilang, `anak.id` yang tidak sama dengan `kunjungan.id_anak`,
`orangtua.id` yang tidak sama dengan `anak.id_orangtua`, key `orangtua` tanpa key
`anak`, dan orangtua baru tanpa data anak baru.

**Tindakan.** Susun ulang payload mengikuti salah satu bentuk yang sah. Pesan
untuk kasus rujuk-silang mencantumkan kedua nilai yang berselisih.

### `1004` REQUIRED_FIELD_MISSING · HTTP 422

> `kolom wajib tidak boleh kosong: <nama_kolom>`

Database menolak karena kolom `NOT NULL` menerima nilai kosong. Umumnya berarti
ada field wajib yang lolos validasi aplikasi tapi tetap kosong sampai ke tabel.

**Tindakan.** Isi kolom yang disebut. Kalau kolom itu tidak ada di kontrak
payload, laporkan — kemungkinan ada ketidakcocokan antara validator dan skema.

### `1005` REFERENCE_NOT_FOUND · HTTP 422

> Referensi tidak ditemukan

Foreign key menunjuk baris yang tidak ada. Paling sering `kunjungan.id_anak`
atau `anak.id_orangtua` merujuk data yang belum pernah dikirim.

**Tindakan.** Kirim data induknya lebih dulu, atau gabungkan dalam satu payload
bentuk `REGISTRASI_LENGKAP` / `ANAK_BARU`.

### `1006` CONSTRAINT_VIOLATION · HTTP 422

> Data tidak memenuhi aturan database

Pelanggaran aturan database di luar kategori di atas — check constraint,
pelanggaran tipe, dan nilai di luar batas kolom: angka yang melampaui presisi
kolom (mis. `nilai_angka` di atas 99999999.9999) atau teks yang melebihi
panjang maksimum.

Nilai enum berikut tidak diperiksa aplikasi, melainkan dijaga database:

| kolom | nilai yang diterima |
|---|---|
| `diagnosa.jenis` | `diagnosis`, `alergi` — kosong berarti `diagnosis` |
| `layanan.jenis` | `procedure`, `medication_dispense`, `nutrition_order`, `immunization`, `service_request` |
| `rujukan.jenis` | `rujukan`, `rujuk_balik`, `internal` — kosong berarti disimpulkan sendiri |

Seluruh transaksi dibatalkan, jadi tidak ada sebagian data yang terlanjur masuk.

**Tindakan.** Kalau nilainya ada di tabel di atas, laporkan beserta payload —
kemungkinan aturan database dan kontrak payload tidak sinkron. Kalau di luar
tabel, perbaiki nilainya.

### `1007` PARAM_REQUIRED · HTTP 422

> Parameter harus dikirim

Parameter path yang wajib terbaca kosong. Sifatnya penjaga: URL yang segmennya
benar-benar hilang, seperti `/api/orangtua//anak`, tidak sampai ke handler dan
dijawab `404` oleh router.

**Tindakan.** Sertakan parameter yang disebut pada URL.

### `1008` FORBIDDEN · HTTP 403

> `endpoint ini hanya untuk faskes, bukan jakantro` · `Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan` · `Tidak bisa verifikasi akses, posyandu tidak terhubung dengan puskesmas`

Key Anda sah dan dikenali, tetapi yang Anda minta di luar kewenangan key itu.
Dua sebab yang mungkin:

- **jalur yang salah** — key jakantro dipakai memanggil
  `POST /api/faskes/pemeriksaan`, yang hanya untuk puskesmas dan RS
- **data milik pihak lain** — posyandu atau faskes yang diminta tidak terhubung
  dengan wilayah kerja Anda
- **anak atau orangtua di luar kewenangan** — di luar wilayah kerja Anda dan
  tidak sedang dirujuk ke faskes Anda (lihat
  [hak akses](?doc=faskes#hak-akses-anak-dan-orangtua))
- **encounter milik faskes lain** — `kunjungan.id_satusehat` atau
  `ref_encounter` menunjuk kunjungan faskes lain, atau encounter yang belum
  tersimpan sehingga kepemilikannya tidak bisa diverifikasi

Bedanya dengan `4002`: `4002` menolak karena role Anda tidak boleh memanggil
endpoint itu sama sekali; `1008` sudah lolos lapisan role, lalu ditolak aturan
kewenangan di dalam handler.

**Tindakan.** Jangan retry. Pastikan Anda memakai key untuk jalur yang benar,
dan posyandu yang dimaksud memang terhubung ke faskes Anda.

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

---

## 6xxx — payload sah tapi sengaja tidak disimpan

### `6001` TIDAK_DISIMPAN · HTTP 422

> Data tidak memenuhi kriteria pemantauan stunting

Payload Anda benar dan lolos validasi, tapi service memang tidak menyimpannya.
Hanya muncul di `POST /api/faskes/pemeriksaan`.

Tiga sebab, dan pesannya menyebutkan yang mana:

```json
{
  "httpCode": 422,
  "errorCode": 6001,
  "errorName": "TIDAK_DISIMPAN",
  "message": "anak bukan sasaran pemantauan, umurnya sudah lewat 5 tahun"
}
```

```json
{
  "httpCode": 422,
  "errorCode": 6001,
  "errorName": "TIDAK_DISIMPAN",
  "message": "tidak ada tanda stunting dan riwayat terakhir bukan stunting"
}
```

```json
{
  "httpCode": 422,
  "errorCode": 6001,
  "errorName": "TIDAK_DISIMPAN",
  "message": "tidak ada tanda stunting dan anak belum punya riwayat"
}
```

Penolakan tidak meninggalkan apa pun di database — anak baru tidak jadi dibuat.

Service ini hanya menyimpan data medis anak yang sedang dipantau: bundle yang
membawa tanda stunting, atau anak yang kunjungan terakhirnya berstatus stunting.
Pemeriksaan anak sehat memang tidak disimpan.

**Tindakan.** Bukan kesalahan Anda dan tidak bisa diperbaiki dengan mengubah
payload. Jangan retry. Perlakukan sebagai "diterima, tidak relevan".

---

Untuk bentuk payload yang menghasilkan error-error di atas, lihat
[Jalur Jakantro](?doc=jakantro), [Jalur Faskes](?doc=faskes), dan
[Contoh Payload](?doc=contoh).
