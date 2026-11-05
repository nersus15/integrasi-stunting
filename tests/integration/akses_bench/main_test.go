package akses_bench_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

// Hanya memakai API yang sama di semua versi (Service.AksesAnak/AksesOrangtua dan
// kunci cache berupa string), supaya file ini bisa dijalankan terhadap commit lama.

const seed = `
INSERT INTO stunting.faskes (id, satusehat_id, nama, jenis, wilayah) VALUES
 ('UJI-BENCH-PKM', 'uji-bench-pkm', 'PKM Bench', 'PUSKESMAS', '31.74.02.0000'),
 ('UJI-BENCH-RS',  'uji-bench-rs',  'RS Bench',  'RS',        '31.71.01.1001');
INSERT INTO stunting.posyandu (id, id_puskesmas, id_kelurahan, nama, rt, rw)
VALUES ('UJI-BENCH-POS', 'UJI-BENCH-PKM', '3174021001', 'Pos Bench', '001', '001');
INSERT INTO stunting.orangtua (id, id_posyandu, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia, created_by)
VALUES ('uji-bench-ortu', 'UJI-BENCH-POS', '3174029900000001', '3174029900000002', 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Bench', 1, 'jakantro');
INSERT INTO stunting.anak (id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, created_by) VALUES
 ('uji-bench-anak-1', 'uji-bench-ortu', 'Anak Satu', '3174029900000003', '2025-01-01', 'L', 1, 1, 3, 50, 33, 'jakantro'),
 ('uji-bench-anak-2', 'uji-bench-ortu', 'Anak Dua',  '3174029900000004', '2025-01-01', 'P', 2, 1, 3, 50, 33, 'jakantro');
INSERT INTO stunting.rujukan (id, id_anak, jenis, id_faskes_asal, id_faskes_tujuan, ref_faskes_asal, ref_faskes_tujuan, tanggal)
VALUES ('uji-bench-rj', 'uji-bench-anak-2', 'rujukan', 'UJI-BENCH-PKM', 'UJI-BENCH-RS', 'uji-bench-pkm', 'uji-bench-rs', '2026-01-01');
`

var (
	wiring  *testutil.Wiring
	db      *sql.DB
	counter *testutil.QueryCounter
	memory  *testutil.CountingMemory
)

func cleanupSeed() {
	for _, q := range []string{
		`DELETE FROM stunting.rujukan WHERE id LIKE 'uji-bench-%'`,
		`DELETE FROM stunting.anak WHERE id LIKE 'uji-bench-%'`,
		`DELETE FROM stunting.orangtua WHERE id LIKE 'uji-bench-%'`,
		`DELETE FROM stunting.posyandu WHERE id LIKE 'UJI-BENCH-%'`,
		`DELETE FROM stunting.faskes WHERE id LIKE 'UJI-BENCH-%'`,
	} {
		_, _ = db.Exec(q)
	}
}

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres tidak hidup, benchmark akses dilewati")
		os.Exit(0)
	}
	db = testutil.OpenDB()
	cleanupSeed()
	if _, err := db.Exec(seed); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}

	app, err := testutil.StartApp(2100)
	if err == nil {
		wiring, err = app.Wiring("")
	}
	if err == nil {
		counter, err = app.InstallQueryCounter()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}
	var ok bool
	if memory, ok = wiring.Memory.(*testutil.CountingMemory); !ok {
		fmt.Fprintf(os.Stderr, "cache tidak dibungkus CountingMemory (%T)\n", wiring.Memory)
		os.Exit(1)
	}

	code := m.Run()
	cleanupSeed()
	app.Stop()
	db.Close()
	os.Exit(code)
}
