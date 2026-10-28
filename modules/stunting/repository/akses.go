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

// orgid tujuan rujukan keluar/rujuk balik terakhir; ada false kalau anak belum pernah dirujuk
func (d *StuntingRepository) TujuanRujukanTerakhir(idAnak string) (tujuan string, ada bool, err error) {
	semua, err := d.TujuanRujukan([]string{idAnak})
	if err != nil {
		return "", false, err
	}
	tujuan, ada = semua[idAnak]
	return tujuan, ada, nil
}

func (d *StuntingRepository) TujuanRujukan(idAnak []string) (map[string]string, error) {
	tujuan := make(map[string]string, len(idAnak))

	// cache per anak: "-" belum pernah dirujuk, "+orgid" sudah
	var belum []string
	for _, id := range idAnak {
		if v := d.GetStringByKey(keyRujukanTerakhir(id)); v != "" {
			if v != "-" {
				tujuan[id] = strings.TrimPrefix(v, "+")
			}
			continue
		}
		belum = append(belum, id)
	}
	if len(belum) == 0 {
		return tujuan, nil
	}

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	var hasil []struct {
		IdAnak string `bun:"id_anak"`
		Tujuan string `bun:"tujuan"`
	}
	// model memberi alias rj dan filter soft delete
	err := bunDB.NewSelect().
		Model((*entity.Rujukan)(nil)).
		DistinctOn("rj.id_anak").
		ColumnExpr("rj.id_anak").
		ColumnExpr("coalesce(f.satusehat_id, rj.ref_faskes_tujuan) AS tujuan").
		Join("LEFT JOIN stunting.faskes AS f ON f.id = rj.id_faskes_tujuan").
		Where("rj.id_anak IN (?)", bun.List(belum)).
		Where("rj.jenis IN (?)", bun.List([]string{entity.RujukanKeluar, entity.RujukBalik})).
		Where("coalesce(f.satusehat_id, rj.ref_faskes_tujuan) IS NOT NULL").
		OrderExpr(`rj.id_anak, coalesce(rj.tanggal, rj."createdAt") DESC, rj."createdAt" DESC`).
		Scan(ctx, &hasil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error(fmt.Sprintf("TujuanRujukan (%v): ", belum) + err.Error())
		return nil, exceptions.Classify(err)
	}

	for _, h := range hasil {
		tujuan[h.IdAnak] = h.Tujuan
	}
	for _, id := range belum {
		simpan := "-"
		if t, ok := tujuan[id]; ok {
			simpan = "+" + t
		}
		d.SetStringByKey(keyRujukanTerakhir(id), simpan, 24*time.Hour)
	}
	return tujuan, nil
}
