package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
)

// orgid tujuan rujukan keluar/rujuk balik terakhir; ada false kalau anak belum pernah dirujuk.
// rujukan tanpa tujuan tidak memindah akses ke siapa pun, jadi tidak dihitung
func (d *StuntingRepository) TujuanRujukanTerakhir(idAnak string) (tujuan string, ada bool, err error) {
	// "-" belum pernah dirujuk, "+orgid" sudah
	memkey := keyRujukanTerakhir(idAnak)
	if v := d.GetStringByKey(memkey); v != "" {
		return strings.TrimPrefix(v, "+"), v != "-", nil
	}

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return "", false, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	var hasil []string
	err = bunDB.NewRaw(`SELECT coalesce(f.satusehat_id, r.ref_faskes_tujuan) FROM stunting.rujukan r
		LEFT JOIN stunting.faskes f ON f.id = r.id_faskes_tujuan
		WHERE r.id_anak = ? AND r.jenis IN (?, ?) AND r."deletedAt" IS NULL
			AND coalesce(f.satusehat_id, r.ref_faskes_tujuan) IS NOT NULL
		ORDER BY coalesce(r.tanggal, r."createdAt") DESC, r."createdAt" DESC
		LIMIT 1`, idAnak, entity.RujukanKeluar, entity.RujukBalik).Scan(ctx, &hasil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error(fmt.Sprintf("TujuanRujukanTerakhir (%s): ", idAnak) + err.Error())
		return "", false, exceptions.Classify(err)
	}

	simpan := "-"
	if len(hasil) > 0 {
		tujuan, ada, simpan = hasil[0], true, "+"+hasil[0]
	}
	d.SetStringByKey(memkey, simpan, 24*time.Hour)
	return tujuan, ada, nil
}

func (d *StuntingRepository) IdAnakOrangtua(idOrangtua string) ([]string, error) {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	var id []string
	err := bunDB.NewRaw(`SELECT id FROM stunting.anak WHERE id_orangtua = ? AND "deletedAt" IS NULL`, idOrangtua).Scan(ctx, &id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error(fmt.Sprintf("IdAnakOrangtua (%s): ", idOrangtua) + err.Error())
		return nil, exceptions.Classify(err)
	}
	return id, nil
}
