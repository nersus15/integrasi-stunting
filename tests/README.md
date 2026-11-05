# Test

Modul terpisah dari `go.work` supaya bisa merangkai aplikasi sendiri tanpa
memengaruhi build service. Jalankan dengan `GOWORK=off`.

## Menjalankan per bagian

Tidak perlu menjalankan semuanya setiap kali. `uji.sh` menjalankan satu atau
beberapa paket saja, memuat env dari `tests/.test.env`, dan memberi tahu paket
mana yang **dilewati** (paket tanpa database keluar lebih awal dan tetap
tercetak `ok`, jadi mudah dikira lulus).

```bash
cd tests
./uji.sh daftar                    # semua bagian beserta kebutuhannya
./uji.sh update_endpoint           # satu paket
./uji.sh optimasi_query -run Summary        # satu kelompok test di dalam paket
./uji.sh unit/validator unit/zscore         # beberapa paket sekaligus
./uji.sh unit                      # satu kelompok: unit | integration | functional
./uji.sh semua                     # seluruh test
./uji.sh delete_endpoint -run TestDeleteAnakSuccess -v   # -v: log lengkap
```

Keluarannya diringkas seperti jest: status per paket, `✓`/`✗` per test (saat
satu paket dijalankan), lalu untuk setiap test gagal hanya lokasi dan pesan
kegagalannya. Contoh:

```
 GAGAL  functional/optimasi_query (1.1s)
    ✗ TestRegistrasiOrangtuaBaruSatuSelectOrangtua (0.03s)
    ✓ TestRegistrasiOrangtuaSudahAdaDipakaiUlang (0.01s)

  ● TestRegistrasiOrangtuaBaruSatuSelectOrangtua

    registrasi_test.go:51: SELECT orangtua untuk orangtua baru: harap 1 statement, ada 2
         1. SELECT … FROM "stunting"."orangtua" AS "o" WHERE (nik = '…') AND "o"."deletedAt" IS NULL LIMIT 1
         2. SELECT … FROM "stunting"."orangtua" AS "o" WHERE (nik = '…') AND "o"."deletedAt" IS NOT NULL LIMIT 1

Paket: 1 gagal, 0 lulus, 1 total
Test : 1 gagal, 1 lulus, 2 total
```

- Test yang dilewati (`t.Skip`) selalu tampil sebagai `○ NamaTest (dilewati)`
  beserta alasannya, juga saat `./uji.sh semua`.
- Log aplikasi (logger webcore, goose, dsb.) tidak tampil di terminal; isinya ada
  di `tests/.log/<kelompok>-<paket>.log`. Log request Fiber dimatikan di aplikasi
  test. `UJI_LOG=stdout ./uji.sh ...` menampilkan keduanya lagi.
- `-v` memberi keluaran mentah `go test` (tanpa ringkasan).
- Peringkasnya ada di `alat/ringkas`, bisa dipakai langsung:
  `GOWORK=off go test -json -p 1 ./unit/... | go run ./alat/ringkas -daftar`.
- Nama bagian boleh nama folder (`update_endpoint`) atau path
  (`functional/update_endpoint`). `find_orangtua` ada di `integration/` dan
  `functional/`, jadi harus memakai path.
- Argumen yang diawali `-` diteruskan ke `go test` (`-run`, `-v`, `-timeout`, ...).

## Env (`.test.env`)

Semua env umum test diatur di `tests/.test.env`: koneksi database (`PG_*`, atau
`PG_DSN` untuk menimpa semuanya), Redis (`REDIS_*`), `ORGID_FASKES`,
`API_BASE`/`API_HEALTH`, `UJI_LOG`, dan `NO_COLOR`. File ini dibaca
oleh `uji.sh` dan oleh `testutil` saat paket dimuat, jadi berlaku juga untuk
`go test` langsung.

Env dari shell menang atas `.test.env`. Jadi
`PG_PORT=5432 ./uji.sh unit` tetap bisa dipakai tanpa mengubah file. Formatnya
`NAMA=nilai` per baris; baris `#` diabaikan, komentar di ujung baris tidak didukung.

| bagian | kapan dijalankan |
|---|---|
| `unit` | setiap perubahan fungsi murni (validator, konversi, FHIR, z-score) |
| `stream_kafka` | perubahan `service/stream.go` atau pemetaan resource FHIR |
| `update_endpoint`, `kontrak_field` | perubahan endpoint PUT atau field yang boleh diisi |
| `delete_endpoint` | perubahan delete, soft delete, atau cache anak/orangtua |
| `create_kunjungan`, `create_orangtua_anak` | perubahan POST jakantro |
| `optimasi_query` | perubahan summary, registrasi, sapuan turunan, atau cek akses faskes |
| `akses`, `kepemilikan` | perubahan aturan akses (`service.Akses`), hak akses faskes, atau pembuat data |
| `api_stunting` | sebelum rilis: skenario end-to-end lengkap |

