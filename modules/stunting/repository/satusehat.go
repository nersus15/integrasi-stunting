package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

// Encounter bisa datang di bundle terpisah, jadi dicari, bukan dari isi bundle
func (d *StuntingRepository) FindKunjunganBySatusehatId(satusehatId string) (*entity.Kunjungan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id encounter tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Kunjungan)

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Kunjungan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindKunjunganBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindObservasiBySatusehatId(satusehatId string) (*entity.Observasi, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id observation tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Observasi)

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Observasi{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Observasi tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindObservasiBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindEpisodeBySatusehatId(satusehatId string, loadFaskes bool) (*entity.Episode, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id episode tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Episode)

	tabel := entity.Episode{}.TableName()
	kolom := "satusehat_id"

	if loadFaskes {
		tabel = "[me],Faskes"
		kolom = "ep.satusehat_id"
	}

	filter := []port.DbExpression{
		{Expr: kolom + " = ?", Args: []any{satusehatId}},
	}
	if !loadFaskes {
		filter = append(filter, belumDihapus)
	}

	err := d.Connection.FindOne(ctx, tmp, tabel, []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Episode tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindEpisodeBySatusehatId (%v): ", satusehatId) + err.Error())
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindRujukanBySatusehatId(satusehatId string) (*entity.Rujukan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id rujukan tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Rujukan)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Rujukan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Rujukan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindRujukanBySatusehatId (%v): ", satusehatId) + err.Error())
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) UpdateKunjunganBySatusehatId(data *entity.Kunjungan, stream bool) (*entity.Kunjungan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	lama, err := d.FindKunjunganBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var gabungan *entity.Kunjungan
	var input map[string]any

	if stream {
		input = map[string]any{
			"tanggal_pengukuran": data.TanggalPengukuran,
			"tanggal_selesai":    data.TanggalSelesai,
		}
	} else {
		gabungan = lama
		gabungan.Override(*data, false)
		input = StructToMap(gabungan)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu

	d.lupakan(keyKunjunganId(lama.ID))

	return gabungan, nil
}

func (d *StuntingRepository) UpdateObservasiBySatusehatId(data *entity.Observasi, stream bool) (*entity.Observasi, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	lama, err := d.FindObservasiBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var gabungan *entity.Observasi
	var input map[string]any
	if stream {
		input = map[string]any{
			"system":            data.System,
			"kode":              data.Kode,
			"display":           data.Display,
			"kategori":          data.Kategori,
			"nilai_angka":       data.NilaiAngka,
			"satuan":            data.Satuan,
			"nilai_teks":        data.NilaiTeks,
			"nilai_kode":        data.NilaiKode,
			"nilai_kode_system": data.NilaiKodeSystem,
			"nilai_display":     data.NilaiDisplay,
			"interpretasi":      data.Interpretasi,
			"tanggal":           data.Tanggal,
		}
	} else {
		gabungan = lama
		gabungan.Override(*data, false)
		input = StructToMap(gabungan)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu

	return gabungan, nil
}
func (d *StuntingRepository) FindDiagnosaBySatusehatId(satusehatId string) (*entity.Diagnosa, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id diagnosa tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Diagnosa)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Diagnosa{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Diagnosa tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindDiagnosaBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindLayananBySatusehatId(satusehatId string) (*entity.Layanan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id layanan tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Layanan)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Layanan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Layanan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindLayananBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

// waktu update dikembalikan supaya respons menampilkan updated_at yang baru
func (d *StuntingRepository) updateBySatusehatId(tabel string, satusehatId *string, idAnak string, kolom map[string]any) (*time.Time, error) {
	if !utils.IsFilled(satusehatId) {
		return nil, exceptions.KolomWajib.Messagef("Id satusehat harus sudah terisi")
	}
	if len(kolom) == 0 {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	waktu := time.Now()
	kolom["updatedAt"] = waktu

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{*satusehatId}},
	}

	if _, err := d.Connection.Update(d.Context.Context, tabel, filter, kolom); err != nil {
		logger.Error(fmt.Sprintf("updateBySatusehatId %s (%s): ", tabel, *satusehatId) + err.Error())
		return nil, err
	}

	d.LupakanRiwayatAnak(idAnak)

	return &waktu, nil
}

func (d *StuntingRepository) UpdateDiagnosaBySatusehatId(data *entity.Diagnosa, stream bool) (*entity.Diagnosa, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindDiagnosaBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":               data.Jenis,
			"system":              data.System,
			"kode":                data.Kode,
			"display":             data.Display,
			"kategori":            data.Kategori,
			"kritikalitas":        data.Kritikalitas,
			"clinical_status":     data.ClinicalStatus,
			"verification_status": data.VerificationStatus,
			"onset":               data.Onset,
			"tanggal_catat":       data.TanggalCatat,
		}
	} else {
		lama.Override(*data, false)
		input = StructToMap(lama)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu

	if stream {
		return nil, nil
	}
	return lama, nil
}

func (d *StuntingRepository) UpdateLayananBySatusehatId(data *entity.Layanan, stream bool) (*entity.Layanan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindLayananBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":    data.Jenis,
			"system":   data.System,
			"kode":     data.Kode,
			"display":  data.Display,
			"kategori": data.Kategori,
			"status":   data.Status,
			"jumlah":   data.Jumlah,
			"satuan":   data.Satuan,
			"tanggal":  data.Tanggal,
			"catatan":  data.Catatan,
		}
	} else {
		lama.Override(*data, false)
		input = StructToMap(lama)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu

	if stream {
		return nil, nil
	}
	return lama, nil
}

func (d *StuntingRepository) UpdateRujukanBySatusehatId(data *entity.Rujukan, stream bool) (*entity.Rujukan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindRujukanBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":     data.Jenis,
			"system":    data.System,
			"kode":      data.Kode,
			"display":   data.Display,
			"status":    data.Status,
			"prioritas": data.Prioritas,
			"alasan":    data.Alasan,
			"tanggal":   data.Tanggal,
		}
	} else {
		lama.Override(*data, false)
		input = StructToMap(lama)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu
	// pemegang akses ikut rujukan terakhir
	d.LupakanRujukanTerakhir(lama.IDAnak)

	if stream {
		return nil, nil
	}
	return lama, nil
}

func (d *StuntingRepository) UpdateEpisodeBySatusehatId(data *entity.Episode, stream bool) (*entity.Episode, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindEpisodeBySatusehatId(utils.Nilai(data.SatusehatId), false)
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"system":  data.System,
			"kode":    data.Kode,
			"display": data.Display,
			"status":  data.Status,
			"mulai":   data.Mulai,
			"selesai": data.Selesai,
		}
	} else {
		lama.Override(*data, false)
		input = StructToMap(lama)
	}

	waktu, err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input)
	if err != nil {
		return nil, err
	}
	lama.UpdatedAt = waktu

	if stream {
		return nil, nil
	}
	return lama, nil
}
