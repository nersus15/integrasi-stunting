# Contoh Payload

Payload lengkap siap salin untuk setiap endpoint POST. Aturan tiap field dan
bentuk responsnya ada di halaman [Jalur jakantro](?doc=jakantro) dan
[Jalur faskes](?doc=faskes) — halaman ini sengaja hanya berisi payload.

Semua contoh memakai id yang sama supaya bisa dijalankan berurutan sebagai satu
alur: orangtua `8f3a1c40…`, anak `b7d9e254…`.

---

# Jalur jakantro

## POST /api/kunjungan — bentuk 1

Orangtua + anak + kunjungan sekaligus, untuk keluarga yang belum pernah dikirim.

```json
{
  "orangtua": {
    "id": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
    "id_posyandu": "POS-JAKARTA-01",
    "no_kk": "3175042509870001",
    "nik": "3175040101900002",
    "nama_ayah": "Ahmad Suryana",
    "nama_ibu": "Dewi Lestari",
    "telepon": "081234567890",
    "rt": "004",
    "rw": "007",
    "alamat": "Jl. Kenanga No. 12",
    "kia": 1
  },
  "anak": {
    "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
    "nik": "3175041503240003",
    "nama": "Rizky Ramadhan",
    "tanggal_lahir": "2024-03-15",
    "jenis_kelamin": "L",
    "anak_ke": 1,
    "imd": 1,
    "bb_lahir": 3.2,
    "tb_lahir": 49.5,
    "lk_lahir": 34.0,
    "source_data": "jakantro"
  },
  "kunjungan": {
    "id": "c1e48a72-9d35-4b80-a6f3-52d7e9418b04",
    "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "tanggal_pengukuran": "2026-08-14",
    "cara_ukur": "telentang",
    "berat_badan": 9.4,
    "tinggi_badan": 76.2,
    "lingkar_lengan": 13.5,
    "lingkar_kepala": 45.1
  }
}
```

## POST /api/kunjungan — bentuk 2

Anak + kunjungan, untuk anak baru pada orangtua yang sudah ada.

```json
{
  "anak": {
    "id": "e2c74a18-5d93-4f60-b287-3ca9f1e05d47",
    "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
    "nik": "3175041503240004",
    "nama": "Aisyah Nur",
    "tanggal_lahir": "2025-01-20",
    "jenis_kelamin": "P",
    "anak_ke": 2,
    "imd": 1,
    "bb_lahir": 2.9,
    "tb_lahir": 47.0,
    "lk_lahir": 33.0,
    "source_data": "jakantro"
  },
  "kunjungan": {
    "id": "f4a92b57-8e31-4d06-9c85-7b1e2f6a3d90",
    "id_anak": "e2c74a18-5d93-4f60-b287-3ca9f1e05d47",
    "tanggal_pengukuran": "2026-08-14",
    "berat_badan": 7.1
  }
}
```

## POST /api/kunjungan — bentuk 3

Kunjungan saja, bentuk flat tanpa key pembungkus.

```json
{
  "kunjungan": {
    "id": "a6f18c93-2b47-4e50-8d31-9c07a5e2f8b6",
    "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "tanggal_pengukuran": "2026-09-11",
    "berat_badan": 9.8
  }
}
```

## POST /api/kesehatan

Bentuk flat. Tiga bentuk payloadnya sama dengan `/api/kunjungan`, bedanya hanya nama key dan field tanggal.

```json
{
  "id": "d5b3f169-8c47-42ae-b91d-6f0a3e28c7d5",
  "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "tanggal_pemantauan": "2026-08-14",
  "tbc_batuk": 0,
  "tbc_demam": 0,
  "tbc_bb": 0,
  "tbc_kontak": 0,
  "layanan_asi_eks": 1,
  "layanan_imunisasi": 1,
  "layanan_vit_a": 1,
  "penyuluhan_edukasi": 1
}
```

## POST /api/orangtua

Membuat orangtua tanpa anak maupun kunjungan.

```json
{
  "id": "3d6b8f21-7a45-4c92-b0e8-1f5a9c73d264",
  "id_posyandu": "POS-JAKARTA-01",
  "no_kk": "3175042509870009",
  "nik": "3175040101900011",
  "nama_ayah": "Bambang Wijaya",
  "nama_ibu": "Sri Handayani",
  "telepon": "081298765432",
  "rt": "002",
  "rw": "005",
  "alamat": "Jl. Melati No. 8",
  "kia": 1
}
```

## POST /api/anak

Membuat anak pada orangtua yang sudah ada.

```json
{
  "id": "9b4e7c02-1f83-4a56-8d97-6e2b0f5a3c81",
  "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
  "nik": "3175041503240007",
  "nama": "Fajar Pratama",
  "tanggal_lahir": "2025-06-02",
  "jenis_kelamin": "L",
  "anak_ke": 3,
  "imd": 1,
  "bb_lahir": 3.0,
  "tb_lahir": 48.0,
  "lk_lahir": 33.5,
  "source_data": "jakantro"
}
```

---

# Jalur faskes

## POST /api/faskes/pemeriksaan

Satu kunjungan beserta seluruh data medisnya. Kirim setelah data Anda diterima
SatuSehat, supaya `id_satusehat` tiap resource bisa disertakan — itu yang
membuat kiriman ulang tidak menggandakan.

`id` tidak dikirim: faskes tidak menentukan id, sistem yang membangkitkannya.

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

