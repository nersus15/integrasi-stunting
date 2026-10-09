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
| `observasi` | — | `id_satusehat`, `system`, dan `kode` wajib. Tanpa `system`, kodenya tidak bisa dipetakan ke kolom kunjungan |
| `diagnosa` | — | `id_satusehat` dan `kode` wajib. `jenis` `diagnosis` (bawaan) atau `alergi` |
| `layanan` | — | `id_satusehat` dan `jenis` wajib |
| `rujukan` | — | `id_satusehat` wajib. `jenis` boleh dikosongkan, arahnya disimpulkan dari faskes asal dan tujuan |
| `episode` | — | `id_satusehat` wajib. Disambung ke kunjungan lewat `kunjungan.ref_episode` |

Setiap key dijelaskan di [Field per bagian](#field-per-bagian). Payload lengkap
yang sudah diuji ada di [Contoh Payload](?doc=contoh#jalur-faskes).

### Field per bagian

Daftar ini lengkap: field yang tidak tercantum tidak dikenali dan diabaikan.

Yang ditentukan sistem selalu diabaikan kalau dikirim: `id`, `id_anak`,
`id_kunjungan`, `id_induk`, `id_faskes`, `id_faskes_asal`, `id_faskes_tujuan`,
`id_episode`, `id_rujukan`, `stunting`, dan kolom audit (`created_at`,
`updated_at`, `deleted_at`, `updated_by`, `deleted_by`). Faskes dirujuk lewat
`ref_*` berisi **id Organization SatuSehat apa adanya**, mis. `100011` — tanpa
awalan `Organization/`.

#### `anak`

Hanya dipakai untuk mencocokkan anak, dan untuk membuat anak baru kalau belum
terdaftar. Data anak yang sudah ada tidak diubah oleh endpoint ini; pakai
[PUT /api/anak](#put-apianak-dan-put-apiorangtua).

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | salah satu | IHS Number pasien. Kunci pencocokan utama |
| `nik` | 16 digit angka | salah satu | dipakai kalau `id_satusehat` belum cocok |
| `nama` | teks, maks 255 | anak baru | |
| `tanggal_lahir` | `YYYY-MM-DD` | anak baru | penentu kriteria balita; anak di atas 5 tahun ditolak `422` `6001` |
| `jenis_kelamin` | `L`/`P` | anak baru | |
| `anak_ke` | bilangan bulat | — | urutan kelahiran |
| `imd` | `0`/`1` | — | mendapat inisiasi menyusu dini |
| `bb_lahir`, `tb_lahir`, `lk_lahir` | angka | — | berat (kg), panjang (cm), dan lingkar kepala (cm) saat lahir |
| `source_data` | teks, maks 36 | — | penanda sumber data |
| `status_aktif` | teks, maks 20 | — | bawaan `aktif`. Disimpan apa adanya |

`id_orangtua` diabaikan: faskes tidak menentukan keluarga anak. Anak baru tanpa
`nama`, `tanggal_lahir`, atau `jenis_kelamin` dijawab `422` `1002`.

#### `kunjungan`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id Encounter. Inilah yang membuat kiriman ulang tidak menggandakan |
| `tanggal_pengukuran` | `YYYY-MM-DD` | ya | tanggal kunjungan dimulai |
| `tanggal_selesai` | `YYYY-MM-DD` | — | akhir rawat inap. Kosongkan untuk rawat jalan |
| `cara_ukur` | teks | — | `telentang` atau `berdiri`. Ditimpa otomatis dari kode tinggi badan di observasi |
| `berat_badan` | angka | — | kg. Ditimpa otomatis dari observasi LOINC `29463-7` |
| `tinggi_badan` | angka | — | cm. Ditimpa otomatis dari `8302-2`, `8306-3` (telentang), atau `8308-9` (berdiri) |
| `lingkar_kepala` | angka | — | cm. Ditimpa otomatis dari `9843-4` |
| `lingkar_lengan` | angka | — | cm (LILA). Ditimpa otomatis dari SNOMED `284473002` |
| `lingkar_dada` | angka | — | cm |
| `asi_bulan_0` … `asi_bulan_6` | `0`/`1` | — | ASI eksklusif pada bulan ke-0 sampai ke-6 |
| `vit_biru`, `vit_merah` | `0`/`1` | — | kapsul vitamin A biru (6–11 bulan) dan merah (12–59 bulan) |
| `pitting_edema` | `0`/`1` | — | ditemukan edema pitting |
| `kelas_ibu_balita` | `0`/`1` | — | ibu mengikuti kelas ibu balita |
| `status_*`, `zscore_*` | teks, angka | — | z-score dan status gizi BB/U, TB/U, BB/TB. Per indeks, kirim di sini **atau** sebagai [observasi z-score](#z-score-dan-status-gizi), jangan keduanya |
| `ref_episode` | teks | — | `id_satusehat` episode yang menaungi kunjungan ini, di array `episode` atau yang sudah tersimpan |
| `ref_rujukan` | teks | — | `id_satusehat` rujukan yang mendasari kunjungan ini |

Faskes kunjungan selalu faskes pemilik API key. `stunting` ditentukan sistem
dari diagnosa dan interpretasi observasi.

#### `observasi[]`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id Observation. Kunci pencocokan |
| `system` | teks | ya | URI terminologi, mis. `http://loinc.org` |
| `kode` | teks | ya | kode di dalam system itu |
| `display` | teks | — | nama terbaca manusia |
| `kategori` | teks | — | mis. `vital-signs`, `exam`, `survey`. Bebas, tidak divalidasi |
| `nilai_angka` + `satuan` | angka + teks | salah satu | untuk hasil terukur |
| `nilai_kode` + `nilai_kode_system` + `nilai_display` | teks | salah satu | untuk hasil berupa kode |
| `nilai_teks` | teks | salah satu | untuk hasil deskriptif |
| `interpretasi` | teks | — | `N`, `L`, `H`, `A`, atau kode interpretasi z-score di bawah |
| `tanggal` | waktu | — | waktu pengukuran, kalau berbeda dari tanggal kunjungan |
| `ref_encounter` | teks | — | `id_satusehat` kunjungan lain yang Anda miliki. Kosong berarti kunjungan di payload ini |
| `component[]` | array | — | anak observasi untuk hasil panel. Strukturnya sama; `tanggal` yang kosong mewarisi induknya |

Ketiga bentuk nilai bersifat pilih salah satu. Observasi panel boleh tidak
punya nilai sama sekali — nilainya ada di `component`.

#### Z-score dan status gizi

Service tidak menghitung z-score. Nilainya diambil dari observasi z-score
SatuSehat (modul Gizi), yang kodenya SNOMED dengan `nilai_angka` berisi z-score:

| `kode` | indeks | mengisi |
|---|---|---|
| `1153593003` | BB/U | `zscore_bbu`, `status_bbu` |
| `1153590000`, `1153604005` | PB/U (0–24 bulan), TB/U (24–60 bulan) | `zscore_tbu`, `status_tbu` |
| `1153598007`, `1153600001` | BB/PB (0–24 bulan), BB/TB (24–60 bulan) | `zscore_bbtb`, `status_bbtb` |
| `1153596006`, `1153594009` | IMT/U, lingkar kepala/U | tidak ada kolom kunjungan, hanya tersimpan sebagai observasi |

`status_*` diambil dari `interpretasi`, sesuai Lampiran 3 terminologi Gizi
SatuSehat. Kodenya hanya dibaca di dalam indeksnya sendiri, karena kode seperti
`248324001` dipakai bersama oleh beberapa indeks.

| kolom | `interpretasi` → status |
|---|---|
| `status_bbu` | `OI000007` Berat Badan Sangat Kurang, `248342006` Berat Badan Kurang, `43664005` Berat Badan Normal, `OI000010` Risiko Berat Badan Lebih |
| `status_tbu` | `OI000011` Sangat Pendek, `444000005` Pendek, `17489000` Normal, `83077003` Tinggi |
| `status_bbtb` | `OI000001` Gizi Buruk, `248325000` Gizi Kurang, `248324001` Gizi Baik, `OI000004` Risiko Gizi Lebih, `238131007` Gizi Lebih, `414915002` Obesitas |

Z-score dan status satu indeks selalu ditulis berpasangan. Observasi z-score
tanpa `interpretasi`, atau dengan kode di luar tabel di atas, membuat statusnya
`NULL` — status lama tidak dipertahankan.

Untuk satu indeks, z-score boleh datang dari `kunjungan` **atau** dari observasi,
tidak keduanya. Mengirim `kunjungan.zscore_tbu` atau `kunjungan.status_tbu`
bersama observasi `1153604005` dijawab `422` `1003`, meskipun nilainya sama.
Indeks yang berbeda boleh dicampur, seperti pada
[contoh lengkap](?doc=contoh#jalur-faskes).

#### `diagnosa[]`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id Condition atau AllergyIntolerance. Kunci pencocokan |
| `jenis` | teks | — | `diagnosis` atau `alergi`. Kosong berarti `diagnosis` |
| `system` | teks | — | ICD-10 untuk diagnosis, SNOMED untuk alergi |
| `kode` | teks | ya | |
| `display` | teks | — | |
| `kategori` | teks | — | diagnosis: `problem-list-item`, `encounter-diagnosis`. Alergi: `food`, `medication`, `environment`, `biologic` |
| `kritikalitas` | teks | — | `low`, `high`, `unable-to-assess` |
| `clinical_status` | teks | — | diagnosis: `active`, `recurrence`, `relapse`, `inactive`, `remission`, `resolved`. Alergi: `active`, `inactive`, `resolved` |
| `verification_status` | teks | — | diagnosis: `unconfirmed`, `provisional`, `differential`, `confirmed`, `refuted`, `entered-in-error`. Alergi: `unconfirmed`, `confirmed`, `refuted`, `entered-in-error` |
| `onset` | `YYYY-MM-DD` | — | kapan mulai dirasakan |
| `tanggal_catat` | `YYYY-MM-DD` | — | kapan dicatat di rekam medis |
| `ref_encounter` | teks | — | sama dengan observasi |

#### `layanan[]`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id Procedure, MedicationDispense, NutritionOrder, Immunization, atau ServiceRequest |
| `jenis` | teks | ya | `procedure`, `medication_dispense`, `nutrition_order`, `immunization`, `service_request` |
| `system`, `kode`, `display` | teks | — | terminologi layanan |
| `kategori` | teks | — | bebas, tidak divalidasi |
| `status` | teks | — | bergantung `jenis`, lihat tabel di bawah |
| `jumlah` + `satuan` | angka + teks | — | dosis atau kuantitas, untuk obat dan imunisasi |
| `tanggal` | waktu | — | kapan diberikan |
| `catatan` | teks | — | teks bebas |
| `ref_encounter` | teks | — | sama dengan observasi |

| `jenis` | `status` yang sah |
|---|---|
| `procedure` | `preparation`, `in-progress`, `not-done`, `on-hold`, `stopped`, `completed`, `entered-in-error`, `unknown` |
| `medication_dispense` | `preparation`, `in-progress`, `cancelled`, `on-hold`, `completed`, `entered-in-error`, `stopped`, `declined`, `unknown` |
| `nutrition_order`, `service_request` | `draft`, `active`, `on-hold`, `revoked`, `completed`, `entered-in-error`, `unknown` |
| `immunization` | `completed`, `entered-in-error`, `not-done` |

#### `rujukan[]`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id ServiceRequest |
| `jenis` | teks | — | `rujukan`, `rujuk_balik`, atau `internal`. Kosong berarti disimpulkan sendiri |
| `ref_faskes_asal` | teks | — | id Organization faskes perujuk. Kosong atau tidak dikenal berarti faskes kunjungan ini |
| `ref_faskes_tujuan` | teks | — | id Organization faskes tujuan. Harus sudah terdaftar; kalau belum, dijawab `422` `1005` |
| `system`, `kode`, `display` | teks | — | jenis layanan yang dirujuk |
| `status` | teks | — | `draft`, `active`, `on-hold`, `revoked`, `completed`, `entered-in-error`, `unknown` |
| `prioritas` | teks | — | `routine`, `urgent`, `asap`, `stat` |
| `alasan` | teks | — | indikasi rujukan |
| `tanggal` | waktu | — | tanggal surat rujukan dibuat |
| `ref_encounter` | teks | — | sama dengan observasi |

Kalau `jenis` dikosongkan, arahnya disimpulkan: tanpa `ref_faskes_tujuan` atau
tujuannya sama dengan asal jadi `internal`; RS ke puskesmas jadi `rujuk_balik`;
selebihnya `rujukan`. Sebutkan `jenis` secara eksplisit kalau Anda tahu
persisnya — nilai yang Anda kirim selalu menang atas kesimpulan itu.

#### `episode[]`

| key | tipe | wajib | keterangan |
|---|---|---|---|
| `id_satusehat` | teks, maks 36 | ya | id EpisodeOfCare. Dipakai `kunjungan.ref_episode` untuk menyambung |
| `ref_faskes` | teks | — | id Organization pengelola episode. Kalau belum terdaftar, faskes episode dibiarkan kosong |
| `system`, `kode`, `display` | teks | — | jenis episode, mis. CodeSystem episodeofcare-type Kemkes |
| `status` | teks | — | `planned`, `waitlist`, `active`, `onhold`, `finished`, `cancelled`, `entered-in-error` |
| `mulai`, `selesai` | `YYYY-MM-DD` | — | rentang episode |

Nilai enum di luar daftar dijawab `422` `1002`, dengan pesan yang menyebut field
dan nilai yang sah.

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

Z-score satu indeks yang dikirim di `kunjungan` sekaligus sebagai observasi
z-score dijawab `422` `1003`, meskipun nilainya sama. Pilih salah satu sumber
per indeks:

```json
{
  "httpCode": 422,
  "errorCode": 1003,
  "errorName": "PAYLOAD_SHAPE_INVALID",
  "message": "observasi[4] (1153590000) sudah membawa zscore_tbu; kunjungan.zscore_tbu dan kunjungan.status_tbu jangan dikirim bersamaan"
}
```

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

`id_satusehat` **wajib** — itu satu-satunya kunci pencarian barisnya. Body-nya
memakai field yang sama dengan bagian yang bersangkutan di
[Field per bagian](#field-per-bagian), dengan syarat minimal berikut:

| `:resourceName` | wajib dikirim |
|---|---|
| `kunjungan` | `id_satusehat`, `tanggal_pengukuran` |
| `observasi` | `id_satusehat`, `system`, `kode` |
| `diagnosa` | `id_satusehat`, `kode` |
| `layanan` | `id_satusehat`, `jenis` (harus nilai yang sah, tapi tidak mengubah jenis yang tersimpan) |
| `rujukan`, `episode` | `id_satusehat` |

Enum yang bergantung pada `jenis` diperiksa terhadap jenis yang **tersimpan**,
bukan yang dikirim.

Field yang tidak dikirim dipertahankan. Untuk `kunjungan`, ini penting: PUT
Encounter hanya membawa periode kunjungan, sehingga berat dan tinggi badan yang
berasal dari Observation tidak ikut terhapus.

PUT `observasi` yang kodenya terangkat ke kolom kunjungan (berat, tinggi,
lingkar, z-score) ikut memperbarui kolom itu di kunjungannya, sama seperti PUT
Observation yang datang lewat Kafka. Kolom kunjungan lain tidak tersentuh.

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

