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

func (d *StuntingRepository) UpdateOrangTuaById(orangtua *entity.Orangtua, ctx context.Context, cariDulu bool) (*entity.Orangtua, error) {
	if orangtua == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if !utils.IsStrFilled(orangtua.ID) {
		return nil, exceptions.Validasi.Messagef("id orangtua tidak boleh kosong")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if cariDulu {
		if _, err := d.FindOrangTuaById(orangtua.ID, false); err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	filter := []port.DbExpression{{Expr: "id", Args: []any{orangtua.ID}}}
	if _, err := d.Connection.UpdateOne(ctx, orangtua.TableName(), filter, orangtua); err != nil {
		return nil, err
	}

	d.LupakanOrangtua(orangtua)
	return orangtua, nil
}

func (d *StuntingRepository) FindOrangtuaBySatusehatId(satusehatId string, loadcache bool) (*entity.Orangtua, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id orang tua tidak boleh kosong")
	}
	filter := []port.DbExpression{{Expr: "satusehat_id = ?", Args: []any{satusehatId}}, belumDihapus}
	return d.cariOrangtua(entity.Orangtua{}.TableName(), filter, keyOrangtuaSatusehat(satusehatId), loadcache, "FindOrangtuaBySatusehatId",
		exceptions.TidakDitemukan.WithMessage("Data Orangtua tidak ditemukan", sql.ErrNoRows))
}

func (d *StuntingRepository) CreateOrangTua(orangtua *entity.Orangtua, ctx context.Context) (*entity.Orangtua, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	// satu pencarian: aktif dipakai ulang, terhapus dipulihkan, belum ada di-insert
	lama, err := d.FindOrangTuaByNikWithDeleted(ctx, orangtua.NIK)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if lama != nil && lama.DeletedAt == nil {
		return lama, nil
	}
	if lama != nil {
		bunDB, ok := d.Connection.GetConnection().(*bun.DB)
		if !ok {
			return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
		}
		pulih, err := pulihkanOrangtuaTerhapus(ctx, bunDB, lama, orangtua)
		if err != nil {
			return nil, err
		}
		d.LupakanOrangtua(pulih)
		return pulih, nil
	}

	if _, err := d.Connection.InsertOne(ctx, orangtua.TableName(), orangtua); err != nil {
		return nil, err
	}
	if !utils.IsStrFilled(orangtua.ID) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}
	d.LupakanOrangtua(orangtua)
	return orangtua, nil
}

func CreateOrangTuaTx(ctx context.Context, tx bun.IDB, orangtua *entity.Orangtua) (*entity.Orangtua, error) {
	if orangtua == nil {
		return nil, nil
	}

	// satu pencarian: aktif dipakai ulang, terhapus dipulihkan, belum ada di-insert
	lama := new(entity.Orangtua)
	err := tx.NewSelect().Model(lama).WhereAllWithDeleted().Where("nik = ?", orangtua.NIK).Limit(1).Scan(ctx)
	switch {
	case err == nil && lama.DeletedAt == nil:
		return lama, nil
	case err == nil:
		return pulihkanOrangtuaTerhapus(ctx, tx, lama, orangtua)
	case !errors.Is(err, sql.ErrNoRows):
		return nil, exceptions.Internal.WithMessage("gagal mencari data orangtua", err)
	}

	if _, err := tx.NewInsert().Model(orangtua).Exec(ctx); err != nil {
		return nil, err
	}
	return orangtua, nil
}

// nik milik orangtua terhapus: baris lama dipulihkan dengan data baru
func pulihkanOrangtuaTerhapus(ctx context.Context, tx bun.IDB, lama, baru *entity.Orangtua) (*entity.Orangtua, error) {
	lama.Override(*baru, false)
	now := time.Now()
	lama.DeletedAt, lama.DeletedBy, lama.UpdatedAt = nil, nil, &now

	if _, err := tx.NewUpdate().Model(lama).WherePK().WhereAllWithDeleted().Exec(ctx); err != nil {
		return nil, err
	}
	return lama, nil
}

