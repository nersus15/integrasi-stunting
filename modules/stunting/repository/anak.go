package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

func (d *StuntingRepository) FindAnakBySatusehatId(satusehatId string, loadCache bool) (*entity.Anak, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id anak tidak boleh kosong")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	anak := new(entity.Anak)
	memkey := keyAnakSatusehat(satusehatId)
	if loadCache && d.Memory.Get(memkey, anak) {
		return anak, nil
	}

	filter := []port.DbExpression{{Expr: "satusehat_id = ?", Args: []any{satusehatId}}, belumDihapus}
	err := d.Connection.FindOne(ctx, anak, entity.Anak{}.TableName(), []string{"*"}, filter, map[string]int{})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindAnakBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	if loadCache {
		d.Memory.Set(memkey, anak, ttlPanjang)
	}
	return anak, nil
}

// mengisi id_orangtua pada anak yang masih kosong, idempoten
func (d *StuntingRepository) HubungkanAnakKeOrangTua(satusehatIdAnak string, idOrangtua string, ctx context.Context) (int64, error) {
	if !utils.IsStrFilled(satusehatIdAnak) {
		return 0, exceptions.Validasi.Messagef("satusehat id anak tidak boleh kosong")
	}
	if !utils.IsStrFilled(idOrangtua) {
		return 0, exceptions.Validasi.Messagef("id orangtua tidak boleh kosong")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	filter := []port.DbExpression{
		{
			Expr: `satusehat_id = ? AND id_orangtua IS NULL AND "deletedAt" IS NULL`,
			Args: []any{satusehatIdAnak},
		},
	}

	// map biasa: struct menimpa semua kolom, port.DbMap gagal type assertion di bun
	data := map[string]any{
		"id_orangtua": idOrangtua,
		"updatedAt":   time.Now(),
	}

	jml, err := d.Connection.Update(ctx, entity.Anak{}.TableName(), filter, data)
	if err == nil && jml > 0 {
		// id_orangtua ikut di cache anak, jadi semua kunci anak dibuang termasuk nik dan urutan
		d.lupakan(keyAnakSatusehat(satusehatIdAnak), keyListAnak(idOrangtua))
		if a, e := d.FindAnakBySatusehatId(satusehatIdAnak, false); e == nil && a != nil {
			d.LupakanAnak(a)
		}
	}

	return jml, err
}

func (d *StuntingRepository) CreateAnak(anak *entity.Anak, ctx context.Context) (*entity.Anak, error) {
	if anak == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if lama, err := d.FindAnak(nil, anak.NIK, &anak.IDOrangtua, &anak.AnakKe, false); err == nil {
		if lama.IDOrangtua != anak.IDOrangtua {
			return nil, exceptions.AnakMilikOrangtuaLain.WithMessage("Anak sudah terdaftar pada orangtua yang berbeda", err)
		}
		return lama, nil
	}

	if _, err := d.Connection.InsertOne(ctx, anak.TableName(), anak); err != nil {
		return nil, err
	}
	if !utils.IsStrFilled(anak.ID) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}
	d.LupakanAnak(anak)
	return anak, nil
}

func (d *StuntingRepository) UpdateAnakById(anak *entity.Anak, ctx context.Context, cariDulu bool) (*entity.Anak, error) {
	if anak == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	if cariDulu {
		if _, err := d.FindAnak(utils.PtrStr(anak.ID, true), anak.NIK, &anak.IDOrangtua, &anak.AnakKe, false); err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	filter := []port.DbExpression{{Expr: "id", Args: []any{anak.ID}}}
	if _, err := d.Connection.UpdateOne(ctx, anak.TableName(), filter, anak); err != nil {
		return nil, err
	}
	if !utils.IsStrFilled(anak.ID) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}
	d.LupakanAnak(anak)
	return anak, nil
}

func CreateAnakTx(ctx context.Context, tx bun.IDB, anak *entity.Anak) (*entity.Anak, error) {
	if anak == nil {
		return nil, nil
	}

	whereClause, whereArgs := "nik = ?", []any{anak.NIK}
	if !utils.IsFilled(anak.NIK) {
		if !utils.IsStrFilled(anak.IDOrangtua) || !utils.IsStrFilled(strconv.Itoa(int(anak.AnakKe))) {
			return nil, exceptions.Validasi.Messagef("untuk mencari anak sebutkan nik anak, atau id orangtua beserta urutan anak")
		}
		whereClause, whereArgs = "id_orangtua = ? AND anak_ke = ?", []any{anak.IDOrangtua, anak.AnakKe}
	}

	lama := new(entity.Anak)
	err := tx.NewSelect().Model(lama).Where(whereClause, whereArgs...).Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.NewInsert().Model(anak).Exec(ctx); err != nil {
			return nil, err
		}
		return anak, nil
	}
	if err != nil {
		return nil, exceptions.Internal.WithMessage("gagal mencari data anak", err)
	}

	if lama.IDOrangtua != anak.IDOrangtua {
		return nil, exceptions.AnakMilikOrangtuaLain.Messagef("nik anak %s sudah terdaftar pada orangtua yang berbeda", *anak.NIK)
	}
	return lama, nil
}

