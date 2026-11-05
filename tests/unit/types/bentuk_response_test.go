package types_test

import (
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jsonKeys(t *testing.T, v any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(v)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

// --- POSITIVE CASE ---

func TestResponse_MemakaiSnakeCase(t *testing.T) {
	tests := []struct {
		nama  string
		nilai any
		wajib []string
	}{
		{
			nama:  "Orangtua",
			nilai: types.Orangtua{},
			wajib: []string{"id", "id_satusehat", "id_posyandu", "no_kk", "nama_ayah",
				"nama_ibu", "nik", "telepon", "rt", "rw", "alamat", "kia",
				"usia_hamil", "kia_bayi_kecil", "created_at", "updated_at"},
		},
		{
			nama:  "Anak",
			nilai: types.Anak{},
			wajib: []string{"id", "id_satusehat", "id_orangtua", "nama", "nik",
				"tanggal_lahir", "jenis_kelamin", "anak_ke", "imd",
				"bb_lahir", "tb_lahir", "lk_lahir", "status_aktif", "created_at"},
		},
		{
			nama:  "Kunjungan",
			nilai: types.Kunjungan{},
			wajib: []string{"id", "id_anak", "tanggal_pengukuran", "cara_ukur",
				"berat_badan", "tinggi_badan", "lingkar_lengan", "lingkar_kepala",
				"lingkar_dada", "asi_bulan_0", "asi_bulan_6", "vit_biru", "vit_merah",
				"pitting_edema", "kelas_ibu_balita", "status_bbu",
				"zscore_bbtb", "status_bbtb", "zscore_bbu",
				"source_data", "created_at"},
		},
		{
			nama:  "Kesehatan",
			nilai: types.Kesehatan{},
			wajib: []string{"id", "id_anak", "tanggal_pemantauan", "tbc_batuk",
				"tbc_demam", "tbc_bb", "tbc_kontak", "layanan_asi_eks", "layanan_mpasi",
				"layanan_imunisasi", "layanan_vit_a", "layanan_obat_cacing",
				"layanan_mt_pangan", "penyuluhan_edukasi", "penyuluhan_rujukan"},
		},
		{
			nama:  "KunjunganAnak",
			nilai: types.KunjunganAnak{},
			wajib: []string{"id_orangtua", "id_anak", "orangtua", "anak", "kunjungan"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			m := jsonKeys(t, tc.nilai)
			for _, k := range tc.wajib {
				assert.Contains(t, m, k, "field %q harus ada pada respons %s", k, tc.nama)
			}
		})
	}
}

func TestResponse_TidakAdaFieldPascalCase(t *testing.T) {
	value := []any{
		types.Orangtua{}, types.Anak{}, types.Kunjungan{},
		types.Kesehatan{}, types.KunjunganAnak{},
	}

	for _, v := range value {
		for k := range jsonKeys(t, v) {
			assert.Equal(t, k, toLower(k),
				"field respons %q masih memakai nama field Go, bukan snake_case", k)
		}
	}
}

func TestResponse_IdSatusehatNullUntukAnakDariJakantro(t *testing.T) {
	// jakantro tidak bisa membuat id satusehat; puskesmas yang mengisinya nanti
	var a *types.Anak
	a = a.FromEntity(&entity.Anak{ID: testutil.IdAnak, SatusehatId: ""})

	require.NotNil(t, a)
	assert.Nil(t, a.IdSatusehat)

	m := jsonKeys(t, a)
	require.Contains(t, m, "id_satusehat")
	assert.Nil(t, m["id_satusehat"], "harus terkirim sebagai null, bukan string kosong")
}

func TestResponse_IdSatusehatTerisiUntukAnakDariPuskesmas(t *testing.T) {
	var a *types.Anak
	a = a.FromEntity(&entity.Anak{ID: testutil.IdAnak, SatusehatId: "SS-001"})

	require.NotNil(t, a)
	require.NotNil(t, a.IdSatusehat)
	assert.Equal(t, "SS-001", *a.IdSatusehat)
}

func TestResponse_OrangtuaIdSatusehatIkutAturanYangSama(t *testing.T) {
	var o *types.Orangtua

	kosong := o.FromEntity(&entity.Orangtua{ID: testutil.IdOrangtua, SatusehatId: ""})
	require.NotNil(t, kosong)
	assert.Nil(t, kosong.IdSatusehat, "orangtua harus mengikuti aturan yang sama dengan anak")

	terisi := o.FromEntity(&entity.Orangtua{ID: testutil.IdOrangtua, SatusehatId: "SS-002"})
	require.NotNil(t, terisi)
	require.NotNil(t, terisi.IdSatusehat)
	assert.Equal(t, "SS-002", *terisi.IdSatusehat)
}

func toLower(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 'A' && r <= 'Z' {
			out[i] = r + 32
		}
	}
	return string(out)
}
