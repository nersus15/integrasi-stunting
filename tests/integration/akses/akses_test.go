package akses_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/service"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	wiring  *testutil.Wiring
	db      *sql.DB
	counter *testutil.QueryCounter
)

const seed = `
INSERT INTO stunting.faskes (id, id_induk, satusehat_id, nama, jenis, wilayah) VALUES
 ('UJI-AKSES-PKM-A',    NULL, 'uji-pkm-a',    'PKM A',    'PUSKESMAS', '31.74.02.0000'),
 ('UJI-AKSES-PUSTU-A',  NULL, 'uji-pustu-a',  'PUSTU A',  'PUSTU',     '31.74.02.1001'),
 ('UJI-AKSES-PUSTU-B',  NULL, 'uji-pustu-b',  'PUSTU B',  'PUSTU',     '31.74.03.1001'),
 ('UJI-AKSES-PKM-L',    NULL, 'uji-pkm-l',    'PKM L',    'PUSKESMAS', '31.74.07.0000'),
 ('UJI-AKSES-PUSTU-X',  NULL, 'uji-pustu-x',  'PUSTU X',  'PUSTU',     NULL),
 ('UJI-AKSES-INDUK',    NULL, 'uji-induk',    'Induk',    'PUSKESMAS', '31.74.06.0000'),
 ('UJI-AKSES-ANAK',     'UJI-AKSES-INDUK', NULL, 'Cabang', 'PUSTU', '31.74.06.1001'),
 ('UJI-AKSES-RS',       NULL, 'uji-rs',       'RS',       'RS',        '31.74.09.1001'),
 ('UJI-AKSES-RS-LUAR',  NULL, 'uji-rs-luar',  'RS Luar',  'RS',        '31.71.01.1001');
INSERT INTO stunting.posyandu (id, id_puskesmas, id_kelurahan, nama, rt, rw) VALUES
 ('UJI-POS-A',     'UJI-AKSES-PKM-A', '3174021001', 'Pos A',     '001', '001'),
 ('UJI-POS-ANAK',  'UJI-AKSES-ANAK',  '3174061001', 'Pos Cab',   '001', '001'),
 ('UJI-POS-LEPAS', NULL,              '3174071001', 'Pos Lepas', '001', '001');
INSERT INTO stunting.orangtua (id, id_posyandu, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia, created_by)
VALUES ('uji-akses-ortu', 'UJI-POS-A', '3174020000000901', '3174020000000902', 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Uji', 1, 'jakantro'),
 ('uji-akses-ortu-kosong', 'UJI-POS-A', '3174020000000931', '3174020000000932', 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Uji', 1, 'jakantro');
INSERT INTO stunting.anak (id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, created_by) VALUES
 ('uji-akses-anak',  'uji-akses-ortu', 'Anak',  '3174020000000903', '2025-01-01', 'L', 1, 1, 3, 50, 33, 'jakantro'),
 ('uji-akses-yatim', NULL,             'Yatim', '3174020000000904', '2025-01-01', 'P', 1, 1, 3, 50, 33, 'uji-pustu-b');
INSERT INTO stunting.kunjungan (id, id_anak, tanggal_pengukuran, id_faskes, created_by)
VALUES ('uji-akses-kunj-yatim', 'uji-akses-yatim', '2026-01-01', 'UJI-AKSES-PKM-A', 'uji-pkm-a');
`

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		os.Exit(0)
	}
	db = testutil.OpenDB()
	cleanupSeed()
	if _, err := db.Exec(seed); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed akses: %v\n", err)
		os.Exit(1)
	}

	app, err := testutil.StartApp(2096)
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

	forgetSeedCache()
	code := m.Run()
	cleanupSeed()
	app.Stop()
	db.Close()
	os.Exit(code)
}

func cleanupSeed() {
	_, _ = db.Exec(`DELETE FROM stunting.rujukan WHERE id LIKE 'uji-akses-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.kunjungan WHERE id LIKE 'uji-akses-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id LIKE 'uji-akses-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id LIKE 'uji-akses-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.posyandu WHERE id LIKE 'UJI-POS-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.faskes WHERE id LIKE 'UJI-AKSES-%'`)
}

