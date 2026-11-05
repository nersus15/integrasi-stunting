package kepemilikan_stream_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataBaruDariStreamDicatatSatusehat(t *testing.T) {
	setup(t)
	send(t, "kep-baru", "POST", pasien("Anak Stream"), ibu("Ibu Stream"), encounter(), tinggiStunting())

	assert.Equal(t, "satusehat|satusehat", scalar(t, `SELECT trim(source_data) || '|' || created_by FROM stunting.anak WHERE nik = $1`, nikAnak))
	assert.Equal(t, "satusehat|satusehat", scalar(t, `SELECT trim(source_data) || '|' || created_by FROM stunting.orangtua WHERE nik = $1`, nikIbu))
	assert.Equal(t, "satusehat|"+orgUji, scalar(t, `SELECT trim(source_data) || '|' || created_by FROM stunting.kunjungan WHERE satusehat_id = $1`, encUji),
		"pembuat kunjungan dari stream adalah faskes pengirim Encounter")
}

func TestPatientMenimpaIdentitasMeskiSudahDiubahJakantro(t *testing.T) {
	setup(t)
	send(t, "kep-awal", "POST", pasien("Anak Stream"), encounter(), tinggiStunting())

	_, err := db.Exec(`UPDATE stunting.anak SET nama = 'Diubah Jakantro', updated_by = 'jakantro' WHERE nik = $1`, nikAnak)
	assert.NoError(t, err)

	send(t, "kep-put-1", "PUT", pasien("Nama Dari SatuSehat"))
	assert.Equal(t, "Nama Dari SatuSehat", scalar(t, `SELECT nama FROM stunting.anak WHERE nik = $1`, nikAnak))

	send(t, "kep-put-2", "PUT", pasien("Nama Kedua Dari SatuSehat"))
	assert.Equal(t, "Nama Kedua Dari SatuSehat", scalar(t, `SELECT nama FROM stunting.anak WHERE nik = $1`, nikAnak),
		"Patient berikutnya tetap menimpa, tidak hanya sekali")
}

func TestRelatedPersonMenimpaIdentitasMeskiSudahDiubahJakantro(t *testing.T) {
	setup(t)
	send(t, "kep-awal-rp", "POST", pasien("Anak Stream"), ibu("Ibu Stream"), encounter(), tinggiStunting())

	_, err := db.Exec(`UPDATE stunting.orangtua SET nama_ibu = 'Diubah Jakantro', updated_by = 'jakantro' WHERE nik = $1`, nikIbu)
	assert.NoError(t, err)

	send(t, "kep-put-rp", "PUT", ibu("Ibu Dari SatuSehat"))
	assert.Equal(t, "Ibu Dari SatuSehat", scalar(t, `SELECT nama_ibu FROM stunting.orangtua WHERE nik = $1`, nikIbu))
}

func TestPembuatTidakBerubahSaatStreamMemperbarui(t *testing.T) {
	setup(t)
	send(t, "kep-awal-pb", "POST", pasien("Anak Stream"), encounter(), tinggiStunting())
	_, err := db.Exec(`UPDATE stunting.anak SET created_by = 'jakantro', source_data = 'jakantro:kader-01' WHERE nik = $1`, nikAnak)
	assert.NoError(t, err)

	send(t, "kep-put-pb", "PUT", pasien("Nama Baru"))
	assert.Equal(t, "jakantro|jakantro:kader-01", scalar(t, `SELECT created_by || '|' || trim(source_data) FROM stunting.anak WHERE nik = $1`, nikAnak))
}
