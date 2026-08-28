# Katalog Error — Modul Stunting

Dokumen ini adalah kontrak error antara service dan pemanggilnya (jakantro).

Sumber kebenarannya ada di dua tempat:

- [`modules/stunting/helper/Exceptions/error.go`](modules/stunting/helper/Exceptions/error.go) — error dari modul (`1xxx`, `2xxx`, `3xxx`, `5xxx`)
- [`libraries/webcore/port/auth/authz.go`](libraries/webcore/port/auth/authz.go) — error autentikasi dan otorisasi (`4xxx`), terbit sebelum request sampai ke modul

**Setiap perubahan pada error di kedua tempat itu harus diikuti perubahan di dokumen ini.**

Terakhir diverifikasi 2026-08-27: kelompok `4xxx` diuji langsung ke service berjalan
untuk keempat kondisi (tanpa kredensial, kunci tidak dikenali, hak kurang, hak cukup)
dan lewat lima skenario unit di `tests/functional/auth/`. Kelompok lain terakhir
diverifikasi 2026-08-24 lewat 94 skenario terhadap keempat endpoint.

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

Pada `environment: development` ada dua field tambahan: `details` berisi pesan yang
sama dengan `message`, dan `stack` berisi jejak Go sekitar 60 baris. Keduanya
dibuang otomatis pada environment lain. Pesan mentah dari Postgres tidak pernah
ikut ke klien — nama constraint dan SQLSTATE hanya masuk ke log.

## `httpCode` dan `errorCode` berbeda

`httpCode` menentukan sikap umum. `errorCode` lebih rinci dan stabil: beberapa
kesalahan berbagi status HTTP yang sama tetapi menuntut penanganan berbeda.
`1002` dan `1005` sama-sama HTTP 400, tapi yang pertama berarti isi field salah
dan yang kedua berarti data rujukan belum dikirim.

Digit pertama `errorCode` menyatakan tindakan yang diharapkan:

| kelompok | arti | sikap pemanggil |
|---|---|---|
| `1xxx` | payload perlu diperbaiki | jangan diulang apa adanya, perbaiki dulu |
| `2xxx` | data yang dicari tidak ada | lanjutkan sesuai alur, bukan kegagalan |
| `3xxx` | bentrok dengan data tersimpan | selesaikan duplikasinya, jangan paksa ulang |
| `4xxx` | kredensial atau hak | `4001` perbaiki kunci, `4002` minta hak — jangan diulang |
| `5xxx` | masalah di sisi layanan | `5002` boleh dicoba ulang, `5001` laporkan |

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

Pencarian tidak menemukan baris yang cocok. Ini jawaban normal, bukan kegagalan.

`Classify` mengenali dua sumber: `utils.ErrTidakDitemukan` yang diterjemahkan
repository, dan `sql.ErrNoRows` mentah sebagai jaring pengaman — sehingga
repository yang lupa menerjemahkan tetap menghasilkan 404, bukan 500.

**Tindakan.** Lanjutkan sesuai alur — biasanya berarti data perlu didaftarkan
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

## 4xxx — kredensial atau hak

Kelompok ini terbit di lapisan autentikasi, **sebelum** request sampai ke modul.
Karena itu ia bisa muncul pada endpoint mana pun, termasuk yang payloadnya sah.

Ada tiga tempat yang bisa menerbitkannya, semuanya lewat `auth.Deny` sehingga
bentuk response-nya seragam:

| tempat | berlaku untuk | sumber aturan |
|---|---|---|
| `AuthN.GetAuthenticatonHandler` | semua route di bawah `PathPrefix` | tabel `access.auth_resources` |
| `middleware.RoleRequired` | route yang dipasangi, RBAC | daftar peran di kode |
| `middleware.PermissionRequired` | route yang dipasangi, ABAC | kebijakan pemanggil |

Dua yang terakhir berjalan **sesudah** yang pertama. Bila satu path diatur di
tabel sekaligus dipasangi middleware, pemanggil harus memenuhi keduanya — yang
paling ketat menang. Hindari menyatakan aturan yang sama di dua tempat.

Pembedaan `4001` dan `4002` penting: keduanya penolakan, tapi menuntut tindakan
yang berlawanan. Sebelum 2026-08-27 keduanya dijawab `401`, sehingga pemanggil
yang menerima "User access denied" akan menyimpulkan kuncinya salah lalu mencoba
ulang selamanya.

### `4001` UNAUTHORIZED · HTTP 401

> `Authorization header required` · `User not found: <kunci>` · `Invalid or expired token`

Layanan tidak mengenali pemanggil. Kredensial tidak dikirim, bentuknya salah,
tidak terdaftar, atau sudah kedaluwarsa.

Header yang diterima: `X-API-Key: <kunci>` atau `Authorization: APIKey <kunci>`
untuk `auth.type: apikey`, dan `Authorization: Bearer <token>` untuk `jwt`.

**Tindakan.** Periksa kunci atau ambil token baru. Mengulang dengan kredensial
yang sama akan selalu gagal.