// baris seed bisa tersisa di cache dari run sebelumnya
func forgetSeedCache() {
	for _, id := range []string{"UJI-POS-A", "UJI-POS-ANAK", "UJI-POS-LEPAS"} {
		wiring.Repository.Lupakan("posyandu::" + id)
	}
	for _, orgid := range []string{"uji-pkm-a", "uji-pustu-a", "uji-pustu-b", "uji-pkm-l", "uji-pustu-x", "uji-induk", "uji-rs", "uji-rs-luar"} {
		wiring.Repository.Lupakan("Faskes::" + orgid)
	}
	for _, id := range []string{"uji-akses-anak", "uji-akses-yatim"} {
		wiring.Repository.LupakanAnak(&entity.Anak{ID: id})
		wiring.Repository.LupakanRujukanTerakhir(id)
	}
	wiring.Repository.LupakanOrangtua(&entity.Orangtua{ID: "uji-akses-ortu"})
	wiring.Repository.LupakanOrangtua(&entity.Orangtua{ID: "uji-akses-ortu-kosong"})
}

func strPtr(v string) *string { return &v }

// orangtua tanpa anak di posyandu tertentu: aksesnya murni akses bawaan posyandu
func TestAksesBawaanPosyandu(t *testing.T) {
	cases := []struct {
		nama     string
		posyandu *string
		orgid    string
		boleh    bool
	}{
		{"orgid kosong", strPtr("UJI-POS-A"), "", false},
		{"jakantro", strPtr("UJI-POS-A"), "jakantro", true},
		{"orangtua belum terikat posyandu", nil, "uji-pkm-a", false},
		{"posyandu tidak dikenal", strPtr("UJI-POS-HANTU"), "uji-pkm-a", false},
		{"faskes pemanggil tidak dikenal", strPtr("UJI-POS-A"), "uji-tidak-ada", false},
		{"puskesmas yang terikat posyandu", strPtr("UJI-POS-A"), "uji-pkm-a", true},
		{"induk dari cabang yang terikat posyandu", strPtr("UJI-POS-ANAK"), "uji-induk", true},
		{"pustu sekecamatan, posyandu terikat puskesmas lain", strPtr("UJI-POS-A"), "uji-pustu-a", false},
		{"puskesmas lain", strPtr("UJI-POS-A"), "uji-pustu-b", false},
		{"posyandu tanpa puskesmas, puskesmas sekecamatan", strPtr("UJI-POS-LEPAS"), "uji-pkm-l", true},
		{"posyandu tanpa puskesmas, puskesmas kecamatan lain", strPtr("UJI-POS-LEPAS"), "uji-pkm-a", false},
		{"posyandu tanpa puskesmas, pustu tanpa wilayah", strPtr("UJI-POS-LEPAS"), "uji-pustu-x", false},
		{"RS sekabupaten", strPtr("UJI-POS-A"), "uji-rs", true},
		{"RS sekabupaten, posyandu tanpa puskesmas", strPtr("UJI-POS-LEPAS"), "uji-rs", true},
		{"RS kabupaten lain", strPtr("UJI-POS-A"), "uji-rs-luar", false},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			err := wiring.Service.Akses(k.orgid, service.AksiBaca, service.Sasaran{Jenis: "orangtua", IdOrangtua: "uji-akses-ortu-kosong", IdPosyandu: k.posyandu})
			if k.boleh {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, 403, exceptions.Classify(err).HttpCode)
		})
	}
}

