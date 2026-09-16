# Jalur jakantro

> **Catatan.** Bagian ini untuk posyandu. Kalau Anda puskesmas atau RS, yang
> berlaku adalah [Jalur faskes](?doc=faskes).

## POST /api/kunjungan

Endpoint utama. Menerima tiga bentuk payload — service menyimpulkan sendiri
entitas mana yang perlu dibuat dari key yang Anda kirim.

### Bentuk 1 — orangtua + anak + kunjungan sekaligus

Untuk anak yang keluarganya belum pernah dikirim. Ketiganya masuk dalam satu
transaction; kalau ada yang gagal, tidak ada yang tersimpan.

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

`201 Created`. Field `orangtua` dan `anak` ikut terisi karena keduanya memang
baru dibuat:

```json
{
  "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
  "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "orangtua": {
    "id": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
    "id_satusehat": null,
    "id_posyandu": "POS-JAKARTA-01",
    "no_kk": "3175042509870001",
    "nama_ayah": "Ahmad Suryana",
    "nama_ibu": "Dewi Lestari",
    "nik": "3175040101900002",
    "telepon": "081234567890",
    "rt": "004",
    "rw": "007",
    "alamat": "Jl. Kenanga No. 12",
    "kia": 1,
    "created_at": "2026-08-28T04:04:51.489267Z",
    "updated_at": null,
    "deleted_at": null,
    "source_data": "20f11c5c-5a29-11f0-a136-5749ce66d036",
    "usia_hamil": null,
    "kia_bayi_kecil": null
  },
  "anak": {
    "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
    "nama": "Rizky Ramadhan",
    "nik": "3175041503240003",
    "tanggal_lahir": "2024-03-15",
    "jenis_kelamin": "L",
    "anak_ke": 1,
    "status_aktif": "aktif",
    "created_at": "2026-08-28T04:04:51.489267Z"
  },
  "kunjungan": {
    "id": "c1e48a72-9d35-4b80-a6f3-52d7e9418b04",
    "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "tanggal_pengukuran": "2026-08-14",
    "cara_ukur": "telentang",
    "berat_badan": 9.4,
    "tinggi_badan": 76.2,
    "lingkar_lengan": 13.5,
    "lingkar_kepala": 45.1,
    "created_at": "2026-08-28T04:04:51.489267Z"
  }
}
```

Response asli memuat semua kolom termasuk yang `null`; contoh di atas dipangkas
supaya terbaca.

### Bentuk 2 — anak + kunjungan

Orangtua sudah ada. Key `orangtua` boleh disertakan berisi `id` saja, boleh juga
dihilangkan sama sekali.

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

Response `201`, dengan `orangtua: null` dan `id_orangtua: null` karena tidak ada
orangtua yang dibuat.

### Bentuk 3 — kunjungan saja

Anak sudah ada. Boleh nested:

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

Boleh juga flat di root — berguna kalau Anda kirim per-record dari queue:

```json
{
  "id": "a6f18c93-2b47-4e50-8d31-9c07a5e2f8b6",
  "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "tanggal_pengukuran": "2026-09-11",
  "berat_badan": 9.8
}
```

Keduanya `201` dengan `orangtua` dan `anak` bernilai `null`.

### Kombinasi yang ditolak

Semuanya `400` dengan `errorCode` `1003`:

| yang salah | pesan |
|---|---|
| Bentuk flat disertai key `orangtua`/`anak`/`kunjungan` | `bentuk flat hanya untuk kunjungan saja` |
| Tidak ada key `kunjungan` sama sekali | `data kunjungan wajib dikirim` |
| `anak.id` ≠ `kunjungan.id_anak` | `anak.id (...) tidak sama dengan kunjungan.id_anak (...)` |
| `orangtua.id` ≠ `anak.id_orangtua` | `orangtua.id (...) tidak sama dengan anak.id_orangtua (...)` |
| Key `orangtua` tanpa key `anak` | `key orangtua hanya boleh dikirim bersama key anak` |
| Orangtua baru (bawa `nik`/`no_kk`) tanpa data anak baru | `orangtua baru harus disertai data anak baru` |

`id_anak` yang tidak ada di database kena `400` `1005 REFERENCE_NOT_FOUND` —
itu foreign key, bukan validasi payload.

### NIK anak boleh kosong

`anak.nik` boleh string kosong, berisi spasi, atau key-nya dihilangkan.
Ketiganya disimpan `NULL` dan tidak bentrok dengan unique constraint. NIK
orangtua tidak boleh kosong.

---

## POST /api/kesehatan

Aturan payloadnya identik dengan `/api/kunjungan` — tiga bentuk yang sama, cross
check id yang sama, pesan error yang sama. Bedanya cuma nama key (`kesehatan`)
dan nama field tanggal (`tanggal_pemantauan`).

Contoh bentuk flat:

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

`201 Created`:

```json
{
  "id_orangtua": null,
  "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "orangtua": null,
  "anak": null,
  "kesehatan": {
    "id": "d5b3f169-8c47-42ae-b91d-6f0a3e28c7d5",
    "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "tanggal_pemantauan": "2026-08-14",
    "tbc_batuk": 0,
    "tbc_demam": 0,
    "tbc_bb": 0,
    "tbc_kontak": 0,
    "layanan_asi_eks": 1,
    "layanan_mpasi": null,
    "layanan_imunisasi": 1,
    "layanan_vit_a": 1,
    "layanan_obat_cacing": null,
    "layanan_mt_pangan": null,
    "penyuluhan_edukasi": 1,
    "penyuluhan_rujukan": null,
    "created_at": "2026-08-28T04:04:51.501846Z",
    "updated_at": null,
    "deleted_at": null,
    "source_data": "20f11c5c-5a29-11f0-a136-5749ce66d036"
  }
}
```

Field `tbc_*`, `layanan_*`, dan `penyuluhan_*` semuanya smallint nullable —
dipakai sebagai flag `0`/`1`.

---

## POST /api/orangtua

Satu orangtua, tanpa nested. Semua field kecuali `source_data`, `usia_hamil`,
dan `kia_bayi_kecil` wajib diisi.

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

`201` mengembalikan object orangtua. Kalau `nik` sudah terdaftar, service
mengembalikan data yang sudah ada — bukan error — jadi endpoint ini idempotent
terhadap NIK.

`no_kk` harus 16 digit angka dan unik. Bentrok `no_kk` kena `409` `3003`.

---

## POST /api/anak

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

`201` mengembalikan object anak. Kalau anak dengan NIK yang sama sudah terdaftar
di bawah `id_orangtua` berbeda, kena `409` `3006` — itu tanda dua sumber data
tidak sepakat anak ini milik siapa, jangan di-retry.

---

---