## Menjalankan semuanya

```bash
cd tests
GOWORK=off go test -p 1 ./...
```

> **Penting.** `-p 1` wajib. Seluruh paket memakai satu database dev yang sama
> dan sebagian berbagi id fixture, jadi menjalankannya paralel membuat mereka
> saling menghapus data. Tanpa `-p 1` kegagalannya tidak konsisten.

Paket yang menyentuh database dilewati (bukan gagal) kalau Postgres dari
`.test.env` tidak hidup.

## Key API test

Seluruh test memakai empat key tetap dari `testutil/keys.go`. `StartApp` membuat
user dan key-nya di `access.auth_users`/`auth_apikey` kalau belum ada, jadi tidak
perlu menyiapkan apa pun.

| konstanta | key | grup | role |
|---|---|---|---|
| `KeyJakantro` | `key-jakantro` | `jakantro` | read, write |
| `KeyFaskesA` | `key-faskes-A` | `orgid:100011` (puskesmas, `OrgFaskesA`) | read, write |
| `KeyFaskesB` | `key-faskes-B` | `orgid:100009864` (puskesmas di luar wilayah A, `OrgFaskesB`) | read, write |
| `KeyReadonly` | `key-readonly` | `orgid:100011` | read |

`functional/update_endpoint` menambah key khusus untuk RS dan key tanpa grup;
keduanya dihapus di akhir paket. Environment `test-*` di `api-collection` memakai
key yang sama.

Paket yang butuh HTTP (`functional/api_stunting`, `functional/auth`, dan lainnya)
menjalankan aplikasi in-process. Untuk menguji service yang sudah berjalan, isi
`API_BASE` saat menjalankan paketnya, bukan di `.test.env`; `API_HEALTH` bawaannya
diturunkan dari `API_BASE`:

```bash
API_BASE=http://localhost:2025/api ./uji.sh auth
```

`api_stunting` tetap mengosongkan tabel `orangtua`, `anak`, `kunjungan`, dan
`kesehatan` di database `.test.env` walaupun menembak service luar.

## Cakupan

```bash
GOWORK=off go test -p 1 -count=1 \
  -coverpkg=github.com/nersus15/integrasi/mod-stunting/... \
  -coverprofile=cakupan.out ./...
go tool cover -func=cakupan.out | tail -1
```

Aplikasi yang dijalankan sebagai proses terpisah tidak ikut terukur, karena itu
paket yang butuh HTTP menjalankan aplikasi di dalam proses test lewat
`testutil.StartApp(port)`.

## Susunan

| folder | isi | butuh |
|---|---|---|
| `unit/` | fungsi murni: validator, deteksi payload, konversi entity/types, `Override`, pemetaan FHIR, utils | — |
| `integration/` lama | repository dan service dengan database palsu | — |
| `integration/stream_kafka` | pipeline Kafka lewat `KafkaHandler.Run` tanpa broker: skenario 1, PUT per resource, bundle khusus, pencatatan pesan gagal dan retry | database |
| `unit/zscore` | pemetaan Observation z-score SatuSehat ke `zscore_*`/`status_*` kunjungan. Kontraknya di `ZSCORE_CONTRACT.md` | — |
| `unit/delete` | repository dan service delete dengan database palsu: soft delete lewat UPDATE, 404, error database, cache | — |
| `integration/akses` | `VerifyAccess` dengan faskes/posyandu buatan (wilayah, induk, RS) | database |
| `integration/klasifikasi` | `exceptions.Classify` terhadap galat Postgres sungguhan | database |
| `integration/repository_db` | fungsi repository yang tidak lewat endpoint | database |
| `functional/api_stunting` | 155 skenario end-to-end lewat HTTP | database + 3 kunci |
| `functional/update_endpoint` | seluruh endpoint PUT dan hak aksesnya | database + key test |
| `functional/kontrak_field` | field yang boleh diisi pengirim di setiap endpoint tulis, kontraknya di `FIELD_CONTRACT.md` | database + key test |
| `functional/delete_endpoint` | `DELETE /api/anak/:id` dan `/api/orangtua/:id`: soft delete, cache, hapus ulang, hak akses | database + key test |
| `functional/optimasi_query` | jumlah statement SQL di jalur yang dioptimasi (summary, registrasi, sapuan turunan, cek akses faskes) beserta perilakunya; summary dikunci dengan golden `testdata/summary.golden.json` | database + key test |
| `functional/auth` | identitas dan kunci tidak tertukar antar request; `TestParalelIdentitasTidakTertukar` memakai aplikasi in-process | database + key test |
| `unit/kepemilikan`, `functional/kepemilikan`, `integration/kepemilikan_stream` | `created_by`, standar `source_data`, dan hak ubah/hapus menurut pembuat data. Kontraknya di `KEPEMILIKAN_CONTRACT.md` | database + key test |
| `functional/*` lainnya | aplikasi dirangkai in-process, diuji lewat `app.Web.Test` | database |