### `4002` FORBIDDEN · HTTP 403

> `User access denied`

Pemanggil **dikenali dengan benar**, tetapi haknya kurang untuk kombinasi method
dan path tersebut. Aturannya ada di tabel `access.auth_resources`, dan peran
pemanggil di `access.auth_users.roles`.

**Tindakan.** Jangan diulang — hasilnya akan sama. Ini bukan masalah kredensial
melainkan hak. Minta penambahan peran ke pengelola, atau pastikan memang endpoint
itu yang dimaksud.

Resource yang tidak punya baris aturan diizinkan secara bawaan, jadi `4002`
hanya muncul pada endpoint yang memang dibatasi — baik lewat tabel maupun lewat
`RoleRequired`/`PermissionRequired` di pendaftaran route.

Satu hal yang tidak terlihat dari response: `PermissionRequired` **tidak**
meluluskan kebijakan ABAC bersyarat, karena syaratnya butuh atribut resource
yang belum tersedia di titik itu. Kebijakan bersyarat harus dinyatakan lewat
`access.auth_resources`. Kalau pemanggil yakin kebijakannya mengizinkan tapi
tetap menerima `4002`, itu penyebab yang pertama perlu diperiksa.

---

## 5xxx — masalah di sisi layanan

### `5001` INTERNAL_ERROR · HTTP 500

> Terjadi kesalahan

Kesalahan yang tidak terduga. Sebab aslinya hanya masuk ke log, tidak dikirim
ke klien.

Kode ini juga dipakai lapisan otorisasi ketika jalurnya sendiri yang rusak —
query ke `access.auth_resources` gagal, atau tipe kontrol akses user dan
resource tidak cocok (RBAC vs ABAC). Ini sengaja dibedakan dari `4002`: yang
satu berarti pemanggil tidak berhak, yang lain berarti layanan tidak bisa
menentukan berhak atau tidak.

**Tindakan.** Laporkan beserta waktu kejadian dan payload. Mengulang biasanya
tidak menolong.

### `5002` SERVICE_UNAVAILABLE · HTTP 503

> Layanan sedang sibuk, silakan coba lagi

Batas waktu terlampaui — context deadline atau `statement_timeout` Postgres.
Transaksi dibatalkan, tidak ada data setengah tersimpan.

**Tindakan.** Boleh dicoba ulang setelah jeda. Satu-satunya kode yang memang
dirancang untuk diulang.

---

## Status verifikasi

| cara diuji | kode |
|---|---|
| lewat HTTP ke service berjalan | `1001` `1002` `1003` `1005` `2001` `3001` `3003` `3006` `4001` `4002` |
| lewat `Classify` dengan error Postgres sungguhan | `1004` `1005` `3001` `3002` `3003` `3004` |
| lewat `Classify` dengan error lapisan aplikasi | `2001` `5001` `5002` |
| lewat rantai auth dengan store palsu | `4001` `4002` `5001` |
| belum terpicu | `1006` `3005` |

`3002` dan `3004` tidak muncul lewat HTTP karena service mencari lebih dulu dan
memakai data yang sudah ada — keduanya hanya tercapai pada kondisi balapan, jadi
diuji langsung ke `Classify` dengan pelanggaran constraint sungguhan.

Dua yang belum terpicu: `1006` menuntut pelanggaran aturan database di luar
kategori yang sudah dipetakan, dan `3005` menuntut constraint unik yang belum
punya pemetaan khusus. Keduanya memang jalur cadangan.

---

## Menambah atau mengubah error

Untuk kelompok `4xxx` sumbernya bukan `error.go` melainkan konstanta di
[`port/auth/authz.go`](libraries/webcore/port/auth/authz.go), dan penolakannya
dirakit di `adapter/auth/authn/authn.go`. Penanda `auth.ErrAccessDenied` yang
memisahkan `4002` dari `5001` — bila penolakan baru dibuat, pastikan ia
membungkus penanda itu supaya tidak salah jatuh ke `5001`.

1. Tambahkan `Kind` baru di `error.go`, pada kelompok yang sesuai dengan
   tindakan yang diharapkan pemanggil — bukan sekadar yang cocok status HTTP-nya.
2. Kalau berasal dari Postgres, tambahkan pemetaannya di `Classify`
   berdasarkan nama constraint. Nama constraint jadi bagian dari kontrak, jadi
   jangan diubah tanpa memperbarui dokumen ini.
3. Kalau berupa aturan bisnis, kembalikan langsung dari repository atau service
   memakai `Kind.New`, `Kind.WithMessage`, atau `Kind.Messagef`. `Classify`
   tidak akan menimpanya.
4. **Perbarui dokumen ini** — tambahkan entri lengkap dengan deskripsi dan
   tindakan, lalu sesuaikan tabel status verifikasi.
5. `errorCode` dan `errorName` yang sudah dipakai tidak boleh berubah artinya.
   Kalau maknanya bergeser, terbitkan kode baru dan tandai yang lama usang.
