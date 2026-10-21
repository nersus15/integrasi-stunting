package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
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
func (d *StuntingRepository) Lupakan(keys ...string) {
	d.lupakan(keys...)
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

func (d *StuntingRepository) LupakanRiwayatAnak(idAnak string) {
	if idAnak == "" {
		return
	}
	d.lupakan(keySummaryAnak(idAnak), keyListKunjungan(idAnak), keyListKesehatan(idAnak))
}

func (d *StuntingRepository) LupakanKunjungan(k *entity.Kunjungan) {
	if k != nil {
		d.lupakan(keyKunjunganId(k.ID))
		d.LupakanRiwayatAnak(k.IDAnak)
	}
}

func (d *StuntingRepository) LupakanKesehatan(h *entity.Kesehatan) {
	if h != nil {
		d.lupakan(keyKesehatanId(h.ID))
		d.LupakanRiwayatAnak(h.IDAnak)
	}
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
func (d *StuntingRepository) GetBool(key string) (nilai bool, ada bool) {
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

func (d *StuntingRepository) GetKey(resource, diff string, varian *string) *string {
	switch strings.ToLower(resource) {
	case "anak", "orangtua":
		if !utils.IsFilled(varian) {
			return nil
		}

		switch utils.StrPtr(varian) {
		case "id":
			if strings.EqualFold(resource, "anak") {
				return utils.PtrStr(keyAnakId(diff), true)
			}
			return utils.PtrStr(keyOrangtuaId(diff), true)

		case "nik":
			if strings.EqualFold(resource, "anak") {
				return utils.PtrStr(keyAnakNik(diff), true)
			}
			return utils.PtrStr(keyOrangtuaNik(diff), true)

		case "satusehat":
			if strings.EqualFold(resource, "anak") {
				return utils.PtrStr(keyAnakSatusehat(diff), true)
			}
			return utils.PtrStr(keyOrangtuaSatusehat(diff), true)
		}

	case "kunjungan":
		return utils.PtrStr(keyKunjunganId(diff), true)

	case "kesehatan":
		return utils.PtrStr(keyKesehatanId(diff), true)
	}

	return nil
}