func (d *StuntingRepository) FindOrangTuaByNik(nik string, loadCache bool) (*entity.Orangtua, error) {
	filter := []port.DbExpression{{Expr: "nik = ?", Args: []any{nik}}, belumDihapus}
	return d.cariOrangtua(entity.Orangtua{}.TableName(), filter, keyOrangtuaNik(nik), loadCache, "FindOrangTuaByNik", exceptions.ErrTidakDitemukan)
}

// entity, karena pemulihan meng-UPDATE seluruh kolom
func (d *StuntingRepository) FindOrangTuaByNikWithDeleted(ctx context.Context, nik string) (*entity.Orangtua, error) {
	temp := new(entity.Orangtua)
	filter := []port.DbExpression{{Expr: "nik = ?", Args: []any{nik}}}

	// FindOne dengan nama tabel tidak memakai tag soft_delete, jadi baris terhapus ikut terbaca
	if err := d.Connection.FindOne(ctx, temp, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{}); err != nil {
		return nil, err
	}
	return temp, nil
}

func (d *StuntingRepository) FindOrangTuaById(id string, loadCache bool) (*entity.Orangtua, error) {
	if !utils.IsStrFilled(id) {
		return nil, exceptions.KolomWajib.Messagef("Id orangtua tidak boleh kosong")
	}
	filter := []port.DbExpression{{Expr: "id", Args: []any{id}}}
	return d.cariOrangtua("[me]", filter, keyOrangtuaId(id), loadCache, "FindOrangTuaById", exceptions.ErrTidakDitemukan)
}

// orangtua aktif menurut filter, cache opsional
func (d *StuntingRepository) cariOrangtua(tabel string, filter []port.DbExpression, memkey string, loadCache bool, label string, tidakAda error) (*entity.Orangtua, error) {
	orangtua := new(entity.Orangtua)
	if loadCache && d.Memory.Get(memkey, orangtua) {
		return orangtua, nil
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	err := d.Connection.FindOne(ctx, orangtua, tabel, []string{"*"}, filter, map[string]int{})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, tidakAda
	}
	if err != nil {
		logger.Error(fmt.Sprintf("%s (%v): ", label, filter[0].Args) + helper.ToLogJSON(err))
		return nil, err
	}

	if loadCache {
		d.Memory.Set(memkey, orangtua, ttlPanjang)
	}
	return orangtua, nil
}

// baris terhapus tidak pernah masuk cache; cache hanya untuk data aktif
func (d *StuntingRepository) FindDeletedOrangtua(id string) (*entity.Orangtua, error) {
	if !utils.IsStrFilled(id) {
		return nil, exceptions.KolomWajib.Messagef("Id orangtua tidak boleh kosong")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	orangtua := new(entity.Orangtua)
	filter := []port.DbExpression{{Expr: "id", Args: []any{id}}}
	err := d.Connection.FindOneDeleted(ctx, orangtua, "[me]", []string{"*"}, filter, map[string]int{})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindDeletedOrangtua (%v): ", id) + err.Error())
		return nil, err
	}
	return orangtua, nil
}

func (d *StuntingRepository) DeleteOrangTua(id *string) error {
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	_, err := d.Connection.DeleteOne(d.Context.Context, entity.Orangtua{}.TableName(), filter)

	if err != nil {
		logger.Error(fmt.Sprintf("Delete Orangtua (%s): ", *id) + helper.ToLogJSON(err))
		return err
	}

	d.lupakan(keyOrangtuaId(*id), keyListAnak(*id))

	return nil
}
func (d *StuntingRepository) DeleteOrangtuaByID(id string, orgid string) error {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*5)
	defer cancel()

	deletedAt := time.Now()
	deletedBy := orgid

	input := map[string]any{
		"deletedAt":  deletedAt,
		"deleted_by": deletedBy,
	}

	filter := []port.DbExpression{
		{Expr: "id", Args: []any{id}},
	}
	filter = append(filter, belumDihapus)

	affected, err := d.Connection.UpdateOne(ctx, entity.Orangtua{}.TableName(), filter, input)

	if err != nil {
		return err
	}

	if affected < 1 {
		return exceptions.TidakDitemukan.Messagef("Orangtua dengan id #%s tidak ditemukan", id)
	}

	return nil
}
