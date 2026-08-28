package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/entity"
	exceptions "github.com/nersus15/integrasi/mod-stunting/helper/Exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

type StuntingRepository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Memory     port.ICacheMemory
}

type TransactiocPayload struct {
}

func NewStuntingRepository(ctx *core.AppContext, cfg *config.ModuleConfig, conn port.IDatabase, mem port.ICacheMemory) *StuntingRepository {
	return &StuntingRepository{
		Connection: conn,
		Context:    ctx,
		Config:     cfg,
		Memory:     mem,
	}
}

func (d *StuntingRepository) CreateOrangTua(orangtua *entity.Orangtua, ctx context.Context) (*types.Orangtua, error) {
	// find before insert
	r, err := d.FindOrangTuaByNik(orangtua.NIK)

	if err == nil {
		return r, nil
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	var res *types.Orangtua
	_, err = d.Connection.InsertOne(ctx, orangtua.TableName(), orangtua)

	if err != nil {
		return nil, err
	}

	res = res.FromEntity(orangtua)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}
	return res, nil
}

func (d *StuntingRepository) CreateAnak(anak *entity.Anak, ctx context.Context) (*types.Anak, error) {
	if anak == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	// Cari dulu
	res, err := d.FindAnak(nil, anak.NIK, &anak.IDOrangtua, &anak.AnakKe)

	if err == nil {
		if res.IdOrangtua != anak.IDOrangtua {
			return nil, exceptions.AnakMilikOrangtuaLain.WithMessage("Anak sudah terdaftar pada orangtua yang berbeda", err)
		}
		return res, nil
	}

	_, err = d.Connection.InsertOne(ctx, anak.TableName(), anak)

	if err != nil {
		return nil, err
	}

	res = res.FromEntity(anak)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	return res, nil
}

func (d *StuntingRepository) CreateKunjungan(kunjungan *entity.Kunjungan, ctx context.Context) (*types.Kunjungan, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	var res *types.Kunjungan
	_, err := d.Connection.InsertOne(ctx, kunjungan.TableName(), kunjungan)

	if err != nil {
		return nil, err
	}

	res = res.FromEntity(kunjungan)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	return res, nil
}

func (d *StuntingRepository) KunjunganTransaction(orangtua *entity.Orangtua, anak *entity.Anak, kunjungan *entity.Kunjungan) (*types.Orangtua, *types.Anak, *types.Kunjungan, error) {
	db := d.Connection.GetConnection()
	bunDB, ok := db.(*bun.DB)

	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return nil, nil, nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	var resOrangtua *types.Orangtua
	var resAnak *types.Anak
	var resKunjungan *types.Kunjungan
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	err := bunDB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var err error
		// run orang tua
		if orangtua != nil {
			resOrangtua, err = CreateOrangTuaTx(ctx, tx, orangtua)
			if err != nil {
				return err
			}

			anak.IDOrangtua = resOrangtua.Id
		}

		if anak != nil {
			resAnak, err = CreateAnakTx(ctx, tx, anak)
			if err != nil {
				return err
			}

			kunjungan.IDAnak = resAnak.Id
		}

		resKunjungan, err = CreateKunjunganTx(ctx, tx, kunjungan)

		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, nil, nil, err
	}

	return resOrangtua, resAnak, resKunjungan, nil
}

func CreateOrangTuaTx(ctx context.Context, tx bun.IDB, orangtua *entity.Orangtua) (*types.Orangtua, error) {
	if orangtua == nil {
		return nil, nil
	}
	var res *types.Orangtua
	temp := new(entity.Orangtua)

	// Cari dulu
	err := tx.NewSelect().Model(temp).Where("nik = ?", orangtua.NIK).Limit(1).Scan(ctx)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.Internal.WithMessage("gagal mencari data orangtua", err)
		}

		// Insert
		_, err1 := tx.NewInsert().Model(orangtua).Exec(ctx)
		if err1 != nil {
			return nil, err1
		}

		return res.FromEntity(orangtua), nil
	}

	return res.FromEntity(temp), nil
}

