# API Integrasi Stunting

Service ini menyatukan data pemantauan stunting dari tiga sumber — posyandu,
puskesmas, dan rumah sakit — ke dalam satu database ildki.

Semua endpoint ada di bawah `/api` dan butuh API key:

```
POST http://<host>/api/kunjungan
X-API-Key: <key>
Content-Type: application/json
```

Yang di luar `/api` — `/stunting/health`, `/stunting/info`, `/stunting/docs` —
bebas diakses tanpa key.

---

## Pilih jalur Anda

Siapa Anda menentukan endpoint mana yang dipakai. Isi dan aturannya berbeda,
jangan tertukar.

| | **Jakantro (posyandu)** | **Faskes (puskesmas & RS)** |
|---|---|---|
| halaman | [Jalur Jakantro](?doc=jakantro) | [Jalur Faskes](?doc=faskes) |
| endpoint kirim | `POST` pada `/api/kunjungan`, `/api/kesehatan`, `/api/orangtua`, `/api/anak`; `PUT` pada `/api/kunjungan/:id`, `/api/kesehatan/:id`, `/api/anak` | `POST /api/faskes/pemeriksaan`, `PUT /api/faskes/:resource` |
| siapa menentukan `id` | **Anda** | **sistem** |
| kunci pencocokan | `id` kiriman Anda | `id_satusehat` (IHS id) |
| identitas pengirim | group `jakantro` di API key | group `orgid:<satusehat id>` di API key |
| isi | kunjungan posyandu dan kuisioner | kunjungan, observasi, diagnosa, layanan, rujukan |

Faskes punya pilihan kedua: kirim ke SatuSehat lewat IL, lalu data mengalir
sendiri ke sini lewat Kafka tanpa memanggil endpoint apa pun. Kedua jalur boleh
dipakai bersamaan — `id_satusehat` yang membuat datanya menyatu, bukan berganda.

Membacanya sama untuk keduanya: [endpoint GET](?doc=baca) tidak membedakan asal
data. Di halaman itu juga ada [antrean pesan kafka yang gagal](?doc=baca#antrean-pesan-kafka-yang-gagal),
untuk memproses ulang data SatuSehat yang sempat gagal masuk.

---

## Aturan `id`

> **Penting.** Ini sumber kebingungan paling sering. Aturannya berlawanan antara
> kedua jalur.

**Jakantro menentukan `id` sendiri.** Kirim UUID — atau string apa pun ≤ 36
karakter — dan itu yang jadi primary key. Mengirim `id` yang sama dua kali dapat
`409` `3001 DUPLICATE_ID`. Ini wajib karena jakantro bukan faskes: tanpa IHS id,
tidak ada cara lain mencocokkan kiriman dengan data yang sudah tersimpan.

**Faskes tidak mengirim `id` sama sekali.** Sistem yang membangkitkannya, dan
pencocokan memakai `id_satusehat`. Karena itu `kunjungan.id_satusehat` wajib,
dan kiriman ulang aman — tidak menggandakan data.

`id` yang telanjur dikirim ke `POST /api/faskes/pemeriksaan` diabaikan.
Sebaliknya, `id` yang hilang di endpoint jakantro ditolak:

```json
{
  "httpCode": 422,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "kunjungan.id wajib dikirim"
}
```

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

Ada dua cara ditolak, dan bedanya menentukan tindakan Anda:

- **`401` `4001`** — key tidak dikirim, salah bentuk, atau tidak terdaftar.
  Perbaiki key, lalu retry.
- **`403` `4002`** — key sah, tapi role Anda tidak diizinkan untuk method + path
  itu. Retry tidak menolong; minta tambahan role ke pengelola.

Role dibaca dari database tiap request, bukan dari key. Jadi perubahan role
berlaku tanpa ganti key, dengan jeda maksimal 30 detik karena ada cache. Aturan
per-endpoint punya cache sendiri, 60 detik.

Selengkapnya, termasuk `403` `1008` yang berbeda sebab, ada di
[Katalog Error](?doc=errors).

---

## Aturan yang berlaku di semua endpoint

**Tanggal selalu `YYYY-MM-DD`.** Format lain ditolak `422`. Timestamp
(`created_at`, `updated_at`, `deleted_at`) dikirim balik sebagai RFC 3339 UTC.

**Field yang tidak dikirim jadi `null`,** kecuali `source_data` yang punya
default di database.

**Kolom id bertipe `varchar`,** jadi nilainya kembali apa adanya — sepanjang
yang Anda kirim, tanpa tambahan apa pun. Bandingkan langsung, tidak perlu
`TRIM()`.

Tiga kolom masih bertipe `character` dan **dipadding spasi sampai panjang
penuh**:

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

---

## Bentuk error

Semua error punya bentuk yang sama:

```json
{
  "httpCode": 422,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "anak.id (\"a1\") tidak sama dengan kunjungan.id_anak (\"a2\")"
}
```

`httpCode` menentukan sikap umum, `errorCode` lebih spesifik dan lebih stabil —
percabangkan handler Anda pada `errorCode`. Di environment `development` ada
tambahan `details` dan `stack`; keduanya hilang di environment lain.

Tabel acuan lengkap beserta arti dan tindak lanjut tiap kode ada di
[Katalog Error](?doc=errors).

---

## Laporan pengujian

Contoh di dokumentasi sengaja dibatasi supaya terbaca. Kalau butuh lebih banyak
kasus — terutama kombinasi yang ditolak — ada laporan pengujian berisi **155
skenario** terhadap seluruh endpoint, lengkap dengan payload yang dikirim,
status code, dan response utuh untuk masing-masing.

[Lihat laporan pengujian](?doc=laporan)

Laporan itu dihasilkan dari test yang ditembakkan ke service sungguhan, jadi
isinya bukan contoh yang diketik tangan — dan versi yang Anda lihat di sini ikut
tertanam di biner service, jadi selalu sesuai dengan versi yang sedang berjalan.

Untuk membangkitkan ulang:

```bash
cd tests
GOWORK=off KUNCI_WRITE=<key> KUNCI_READ=<key> KUNCI_FASKES=<key> go test -p 1 ./functional/api_stunting/
```

`KUNCI_FASKES` adalah kunci ber-group `orgid:<satusehat id>`, dipakai skenario
jalur faskes.
