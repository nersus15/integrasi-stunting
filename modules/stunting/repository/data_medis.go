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
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

func (d *StuntingRepository) CreateObservasi(observasi *entity.Observasi, ctx context.Context) (*entity.Observasi, error) {
	if observasi == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, observasi.TableName(), observasi); err != nil {
		return nil, err
	}
	return observasi, nil
}

func (d *StuntingRepository) CreateDiagnosa(diagnosa *entity.Diagnosa, ctx context.Context) (*entity.Diagnosa, error) {
	if diagnosa == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, diagnosa.TableName(), diagnosa); err != nil {
		return nil, err
	}
	return diagnosa, nil
}

func (d *StuntingRepository) CreateLayanan(layanan *entity.Layanan, ctx context.Context) (*entity.Layanan, error) {
	if layanan == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, layanan.TableName(), layanan); err != nil {
		return nil, err
	}
	return layanan, nil
}

// mengisi id_kunjungan pada baris yang menunggu Encounter-nya, idempoten
func (d *StuntingRepository) HubungkanKeKunjungan(tabel string, idKunjungan string, refEncounter string, ctx context.Context) (int64, error) {
	if !utils.IsStrFilled(idKunjungan) || !utils.IsStrFilled(refEncounter) {
		return 0, exceptions.Validasi.Messagef("id kunjungan dan ref encounter tidak boleh kosong")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	filter := []port.DbExpression{
		{Expr: `ref_encounter = ? AND id_kunjungan IS NULL AND "deletedAt" IS NULL`, Args: []any{refEncounter}},
	}

	data := map[string]any{"id_kunjungan": idKunjungan, "updatedAt": time.Now()}

	return d.Connection.Update(ctx, tabel, filter, data)
}

func (d *StuntingRepository) CreateEpisode(episode *entity.Episode, ctx context.Context) (*entity.Episode, error) {
	if episode == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, episode.TableName(), episode); err != nil {
		return nil, err
	}
	return episode, nil
}

type DataMedis struct {
	AnakBaru     *entity.Anak
	Kunjungan    *entity.Kunjungan
	RefEncounter *string
	Episode      []*entity.Episode
	Observasi    []*entity.Observasi
	Komponen     []*entity.Observasi
	Diagnosa     []*entity.Diagnosa
	Layanan      []*entity.Layanan
	Rujukan      []*entity.Rujukan

	IdKunjunganStatus *string
	Stunting          *int

	IdKunjunganUkur *string
	KolomUkur       map[string]any
}

var tabelTurunan = []string{"stunting.observasi", "stunting.diagnosa", "stunting.layanan", "stunting.rujukan"}

// UPDATE di dalam WITH tetap jalan walau hasilnya tidak dibaca
var sapuTurunan = func() string {
	bagian := make([]string, len(tabelTurunan))
	for i, tabel := range tabelTurunan {
		bagian[i] = fmt.Sprintf(`t%d AS (UPDATE %s SET id_kunjungan = ?, "updatedAt" = ? WHERE ref_encounter = ? AND id_kunjungan IS NULL AND "deletedAt" IS NULL)`, i, tabel)
	}
	return "WITH " + strings.Join(bagian, ", ") + " SELECT 1"
}()