func CreateAnakTx(ctx context.Context, tx bun.IDB, anak *entity.Anak) (*types.Anak, error) {
	if anak == nil {
		return nil, nil
	}
	var res *types.Anak
	temp := new(entity.Anak)

	// Cari dulu
	whereClause := ""
	var whereArgs []any
	if utils.IsFilled(anak.NIK) {
		whereClause = "nik = ?"
		whereArgs = []any{anak.NIK}
	} else {
		if !utils.IsStrFilled(anak.IDOrangtua) || !utils.IsStrFilled(strconv.Itoa(int(anak.AnakKe))) {
			return nil, exceptions.Validasi.Messagef("untuk mencari anak sebutkan nik anak, atau id orangtua beserta urutan anak")
		}

		whereClause = "id_orangtua = ? AND anak_ke = ?"
		whereArgs = []any{anak.IDOrangtua, anak.AnakKe}
	}

	err := tx.NewSelect().Model(temp).Where(whereClause, whereArgs...).Limit(1).Scan(ctx)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.Internal.WithMessage("gagal mencari data anak", err)
		}
		// Insert
		_, err1 := tx.NewInsert().Model(anak).Exec(ctx)
		if err1 != nil {
			return nil, err1
		}

		return res.FromEntity(anak), nil
	}

	// cek dulu apakah id yang ditemukan sama dengan yang dikirim
	if temp.IDOrangtua != anak.IDOrangtua {
		return nil, exceptions.AnakMilikOrangtuaLain.Messagef("nik anak %s sudah terdaftar pada orangtua yang berbeda", *anak.NIK)
	}

	// retrun data yang dietemukan
	return res.FromEntity(temp), nil
}

func CreateKunjunganTx(ctx context.Context, tx bun.IDB, kunjungan *entity.Kunjungan) (*types.Kunjungan, error) {
	if kunjungan == nil {
		return nil, nil
	}
	var res *types.Kunjungan

	_, err := tx.NewInsert().Model(kunjungan).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return res.FromEntity(kunjungan), nil
}

func (d *StuntingRepository) FindOrangTuaByNik(nik string) (*types.Orangtua, error) {
	orangtua := new(entity.Orangtua)

	var res *types.Orangtua

	filter := []port.DbExpression{
		{
			Expr: "nik = ?",
			Args: []any{nik},
		},
	}

	err := d.Connection.FindOne(d.Context.Context, orangtua, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindOrangTuaByNik (%v): ", nik) + helper.ToLogJSON(err))
		return nil, err
	}

	return res.FromEntity(orangtua), nil
}

func (d *StuntingRepository) FindOrangTuaById(id string) (*types.Orangtua, error) {
	orangtua := new(entity.Orangtua)
	var res *types.Orangtua

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	err := d.Connection.FindOne(d.Context.Context, orangtua, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindOrangTuaById (%v): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	return res.FromEntity(orangtua), nil
}

func (d *StuntingRepository) FindAnak(id *string, nik *string, orangtua *string, urutan *int16) (*types.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	res := new(types.Anak)
	tmp := new(entity.Anak)

	filter := make([]port.DbExpression, 0)

	if utils.IsFilled(id) {
		filter = append(filter, port.DbExpression{
			Expr: "id",
			Args: []any{id},
		})
	} else if utils.IsFilled(nik) {
		filter = append(filter, port.DbExpression{
			Expr: "nik",
			Args: []any{nik},
		})
	} else {
		if !utils.IsFilled(orangtua) || urutan == nil {
			return nil, exceptions.Validasi.Messagef("untuk mencari anak sebutkan id anak, nik anak, atau id orangtua beserta urutan anak")
		}

		filter = append(filter, port.DbExpression{
			Expr: "id_orangtua = ? AND anak_ke = ?",
			Args: []any{orangtua, urutan},
		})
	}
	sort := map[string]int{
		"anak_ke": 1,
	}

	err := d.Connection.FindOne(ctx, tmp, entity.Anak{}.TableName(), []string{}, filter, sort)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
		}

		logger.Error(fmt.Sprintf("Find Anak (id=%v, nik=%v, id_orangtua=%v, anak_ke=%v): ", utils.Nilai(id), utils.Nilai(nik), utils.Nilai(orangtua), urutan) + helper.ToLogJSON(err))
		return nil, err
	}

	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) ListAnak(orangtua, nikorangtua, nokk *string) (*types.ListAnak, error) {
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
	var a *types.ListAnak

	res := a.FromEntity(tmp)
	return res, nil
}

func (d *StuntingRepository) ListKunjunganAnak(id string) (*types.KunjunganAnakArray, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

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
	var res *types.KunjunganAnakArray
	return res.FromEntity(anak), nil
}

// ListKesehatanAnak mengambil satu anak beserta seluruh riwayat kesehatannya.
// Memakai bun langsung, bukan lib-sql: relasi has-many tidak bisa dipenuhi
// lewat Scan(ctx, dest) -- bun menolaknya dan meminta Model.
func (d *StuntingRepository) ListKesehatanAnak(id string) (*types.KesehatanAnakArray, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

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

	var res *types.KesehatanAnakArray
	return res.FromEntity(anak), nil
}

// SummaryAnak mengambil anak beserta dua riwayatnya sekaligus. Kedua relasi
// has-many dijalankan bun sebagai query terpisah, jadi biayanya tiga round
// trip -- masih lebih murah daripada dua panggilan HTTP dari pemanggil.
func (d *StuntingRepository) SummaryAnak(id string) (*types.SummaryAnak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

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
		Relation("Kesehatan", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("k.tanggal_pemantauan ASC")
		}).
		Where("a.id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Anak dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Summary Anak (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	var res *types.SummaryAnak
	return res.FromEntity(anak), nil
}

