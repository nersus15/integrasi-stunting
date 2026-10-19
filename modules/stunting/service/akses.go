package service

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/webcore-go/webcore/infra/logger"
)

func (s *StuntingService) VerifyAccess(idposyandu, orgid *string) error {

	if !utils.IsFilled(orgid) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}

	// null berarti belum terikat posyandu, beda dari posyandu tidak dikenal
	if !utils.IsFilled(idposyandu) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, orangtua belum terikat posyandu", nil)
	}

	posyandu, err := s.Repository.FindPosyanduDetail(*idposyandu)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
		}
		return exceptions.Forbidden.WithMessage(err.Error(), err)
	}

	if posyandu == nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
	}

	// jika puskesmas/RS
	if utils.IsFilled(orgid) && *orgid != "jakantro" {
		if posyandu.Puskesmas == nil {
			return exceptions.Forbidden.WithMessage("Tidak bisa verifikasi akses, posyandu tidak terhubung dengan puskesmas", nil)
		}
		valid := false
		satusehatid := posyandu.Puskesmas.SatusehatID
		induk := posyandu.Puskesmas.Induk
		if satusehatid != nil {
			if *satusehatid == *orgid {
				valid = true
			}
		} else if induk != nil {
			if induk.SatusehatID != nil && *induk.SatusehatID == *orgid {
				valid = true
			}
		}

		if !valid {
			if !utils.IsFilled(posyandu.Puskesmas.Wilayah) {
				return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes posyandu kosong", nil)
			}

			// Cari wilayah kerja dari faskes user
			faskesSaya, err := s.Repository.FindFaskesByOrgid(*orgid)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
				}
				return exceptions.Forbidden.WithMessage(err.Error(), err)
			}

			if faskesSaya == nil {
				return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
			}

			if slices.Contains([]string{"PUSTU", "PUSKESMAS"}, strings.ToUpper(faskesSaya.Jenis)) {
				if !utils.IsFilled(faskesSaya.Wilayah) {
					return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes user kosong", nil)
				}
				levelWilayah := entity.LevelWilayah(posyandu.Puskesmas.Wilayah)

				switch levelWilayah {
				case 3: // kecamatan
					if utils.Substr(*faskesSaya.Wilayah, 0, 9)+"0000" == *posyandu.Puskesmas.Wilayah {
						valid = true
					}
				case 4:
					if utils.Substr(*faskesSaya.Wilayah, 0, 9)+"0000" == utils.Substr(*posyandu.Puskesmas.Wilayah, 0, 9)+"0000" {
						valid = true
					}
				}
			} else {
				// Untuk RS, jangan bandingkan wilayah kerja tapi rujukan (saat ini belum ada)
			}

		}

		if !valid {
			return exceptions.Forbidden.WithMessage("Tidak bisa akses data", nil)
		}
	}

	return nil
}

func (s *StuntingService) orgidPemilik(refEncounter *string, idKunjungan *string) (string, error) {
	if utils.IsFilled(refEncounter) {
		t, err := s.Repository.OrgIdByEncounterSId(*refEncounter)
		if err == nil {
			return utils.StrPtr(t), nil
		}
	}

	if utils.IsFilled(idKunjungan) {
		t, err := s.Repository.OrgIdByKunjunganId(*idKunjungan)
		if err == nil {
			return utils.StrPtr(t), nil
		}
	}

	return "", exceptions.Forbidden.Messagef("Tidak Bisa Verifikasi Akses: pemilik data tidak diketahui")
}

func (s *StuntingService) VerifyEncounterMilik(orgid, refEncounter string, bolehBelumAda bool) error {
	pemilik, err := s.Repository.OrgIdByEncounterSId(refEncounter)
	if err != nil {
		if exceptions.Classify(err).ErrorCode != exceptions.TidakDitemukan.ErrorCode {
			return err
		}
		if bolehBelumAda {
			return nil
		}
		return exceptions.Forbidden.Messagef("encounter %s belum tersimpan, kepemilikannya tidak bisa diverifikasi", refEncounter)
	}

	if utils.StrPtr(pemilik) != orgid {
		return exceptions.Forbidden.Messagef("encounter %s milik faskes lain", refEncounter)
	}

	return nil
}

