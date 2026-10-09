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

Satu kunjungan beserta seluruh data medisnya. Payload di bawah sengaja dibuat
selengkap mungkin: setiap array terisi, setiap varian nilai observasi muncul,
kedua jenis diagnosa ada, keempat jenis layanan ada.

> **Penting.** Kirim setelah data Anda diterima SatuSehat, supaya `id_satusehat`
> tiap resource bisa disertakan — itu yang membuat kiriman ulang tidak
> menggandakan. `id` tidak dikirim: faskes tidak menentukan id, sistem yang
> membangkitkannya.

```json
{
 "anak": {
  "id_satusehat": "P99001100022",
  "nik": "3175042003240007",
  "nama": "Siti Aminah",
  "tanggal_lahir": "2024-05-20",
  "jenis_kelamin": "P",
  "anak_ke": 1
 },
 "kunjungan": {
  "id_satusehat": "enc-lengkap-01",
  "tanggal_pengukuran": "2026-05-12",
  "tanggal_selesai": "2026-05-15",
  "cara_ukur": "telentang",
  "berat_badan": 8.9,
  "tinggi_badan": 76.4,
  "lingkar_lengan": 12.1,
  "lingkar_kepala": 45.3,
  "lingkar_dada": 46.0,
  "asi_bulan_0": 1,
  "asi_bulan_1": 1,
  "asi_bulan_2": 1,
  "asi_bulan_3": 1,
  "asi_bulan_4": 1,
  "asi_bulan_5": 1,
  "asi_bulan_6": 1,
  "vit_biru": 0,
  "vit_merah": 1,
  "pitting_edema": 0,
  "kelas_ibu_balita": 1,
  "ref_episode": "ep-lengkap-01",
  "status_bbu": "Berat Badan Sangat Kurang",
  "zscore_bbu": -3.05,
  "status_bbtb": "Gizi Kurang",
  "zscore_bbtb": -1.94
 },
 "episode": [
  {
   "id_satusehat": "ep-lengkap-01",
   "ref_faskes": "100011",
   "system": "http://terminology.kemkes.go.id/CodeSystem/episodeofcare-type",
   "kode": "TB-SO",
   "display": "Tuberkulosis Sensitif Obat",
   "status": "active",
   "mulai": "2026-05-12",
   "selesai": "2026-11-12"
  }
 ],
 "observasi": [
  {
   "id_satusehat": "obs-lkp-bb",
   "system": "http://loinc.org",
   "kode": "29463-7",
   "display": "Body weight",
   "kategori": "vital-signs",
   "nilai_angka": 8.9,
   "satuan": "kg",
   "interpretasi": "OI000007",
   "tanggal": "2026-05-12T02:10:00Z"
  },
  {
   "id_satusehat": "obs-lkp-tb",
   "system": "http://loinc.org",
   "kode": "8306-3",
   "display": "Body height (lying)",
   "kategori": "vital-signs",
   "nilai_angka": 76.4,
   "satuan": "cm",
   "interpretasi": "OI000011",
   "tanggal": "2026-05-12T02:10:00Z"
  },
  {
   "id_satusehat": "obs-lkp-lk",
   "system": "http://loinc.org",
   "kode": "9843-4",
   "display": "Head circumference",
   "kategori": "vital-signs",
   "nilai_angka": 45.3,
   "satuan": "cm",
   "tanggal": "2026-05-12T02:10:00Z"
  },
  {
   "id_satusehat": "obs-lkp-lila",
   "system": "http://snomed.info/sct",
   "kode": "284473002",
   "display": "Mid upper arm circumference",
   "kategori": "exam",
   "nilai_angka": 12.1,
   "satuan": "cm",
   "tanggal": "2026-05-12T02:10:00Z"
  },
  {
   "id_satusehat": "obs-lkp-z-pbu",
   "system": "http://snomed.info/sct",
   "kode": "1153590000",
   "display": "Length for age z score",
   "kategori": "exam",
   "nilai_angka": -3.21,
   "satuan": "{Zscore}",
   "interpretasi": "OI000011",
   "tanggal": "2026-05-12T02:10:00Z"
  },
  {
   "id_satusehat": "obs-lkp-kode",
   "system": "http://snomed.info/sct",
   "kode": "363870007",
   "display": "Mental state, behavior / psychosocial function observable",
   "kategori": "survey",
   "nilai_kode": "OV000318",
   "nilai_kode_system": "http://terminology.kemkes.go.id/CodeSystem/clinical-term",
   "nilai_display": "Dilakukan stimulasi tumbuh kembang",
   "tanggal": "2026-05-12T02:20:00Z"
  },
  {
   "id_satusehat": "obs-lkp-teks",
   "system": "http://loinc.org",
   "kode": "75275-8",
   "display": "Nutrition assessment note",
   "kategori": "survey",
   "nilai_teks": "Nafsu makan menurun sejak dua minggu terakhir",
   "tanggal": "2026-05-12T02:25:00Z"
  },
  {
   "id_satusehat": "obs-lkp-panel",
   "system": "http://snomed.info/sct",
   "kode": "1156892006",
   "display": "Nutrition assessment",
   "kategori": "survey",
   "tanggal": "2026-05-12T02:30:00Z",
   "component": [
    {
     "id_satusehat": "obs-lkp-c1",
     "system": "http://snomed.info/sct",
     "kode": "709261005",
     "display": "Assessment of breastfeeding",
     "nilai_angka": 1
    },
    {
     "id_satusehat": "obs-lkp-c2",
     "system": "http://snomed.info/sct",
     "kode": "710999009",
     "display": "Monitoring food intake",
     "nilai_angka": 0
    }
   ]
  }
 ],
 "diagnosa": [
  {
   "id_satusehat": "cond-lkp-01",
   "jenis": "diagnosis",
   "system": "http://hl7.org/fhir/sid/icd-10",
   "kode": "E45",
   "display": "Nutritional stunting",
   "kategori": "encounter-diagnosis",
   "clinical_status": "active",
   "verification_status": "confirmed",
   "onset": "2026-03-01",
   "tanggal_catat": "2026-05-12"
  },
  {
   "id_satusehat": "alg-lkp-01",
   "jenis": "alergi",
   "system": "http://snomed.info/sct",
   "kode": "227493005",
   "display": "Cow milk",
   "kategori": "food",
   "kritikalitas": "high",
   "clinical_status": "active",
   "verification_status": "confirmed",
   "onset": "2026-02-01",
   "tanggal_catat": "2026-02-02"
  }
 ],
 "layanan": [
  {
   "id_satusehat": "proc-lkp-01",
   "jenis": "procedure",
   "system": "http://snomed.info/sct",
   "kode": "441041000124100",
   "display": "Counseling about nutrition",
   "kategori": "409063005",
   "status": "completed",
   "tanggal": "2026-05-12T03:00:00Z",
   "catatan": "Konseling gizi bersama ibu"
  },
  {
   "id_satusehat": "md-lkp-01",
   "jenis": "medication_dispense",
   "system": "http://sys-ids.kemkes.go.id/kfa",
   "kode": "93000271",
   "display": "F-100 Therapeutic Milk",
   "status": "completed",
   "jumlah": 30,
   "satuan": "SAC",
   "tanggal": "2026-05-12T03:30:00Z"
  },
  {
   "id_satusehat": "no-lkp-01",
   "jenis": "nutrition_order",
   "system": "http://snomed.info/sct",
   "kode": "435801000124108",
   "display": "High protein diet",
   "status": "active",
   "tanggal": "2026-05-12T03:40:00Z"
  },
  {
   "id_satusehat": "imm-lkp-01",
   "jenis": "immunization",
   "system": "http://sys-ids.kemkes.go.id/kfa",
   "kode": "93000148",
   "display": "Vitamin A 200.000 IU",
   "status": "completed",
   "jumlah": 1,
   "satuan": "KAP",
   "tanggal": "2026-05-12T03:50:00Z"
  }
 ],
 "rujukan": [
  {
   "id_satusehat": "sr-lkp-keluar",
   "jenis": "rujukan",
   "ref_faskes_asal": "100011",
   "ref_faskes_tujuan": "200022",
   "system": "http://snomed.info/sct",
   "kode": "737481003",
   "display": "Inpatient care management",
   "status": "active",
   "prioritas": "urgent",
   "alasan": "Nutritional stunting",
   "tanggal": "2026-05-15T01:00:00Z"
  }
 ]
}
```

