# Jalur faskes

Bagian ini untuk puskesmas dan rumah sakit — jalur masuk langsung, alternatif
dari SatuSehat → IL → Kafka. Dipakai faskes yang tidak ingin melewatkan datanya
lewat IL.

Kedua jalur boleh dipakai bersamaan. `id_satusehat` yang membuat datanya
menyatu, bukan berganda.

## POST /api/faskes/pemeriksaan

Satu kunjungan beserta seluruh data medisnya, dalam satu transaksi.

> **Penting.** Kirim **setelah** data Anda diterima SatuSehat. Saat itulah IHS
> id tiap resource sudah Anda pegang dan bisa disertakan — dan itulah yang
> membuat kiriman ini menyatu dengan data yang datang lewat Kafka.

> **Awas.** Faskes pengirim diambil dari **API key**, bukan dari body. Key Anda
> harus milik user ber-group `orgid:<satusehat id faskes>`. `kunjungan.id_faskes`
> yang Anda kirim diabaikan, dan key jakantro ditolak `403` `1008`.

Kalau anaknya sudah terdaftar, Anda harus punya
[hak akses](#hak-akses-anak-dan-orangtua) atas anak itu; kalau tidak, dijawab
`403` `1008`. Anak yang belum terdaftar dibuat baru tanpa pemeriksaan ini.

Encounter yang dirujuk juga harus milik Anda, dan pelanggarannya dijawab `403`
`1008`:

- `kunjungan.id_satusehat` yang sudah tersimpan harus kunjungan faskes Anda.
  Yang belum tersimpan dibuat baru.
- `ref_encounter` di observasi, diagnosa, layanan, dan rujukan yang menunjuk
  encounter lain harus menunjuk encounter faskes Anda yang sudah tersimpan.

### Bentuk payload

| key | wajib | keterangan |
|---|---|---|
| `anak` | ya | `nik` atau `id_satusehat` salah satu. Kalau anaknya belum ada, `nama`, `tanggal_lahir`, dan `jenis_kelamin` juga perlu supaya bisa dibuat |
| `kunjungan` | ya | `id_satusehat` dan `tanggal_pengukuran` wajib. `id` dan `id_faskes` diabaikan |
| `observasi` | — | `system` dan `kode` wajib. Tanpa `system`, kodenya tidak bisa dipetakan ke kolom kunjungan |
| `diagnosa` | — | `id_satusehat` dan `kode` wajib. `jenis` `diagnosis` (bawaan) atau `alergi` |
| `layanan` | — | `jenis` wajib |
| `rujukan` | — | `jenis` boleh dikosongkan, arahnya disimpulkan dari faskes asal dan tujuan |
| `episode` | — | disambung ke kunjungan lewat `kunjungan.ref_episode` |

Penjelasan setiap key beserta payload lengkap yang sudah diuji ada di
[Contoh Payload](?doc=contoh#jalur-faskes).

### Response

`201` — kedua `id` dibangkitkan sistem:

```json
{
  "id_anak": "a1b2c3d4-2222-4aaa-8bbb-000000000002",
  "id_kunjungan": "c68e7772-f6dd-4a54-b17e-e99a6f4c42fc"
}
```

`201` selalu berarti tersimpan. Kalau datanya tidak memenuhi kriteria pemantauan,
jawabannya `422` `6001`:

```json
{
  "httpCode": 422,
  "errorCode": 6001,
  "errorName": "TIDAK_DISIMPAN",
  "message": "tidak ada tanda stunting dan riwayat terakhir bukan stunting"
}
```

Yang memicunya: anaknya sudah lewat 5 tahun, atau pemeriksaan ini tidak membawa
tanda stunting sementara anaknya juga belum pernah berstatus stunting. Bukan
kesalahan payload — jangan retry. Penolakan tidak meninggalkan apa pun di
database.

## PUT /api/faskes/:resourceName

Memperbarui satu resource yang sudah tersimpan. `:resourceName` salah satu dari
`kunjungan`, `observasi`, `diagnosa`, `layanan`, `rujukan`, `episode`.

```json
PUT /api/faskes/diagnosa

{
  "id_satusehat": "cond-pkm-01",
  "kode": "E45",
  "display": "Nutritional stunting (revisi)"
}
```

`id_satusehat` **wajib** — itu satu-satunya kunci pencarian barisnya.

Field yang tidak dikirim dipertahankan. Untuk `kunjungan`, ini penting: PUT
Encounter hanya membawa periode kunjungan, sehingga berat dan tinggi badan yang
berasal dari Observation tidak ikut terhapus.

| jawaban | arti |
|---|---|
| `200` | tersimpan; body berisi baris hasil penggabungan |
| `403` `1008` | resource itu milik faskes lain |
| `404` `2001` | `:resourceName` tidak dikenal, atau barisnya belum pernah tersimpan |
| `422` `1002` | `id_satusehat` tidak dikirim, atau isinya tidak memenuhi aturan |

Anda hanya bisa memperbarui resource yang faskes-nya sama dengan pemilik API key.

### Field yang tidak bisa diubah

Field berikut diabaikan kalau dikirim. Semuanya menentukan milik siapa data itu,
di mana posisinya, atau dihitung sistem — bukan isi klinis.

| resource | field |
|---|---|
| semua | `id`, `id_satusehat`, `id_anak`, `created_at`, `updated_at` |
| `kunjungan` | `id_faskes`, `id_episode`, `ref_episode`, `id_rujukan`, `ref_rujukan`, `stunting` |
| `observasi` | `id_kunjungan`, `id_induk`, `ref_encounter` |
| `diagnosa` | `id_kunjungan`, `ref_encounter`, `jenis`, `tanggal_catat` |
| `layanan` | `id_kunjungan`, `ref_encounter`, `jenis` |
| `rujukan` | `id_kunjungan`, `ref_encounter`, `jenis`, `tanggal`, `id_faskes_asal`, `ref_faskes_asal`, `id_faskes_tujuan`, `ref_faskes_tujuan` |
| `episode` | `id_faskes` |

`jenis` dan `tanggal` rujukan dikunci karena [hak akses](#hak-akses-anak-dan-orangtua)
bergantung pada urutan rujukan dan rujuk balik. Kalau salah satunya keliru,
perbaiki di SatuSehat — perubahan yang datang lewat Kafka tetap diterima.

---

### Kenapa `id_satusehat` wajib

Kolom `satusehat_id` unik di tabel kunjungan, observasi, diagnosa, layanan, dan
rujukan. Itulah kunci pencocokan yang menggantikan `id` — yang di jalur jakantro
harus ditentukan pengirim, tapi di sini dibangkitkan sistem.

Konsekuensinya:

- kiriman ulang tidak menggandakan data. Payload yang sama persis dikirim
  berkali-kali tetap dijawab `201`, jumlah barisnya tidak bertambah
- data yang sama dari Kafka dan dari endpoint ini menyatu ke baris yang sama
- rujukan dan kunjungan yang memenuhinya tetap bisa saling tertaut

### Yang tidak berbeda dari jalur SatuSehat

Penentuan stunting, pengangkatan observasi ke kolom kunjungan, dan penyambungan
rujukan memakai jalan yang sama persis — tidak ada logika terpisah untuk jalur
ini. Hasil bacanya pun identik; lihat [endpoint GET](?doc=baca).

---

## PUT /api/anak dan PUT /api/orangtua

Faskes memakai endpoint yang sama dengan jakantro — lihat
[PUT /api/anak](?doc=jakantro#put-apianak) dan
[PUT /api/orangtua](?doc=jakantro#put-apiorangtua). Hanya `id` yang wajib,
field yang tidak dikirim dipertahankan.

Data identitas dari faskes dianggap lebih terpercaya, jadi `nik` dan `no_kk`
boleh Anda perbaiki. Dua field diabaikan kalau dikirim faskes, karena keduanya
menentukan siapa yang berhak atas data itu:

| field | endpoint | alasan |
|---|---|---|
| `id_posyandu` | `PUT /api/orangtua` | memindah keluarga ke wilayah lain |
| `id_orangtua` | `PUT /api/anak` | memindah anak ke keluarga lain |

## Hak akses anak dan orangtua

Berlaku untuk `GET /api/anak/:id`, `GET /api/orangtua`, `PUT /api/anak`,
`PUT /api/orangtua`, dan `POST /api/faskes/pemeriksaan` untuk anak yang sudah
terdaftar. Akses diberikan kalau salah satu terpenuhi:

| pemanggil | syarat |
|---|---|
| jakantro | selalu |
| puskesmas dan pustu | posyandu keluarga itu ada di wilayah kerja Anda |
| faskes tujuan rujukan (umumnya RS) | ada rujukan untuk anak itu dengan tujuan faskes Anda, dan belum ada rujuk balik dari faskes Anda sesudahnya |
| faskes yang mencatat kunjungan | hanya untuk anak yang belum terhubung ke orangtua |

Untuk orangtua, rujukan salah satu anaknya sudah cukup. Rujuk balik mencabut
akses; rujukan baru sesudahnya memberikannya lagi. Rujukan `internal` tidak
memberi akses.

