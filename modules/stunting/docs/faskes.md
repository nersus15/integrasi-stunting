# Jalur faskes

Bagian ini untuk puskesmas dan RS. API key Anda harus ber-group
`orgid:<satusehat id faskes>`; key jakantro ditolak `403`.

## POST /api/faskes/pemeriksaan

Jalur masuk langsung untuk puskesmas dan RS, alternatif dari jalur SatuSehat →
IL → Kafka. Dipakai faskes yang tidak ingin melewatkan datanya lewat IL.

> **Penting.** Kirim ini **setelah** data Anda diterima SatuSehat, supaya IHS id
> tiap resource sudah Anda pegang dan bisa disertakan. Itu yang membuat data dari
> kedua jalur menyatu, bukan berganda.

> **Awas.** Faskes pengirim diambil dari **API key**, bukan dari body. Key Anda
> harus milik user ber-group `orgid:<satusehat id faskes>`; `kunjungan.id_faskes`
> yang Anda kirim diabaikan, dan key jakantro ditolak `403`.

```json
{
  "anak": {
    "nik": "3175042003240007",
    "id_satusehat": "P99001100022",
    "nama": "Siti Aminah",
    "tanggal_lahir": "2024-05-20",
    "jenis_kelamin": "P"
  },
  "kunjungan": {
    "id_satusehat": "enc-langsung-01",
    "tanggal_pengukuran": "2026-04-10",
    "tanggal_selesai": "2026-04-10",
    "cara_ukur": "telentang",
    "berat_badan": 8.4,
    "tinggi_badan": 75.2
  },
  "observasi": [
    { "id_satusehat": "obs-langsung-bb", "system": "http://loinc.org", "kode": "29463-7",
      "nilai_angka": 8.4, "satuan": "kg", "interpretasi": "OI000007" }
  ],
  "diagnosa": [
    { "id_satusehat": "cond-01", "jenis": "diagnosis",
      "system": "http://hl7.org/fhir/sid/icd-10", "kode": "E45" }
  ],
  "layanan":  [ ],
  "rujukan":  [ ],
  "episode":  [ ]
}
```

| key | keterangan |
|---|---|
| `anak` | **`nik` atau `id_satusehat` wajib salah satu.** Kalau anaknya belum ada, `nama`, `tanggal_lahir`, dan `jenis_kelamin` juga wajib supaya bisa dibuat |
| `kunjungan` | **`id_satusehat` dan `tanggal_pengukuran` wajib.** `id` dan `id_faskes` diabaikan |
| `observasi` | `system` dan `kode` wajib — tanpa `system`, kodenya tidak bisa dipetakan ke kolom kunjungan |
| `diagnosa` | `kode` wajib. `jenis` `diagnosis` (bawaan) atau `alergi` |
| `layanan` | `jenis` wajib |
| `rujukan`, `episode` | opsional, bentuknya sama dengan yang muncul di summary |

Response `201` — `id` keduanya dibangkitkan sistem:

```json
{
  "id_anak": "a1b2c3d4-2222-4aaa-8bbb-000000000002",
  "id_kunjungan": "c68e7772-f6dd-4a54-b17e-e99a6f4c42fc"
}
```

`201` selalu berarti tersimpan. Kalau datanya tidak memenuhi kriteria pemantauan,
jawabannya `422` `6001 TIDAK_DISIMPAN`, bukan `201`:

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
database. Selengkapnya di [katalog error](?doc=errors).

Kriteria itu sama persis dengan jalur SatuSehat: penentuan stunting, pengangkatan
observasi ke kolom kunjungan, dan penyambungan rujukan memakai jalan yang sama —
tidak ada logika terpisah untuk jalur ini.

### Kenapa `id_satusehat` penting

Kolom `satusehat_id` unik di tabel kunjungan, observasi, diagnosa, layanan, dan
rujukan. Dengan IHS id disertakan:

- kiriman ulang tidak menggandakan data
- kalau data yang sama juga tiba lewat Kafka, keduanya menyatu ke baris yang sama
- rujukan dan kunjungan tujuannya tetap bisa saling tertaut

Berbeda dari jakantro, **faskes tidak menentukan `id` sendiri**. Jakantro butuh
itu karena tidak punya satusehat id sehingga tidak bisa mencocokkan data lama.
Faskes punya, jadi `id` dibangkitkan sistem dan pencocokan memakai
`id_satusehat` — karena itu ia wajib.

---
