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
| endpoint kirim | `POST` pada `/api/kunjungan`, `/api/kesehatan`, `/api/orangtua`, `/api/anak`; `PUT` pada `/api/kunjungan/:id`, `/api/kesehatan/:id`, `/api/anak`, `/api/orangtua`; `DELETE` pada `/api/orangtua/:id`, `/api/anak/:id`, `/api/kunjungan/:id`, `/api/kesehatan/:id` | `POST` pada `/api/faskes/pemeriksaan`, `/api/faskes/orangtua`, `/api/faskes/anak`; `PUT /api/faskes/:resource` (termasuk `anak` dan `orangtua`) |
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

`id` yang telanjur dikirim ke endpoint faskes diabaikan.
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

**Field yang tidak dikirim jadi `null`,** kecuali `kia` dan `imd` yang menjadi
`0`, serta `source_data` yang selalu diisi sistem (lihat
[asal data](#asal-data-dan-kolom-audit)).

**Kolom yang diisi sistem diabaikan kalau dikirim:** `created_at`,
`updated_at`, `deleted_at`, `created_by`, `updated_by`, dan `deleted_by`, di
semua endpoint.
Daftar lengkap field yang diterima tiap endpoint ada di
[Jalur jakantro](?doc=jakantro#field-per-entitas) dan
[Jalur faskes](?doc=faskes#field-per-bagian).

**Kolom id bertipe `varchar`,** jadi nilainya kembali apa adanya — sepanjang
yang Anda kirim, tanpa tambahan apa pun. Bandingkan langsung, tidak perlu
`TRIM()`.

`rt`, `rw` (orangtua), dan `jenis_kelamin` masih bertipe `character`, jadi
dipadding spasi sampai 3 dan 1 karakter. Nilainya memang selalu sepanjang itu,
jadi biasanya tidak terasa.

---

## Aturan akses

Kunci API menentukan jalur Anda: grup `jakantro`, atau `orgid:<satusehat id>`
untuk faskes. Jalur yang salah ditolak `403` `1008` sebelum aturan di bawah
berlaku. Data dari SatuSehat masuk lewat Kafka, bukan lewat endpoint.

| data | aksi | jakantro | faskes | SatuSehat (Kafka) |
|---|---|---|---|---|
| orangtua | baca | semua | perlu [hak akses](#hak-akses-faskes), juga lewat `nik` | — |
| | buat | `POST /api/orangtua`, key `orangtua` di POST kunjungan/kesehatan | `POST /api/faskes/orangtua` | dari RelatedPerson |
| | ubah | semua field; untuk orangtua buatan faskes/SatuSehat, `nik` diabaikan | `PUT /api/faskes/orangtua`, perlu hak akses; `id_posyandu` diabaikan | RelatedPerson selalu menimpa identitas |
| | hapus | hanya buatan sendiri | — | — |
| anak | baca | semua | perlu hak akses, juga lewat `nik` | — |
| | buat | `POST /api/anak`, key `anak` di POST kunjungan/kesehatan | `POST /api/faskes/anak` (perlu hak akses atas `id_orangtua`), atau anak baru dari pemeriksaan | dari Patient |
| | ubah | semua field; untuk anak buatan faskes/SatuSehat, `nik` dan `id_orangtua` diabaikan | `PUT /api/faskes/anak`, perlu hak akses; `id_orangtua` diabaikan | Patient selalu menimpa identitas |
| | hapus | hanya buatan sendiri | — | — |
| kunjungan | baca | semua | perlu hak akses atas anaknya | — |
| | buat | POST kunjungan, untuk anak siapa pun | `POST /api/faskes/pemeriksaan`, perlu hak akses atas anaknya | dari Encounter |
| | ubah | hanya buatan sendiri | `PUT /api/faskes/kunjungan`, hanya pembuatnya | PUT Encounter |
| | hapus | hanya buatan sendiri | — | — |
| data medis (observasi, diagnosa, layanan, rujukan, episode) | baca | semua | lewat kunjungan anaknya | — |
| | ubah | — | `PUT /api/faskes/:resource`, hanya pembuat kunjungannya | PUT per resource |
| kesehatan | baca | semua | perlu hak akses atas anaknya | — |
| | buat, ubah, hapus | hanya buatan sendiri | — | — |

Penolakan selalu `403` `1008`. `id_satusehat` dan `source_data` kiriman
jakantro diabaikan.

### Hak akses faskes

Faskes punya hak atas anak dan orangtua yang ia buat sendiri, dan atas pasien
yang sedang menjadi tanggung jawabnya:

| kondisi anak | yang memegang hak |
|---|---|
| sudah pernah dirujuk | faskes tujuan rujukan terakhir (`rujukan` atau `rujuk_balik`) |
| belum pernah dirujuk | puskesmas yang terikat dengan posyandu keluarganya; kalau posyandu itu tidak terikat puskesmas, puskesmas/pustu sekecamatan dengan kelurahan posyandu. RS yang sekabupaten/kota dengan posyandu juga memegang hak |

Hak ini dipakai untuk baca, `PUT /api/faskes/anak` dan `/orangtua`, `POST
/api/faskes/anak` (atas `id_orangtua`), dan pemeriksaan baru. Rujukan
memindahkan hak ke tujuannya: faskes yang merujuk, termasuk puskesmas posyandu,
kehilangan hak sampai pasien dirujuk balik kepadanya. Pembuat data tidak pernah
kehilangan hak. Rujukan `internal` dan rujukan tanpa tujuan tidak memindahkan
hak.

Orangtua mengikuti anaknya: faskes memegang hak atas orangtua kalau memegang
hak atas salah satu anaknya. Orangtua yang belum punya anak dipegang faskes
posyandunya.

Kunjungan, kesehatan, dan data medis tetap hanya bisa diubah pembuatnya. Faskes
tidak pernah menghapus identitas.

---

## Asal data dan kolom audit

Setiap orangtua, anak, kunjungan, dan kesehatan mencatat pembuatnya di
`created_by` dan jalur masuknya di `source_data`. Keduanya diisi sistem dan
tidak pernah berubah; nilai kiriman diabaikan.

| jalur | `created_by` | `source_data` |
|---|---|---|
| jakantro | `jakantro` | `jakantro:<source_data kiriman>`, atau `jakantro` kalau tidak dikirim |
| faskes | orgid faskes | `faskes:<orgid>` |
| SatuSehat | anak dan orangtua: `satusehat`; kunjungan: orgid faskes pengirim Encounter (`satusehat` kalau tidak ada) | `satusehat` |

`source_data` tersimpan paling panjang 50 karakter, jadi kiriman jakantro paling
panjang 41 karakter (UUID muat); lebih dari itu `422` `1002`. Data medis tidak
punya `created_by`; pemiliknya mengikuti kunjungan.

`updated_by` mencatat siapa yang terakhir mengubah:

| jalur | diisi saat |
|---|---|
| jakantro | `jakantro`, di PUT orangtua, anak, kunjungan, dan kesehatan |
| faskes | orgid, di `PUT /api/faskes/orangtua`, `/anak`, dan `/kunjungan` |
| SatuSehat | `satusehat`, saat Patient atau RelatedPerson menimpa identitas |

Tidak mengubah `updated_by`: PUT Encounter dari SatuSehat (hanya tanggal),
`anak_ke` dari Observation, penautan anak ke orangtua, kolom ukur kunjungan dari
observasi, dan pemulihan orangtua yang terhapus. Data medis tidak punya
`updated_by`; yang berubah hanya `updated_at`.

`deleted_by` diisi `jakantro` saat jakantro menghapus orangtua, anak,
kunjungan, atau kesehatan. Faskes dan SatuSehat tidak menghapus. Mendaftarkan
lagi NIK orangtua yang terhapus memulihkan barisnya dan mengosongkan
`deleted_at` serta `deleted_by`.

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

