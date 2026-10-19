package repository

import (
	"fmt"
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

func keyOrangtuaId(id string) string                 { return "OrangtuaByID::" + id }
func keyOrangtuaSatusehat(satusehatid string) string { return "ORANGTUA:SATUSEHAT:" + satusehatid }
func keyOrangtuaNik(nik string) string               { return "ORTU:NIK:" + nik }

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
func keyAksesKlinis(orgid, idAnak, idOrangtua string) string {
	return fmt.Sprintf("AKSESKLINIS:%s:%s:%s", orgid, idAnak, idOrangtua)
}

var JedaLupakanUlang = 12 * time.Second

func (d *StuntingRepository) lupakan(keys ...string) {
	var dihapus []string
	for _, k := range keys {
		if k == "" {
			continue
		}
		_ = d.Memory.Delete(k)
		dihapus = append(dihapus, k)
	}
	if len(dihapus) > 0 && JedaLupakanUlang > 0 {
		time.AfterFunc(JedaLupakanUlang, func() {
			for _, k := range dihapus {
				_ = d.Memory.Delete(k)
			}
		})
	}
}

func (d *StuntingRepository) LupakanAnak(anak *entity.Anak) {
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

func (d *StuntingRepository) LupakanAnakTypes(a *types.Anak) {
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

	d.LupakanAnak(e)
}

func (d *StuntingRepository) LupakanRiwayatAnak(idAnak string) {
	if idAnak == "" {
		return
	}
	d.lupakan(keySummaryAnak(idAnak), keyListKunjungan(idAnak), keyListKesehatan(idAnak))
}

func (d *StuntingRepository) LupakanOrangtua(o *entity.Orangtua) {
	if o == nil {
		return
	}
	d.lupakan(keyOrangtuaId(o.ID), keyOrangtuaNik(o.NIK), keyListAnak(o.ID), keyOrangtuaSatusehat(o.SatusehatId))
}

func (d *StuntingRepository) LupakanAksesKlinis(orgid, idAnak, idOrangtua string) {
	d.lupakan(keyAksesKlinis(orgid, idAnak, idOrangtua))
}
func (d *StuntingRepository) GetStringByKey(key string) string {
	out := new(string)
	if ok := d.Memory.Get(key, out); !ok {
		return ""
	}
	return utils.StrPtr(out)
}
func (d *StuntingRepository) GetBool(key string) (bool, foun bool) {
	out := new(bool)
	if ok := d.Memory.Get(key, out); !ok {
		return false, false
	}
	return *out, true

}
func (d *StuntingRepository) SetStringByKey(key, value string, ttl time.Duration) {
	d.Memory.Set(key, value, ttl)
}
func (d *StuntingRepository) SetBool(key string, value bool, ttl time.Duration) {
	d.Memory.Set(key, value, ttl)
}