Paket yang menjalankan aplikasi memakai port berbeda supaya tidak bertabrakan:
`api_stunting` 2098, `auth` 2085, `kepemilikan` 2083, `kepemilikan_stream` 2082, `stream_kafka` 2099, `update_endpoint` 2097, `akses` 2096,
`repository_db` 2095, `kontrak_field` 2091, `delete_endpoint` 2090, `create_kunjungan` 2089,
`create_orangtua_anak` 2088, `find_orangtua` 2087, `optimasi_query` 2086. Kafka selalu dimatikan di sana.

`functional` memakai **database sungguhan**, bukan palsu: `Module.Init`
menjalankan migrasi goose yang menuntut koneksi `*bun.DB`. Migrasinya
idempoten, jadi aman dijalankan berulang.

## Data fixture

`testutil.CleanFixtures()` menghapus baris yang dibuat fixture, dan
`SeedOrangtua` / `SeedOrangtuaAndAnak` menyiapkan prasyarat untuk bentuk
payload yang mengandaikan datanya sudah ada. Seluruh fixture memakai id tetap,
jadi tiap kasus harus berangkat dari keadaan yang sama.

`functional/api_stunting` mengosongkan `stunting.kunjungan`, `kesehatan`,
`anak`, dan `orangtua` sebelum mulai. Tabel acuan `faskes` dan `posyandu` tidak
disentuh.

Paket yang memakai skenario 1 (`simulasi/skenario1`) menjalankan seed faskes
dan posyandunya, lalu menghapus anak, orangtua, dan episode skenario itu
setelah tiap test. `integration/akses` menambah baris `UJI-AKSES-*` dan
`UJI-POS-*`, `functional/update_endpoint` menambah pengguna API `uji-cakupan-*`;
keduanya dihapus lagi di akhir paket. Pesan Kafka gagal yang dibuat test
memakai `group_id` sendiri. Satu-satunya pengecualian adalah test endpoint
`/kafka/gagal`, yang harus memakai `group_id` service; test itu dilewati kalau
antrean service sedang berisi pesan lain.

## Laporan

`functional/api_stunting` menulis laporan HTML berisi payload, status code, dan
response tiap skenario ke `tests/functional/api_stunting/laporan.html` (bisa
diganti lewat `LAPORAN_HTML`). Laporan ini tidak lagi disajikan service —
ukurannya terlalu besar untuk ikut ditanam ke biner.

## Menghitung query

`app.InstallQueryCounter()` menempelkan hook bun ke koneksi aplikasi in-process.
`Measure(fn)` mengembalikan statement yang dijalankan selama `fn`, dan
`Match(pattern)` menyaringnya dengan regexp:

```go
q := counter.Measure(func() { request(t, "GET", "/anak/"+id+"/summary", testutil.KeyJakantro, nil) })
assert.LessOrEqual(t, len(q.Match(`^\s*SELECT .* FROM "stunting"\.`)), 8)
```

Golden summary ditulis ulang dengan `TULIS_GOLDEN=1 ./uji.sh optimasi_query -run
TestSummaryBentukResponsTetap`, hanya kalau perubahan bentuk respons memang
disengaja. Untuk jumlah statement per endpoint di bawah beban, lihat
`tests/benchmark` (`./jalankan.sh uji profil`).

## Benchmark cek akses

`integration/akses_bench` mengukur `AksesAnak`/`AksesOrangtua` di dalam proses, dengan Postgres dan
Redis dari `.test.env`, tanpa HTTP. Selain `ns/op`, tiap kasus melaporkan `query/op` (statement DB) dan
`cache-get/op` (GET ke cache; `Wiring` membungkus cache dengan `CountingMemory`). Kasus `dingin`
membuang cache data transaksi sebelum tiap cek; cache master faskes/posyandu tetap hangat.

`uji.sh` meringkas keluaran, jadi jalankan `go test` langsung:

```bash
set -a; . ./.test.env; set +a
GOWORK=off go test ./integration/akses_bench -run '^$' -bench . -benchmem -count 10 | tee baru.txt
benchstat lama.txt baru.txt
```

File benchmark hanya memakai API yang ada di semua versi, jadi bisa dijalankan terhadap commit lama:
buat worktree commit itu, salin `libraries/`, `tests/`, dan `config.yaml` (ketiganya tidak dilacak git)
ke dalamnya, lalu jalankan perintah yang sama di `tests/` worktree tersebut.
