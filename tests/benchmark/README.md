# Benchmark: simulasi satu server terbatas

Mengukur titik jenuh service di mesin lokal tanpa deploy, dengan membatasi sumber daya agar
mendekati server sungguhan (bukan SSD + 8 core laptop).

## Cara kerja

Postgres, Redis, dan app berjalan dalam **satu container ("box")** sehingga berbagi CPU, RAM, dan
disk seperti satu mesin. Batasnya diambil dari `spek-server.env`:

| Spek | Mekanisme |
|---|---|
| CPU, RAM | `docker --cpus`, `--memory` (page cache ikut dihitung; GOMAXPROCS Go mengikuti kuota CPU) |
| IOPS & throughput disk | `--device-{read,write}-{iops,bps}` (cgroup v2 `io.max`) pada `DISK_DEVICE` |
| Latensi fsync HDD | Postgres `commit_delay = DISK_FSYNC_MS` + `commit_siblings = 0`: setiap flush WAL menunggu sekian ms |
| Konfigurasi Postgres/app | diteruskan ke `postgres -c ...` dan `config.yaml` box |

Kafka memakai broker lokal yang sudah ada (`localhost:1092`) dengan topic/group terpisah
(`bench-stunting-fhir` / `bench-stunting`), mewakili broker di luar server.

Yang **tidak** tersimulasi: latensi baca acak HDD saat beban rendah (batas IOPS hanya terasa saat
antre), latensi jaringan klien → server, dan CPU/cache yang berbeda generasi. Hasil di sini adalah
perkiraan; bandingkan antar konfigurasi, bukan angka mutlak.

## Menjalankan

```bash
./jalankan.sh build        # sekali, atau setiap source app berubah
./jalankan.sh siapkan      # sekali, atau setiap SEED_* / migrasi berubah (~2 menit untuk 300 rb anak)
./jalankan.sh uji semua    # http -> kafka -> campuran, ~40 menit
./jalankan.sh bersihkan    # hapus container, volume, topic
```

Mode `uji`: `http`, `kafka`, `campuran` (butuh `CAMPURAN_LAJU_HTTP` dan `CAMPURAN_LAJU_KAFKA`), `semua`,
dan `profil`. `profil` menjalankan setiap operasi 20 kali berurutan (`PROFIL_ULANG`) dengan
`pg_stat_statements` aktif, lalu menulis `profil.md`: jumlah statement per request beserta rinciannya.
Pakai untuk membandingkan sebelum dan sesudah optimasi query.
Setiap run memakai salinan segar volume hasil `siapkan`.

Semua nilai di `spek-server.env` dan `beban.env` bisa ditimpa per run tanpa mengubah file:

```bash
SERVER_CPUS=4 SERVER_MEMORY=8g DISK_FSYNC_MS=0 ./jalankan.sh uji http
```

## Beban

- **HTTP**: open-loop pada laju tetap per tahap (`HTTP_LAJU_AWAL` × `HTTP_FAKTOR` per tahap),
  campuran operasi menurut `BOBOT_*`. Latensi dihitung dari waktu jadwal kirim, jadi antrean di sisi
  klien ikut terhitung.
- **Kafka**: perjalanan anak dari `simulasi/skenario1` (Patient → RelatedPerson → puskesmas → RS →
  rujuk balik → lanjutan) dengan id unik, key partisi = IHS anak. Lag diukur dari offset yang
  di-commit consumer app. Setelah tahap, `KAFKA_BURST_ANAK` perjalanan diantrekan sekaligus untuk
  mengukur kapasitas consumer.
- **Campuran**: HTTP dan Kafka bersamaan pada `CAMPURAN_PERSEN` dari titik jenuh masing-masing.

Tahap dianggap jenuh bila error > `BATAS_ERROR_PERSEN`, p95 > `BATAS_P95_MS`, laju tercapai < 90%
target, atau lag Kafka terus naik.

## Hasil

`hasil/<waktu>-<mode>/`:

- `laporan.md`: tabel per tahap, sumber daya box, rincian per operasi di sekitar titik jenuh.
- `hasil.json`: data mentah.
- `kalibrasi.txt`: fsync/detik terukur di dalam batas IOPS.
- `box.log`: log app dan postgres.
- `env-dipakai.txt`: nilai yang benar-benar dipakai.

Membaca kolom sumber daya:

- `PSI cpu/io/mem`: persen waktu ada proses tertahan menunggu sumber daya itu. Yang paling tinggi
  saat jenuh adalah bottleneck-nya.
- `CPU throttled`: kuota CPU habis.
- `koneksi aktif` mendekati `DB_MAX_OPEN_CONNS`: pool koneksi app yang membatasi.
