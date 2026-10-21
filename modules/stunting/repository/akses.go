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
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
)

func (d *StuntingRepository) AksesKlinis(orgid string, idAnak string, idOrangtua string) (bool, error) {
	if res, found := d.GetBool(keyAksesKlinis(orgid, idAnak, idOrangtua)); found {
		return res, nil
	}

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return false, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	var ada bool
	err := bunDB.NewRaw(`SELECT EXISTS (
		SELECT 1 FROM stunting.rujukan r
		JOIN stunting.anak a ON a.id = r.id_anak AND a."deletedAt" IS NULL
		JOIN stunting.faskes f ON f.id = r.id_faskes_tujuan
		WHERE f.satusehat_id = ? AND r.jenis = ? AND r."deletedAt" IS NULL
			AND (a.id = ? OR a.id_orangtua = ?)
			AND NOT EXISTS (
				SELECT 1 FROM stunting.rujukan b
				WHERE b.id_anak = r.id_anak AND b.jenis = ? AND b.id_faskes_asal = r.id_faskes_tujuan
					AND b."deletedAt" IS NULL
					AND coalesce(b.tanggal, b."createdAt") >= coalesce(r.tanggal, r."createdAt"))
	) OR EXISTS (
		SELECT 1 FROM stunting.kunjungan k
		JOIN stunting.anak a ON a.id = k.id_anak AND a.id_orangtua IS NULL
		JOIN stunting.faskes f ON f.id = k.id_faskes
		WHERE f.satusehat_id = ? AND a.id = ? AND k."deletedAt" IS NULL
	)`, orgid, entity.RujukanKeluar, idAnak, idOrangtua, entity.RujukBalik, orgid, idAnak).Scan(ctx, &ada)

	if err != nil {
		logger.Error(fmt.Sprintf("AksesKlinis (%s): ", orgid) + err.Error())
		return false, exceptions.Classify(err)
	}

	d.SetBool(keyAksesKlinis(orgid, idAnak, idOrangtua), ada, 24*time.Hour)
	return ada, nil
}

// hanya saat id_orangtua anak berubah atau anak dihapus
func (d *StuntingRepository) LupakanAksesPerubahanOrangtua(idAnak string, idOrangtua ...string) {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok || !utils.IsStrFilled(idAnak) {
		return
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 5*time.Second)
	defer cancel()

	var orgids []string
	err := bunDB.NewRaw(`SELECT DISTINCT f.satusehat_id FROM stunting.faskes f
		WHERE f.satusehat_id IS NOT NULL AND f.id IN (
			SELECT id_faskes_asal FROM stunting.rujukan WHERE id_anak = ?
			UNION SELECT id_faskes_tujuan FROM stunting.rujukan WHERE id_anak = ?
			UNION SELECT id_faskes FROM stunting.kunjungan WHERE id_anak = ?)`, idAnak, idAnak, idAnak).Scan(ctx, &orgids)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("LupakanAksesPerubahanOrangtua (" + idAnak + "): " + err.Error())
		return
	}

	for _, orgid := range orgids {
		d.LupakanAksesKlinis(orgid, idAnak, "")
		for _, o := range idOrangtua {
			if utils.IsStrFilled(o) {
				d.LupakanAksesKlinis(orgid, "", o)
			}
		}
	}
}