func (d *StuntingRepository) SimpanDataMedis(data *DataMedis, ctx context.Context) error {
	if data == nil {
		return exceptions.BodyRusak.New(nil)
	}

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 30*time.Second)
		defer cancel()
	}

	err := bunDB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if data.AnakBaru != nil {
			if _, err := tx.NewInsert().Model(data.AnakBaru).Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}

		if len(data.Episode) > 0 {
			if _, err := tx.NewInsert().Model(&data.Episode).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}

		if data.Kunjungan != nil {
			if _, err := tx.NewInsert().Model(data.Kunjungan).Returning("NULL").Exec(ctx); err != nil {
				return err
			}

			// sapu turunan yang lebih dulu tiba di bundle lain
			if utils.IsFilled(data.RefEncounter) {
				now := time.Now()
				args := make([]any, 0, 3*len(tabelTurunan))
				for range tabelTurunan {
					args = append(args, data.Kunjungan.ID, now, *data.RefEncounter)
				}
				if _, err := tx.NewRaw(sapuTurunan, args...).Exec(ctx); err != nil {
					return err
				}
			}
		}

		// kunjungan yang lebih dulu tiba menunggu episode-nya
		for _, ep := range data.Episode {
			if ep.SatusehatId == nil {
				continue
			}
			if _, err := tx.NewUpdate().Table("stunting.kunjungan").
				Set("id_episode = ?", ep.ID).
				Set(`"updatedAt" = ?`, time.Now()).
				Where(`ref_episode = ? AND id_episode IS NULL AND "deletedAt" IS NULL`, *ep.SatusehatId).
				Exec(ctx); err != nil {
				return err
			}
		}

		if data.IdKunjunganStatus != nil && data.Stunting != nil {
			if _, err := tx.NewUpdate().Table("stunting.kunjungan").
				Set("stunting = ?", *data.Stunting).
				Set(`"updatedAt" = ?`, time.Now()).
				Where(`id = ? AND "deletedAt" IS NULL`, *data.IdKunjunganStatus).
				Exec(ctx); err != nil {
				return err
			}
		}

		if data.IdKunjunganUkur != nil {
			if kolom := data.KolomUkur; len(kolom) > 0 {
				q := tx.NewUpdate().Table("stunting.kunjungan").Set(`"updatedAt" = ?`, time.Now())
				for k, v := range kolom {
					q = q.Set("? = ?", bun.Ident(k), v)
				}
				if _, err := q.Where(`id = ? AND "deletedAt" IS NULL`, *data.IdKunjunganUkur).Exec(ctx); err != nil {
					return err
				}
			}
		}

		// induk lebih dulu, komponen menunjuk ke id_induk
		if len(data.Observasi) > 0 {
			if _, err := tx.NewInsert().Model(&data.Observasi).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}
		if len(data.Komponen) > 0 {
			if _, err := tx.NewInsert().Model(&data.Komponen).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}
		if len(data.Diagnosa) > 0 {
			if _, err := tx.NewInsert().Model(&data.Diagnosa).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}
		if len(data.Layanan) > 0 {
			if _, err := tx.NewInsert().Model(&data.Layanan).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}
		if len(data.Rujukan) > 0 {
			if _, err := tx.NewInsert().Model(&data.Rujukan).On("CONFLICT DO NOTHING").Returning("NULL").Exec(ctx); err != nil {
				return err
			}
		}

		// setelah rujukan ada barisnya: sapu kunjungan yang menunggunya
		for _, rj := range data.Rujukan {
			if rj.SatusehatId == nil {
				continue
			}
			if _, err := tx.NewUpdate().Table("stunting.kunjungan").
				Set("id_rujukan = ?", rj.ID).
				Set(`"updatedAt" = ?`, time.Now()).
				Where(`ref_rujukan = ? AND id_rujukan IS NULL AND "deletedAt" IS NULL`, *rj.SatusehatId).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// sapuan bisa mengenai kunjungan lain, jadi seluruh riwayat anak dibuang
	d.LupakanRiwayatAnak(idAnakDataMedis(data))
	if data.Kunjungan != nil {
		d.lupakan(keyKunjunganId(data.Kunjungan.ID))
	}
	if data.IdKunjunganStatus != nil {
		d.lupakan(keyKunjunganId(*data.IdKunjunganStatus))
	}

	return nil
}

// kunjungan bisa nil kalau data medis tiba lebih dulu
func idAnakDataMedis(data *DataMedis) string {
	if data.Kunjungan != nil && data.Kunjungan.IDAnak != "" {
		return data.Kunjungan.IDAnak
	}
	for _, o := range data.Observasi {
		if o.IDAnak != "" {
			return o.IDAnak
		}
	}
	for _, dg := range data.Diagnosa {
		if dg.IDAnak != "" {
			return dg.IDAnak
		}
	}
	for _, l := range data.Layanan {
		if l.IDAnak != "" {
			return l.IDAnak
		}
	}
	for _, r := range data.Rujukan {
		if r.IDAnak != "" {
			return r.IDAnak
		}
	}
	for _, e := range data.Episode {
		if e.IDAnak != "" {
			return e.IDAnak
		}
	}
	return ""
}

// rujukan pertama anak beserta faskes tujuannya
func (d *StuntingRepository) FindRujukanByIdAnak(idanak string) (*entity.Rujukan, error) {
	if !utils.IsStrFilled(idanak) {
		return nil, exceptions.Validasi.Messagef("id anak tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Rujukan)
	filter := []port.DbExpression{
		{Expr: "id_anak = ?", Args: []any{idanak}},
	}

	err := d.Connection.FindOne(ctx, tmp, "[me],FaskesTujuan", []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Rujukan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindRujukanByIdAnak (%v): ", idanak) + err.Error())
		return nil, err
	}
	return tmp, nil
}

func (d *StuntingRepository) CreateRujukan(rujukan *entity.Rujukan, ctx context.Context) (*entity.Rujukan, error) {
	if rujukan == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, rujukan.TableName(), rujukan); err != nil {
		return nil, err
	}
	return rujukan, nil
}
