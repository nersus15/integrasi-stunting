package akses_test

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ortu2  = "uji-akses-ortu-2"
	ortu3  = "uji-akses-ortu-3"
	yatim2 = "uji-akses-yatim-2"
	ihsY2  = "P-UJI-YATIM-2"
)

func setupYatimDirujuk(t *testing.T) {
	t.Helper()
	cleanup := func() {
		_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id = $1`, yatim2)
		_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id IN ($1, $2)`, ortu2, ortu3)
		_, _ = db.Exec(`DELETE FROM stunting.rujukan WHERE id = 'uji-akses-rj-yatim2'`)
		// baris ditulis lewat SQL, jadi cache-nya dibuang manual
		wiring.Repository.LupakanRujukanTerakhir(yatim2)
		wiring.Repository.LupakanAnak(&entity.Anak{ID: yatim2, SatusehatId: ihsY2})
		for _, o := range []string{ortu2, ortu3} {
			wiring.Repository.LupakanOrangtua(&entity.Orangtua{ID: o})
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	_, err := db.Exec(`INSERT INTO stunting.orangtua (id, id_posyandu, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia, created_by) VALUES
		($1, 'UJI-POS-A', '3174020000000911', '3174020000000912', 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Uji', 1, 'jakantro'),
		($2, 'UJI-POS-A', '3174020000000921', '3174020000000922', 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Uji', 1, 'jakantro')`, ortu2, ortu3)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO stunting.anak (id, satusehat_id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, created_by)
		VALUES ($1, $2, NULL, 'Yatim 2', '3174020000000913', '2025-01-01', 'L', 1, 1, 3, 50, 33, 'jakantro')`, yatim2, ihsY2)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO stunting.rujukan (id, id_anak, jenis, id_faskes_asal, id_faskes_tujuan, tanggal)
		VALUES ('uji-akses-rj-yatim2', $1, 'rujukan', 'UJI-AKSES-PKM-A', 'UJI-AKSES-RS-LUAR', '2026-01-01')`, yatim2)
	require.NoError(t, err)
}

func TestAksesOrangtuaSetelahAnakDihubungkan(t *testing.T) {
	setupYatimDirujuk(t)
	require.Error(t, accessOrangtua("uji-rs-luar", ortu2), "belum ada anak yang dirujuk ke RS")
	require.NoError(t, accessOrangtua("uji-pkm-a", ortu2), "belum punya anak, dipegang puskesmas posyandu")

	n, err := wiring.Repository.HubungkanAnakKeOrangTua(ihsY2, ortu2, nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)

	assert.NoError(t, accessOrangtua("uji-rs-luar", ortu2), "anaknya kini dirujuk ke RS")
	assert.Error(t, accessOrangtua("uji-pkm-a", ortu2), "satu-satunya anak sudah dirujuk, akses puskesmas posyandu pindah")
}

func TestAksesOrangtuaSetelahAnakDipindah(t *testing.T) {
	setupYatimDirujuk(t)
	_, err := wiring.Repository.HubungkanAnakKeOrangTua(ihsY2, ortu2, nil)
	require.NoError(t, err)
	require.NoError(t, accessOrangtua("uji-rs-luar", ortu2))
	require.Error(t, accessOrangtua("uji-rs-luar", ortu3))

	_, err = wiring.Service.UpdateAnakById(&entity.Anak{ID: yatim2, IDOrangtua: ortu3}, strPtr("jakantro"))
	require.NoError(t, err)

	assert.Error(t, accessOrangtua("uji-rs-luar", ortu2), "anak yang dirujuk sudah pindah")
	assert.NoError(t, accessOrangtua("uji-rs-luar", ortu3))
}

func TestAksesOrangtuaSetelahAnakDihapus(t *testing.T) {
	setupYatimDirujuk(t)
	_, err := wiring.Repository.HubungkanAnakKeOrangTua(ihsY2, ortu2, nil)
	require.NoError(t, err)
	require.NoError(t, accessOrangtua("uji-rs-luar", ortu2))

	require.NoError(t, wiring.Service.DeleteAnakByID(yatim2, "jakantro"))

	assert.Error(t, accessOrangtua("uji-rs-luar", ortu2), "satu-satunya anak yang dirujuk sudah dihapus, tanpa menunggu job latar")
}
