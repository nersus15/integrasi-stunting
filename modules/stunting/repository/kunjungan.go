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
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

func (d *StuntingRepository) CreateKunjungan(kunjungan *entity.Kunjungan, ctx context.Context) (*entity.Kunjungan, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, kunjungan.TableName(), kunjungan); err != nil {
		return nil, err
	}
	if !utils.IsStrFilled(kunjungan.ID) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	d.LupakanRiwayatAnak(kunjungan.IDAnak)
	d.lupakan(keyKunjunganId(kunjungan.ID))
	return kunjungan, nil
}

func CreateKunjunganTx(ctx context.Context, tx bun.IDB, kunjungan *entity.Kunjungan) (*entity.Kunjungan, error) {
	if kunjungan == nil {
		return nil, nil
	}
	if _, err := tx.NewInsert().Model(kunjungan).Exec(ctx); err != nil {
		return nil, err
	}
	return kunjungan, nil
}

// anak beserta kunjungannya
func (d *StuntingRepository) ListKunjunganAnak(id string) (*entity.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyListKunjungan(id)
	if cached := new(entity.Anak); d.Memory.Get(memkey, cached) {
		return cached, nil
	}

	anak := new(entity.Anak)
	filter := []port.DbExpression{
		{
			Expr: "a.id = ?",
			Args: []any{id},
		},
	}
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	query := bunDB.NewSelect().Model(anak).
		Relation("Kunjungan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("kj.createdAt ASC")
		})

	for _, f := range filter {
		query.Where(f.Expr, f.Args[0])
	}

	err := query.Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Anak dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("List Kunjungan Anak (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}
	d.Memory.Set(memkey, anak, ttlSedang)
	return anak, nil
}

func (d *StuntingRepository) UpdateKunjunganById(id string, input *entity.Kunjungan, cariDulu bool) (*entity.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	if cariDulu {
		if _, err := d.FindKunjunganById(id); err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	_, err := d.Connection.UpdateOne(ctx, entity.Kunjungan{}.TableName(), filter, input)

	if err != nil {
		return nil, err
	}

	d.lupakan(keyKunjunganId(id))
	d.LupakanRiwayatAnak(input.IDAnak)

	return input, nil
}
func (d *StuntingRepository) FindKunjunganById(id string) (*entity.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyKunjunganId(id)
	if cached := new(entity.Kunjungan); d.Memory.Get(memkey, cached) {
		return cached, nil
	}

	tmp := new(entity.Kunjungan)
	filter := []port.DbExpression{
		{Expr: "id = ?", Args: []any{id}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{}, filter, map[string]int{})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Kunjungan dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Find Kunjungan (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	d.Memory.Set(memkey, tmp, ttlSedang)
	return tmp, nil
}

func (d *StuntingRepository) FindKunjunganByIdAnak(idanak string) (*entity.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	tmp := new(entity.Kunjungan)
	filter := []port.DbExpression{
		{Expr: "id_anak = ?", Args: []any{idanak}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{}, filter, map[string]int{"createdAt": 0})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Kunjungan dengan id %s tidak ditemukan", idanak)
		}
		logger.Error(fmt.Sprintf("Find Kunjungan (id: %s): ", idanak) + err.Error())
		return nil, err
	}

	return tmp, nil
}

// stunting NULL berarti belum ditentukan, dilewati
func (d *StuntingRepository) FindStatusStuntingTerakhir(idanak string) (*entity.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	tmp := new(entity.Kunjungan)
	filter := []port.DbExpression{
		{Expr: `id_anak = ? AND stunting IS NOT NULL AND "deletedAt" IS NULL`, Args: []any{idanak}},
	}

	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{}, filter, map[string]int{"createdAt": 0})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Status stunting anak %s belum ada", idanak)
		}
		logger.Error(fmt.Sprintf("FindStatusStuntingTerakhir (%s): ", idanak) + err.Error())
		return nil, err
	}

	return tmp, nil
}

// hanya kolom yang dikirim yang di-SET; nilai nil ditulis NULL
func (d *StuntingRepository) UpdateUkurKunjungan(id string, kolom map[string]any) error {
	if len(kolom) == 0 {
		return nil
	}

	lama, err := d.FindKunjunganById(id)
	if err != nil {
		return err
	}

	kolom["updatedAt"] = time.Now()
	filter := []port.DbExpression{
		{Expr: `id = ? AND "deletedAt" IS NULL`, Args: []any{id}},
	}
	if _, err := d.Connection.Update(d.Context.Context, entity.Kunjungan{}.TableName(), filter, kolom); err != nil {
		logger.Error(fmt.Sprintf("UpdateUkurKunjungan (%s): ", id) + err.Error())
		return err
	}

	d.lupakan(keyKunjunganId(id))
	d.LupakanRiwayatAnak(lama.IDAnak)

	return nil
}

func (d *StuntingRepository) DeleteKunjunganById(id, orgid string) error {
	filter := []port.DbExpression{
		{Expr: "id", Args: []any{id}},
		{Expr: `"deletedAt"`, Args: []any{nil}},
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 5*time.Second)
	defer cancel()

	input := map[string]any{
		"deletedAt":  time.Now(),
		"deleted_by": orgid,
	}
	affected, err := d.Connection.UpdateOne(ctx, entity.Kunjungan{}.TableName(), filter, input)

	if err != nil {
		return exceptions.Classify(err)
	}
	if affected < 1 {
		return exceptions.TidakDitemukan.Messagef("Kunjungan dengan id #%s tidak ditemukan", id)
	}

	return nil
}
