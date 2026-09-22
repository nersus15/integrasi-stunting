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

### Bentuk payload

| key | wajib | keterangan |
|---|---|---|
| `anak` | ya | `nik` atau `id_satusehat` salah satu. Kalau anaknya belum ada, `nama`, `tanggal_lahir`, dan `jenis_kelamin` juga perlu supaya bisa dibuat |
| `kunjungan` | ya | `id_satusehat` dan `tanggal_pengukuran` wajib. `id` dan `id_faskes` diabaikan |
| `observasi` | — | `system` dan `kode` wajib. Tanpa `system`, kodenya tidak bisa dipetakan ke kolom kunjungan |
| `diagnosa` | — | `kode` wajib. `jenis` `diagnosis` (bawaan) atau `alergi` |
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
