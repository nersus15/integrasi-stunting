# Pembacaan

Berlaku untuk kedua jalur.

## GET endpoints

Semua GET butuh API key. Yang tidak ketemu selalu `404` `2001 NOT_FOUND`.

### GET /api/orangtua/:id

| parameter | tipe | keterangan |
|---|---|---|
| `id` | path, opsional | id orangtua |
| `nik` | query | NIK orangtua |
| `nokk` | query | **belum diimplementasikan** — selalu 404 |
| `nama_ayah` | query | **belum diimplementasikan** — selalu 404 |
| `nama_ibu` | query | **belum diimplementasikan** — selalu 404 |

Minimal satu harus diisi; kalau kosong semua kena `400` `1002`.

```
GET /api/orangtua/8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21
GET /api/orangtua?nik=3175040101900002
```

```json
{
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
  "kia_bayi_kecil": null,
  "updated_by": null,
  "deleted_by": null
}
```

### GET /api/orangtua/:id/anak

Orangtua beserta semua anaknya, urut `anak_ke`.

```json
{
  "id": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
  "nama_ayah": "Ahmad Suryana",
  "nama_ibu": "Dewi Lestari",
  "nik": "3175040101900002",
  "anak": [
    {
      "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
      "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
      "nama": "Rizky Ramadhan",
      "nik": "3175041503240003",
      "tanggal_lahir": "2024-03-15",
      "jenis_kelamin": "L",
      "anak_ke": 1,
      "status_aktif": "aktif"
    }
  ]
}
```

Field orangtua di-flatten ke root, bukan dibungkus object `orangtua`. Anak tanpa
riwayat tetap muncul; `anak` berupa `[]` kalau orangtua belum punya anak.

### GET /api/anak/:id

```json
{
  "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
  "id_satusehat": null,
  "id_orangtua": "8f3a1c40-6b2e-4d19-9a77-1e5c8b0d4a21",
  "nama": "Rizky Ramadhan",
  "nik": "3175041503240003",
  "tanggal_lahir": "2024-03-15",
  "jenis_kelamin": "L",
  "anak_ke": 1,
  "imd": 1,
  "bb_lahir": 3.2,
  "tb_lahir": 49.5,
  "lk_lahir": 34,
  "created_at": "2026-08-28T04:04:51.489267Z",
  "updated_at": null,
  "deleted_at": null,
  "source_data": "jakantro                            ",
  "status_aktif": "aktif",
  "updated_by": null,
  "deleted_by": null
}
```

### GET /api/anak/:id/kunjungan

Anak beserta riwayat kunjungannya, urut `createdAt` naik.

```json
{
  "anak": {
    "id": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
    "nama": "Rizky Ramadhan",
    "nik": "3175041503240003",
    "anak_ke": 1
  },
  "kunjungan": [
    {
      "id": "c1e48a72-9d35-4b80-a6f3-52d7e9418b04",
      "id_anak": "b7d9e254-3f81-4c6a-8e12-90ab5d3f7c68",
      "tanggal_pengukuran": "2026-08-14",
      "cara_ukur": "telentang",
      "berat_badan": 9.4,
      "tinggi_badan": 76.2
    }
  ]
}
```

Anak tanpa kunjungan tetap `200`, dengan `kunjungan: []`. Yang `404` hanya kalau
anaknya sendiri tidak ada.

### GET /api/anak/:id/kesehatan

Bentuknya sama, `kunjungan` diganti `kesehatan`, urut `tanggal_pemantauan` naik.

### GET /api/anak/:id/summary

Riwayat lengkap satu anak dari ketiga sumber sekaligus: posyandu (jakantro),
puskesmas, dan RS. Ini endpoint untuk melacak perjalanan penanganan, bukan
sekadar penghemat round trip.

```json
{
  "anak":        { "id": "...", "nama": "Siti Aminah", "satusehat_id": "P99001100022" },
  "episode":     [ ... ],
  "kunjungan":   [ ... ],
  "kesehatan":   [ ... ],
  "layanan":     [ ... ],
  "rujukan":     [ ... ],
  "menggantung": { "observasi": [], "diagnosa": [], "layanan": [], "rujukan": [] }
}
```