func (d *StuntingRepository) FindAnak(id *string, nik *string, orangtua *string, urutan *int16, loadCache bool) (*entity.Anak, error) {
	filter, memkey, err := kriteriaAnak(id, nik, orangtua, urutan)
	if err != nil {
		return nil, err
	}

	anak := new(entity.Anak)
	if loadCache && d.Memory.Get(memkey, anak) {
		return anak, nil
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	filter = append(filter, belumDihapus)
	if err := d.Connection.FindOne(ctx, anak, entity.Anak{}.TableName(), []string{}, filter, map[string]int{"anak_ke": 1}); err != nil {
		return nil, galatCariAnak(err, id, nik, orangtua, urutan)
	}

	if loadCache {
		if err := d.Memory.Set(memkey, anak, ttlPanjang); err != nil {
			logger.Error("FindAnak:Memory.Set: "+memkey, "detail", err)
		}
	}
	return anak, nil
}

// baris terhapus tidak pernah masuk cache; cache hanya untuk data aktif
func (d *StuntingRepository) FindDeletedAnak(id *string, nik *string, orangtua *string, urutan *int16) (*entity.Anak, error) {
	filter, _, err := kriteriaAnak(id, nik, orangtua, urutan)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	anak := new(entity.Anak)
	if err := d.Connection.FindOneDeleted(ctx, anak, "[me]", []string{}, filter, map[string]int{"anak_ke": 1}); err != nil {
		return nil, galatCariAnak(err, id, nik, orangtua, urutan)
	}
	return anak, nil
}

// cari lewat id, nik, atau id orangtua + urutan
func kriteriaAnak(id, nik, orangtua *string, urutan *int16) ([]port.DbExpression, string, error) {
	switch {
	case utils.IsFilled(id):
		return []port.DbExpression{{Expr: "id", Args: []any{id}}}, keyAnakId(*id), nil
	case utils.IsFilled(nik):
		return []port.DbExpression{{Expr: "nik", Args: []any{nik}}}, keyAnakNik(*nik), nil
	case utils.IsFilled(orangtua) && urutan != nil:
		return []port.DbExpression{{Expr: "id_orangtua = ? AND anak_ke = ?", Args: []any{orangtua, urutan}}}, keyAnakUrutan(*orangtua, *urutan), nil
	}
	return nil, "", exceptions.Validasi.Messagef("untuk mencari anak sebutkan id anak, nik anak, atau id orangtua beserta urutan anak")
}

func galatCariAnak(err error, id, nik, orangtua *string, urutan *int16) error {
	if errors.Is(err, sql.ErrNoRows) {
		return exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
	}
	logger.Error(fmt.Sprintf("Find Anak (id=%v, nik=%v, id_orangtua=%v, anak_ke=%v): ", utils.Nilai(id), utils.Nilai(nik), utils.Nilai(orangtua), urutan) + helper.ToLogJSON(err))
	return err
}

// orangtua beserta anak-anaknya
func (d *StuntingRepository) ListAnak(orangtua, nikorangtua, nokk *string) (*entity.Orangtua, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()
	tmp := new(entity.Orangtua)

	filter := make([]port.DbExpression, 0)

	if utils.IsFilled(orangtua) {
		filter = append(filter, port.DbExpression{
			Expr: "o.id",
			Args: []any{orangtua},
		})
	}
	if utils.IsFilled(nikorangtua) {
		filter = append(filter, port.DbExpression{
			Expr: "o.nik",
			Args: []any{nikorangtua},
		})
	}
	if utils.IsFilled(nokk) {
		filter = append(filter, port.DbExpression{
			Expr: "o.no_kk",
			Args: []any{nokk},
		})
	}

	// hanya pencarian lewat id orangtua yang di-cache; nik dan no_kk tidak punya invalidasi
	var memkey string
	if utils.IsFilled(orangtua) && !utils.IsFilled(nikorangtua) && !utils.IsFilled(nokk) {
		memkey = keyListAnak(*orangtua)
		if cached := new(entity.Orangtua); d.Memory.Get(memkey, cached) {
			return cached, nil
		}
	}

	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	query := bunDB.NewSelect().Model(tmp).
		Relation("Anak", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("a.anak_ke ASC")
		})

	for _, f := range filter {
		query.Where(fmt.Sprintf("%s = ?", f.Expr), f.Args[0])
	}

	err := query.Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Orang tua dengan id/nik/nokk %s/%s/%s tidak ditemukan",
				utils.Nilai(orangtua), utils.Nilai(nikorangtua), utils.Nilai(nokk))
		}

		logger.Error(fmt.Sprintf("ListAnak (id_orangtua=%s, nik_orangtua=%s, nokk=%s): ",
			utils.Nilai(orangtua), utils.Nilai(nikorangtua), utils.Nilai(nokk)) + helper.ToLogJSON(err))
		return nil, err
	}
	if memkey != "" {
		d.Memory.Set(memkey, tmp, ttlPanjang)
	}
	return tmp, nil
}

func (d *StuntingRepository) DeleteAnak(id string) error {
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	_, err := d.Connection.DeleteOne(d.Context.Context, entity.Anak{}.TableName(), filter)

	if err != nil {
		logger.Error(fmt.Sprintf("Delete Anak (%s): ", id) + helper.ToLogJSON(err))
		return err
	}

	d.lupakan(keyAnakId(id))
	d.LupakanRiwayatAnak(id)

	return nil
}

func (d *StuntingRepository) DeleteAnakByID(id string, orgid string) error {
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

	affected, err := d.Connection.UpdateOne(ctx, entity.Anak{}.TableName(), filter, input)

	if err != nil {
		return err
	}

	if affected < 1 {
		return exceptions.TidakDitemukan.Messagef("Anak dengan id #%s tidak ditemukan", id)
	}

	return nil
}