func (d *StuntingRepository) FindKunjunganById(id string) (*types.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	tmp := new(entity.Kunjungan)
	filter := []port.DbExpression{
		{Expr: "id = ?", Args: []any{id}},
	}

	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{}, filter, map[string]int{})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Kunjungan dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Find Kunjungan (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	var res *types.Kunjungan
	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) FindKesehatanById(id string) (*types.Kesehatan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	tmp := new(entity.Kesehatan)
	filter := []port.DbExpression{
		{Expr: "id = ?", Args: []any{id}},
	}

	err := d.Connection.FindOne(ctx, tmp, entity.Kesehatan{}.TableName(), []string{}, filter, map[string]int{})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data Kesehatan dengan id %s tidak ditemukan", id)
		}
		logger.Error(fmt.Sprintf("Find Kesehatan (id: %s): ", id) + helper.ToLogJSON(err))
		return nil, err
	}

	var res *types.Kesehatan
	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) CreateKesehatan(kesehatan *entity.Kesehatan, ctx context.Context) (*types.Kesehatan, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	var res *types.Kesehatan
	_, err := d.Connection.InsertOne(ctx, kesehatan.TableName(), kesehatan)

	if err != nil {
		return nil, err
	}

	res = res.FromEntity(kesehatan)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	return res, nil
}

// KesehatanTransaction menyimpan orangtua, anak, dan kesehatan dalam satu
// transaksi. Kembarannya KunjunganTransaction; bila salah satu gagal, tidak
// ada data setengah tersimpan.
func (d *StuntingRepository) KesehatanTransaction(orangtua *entity.Orangtua, anak *entity.Anak, kesehatan *entity.Kesehatan) (*types.Orangtua, *types.Anak, *types.Kesehatan, error) {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return nil, nil, nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	var resOrangtua *types.Orangtua
	var resAnak *types.Anak
	var resKesehatan *types.Kesehatan

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	err := bunDB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var err error

		if orangtua != nil {
			resOrangtua, err = CreateOrangTuaTx(ctx, tx, orangtua)
			if err != nil {
				return err
			}
			anak.IDOrangtua = resOrangtua.Id
		}

		if anak != nil {
			resAnak, err = CreateAnakTx(ctx, tx, anak)
			if err != nil {
				return err
			}
			kesehatan.IDAnak = resAnak.Id
		}

		resKesehatan, err = CreateKesehatanTx(ctx, tx, kesehatan)
		return err
	})

	if err != nil {
		return nil, nil, nil, err
	}

	return resOrangtua, resAnak, resKesehatan, nil
}

func CreateKesehatanTx(ctx context.Context, tx bun.IDB, kesehatan *entity.Kesehatan) (*types.Kesehatan, error) {
	if kesehatan == nil {
		return nil, nil
	}
	var res *types.Kesehatan

	_, err := tx.NewInsert().Model(kesehatan).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return res.FromEntity(kesehatan), nil
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

	return nil
}

func (d *StuntingRepository) DeleteAnak(id *string) error {
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	_, err := d.Connection.DeleteOne(d.Context.Context, entity.Anak{}.TableName(), filter)

	if err != nil {
		logger.Error(fmt.Sprintf("Delete Anak (%s): ", *id) + helper.ToLogJSON(err))
		return err
	}

	return nil
}

func (d *StuntingRepository) StartMigration(DB interface{}, dialect string, service string, command string, dir string, args []string) error {
	bunDB, ok := DB.(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	// Set dialek wajib di sini sebelum eksekusi run
	switch dialect {
	case "sqlite":
		goose.SetDialect("sqlite3")
	case "postgres":
		goose.SetDialect("postgres")
		logger.Info(fmt.Sprintf("Membuat Schema %s Untuk postgres", d.Context.Config.Database.SchemaName))

		_, err := bunDB.ExecContext(d.Context.Context, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", d.Context.Config.Database.SchemaName))
		if err != nil {
			logger.Error(fmt.Sprintf("Gagal membuat schema jakantro: %v", err))
			return err
		}
	}

	if service != "" {
		goose.SetTableName("__migration_" + service + "_logs")
	} else {
		goose.SetTableName("__migration_webcore_logs")
	}

	logger.Info(fmt.Sprintf("Mengeksekusi Goose %s pada folder: %s", command, dir))

	if err := goose.RunContext(d.Context.Context, command, bunDB.DB, dir, args...); err != nil {
		logger.Error(fmt.Sprintf("goose run %s: %v", command, err))
		return err
	}

	return nil
}