Penjelasan setiap key, termasuk nilai yang sah dan field yang diabaikan, ada di
[Field per bagian](?doc=faskes#field-per-bagian).

### Yang ditunjukkan contoh ini

| bagian | isi |
|---|---|
| `kunjungan.tanggal_selesai` | berbeda dari `tanggal_pengukuran` — rawat inap 12–15 Mei. Samakan saja untuk rawat jalan |
| `kunjungan.ref_episode` | menunjuk `episode[0].id_satusehat`; penyambungannya otomatis |
| `observasi[0..3]` | nilai angka bersatuan. Kode BB/TB/LK/LILA yang dikenal ikut terangkat ke kolom kunjungan |
| `observasi[4]` | **z-score** PB/U beserta `interpretasi`, mengisi `zscore_tbu` dan `status_tbu`. Z-score BB/U dan BB/TB dikirim di `kunjungan` — dua sumber boleh dicampur asalkan indeksnya berbeda |
| `observasi[5]` | nilai berupa **kode**, memakai `nilai_kode` + `nilai_kode_system` + `nilai_display` |
| `observasi[6]` | nilai berupa **teks bebas**, memakai `nilai_teks` |
| `observasi[7]` | **panel berkomponen** — induk tanpa nilai, anaknya di `component` |
| `diagnosa` | `jenis` `diagnosis` dan `alergi`. Alergi memakai `kritikalitas` |
| `layanan` | empat dari lima `jenis` yang sah; `service_request` tidak ikut karena rujukan sudah punya arraynya sendiri |
| `rujukan` | `jenis` boleh `rujukan`, `rujuk_balik`, atau `internal`. Faskes tujuan dari `ref_faskes_tujuan`, bukan id internal |
| `episode` | `EpisodeOfCare`, memakai CodeSystem episodeofcare-type milik Kemkes |

Field yang boleh dihilangkan: seluruh array (`observasi`, `diagnosa`, `layanan`,
`rujukan`, `episode`) opsional, begitu juga `tanggal_selesai`, `lingkar_*`,
`cara_ukur`, dan semua `status_*`/`zscore_*`. Yang wajib hanya identitas anak,
`kunjungan.id_satusehat`, dan `kunjungan.tanggal_pengukuran`.

Payload ini benar-benar dikirim ke service saat dokumentasi ditulis, dan
menghasilkan 1 kunjungan, 8 observasi induk + 2 komponen, 2 diagnosa, 4 layanan,
1 rujukan, dan 1 episode.