| key | isi |
|---|---|
| `anak` | identitas anak |
| `episode` | episode perawatan (`EpisodeOfCare`), satu episode bisa mencakup banyak kunjungan di faskes berbeda |
| `kunjungan` | riwayat kunjungan, urut `tanggal_pengukuran` naik. Tiap kunjungan membawa data medisnya sendiri |
| `kesehatan` | kuisioner pemantauan dari posyandu |
| `layanan` | **semua** layanan lintas kunjungan, datar dan kronologis |
| `rujukan` | **semua** rujukan lintas kunjungan, beserta tindak lanjutnya |
| `menggantung` | data medis yang sudah tiba tapi Encounter-nya belum, jadi belum menempel ke kunjungan mana pun |

`layanan` dan `rujukan` sengaja diulang datar di tingkat atas supaya jejak
penanganan dan alur rujukan bisa dibaca sekali lihat, tanpa menelusuri satu per
satu kunjungan. Isinya sama dengan yang tersebar di dalam `kunjungan`.

#### Isi tiap `kunjungan`

```json
{
  "id": "...",
  "id_anak": "...",
  "tanggal_pengukuran": "2026-02-20",
  "cara_ukur": "telentang",
  "berat_badan": 7.5,
  "tinggi_badan": 72.8,
  "lingkar_lengan": null,
  "lingkar_kepala": null,
  "stunting": 1,

  "id_satusehat": "enc-rs-01",
  "id_faskes": "FAS-RS-HARAPAN",
  "id_episode": "...",
  "ref_episode": "ep-stunting-sim1",
  "id_rujukan": "...",
  "ref_rujukan": "sr-rujuk-01",

  "faskes":       { "nama": "RS Harapan Bunda", "jenis": "rs" },
  "episode":      { "kode": "hacc", "status": "active" },
  "atas_rujukan": { "id_satusehat": "sr-rujuk-01", "jenis": "rujukan" },

  "observasi": [ ... ],
  "diagnosa":  [ ... ],
  "layanan":   [ ... ],
  "rujukan":   [ ... ]
}
```

| kolom | arti |
|---|---|
| `tanggal_pengukuran` | tanggal kunjungan. Untuk posyandu ini memang tanggal penimbangan; untuk FHIR ini `Encounter.period.start`, jadi **tanggal mulai** kunjungan |
| `tanggal_selesai` | `Encounter.period.end`. Sama dengan `tanggal_pengukuran` untuk rawat jalan; berbeda untuk rawat inap yang melintasi beberapa hari. `null` untuk kunjungan posyandu |
| `berat_badan`, `tinggi_badan`, `lingkar_lengan`, `lingkar_kepala` | antropometri. `null` berarti tidak diukur, **bukan** nol |
| `cara_ukur` | `telentang` atau `berdiri`, diturunkan dari kode LOINC tinggi badan |
| `stunting` | `1` stunting, `0` tidak, **`null` belum ditentukan**. Kunjungan posyandu dan kunjungan tanpa data klinis selalu `null` — jakantro tidak menetapkan status, dan kunjungan ber-`null` tidak dipakai sebagai acuan riwayat |
| `id_faskes` | `null` berarti posyandu (jakantro). Terisi berarti puskesmas atau RS |
| `faskes` | detail faskes-nya; inilah penanda asal data |
| `id_satusehat` | IHS id `Encounter` |
| `id_episode` / `episode` | episode perawatan yang menaungi kunjungan ini |
| `ref_episode` | IHS id episode dari payload. Terisi sementara `id_episode` masih `null` = episodenya belum tiba |
| `observasi` | observasi di kunjungan ini. `component` bersarang di induknya, jadi tidak muncul dua kali |
| `diagnosa` | diagnosis **dan** alergi. Bedakan lewat kolom `jenis`: `diagnosis` atau `alergi` |
| `layanan` | tindakan, obat, imunisasi, order gizi |

