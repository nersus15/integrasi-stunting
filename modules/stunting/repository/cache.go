package repository

import (
	"strconv"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

const (
	ttlPanjang = 24 * time.Hour
	ttlSedang  = 15 * time.Minute
	ttlPendek  = 60 * time.Second
)

func keyOrangtuaId(id string) string   { return "OrangtuaByID::" + id }
func keyOrangtuaNik(nik string) string { return "ORTU:NIK:" + nik }

func keyAnakId(id string) string         { return "ANAK:ID:" + id }
func keyAnakNik(nik string) string       { return "ANAK:NIK:" + nik }
func keyAnakSatusehat(sid string) string { return "ANAK:SATUSEHAT:" + sid }
func keyAnakUrutan(idOrangtua string, urutan int16) string {
	return "ANAK:ORANGTUA:" + idOrangtua + ":" + strconv.Itoa(int(urutan))
}

func keyListAnak(idOrangtua string) string  { return "LIST:ANAK:ORTU:" + idOrangtua }
func keyListKunjungan(idAnak string) string { return "LIST:KUNJUNGAN:" + idAnak }
func keyListKesehatan(idAnak string) string { return "LIST:KESEHATAN:" + idAnak }
func keySummaryAnak(idAnak string) string   { return "SUMMARY:ANAK:" + idAnak }
func keyKunjunganId(id string) string       { return "KUNJUNGAN:ID:" + id }
func keyKesehatanId(id string) string       { return "KESEHATAN:ID:" + id }

func (d *StuntingRepository) lupakan(keys ...string) {
	for _, k := range keys {
		if k == "" {
			continue
		}
		_ = d.Memory.Delete(k)
	}
}

func (d *StuntingRepository) lupakanAnak(anak *entity.Anak) {
	if anak == nil {
		return
	}

	keys := []string{
		keyAnakId(anak.ID),
		keySummaryAnak(anak.ID),
		keyListKunjungan(anak.ID),
		keyListKesehatan(anak.ID),
	}
	if anak.NIK != nil && *anak.NIK != "" {
		keys = append(keys, keyAnakNik(*anak.NIK))
	}
	if anak.SatusehatId != "" {
		keys = append(keys, keyAnakSatusehat(anak.SatusehatId))
	}
	if anak.IDOrangtua != "" {
		keys = append(keys, keyAnakUrutan(anak.IDOrangtua, anak.AnakKe), keyListAnak(anak.IDOrangtua))
	}

	d.lupakan(keys...)
}

func (d *StuntingRepository) lupakanAnakTypes(a *types.Anak) {
	if a == nil {
		return
	}

	e := &entity.Anak{
		ID:         a.Id,
		IDOrangtua: a.IdOrangtua,
		NIK:        a.Nik,
		AnakKe:     a.AnakKe,
	}
	if a.IdSatusehat != nil {
		e.SatusehatId = *a.IdSatusehat
	}

	d.lupakanAnak(e)
}

func (d *StuntingRepository) lupakanRiwayatAnak(idAnak string) {
	if idAnak == "" {
		return
	}
	d.lupakan(keySummaryAnak(idAnak), keyListKunjungan(idAnak), keyListKesehatan(idAnak))
}

func (d *StuntingRepository) lupakanOrangtua(o *entity.Orangtua) {
	if o == nil {
		return
	}
	d.lupakan(keyOrangtuaId(o.ID), keyOrangtuaNik(o.NIK), keyListAnak(o.ID))
}

func (d *StuntingRepository) GetStringByKey(key string) string {
	out := new(string)
	if ok := d.Memory.Get(key, out); !ok {
		return ""
	}
	return utils.StrPtr(out)
}

func (d *StuntingRepository) SetStringByKey(key, value string, ttl time.Duration) {
	d.Memory.Set(key, value, ttl)
}
