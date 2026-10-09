# Jalur jakantro

> **Catatan.** Bagian ini untuk posyandu. Kalau Anda puskesmas atau RS, yang
> berlaku adalah [Jalur faskes](?doc=faskes).

Semua endpoint di halaman ini mewajibkan Anda mengirim `id` sendiri — lihat
[aturan `id`](?doc=index#aturan-id). Contoh payload siap salin ada di
[Contoh Payload](?doc=contoh).

## Field per entitas

Daftar ini lengkap: field yang tidak tercantum tidak dikenali dan diabaikan.
Kolom **wajib** berlaku untuk POST. Aturan PUT ada di bagian endpoint
masing-masing.

Kolom yang diisi sistem — `created_at`, `updated_at`, `deleted_at`,
`updated_by`, `deleted_by` — diabaikan kalau dikirim, di semua endpoint.

### `orangtua`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id` | teks, maks 36 | ya | id buatan jakantro. Kunci pencarian di `PUT /api/orangtua` |
| `id_posyandu` | teks, maks 36 | ya | posyandu tempat keluarga terdaftar. Menentukan puskesmas yang berhak mengakses keluarga ini |
| `no_kk` | 16 digit angka | ya | nomor kartu keluarga, unik. Bentrok dijawab `409` `3003` |
| `nik` | 16 digit angka | ya | NIK orangtua. POST dengan NIK yang sudah terdaftar mengembalikan data yang ada, bukan error |
| `nama_ayah`, `nama_ibu` | teks, maks 255 | ya | |
| `telepon` | teks, maks 255 | ya | |
| `rt`, `rw` | teks, 3 karakter | ya | dikembalikan dengan padding spasi sampai 3 karakter |
| `alamat` | teks | ya | |
| `kia` | `0`/`1` | — | keluarga memiliki buku KIA. Kosong berarti `0` |
| `usia_hamil` | bilangan bulat ≥ 0 | — | usia kehamilan ibu, kalau sedang hamil |
| `kia_bayi_kecil` | `0`/`1` | — | memiliki buku KIA bayi kecil |
| `source_data` | teks, maks 36 | — | penanda sumber data. Kosong memakai default database. Tidak bisa diubah setelah dibuat |
| `satusehat_id` | — | — | **diabaikan** dari jakantro; diisi jalur SatuSehat |

### `anak`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id` | teks, maks 36 | ya | id buatan jakantro. Kunci pencarian di `PUT /api/anak` |
| `id_orangtua` | teks, maks 36 | ya | `id` orangtua yang sudah ada atau dikirim di payload yang sama |
| `nama` | teks, maks 255 | ya | |
| `nik` | 16 digit angka | — | boleh kosong, lihat [NIK anak boleh kosong](#nik-anak-boleh-kosong) |
| `tanggal_lahir` | `YYYY-MM-DD` | ya | |
| `jenis_kelamin` | `L`/`P` | ya | |
| `anak_ke` | bilangan bulat > 0 | ya | urutan kelahiran dalam keluarga |
| `imd` | `0`/`1` | — | mendapat inisiasi menyusu dini. Kosong berarti `0` |
| `bb_lahir` | angka > 0 | ya | berat lahir, kg |
| `tb_lahir` | angka > 0 | ya | panjang lahir, cm |
| `lk_lahir` | angka > 0 | ya | lingkar kepala saat lahir, cm |
| `source_data` | teks, maks 36 | ya | penanda sumber data, mis. `jakantro`. Tidak bisa diubah setelah dibuat |
| `status_aktif` | teks, maks 20 | — | bawaan `aktif`. Disimpan apa adanya, tidak dipakai logika service |
| `satusehat_id` | — | — | **diabaikan** dari jakantro; diisi jalur SatuSehat |

### `kunjungan`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id` | teks, maks 36 | ya | id buatan jakantro, mis. id pengukuran di sistem Anda |
| `id_anak` | teks, maks 36 | ya | `id` anak yang sudah ada atau dikirim di payload yang sama |
| `tanggal_pengukuran` | `YYYY-MM-DD` | ya | tanggal penimbangan atau pengukuran |
| `tanggal_selesai` | `YYYY-MM-DD` | — | akhir kunjungan. Untuk posyandu biasanya dikosongkan |
| `cara_ukur` | teks | — | `telentang` atau `berdiri` |
| `berat_badan` | angka > 0 | — | kg |
| `tinggi_badan` | angka > 0 | — | panjang atau tinggi badan, cm |
| `lingkar_lengan` | angka > 0 | — | lingkar lengan atas (LILA), cm |
| `lingkar_kepala` | angka > 0 | — | cm |
| `lingkar_dada` | angka > 0 | — | cm |
| `asi_bulan_0` … `asi_bulan_6` | `0`/`1` | — | anak mendapat ASI eksklusif pada bulan ke-0 sampai ke-6 |
| `vit_biru` | `0`/`1` | — | mendapat kapsul vitamin A biru (usia 6–11 bulan) |
| `vit_merah` | `0`/`1` | — | mendapat kapsul vitamin A merah (usia 12–59 bulan) |
| `pitting_edema` | `0`/`1` | — | ditemukan edema pitting |
| `kelas_ibu_balita` | `0`/`1` | — | ibu mengikuti kelas ibu balita |
| `status_bbu`, `status_tbu`, `status_bbtb` | teks | — | status gizi BB/U, TB/U, BB/TB hasil hitungan Anda, mis. `Sangat Pendek`. Labelnya mengikuti [tabel status gizi](?doc=faskes#z-score-dan-status-gizi) |
| `zscore_bbu`, `zscore_tbu`, `zscore_bbtb` | angka | — | z-score pasangan status di atas, disimpan 2 desimal |
| `source_data` | teks, maks 36 | — | penanda sumber data. Tidak bisa diubah setelah dibuat |

`id_satusehat`, `id_faskes`, `stunting`, `ref_episode`, `ref_rujukan`,
`id_episode`, dan `id_rujukan` milik jalur faskes dan **diabaikan** kalau
dikirim jakantro. Kunjungan dari jakantro selalu tetap milik jakantro.

### `kesehatan`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id` | teks, maks 36 | ya | id buatan jakantro |
| `id_anak` | teks, maks 36 | ya | `id` anak yang sudah ada atau dikirim di payload yang sama |
| `tanggal_pemantauan` | `YYYY-MM-DD` | ya | |
| `tbc_batuk` | `0`/`1` | — | skrining TBC: ada gejala batuk |
| `tbc_demam` | `0`/`1` | — | skrining TBC: ada gejala demam |
| `tbc_bb` | `0`/`1` | — | skrining TBC: berat badan turun atau tidak naik |
| `tbc_kontak` | `0`/`1` | — | skrining TBC: kontak dengan pasien TBC |
| `layanan_asi_eks` | `0`/`1` | — | mendapat layanan ASI eksklusif |
| `layanan_mpasi` | `0`/`1` | — | mendapat layanan MP-ASI |
| `layanan_imunisasi` | `0`/`1` | — | mendapat imunisasi |
| `layanan_vit_a` | `0`/`1` | — | mendapat vitamin A |
| `layanan_obat_cacing` | `0`/`1` | — | mendapat obat cacing |
| `layanan_mt_pangan` | `0`/`1` | — | mendapat makanan tambahan pangan lokal |
| `penyuluhan_edukasi` | `0`/`1` | — | orangtua mendapat penyuluhan atau edukasi |
| `penyuluhan_rujukan` | `0`/`1` | — | anak dirujuk |
| `source_data` | teks, maks 36 | — | penanda sumber data. Tidak bisa diubah setelah dibuat |

---|---|---|
| `orangtua` | `id`, `id_posyandu`, `no_kk`, `nik`, `nama_ayah`, `nama_ibu`, `telepon`, `rt`, `rw`, `alamat`, `kia` | `no_kk` dan `nik` 16 digit angka. `kia` hanya `0`/`1`. `usia_hamil` tidak boleh negatif, `kia_bayi_kecil` hanya `0`/`1` |
| `anak` | `id`, `id_orangtua`, `nama`, `tanggal_lahir`, `jenis_kelamin`, `anak_ke`, `imd`, `bb_lahir`, `tb_lahir`, `lk_lahir`, `source_data` | `jenis_kelamin` hanya `L`/`P`. `anak_ke`, `bb_lahir`, `tb_lahir`, `lk_lahir` harus > 0. `imd` hanya `0`/`1`. `nik` opsional, tapi kalau dikirim harus 16 digit |
| `kunjungan` | `id`, `id_anak`, `tanggal_pengukuran` | field ukuran opsional, tapi kalau dikirim harus > 0. `asi_*`, `vit_*`, `pitting_edema`, `kelas_ibu_balita` hanya `0`/`1` |
| `kesehatan` | `id`, `id_anak`, `tanggal_pemantauan` | `tbc_*`, `layanan_*`, `penyuluhan_*` opsional, hanya `0`/`1` |

---

## POST /api/kunjungan

Endpoint utama. Menerima tiga bentuk payload — service menyimpulkan sendiri
entitas mana yang perlu dibuat dari key yang Anda kirim.
Isi key `orangtua`, `anak`, dan `kunjungan` mengikuti
[Field per entitas](#field-per-entitas).

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

Semuanya `422` dengan `errorCode` `1003`:

| yang salah | pesan |
|---|---|
| Bentuk flat disertai key `orangtua`/`anak`/`kunjungan` | `bentuk flat hanya untuk kunjungan saja` |
| Tidak ada key `kunjungan` sama sekali | `data kunjungan wajib dikirim` |
| `anak.id` ≠ `kunjungan.id_anak` | `anak.id (...) tidak sama dengan kunjungan.id_anak (...)` |
| `orangtua.id` ≠ `anak.id_orangtua` | `orangtua.id (...) tidak sama dengan anak.id_orangtua (...)` |
| Key `orangtua` tanpa key `anak` | `key orangtua hanya boleh dikirim bersama key anak` |
| Orangtua baru (bawa `nik`/`no_kk`) tanpa data anak baru | `orangtua baru harus disertai data anak baru` |

`id_anak` yang tidak ada di database kena `422` `1005 REFERENCE_NOT_FOUND` —
itu foreign key, bukan validasi payload.

### NIK anak boleh kosong

`anak.nik` boleh string kosong, berisi spasi, atau key-nya dihilangkan.
Ketiganya disimpan `NULL` dan tidak bentrok dengan unique constraint. NIK
orangtua tidak boleh kosong.

### PUT /api/kunjungan/:id

Memperbarui kunjungan yang sudah tersimpan. `id` pada path adalah id yang Anda
tentukan sendiri saat membuatnya.

```json
{
  "id": "c1e48a72-9d35-4b80-a6f3-52d7e9418b04",
  "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "tanggal_pengukuran": "2026-08-14",
  "berat_badan": 9.8
}
```

`id`, `id_anak`, dan `tanggal_pengukuran` wajib, sama seperti POST. Field
lain mengikuti [tabel `kunjungan`](#kunjungan). Field yang tidak dikirim
**tidak dihapus** — nilainya yang sekarang dipertahankan. Jadi mengirim
`berat_badan` saja tidak akan mengosongkan tinggi badan.

> **Awas.** Kunjungan yang dibuat faskes tidak boleh diubah jakantro.
> Percobaannya dijawab `403` `1008` dengan menyebut faskes pembuatnya.

`id`, `id_anak`, dan `source_data` tidak ikut berubah meski dikirim.

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

### PUT /api/kesehatan/:id

Bentuknya sama dengan `PUT /api/kunjungan/:id`: `id`, `id_anak`, dan
`tanggal_pemantauan` wajib, field yang tidak dikirim dipertahankan, dan `id`,
`id_anak`, serta `source_data` tidak berubah.

---

## POST /api/orangtua

Satu orangtua, tanpa nested. Field-nya ada di [tabel `orangtua`](#orangtua).

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

Satu anak pada orangtua yang sudah ada. Field-nya ada di [tabel `anak`](#anak).

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

### PUT /api/anak

Memperbarui anak yang sudah tersimpan. Anak dicari lewat `id` di body; hanya
`id` yang wajib. Field yang tidak dikirim dipertahankan, field yang dikirim
diperiksa formatnya sama seperti `POST /api/anak`.

```json
{
  "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "nama": "Rizky Ramadhan Putra",
  "nik": "3175041503240003"
}
```

`200` mengembalikan anak hasil pembaruan, `404` `2001` kalau anaknya tidak ada.
Field lain mengikuti [tabel `anak`](#anak). `satusehat_id` dan `source_data`
diabaikan — nilai dari SatuSehat tetap dipertahankan. Mengirim `id_orangtua`
lain memindahkan anak ke orangtua itu. `imd` boleh diubah ke `0`.

### PUT /api/orangtua

Sama dengan `PUT /api/anak`: dicari lewat `id`, hanya `id` yang wajib, field yang
tidak dikirim dipertahankan. Field lain mengikuti [tabel `orangtua`](#orangtua);
`satusehat_id` dan `source_data` diabaikan, `kia` boleh diubah ke `0`.

```json
{
  "id": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
  "telepon": "081299990000",
  "alamat": "Jl. Kenanga No. 14"
}
```

Kedua endpoint ini juga dipakai faskes, dengan batasan yang dijelaskan di
[Jalur faskes](?doc=faskes#put-apianak-dan-put-apiorangtua).