#### `rujukan` vs `atas_rujukan`

Dua-duanya soal rujukan, tapi arahnya berlawanan:

| | arti |
|---|---|
| `kunjungan.rujukan` | rujukan yang **diterbitkan** di kunjungan ini — pasien dikirim ke faskes lain |
| `kunjungan.atas_rujukan` | rujukan yang **dipenuhi** oleh kunjungan ini — pasien datang karena dirujuk |

Contoh dari satu perjalanan pasien:

```
2026-02-05  Puskesmas   rujukan      = [sr-rujuk-01]   <- menerbitkan rujukan ke RS
                        atas_rujukan = null

2026-02-20  RS          rujukan      = []
                        atas_rujukan = sr-rujuk-01     <- datang karena rujukan itu

2026-02-25  RS          rujukan      = [sr-balik-01]   <- menerbitkan rujuk balik
                        atas_rujukan = null

2026-03-15  Puskesmas   rujukan      = []
                        atas_rujukan = sr-balik-01     <- pasien kembali
```

Jadi `sr-rujuk-01` muncul dua kali: sebagai `rujukan` di kunjungan asal, dan
sebagai `atas_rujukan` di kunjungan yang memenuhinya. Sumbernya
`Encounter.basedOn` yang diisi faskes tujuan.

Satu rujukan bisa dipenuhi **lebih dari satu** kunjungan — konsultasi awal lalu
kontrol lanjutan. Karena itu `atas_rujukan` tunggal di sisi kunjungan, sedangkan
di daftar `rujukan` tingkat atas tindak lanjutnya berupa array.

#### Isi tiap `rujukan` (tingkat atas)

```json
{
  "id_satusehat": "sr-rujuk-01",
  "jenis": "rujukan",
  "id_faskes_asal": "FAS-PKM-KENANGA",
  "id_faskes_tujuan": "FAS-RS-HARAPAN",
  "alasan": "Nutritional stunting",
  "prioritas": "urgent",
  "tanggal": "2026-02-05T03:30:00Z",
  "tindakan": [
    { "id": "...", "tanggal_pengukuran": "2026-02-20", "faskes": { "nama": "RS Harapan Bunda" } }
  ]
}
```

| kolom | arti |
|---|---|
| `jenis` | `rujukan` keluar, `rujuk_balik` kembali ke faskes perujuk, atau `internal` |
| `tindakan` | kunjungan yang memenuhi rujukan ini. **Array kosong = rujukan belum ditindaklanjuti** |
| `id_faskes_asal` / `id_faskes_tujuan` | `null` kalau faskes-nya belum terdaftar di tabel faskes |

`jenis` diambil dari `ServiceRequest.category` sesuai Playbook Rujukan: kode
SNOMED `3457005` menandai rujukan pasien, kode Kemkes `SR000007` menandai rujuk
balik. Kalau kategorinya tidak menyebutkan apa-apa dan faskes tujuannya tidak
diketahui — misalnya permintaan lab atau radiologi yang `performer`-nya seorang
Practitioner — jenisnya `internal`, bukan rujukan antar faskes.

Array `tindakan` yang kosong adalah sinyal paling berguna di endpoint ini: anak
dirujuk tapi tidak pernah datang.

#### `menggantung`

Data medis bisa tiba sebelum `Encounter`-nya. Baris seperti itu disimpan dengan
`id_kunjungan` `null` dan ditampung di sini, lalu tersambung sendiri begitu
Encounter-nya masuk. Kalau tidak ditampilkan, data itu tidak terlihat di mana
pun.

### GET /api/anak/kunjungan/:id dan GET /api/anak/kesehatan/:id

Satu record by id. Path param `id` wajib; kalau kosong kena `400` `1007`.

```
GET /api/anak/kunjungan/c1e48a72-9d35-4b80-a6f3-52d7e9418b04
GET /api/anak/kesehatan/d5b3f169-8c47-42ae-b91d-6f0a3e28c7d5
```

Response berupa object kunjungan atau kesehatan, tanpa data anak.

---