func insertRujukan(t *testing.T, id, idAnak, jenis, asal, tujuan, tanggal string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO stunting.rujukan (id, id_anak, jenis, id_faskes_asal, id_faskes_tujuan, tanggal)
		VALUES ($1, $2, $3, $4, $5, $6)`, id, idAnak, jenis, asal, tujuan, tanggal)
	require.NoError(t, err)
	wiring.Repository.LupakanRujukanTerakhir(idAnak)
}

func removeRujukan(t *testing.T) {
	remove := func() {
		_, _ = db.Exec(`DELETE FROM stunting.rujukan WHERE id LIKE 'uji-akses-rj-%'`)
		for _, id := range []string{"uji-akses-anak", "uji-akses-yatim", yatim2} {
			wiring.Repository.LupakanRujukanTerakhir(id)
		}
	}
	remove()
	t.Cleanup(remove)
}

func accessAnakAndOrangtua(orgid string) (errAnak, errOrtu error) {
	return accessAnak(orgid, "uji-akses-anak"), accessOrangtua(orgid, "uji-akses-ortu")
}

// skenario 1: dibuat jakantro, dipegang puskesmas posyandu, lalu dirujuk ke RS
func TestAksesPindahKeTujuanRujukanTerakhir(t *testing.T) {
	removeRujukan(t)

	allowed := func(orgid, message string) {
		t.Helper()
		a, o := accessAnakAndOrangtua(orgid)
		assert.NoError(t, a, "anak, %s: %s", orgid, message)
		assert.NoError(t, o, "orangtua, %s: %s", orgid, message)
	}
	denied := func(orgid, message string) {
		t.Helper()
		a, o := accessAnakAndOrangtua(orgid)
		assert.Equal(t, 403, exceptions.Classify(a).HttpCode, "anak, %s: %s", orgid, message)
		assert.Equal(t, 403, exceptions.Classify(o).HttpCode, "orangtua, %s: %s", orgid, message)
	}

	allowed("uji-pkm-a", "belum dirujuk, puskesmas posyandu")
	allowed("uji-rs", "belum dirujuk, RS sekab/kota posyandu")
	denied("uji-rs-luar", "belum dirujuk")

	insertRujukan(t, "uji-akses-rj-internal", "uji-akses-anak", "internal", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2025-12-01")
	allowed("uji-pkm-a", "rujukan internal tidak memindah akses")
	denied("uji-rs-luar", "rujukan internal tidak memindah akses")

	insertRujukan(t, "uji-akses-rj-1", "uji-akses-anak", "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2026-01-01")
	allowed("uji-rs-luar", "tujuan rujukan")
	denied("uji-pkm-a", "sudah merujuk ke RS, aksesnya pindah")
	denied("uji-rs", "akses bawaan posyandu tercabut setelah ada rujukan")

	insertRujukan(t, "uji-akses-rj-balik", "uji-akses-anak", "rujuk_balik", "UJI-AKSES-RS-LUAR", "UJI-AKSES-PKM-A", "2026-02-01")
	allowed("uji-pkm-a", "tujuan rujuk balik")
	denied("uji-rs-luar", "sudah merujuk balik")

	insertRujukan(t, "uji-akses-rj-2", "uji-akses-anak", "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2026-03-01")
	allowed("uji-rs-luar", "dirujuk lagi")
	denied("uji-pkm-a", "dirujuk lagi")
}

// skenario 3: pembuat tetap punya akses walau pasien dirujuk ke tempat lain
func TestPembuatTetapPunyaAksesSetelahMerujuk(t *testing.T) {
	removeRujukan(t)

	assert.NoError(t, accessAnak("uji-pustu-b", "uji-akses-yatim"), "pembuat")
	assert.Error(t, accessAnak("uji-rs-luar", "uji-akses-yatim"), "belum dirujuk")

	insertRujukan(t, "uji-akses-rj-y1", "uji-akses-yatim", "rujukan", "UJI-AKSES-PUSTU-B", "UJI-AKSES-RS-LUAR", "2026-01-01")
	assert.NoError(t, accessAnak("uji-rs-luar", "uji-akses-yatim"), "tujuan rujukan")
	assert.NoError(t, accessAnak("uji-pustu-b", "uji-akses-yatim"), "pembuat tetap punya akses")

	insertRujukan(t, "uji-akses-rj-y2", "uji-akses-yatim", "rujuk_balik", "UJI-AKSES-RS-LUAR", "UJI-AKSES-PUSTU-B", "2026-02-01")
	assert.Error(t, accessAnak("uji-rs-luar", "uji-akses-yatim"), "sudah merujuk balik")
	assert.NoError(t, accessAnak("uji-pustu-b", "uji-akses-yatim"))
}

func TestPencatatKunjunganBukanPemegangAkses(t *testing.T) {
	removeRujukan(t)
	assert.Error(t, accessAnak("uji-pkm-a", "uji-akses-yatim"), "mencatat kunjungan tidak memberi akses identitas")
}

func TestAksesJakantroDanDataTidakAda(t *testing.T) {
	removeRujukan(t)
	a, o := accessAnakAndOrangtua("jakantro")
	assert.NoError(t, a)
	assert.NoError(t, o)
	assert.NoError(t, accessAnak("jakantro", "uji-akses-yatim"))

	assert.Equal(t, 403, exceptions.Classify(accessAnak("", "uji-akses-anak")).HttpCode)
	assert.Equal(t, 403, exceptions.Classify(accessOrangtua("", "uji-akses-ortu")).HttpCode)
	assert.Equal(t, 404, exceptions.Classify(accessAnak("uji-pkm-a", "tidak-ada")).HttpCode)
	assert.Equal(t, 404, exceptions.Classify(accessOrangtua("uji-pkm-a", "tidak-ada")).HttpCode)
}

func accessAnak(orgid, id string) error {
	return wiring.Service.AksesAnak(orgid, service.AksiBaca, id)
}

func accessOrangtua(orgid, id string) error {
	return wiring.Service.AksesOrangtua(orgid, service.AksiBaca, id)
}