func (s *StuntingService) VerifyAccessForUpdateFaskes(orgid string, jenis string, satusehat_id string) error {
	orgidRegistrar := ""
	memkey := fmt.Sprintf("%s::%s", jenis, satusehat_id)

	if og := s.Repository.GetStringByKey(memkey); og == "" {
		switch jenis {
		case "kunjungan":
			if t, err := s.Repository.OrgIdByEncounterSId(satusehat_id); err != nil {
				return exceptions.Classify(err)
			} else {
				orgidRegistrar = utils.StrPtr(t)
			}
		case "observasi":
			if ob, err := s.Repository.FindObservasiBySatusehatId(satusehat_id); err == nil {
				pemilik, err := s.orgidPemilik(ob.RefEncounter, ob.IDKunjungan)
				if err != nil {
					return err
				}
				orgidRegistrar = pemilik
			} else {
				return exceptions.Classify(err)
			}
		case "diagnosa":
			if ob, err := s.Repository.FindDiagnosaBySatusehatId(satusehat_id); err == nil {
				pemilik, err := s.orgidPemilik(ob.RefEncounter, ob.IDKunjungan)
				if err != nil {
					return err
				}
				orgidRegistrar = pemilik
			} else {
				return exceptions.Classify(err)
			}
		case "layanan":
			if ob, err := s.Repository.FindLayananBySatusehatId(satusehat_id); err == nil {
				pemilik, err := s.orgidPemilik(ob.RefEncounter, ob.IDKunjungan)
				if err != nil {
					return err
				}
				orgidRegistrar = pemilik
			} else {
				return exceptions.Classify(err)
			}
		case "rujukan":
			if ob, err := s.Repository.FindRujukanBySatusehatId(satusehat_id); err == nil {
				pemilik, err := s.orgidPemilik(ob.RefEncounter, ob.IDKunjungan)
				if err != nil {
					return err
				}
				orgidRegistrar = pemilik
			} else {
				return exceptions.Classify(err)
			}
		case "episode":
			if ob, err := s.Repository.FindEpisodeBySatusehatId(satusehat_id, true); err == nil {
				orgidRegistrar = *ob.Faskes.SatusehatID
			} else {
				return exceptions.Classify(err)
			}
		}

		s.Repository.SetStringByKey(memkey, orgidRegistrar, 30*24*time.Hour)
	} else {
		orgidRegistrar = og
	}

	if orgid != orgidRegistrar {
		logger.Info("VerifyAccessForUpdateFaskes", "orgid", orgid, "registrar", orgidRegistrar)
		return exceptions.Forbidden.WithMessage("Tidak bisa akses data", nil)
	}

	return nil
}

func (s *StuntingService) VerifyAksesOrangtua(orgid *string, o *types.Orangtua, idOrangtua, IdSatusehat *string) (*types.Orangtua, error) {
	if !utils.IsFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}
	if *orgid == "jakantro" {
		return nil, nil
	}

	if o == nil {
		// cari hanya jika memang belum dicari
		var orangtua *entity.Orangtua
		var err error
		if IdSatusehat != nil {
			orangtua, err = s.Repository.FindOrangtuaBySatusehatId(utils.StrPtr(IdSatusehat), true)
			if err != nil && errors.Is(err, sql.ErrNoRows) && utils.IsFilled(idOrangtua) {
				// fallback search by id jika id di kirim
				return s.VerifyAksesOrangtua(orgid, nil, idOrangtua, nil)
			}
		} else {
			orangtua, err = s.Repository.FindOrangTuaById(utils.StrPtr(idOrangtua), true)
		}
		if err != nil {
			return nil, exceptions.Classify(err)
		}
		o = (*types.Orangtua)(nil).FromEntity(orangtua)
	}

	if err := s.aksesWilayahAtauKlinis(*orgid, o.IdPosyandu, "", o.Id); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *StuntingService) VerifyAksesAnak(orgid *string, a *types.Anak, idAnak, satusehatId *string) (*types.Anak, error) {
	if !utils.IsFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}
	if *orgid == "jakantro" {
		return nil, nil
	}

	if a == nil {
		// cari hanya jika pemanggil belum mencarinya
		var anak *entity.Anak
		var e error
		if satusehatId != nil {
			anak, e = s.Repository.FindAnakBySatusehatId(utils.StrPtr(satusehatId), false)
			if anak == nil && idAnak != nil && (errors.Is(e, sql.ErrNoRows)) {
				// fallback ke id
				return s.VerifyAksesAnak(orgid, nil, idAnak, nil)
			}
		} else {
			anak, e = s.Repository.FindAnak(idAnak, nil, nil, nil, false)
		}
		if e != nil {
			return nil, exceptions.Classify(e)
		}
		a = (*types.Anak)(nil).FromEntity(anak)
	}

	var idPosyandu *string
	if utils.IsStrFilled(a.IdOrangtua) {
		orangtua, err := s.Repository.FindOrangTuaById(a.IdOrangtua, true)
		if err != nil {
			return nil, exceptions.Classify(err)
		}
		idPosyandu = orangtua.IDPosyandu
	}

	if err := s.aksesWilayahAtauKlinis(*orgid, idPosyandu, a.Id, ""); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *StuntingService) aksesWilayahAtauKlinis(orgid string, idPosyandu *string, idAnak, idOrangtua string) error {
	errWilayah := s.VerifyAccess(idPosyandu, &orgid)
	if errWilayah == nil {
		return nil
	}

	ada, err := s.Repository.AksesKlinis(orgid, idAnak, idOrangtua)
	if err != nil {
		return err
	}
	if ada {
		return nil
	}

	return errWilayah
}
