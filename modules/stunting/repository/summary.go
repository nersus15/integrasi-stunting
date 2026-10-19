package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
)

// anak beserta seluruh riwayatnya, turunan sudah dibagi ke kunjungan masing-masing
func (d *StuntingRepository) SummaryAnak(id string) (*entity.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keySummaryAnak(id)
	if cached := new(entity.Anak); d.Memory.Get(memkey, cached) {
		return cached, nil
	}

	anak := new(entity.Anak)

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	err := bunDB.NewSelect().Model(anak).
		Relation("Kunjungan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("kj.tanggal_pengukuran ASC")
		}).
		Relation("Kunjungan.Faskes").
		Relation("Kunjungan.Episode").
		Relation("Kunjungan.AtasRujukan").
		Relation("Observasi", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("ob.tanggal ASC", "ob.kode ASC")
		}).
		Relation("Diagnosa", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("d.tanggal_catat ASC")
		}).
		Relation("Layanan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("l.tanggal ASC")
		}).
		Relation("Rujukan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("rj.tanggal ASC")
		}).
		Relation("Kesehatan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("k.tanggal_pemantauan ASC")
		}).
		Relation("Episode", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("ep.mulai ASC")
		}).
		Where("a.id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Anak dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Summary Anak (id: %s): ", id) + err.Error())
		return nil, err
	}
	bagiTurunanKeKunjungan(anak)

	d.Memory.Set(memkey, anak, ttlPendek)
	return anak, nil
}

func bagiTurunanKeKunjungan(anak *entity.Anak) {
	kunjungan := make(map[string]*entity.Kunjungan, len(anak.Kunjungan))

	for _, k := range anak.Kunjungan {
		kunjungan[k.ID] = k
	}

	tempat := func(idkunjungan *string) (k *entity.Kunjungan, simpan bool) {
		if idkunjungan == nil {
			return nil, true
		}
		k, ada := kunjungan[*idkunjungan]
		return k, ada
	}

	induk := map[string]*entity.Observasi{}
	var orphanObs []*entity.Observasi

	for _, o := range anak.Observasi {
		if o.IDInduk != nil {
			continue
		}
		k, simpan := tempat(o.IDKunjungan)
		if !simpan {
			continue
		}

		induk[o.ID] = o
		if k != nil {
			k.Observasi = append(k.Observasi, o)
		} else {
			// observasi tanpa kunjungan masuk ke anak.observasi
			orphanObs = append(orphanObs, o)
		}
	}

	// Observasi.component, disimpan di tabel observasi sebagai child dari observation
	for _, o := range anak.Observasi {
		// jika o.IDInduk == nil berarti ini resource observation bukan observation.component
		if o.IDInduk == nil {
			continue
		}
		// component dimasukkan ke induknya
		if p, ada := induk[*o.IDInduk]; ada {
			p.Component = append(p.Component, o)
		}
	}

	for _, o := range induk {
		slices.SortStableFunc(o.Component, func(a, b *entity.Observasi) int {
			return strings.Compare(a.Kode, b.Kode)
		})
	}
	anak.Observasi = orphanObs

	var orphanDiagnosa []*entity.Diagnosa
	for _, d := range anak.Diagnosa {
		if k, simpan := tempat(d.IDKunjungan); k != nil {
			k.Diagnosa = append(k.Diagnosa, d)
		} else if simpan {
			orphanDiagnosa = append(orphanDiagnosa, d)
		}
	}
	anak.Diagnosa = orphanDiagnosa

	// layanan & rujukan: daftar datar anak tetap memuat yang menggantung
	var layanan []*entity.Layanan
	for _, l := range anak.Layanan {
		k, simpan := tempat(l.IDKunjungan)
		if !simpan {
			continue
		}
		if k != nil {
			k.Layanan = append(k.Layanan, l)
		}
		layanan = append(layanan, l)
	}
	anak.Layanan = layanan

	var rujukan []*entity.Rujukan
	for _, r := range anak.Rujukan {
		k, simpan := tempat(r.IDKunjungan)
		if !simpan {
			continue
		}
		if k != nil {
			k.Rujukan = append(k.Rujukan, r)
		}
		rujukan = append(rujukan, r)
	}
	anak.Rujukan = rujukan
}
