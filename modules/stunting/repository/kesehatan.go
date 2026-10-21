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

func (d *StuntingRepository) UpdateKesehatanById(id string, input *entity.Kesehatan, cariDulu bool) (*entity.Kesehatan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	if cariDulu {
		if _, err := d.FindKesehatanById(id); err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	_, err := d.Connection.UpdateOne(ctx, entity.Kesehatan{}.TableName(), filter, input)

	if err != nil {
		return nil, err
	}

	d.lupakan(keyKesehatanId(id))
	d.LupakanRiwayatAnak(input.IDAnak)

	return input, nil
}

// bun langsung, lib-sql tidak mendukung relasi has-many
func (d *StuntingRepository) ListKesehatanAnak(id string) (*entity.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyListKesehatan(id)
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
		Relation("Kesehatan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("k.tanggal_pemantauan ASC")
		}).
		Where("a.id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Anak dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("List Kesehatan Anak (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	d.Memory.Set(memkey, anak, ttlPanjang)
	return anak, nil
}

func (d *StuntingRepository) FindKesehatanById(id string) (*entity.Kesehatan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyKesehatanId(id)
	if cached := new(entity.Kesehatan); d.Memory.Get(memkey, cached) {
		return cached, nil
	}

	tmp := new(entity.Kesehatan)
	filter := []port.DbExpression{
		{Expr: "id = ?", Args: []any{id}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Kesehatan{}.TableName(), []string{}, filter, map[string]int{})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Kesehatan dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Find Kesehatan (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	d.Memory.Set(memkey, tmp, ttlPanjang)
	return tmp, nil
}

func (d *StuntingRepository) CreateKesehatan(kesehatan *entity.Kesehatan, ctx context.Context) (*entity.Kesehatan, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if _, err := d.Connection.InsertOne(ctx, kesehatan.TableName(), kesehatan); err != nil {
		return nil, err
	}
	if !utils.IsStrFilled(kesehatan.ID) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	d.LupakanRiwayatAnak(kesehatan.IDAnak)
	d.lupakan(keyKesehatanId(kesehatan.ID))
	return kesehatan, nil
}

func (d *StuntingRepository) DeleteKesehatanById(id, orgid string) error {
	ctx, cancel := context.WithTimeout(d.Context.Context, 5*time.Second)
	defer cancel()

	filter := []port.DbExpression{{Expr: "id", Args: []any{id}}, belumDihapus}
	input := map[string]any{"deletedAt": time.Now(), "deleted_by": orgid}
	affected, err := d.Connection.UpdateOne(ctx, entity.Kesehatan{}.TableName(), filter, input)
	if err != nil {
		return err
	}
	if affected < 1 {
		return exceptions.TidakDitemukan.Messagef("Kesehatan dengan id #%s tidak ditemukan", id)
	}
	return nil
}

func CreateKesehatanTx(ctx context.Context, tx bun.IDB, kesehatan *entity.Kesehatan) (*entity.Kesehatan, error) {
	if kesehatan == nil {
		return nil, nil
	}
	if _, err := tx.NewInsert().Model(kesehatan).Exec(ctx); err != nil {
		return nil, err
	}
	return kesehatan, nil
}
