package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
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

func (d *StuntingRepository) FindKafkaTransaction(TransactionId string) (*types.KafkaTransaction, error) {
	filter := []port.DbExpression{
		{
			Expr: "transaction_id = ? AND group_id = ? AND resolved_at IS NULL",
			Args: []any{TransactionId, d.Config.Kafka.GroupID},
		},
	}

	tmp := new(entity.FailedTransactions)
	var res *types.KafkaTransaction

	err := d.Connection.FindOne(d.Context.Context, tmp, tmp.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindKafkaTransaction (%s): ", TransactionId) + err.Error())
		return nil, err
	}

	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) FindKafkaTransactions(patientId *string) ([]*types.KafkaTransaction, error) {
	filter := []port.DbExpression{
		{
			Expr: "group_id = ? AND resolved_at IS NULL",
			Args: []any{d.Config.Kafka.GroupID},
		},
	}

	if utils.IsFilled(patientId) {
		filter = append(filter, port.DbExpression{Expr: "patient_id = ?", Args: []any{*patientId}})
	}

	tmp := make([]*entity.FailedTransactions, 0)
	err := d.Connection.Find(d.Context.Context, &tmp, entity.FailedTransactions{}.TableName(),
		[]string{"*"}, filter, map[string]int{"createdAt": 1}, 0, 0)

	if err != nil {
		logger.Error("FindKafkaTransactions: " + err.Error())
		return nil, err
	}

	res := make([]*types.KafkaTransaction, 0, len(tmp))
	for _, v := range tmp {
		var t *types.KafkaTransaction
		res = append(res, t.FromEntity(v))
	}

	return res, nil
}

func (d *StuntingRepository) SaveKafkaTransaction(data types.KafkaTransactionPayload) error {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	e := data.ToEntity()

	_, err := bunDB.NewInsert().Model(e).
		On("CONFLICT (transaction_id, group_id) DO UPDATE").
		Set("message = EXCLUDED.message").
		Set("error = EXCLUDED.error").
		Set("patient_id = EXCLUDED.patient_id").
		Set("attempt = tx.attempt + 1").
		Set("resolved_at = NULL").
		Set(`"updatedAt" = ?`, time.Now()).
		Exec(d.Context.Context)

	if err != nil {
		logger.Error(fmt.Sprintf("SaveKafkaTransaction (%s): ", data.TransactionId) + err.Error())
	}

	return err
}

func (d *StuntingRepository) TandaiKafkaTransactionSelesai(transactionId string) error {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	_, err := bunDB.NewUpdate().Model((*entity.FailedTransactions)(nil)).
		Set("resolved_at = ?", time.Now()).
		Set("error = NULL").
		Set(`"updatedAt" = ?`, time.Now()).
		Where("transaction_id = ? AND group_id = ?", transactionId, d.Config.Kafka.GroupID).
		Exec(d.Context.Context)

	if err != nil {
		logger.Error(fmt.Sprintf("TandaiKafkaTransactionBeres (%s): ", transactionId) + err.Error())
	}

	return err
}

func (d *StuntingRepository) UpdateOrangTuaById(orangtua *entity.Orangtua, ctx context.Context, cariDulu bool) (*types.Orangtua, error) {
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

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{orangtua.ID},
		},
	}

	if cariDulu {
		res, err := d.FindOrangTuaById(orangtua.ID, false)
		if err != nil {
			return nil, exceptions.Classify(err)
		}

		if res == nil {
			return nil, exceptions.TidakDitemukan.Messagef("Orangtua dengan id: #%s tidak ditemukan", orangtua.ID)
		}
	}

	_, err := d.Connection.UpdateOne(ctx, orangtua.TableName(), filter, orangtua)

	if err != nil {
		return nil, err
	}

	var temp *types.Orangtua

	res := temp.FromEntity(orangtua)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	d.LupakanOrangtua(orangtua)

	return res, nil
}

func (d *StuntingRepository) FindOrangtuaBySatusehatId(satusehatId string, loadcache bool) (*types.Orangtua, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id orang tua tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*5)
	defer cancel()

	tmp := new(entity.Orangtua)
	res := new(types.Orangtua)

	memkey := keyOrangtuaSatusehat(satusehatId)

	if loadcache {
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}

	filter := []port.DbExpression{
		{
			Expr: "satusehat_id = ?",
			Args: []any{satusehatId},
		},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Orangtua tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindOrangtuaBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	res = res.FromEntity(tmp)
	if loadcache {
		d.Memory.Set(memkey, res, ttlPanjang)
	}

	return res, nil
}

func (d *StuntingRepository) FindAnakBySatusehatId(satusehatId string, loadCache bool) (*types.Anak, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id anak tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Anak)
	res := new(types.Anak)

	memkey := keyAnakSatusehat(satusehatId)
	if loadCache {
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}

	filter := []port.DbExpression{
		{
			Expr: "satusehat_id = ?",
			Args: []any{satusehatId},
		},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Anak{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindAnakBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	res = res.FromEntity(tmp)
	if loadCache {
		d.Memory.Set(memkey, res, ttlPanjang)
	}

	return res, nil
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
		// id_orangtua ikut tersimpan di cache anak, jadi semua jalan masuk ke
		// baris itu perlu dibuang -- termasuk yang berkunci nik dan urutan
		d.lupakan(keyAnakSatusehat(satusehatIdAnak), keyListAnak(idOrangtua))
		if a, e := d.FindAnakBySatusehatId(satusehatIdAnak, false); e == nil && a != nil {
			d.LupakanAnakTypes(a)
			d.LupakanAksesPerubahanOrangtua(a.Id, idOrangtua)
		}
	}

	return jml, err
}

func (d *StuntingRepository) CreateOrangTua(orangtua *entity.Orangtua, ctx context.Context) (*types.Orangtua, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	// satu pencarian: aktif dipakai ulang, terhapus dipulihkan, belum ada di-insert
	var res *types.Orangtua
	lama, err := d.FindOrangTuaByNikWithDeleted(ctx, orangtua.NIK)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if lama != nil && lama.DeletedAt == nil {
		return res.FromEntity(lama), nil
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
		return res.FromEntity(pulih), nil
	}

	_, err = d.Connection.InsertOne(ctx, orangtua.TableName(), orangtua)

	if err != nil {
		return nil, err
	}

	res = res.FromEntity(orangtua)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	d.LupakanOrangtua(orangtua)

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
	res, err := d.FindAnak(nil, anak.NIK, &anak.IDOrangtua, &anak.AnakKe, false)

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

	d.LupakanAnak(anak)

	return res, nil
}

func (d *StuntingRepository) UpdateAnakById(anak *entity.Anak, ctx context.Context, cariDulu bool) (*types.Anak, error) {
	if anak == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(d.Context.Context, 10*time.Second)
		defer cancel()
	}

	// Cari dulu
	if cariDulu {
		res, err := d.FindAnak(utils.PtrStr(anak.ID, true), anak.NIK, &anak.IDOrangtua, &anak.AnakKe, false)

		if err != nil {
			return nil, exceptions.Classify(err)
		}

		if res == nil {
			return nil, exceptions.TidakDitemukan.Messagef("Anak dengan id: #%s tidak ditemukan", anak.ID)
		}
	}
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{anak.ID},
		},
	}

	_, err := d.Connection.UpdateOne(ctx, anak.TableName(), filter, anak)

	if err != nil {
		return nil, err
	}

	var temp *types.Anak

	res := temp.FromEntity(anak)

	if res == nil || !utils.IsStrFilled(res.Id) {
		return nil, exceptions.Internal.Messagef("data gagal disimpan")
	}

	d.LupakanAnak(anak)

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

	d.LupakanRiwayatAnak(kunjungan.IDAnak)
	d.lupakan(keyKunjunganId(kunjungan.ID))

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
		if kunjungan.SourceData != nil {
			kunjungan.SourceData = utils.PtrStr(fmt.Sprintf("jakantro:%s", utils.StrPtr(kunjungan.SourceData)), true)
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

	d.LupakanOrangtua(orangtua)
	// NIK yang sudah ada dipakai ulang atau dipulihkan: barisnya bisa beda id dari payload
	if orangtua != nil && resOrangtua != nil && resOrangtua.Id != orangtua.ID {
		d.LupakanOrangtua(resOrangtua.ToPayload().ToEntity())
	}
	d.LupakanAnak(anak)
	if kunjungan != nil {
		d.LupakanRiwayatAnak(kunjungan.IDAnak)
		d.lupakan(keyKunjunganId(kunjungan.ID))
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
	err := tx.NewSelect().Model(temp).WhereAllWithDeleted().Where("nik = ?", orangtua.NIK).Limit(1).Scan(ctx)

	// jika ketemu
	if err == nil {
		if temp.DeletedAt == nil {
			// Disini duplikat, tapi gunakan yang sudah ada dan jangan create baru
			return res.FromEntity(temp), nil
		}
		if pulih, err := pulihkanOrangtuaTerhapus(ctx, tx, temp, orangtua); err != nil || pulih != nil {
			return res.FromEntity(pulih), err
		}
	}
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

// NIK yang hanya dipakai orangtua terhapus tidak ditolak: baris itu dipulihkan dan diperbarui dengan data baru
func pulihkanOrangtuaTerhapus(ctx context.Context, tx bun.IDB, lama, baru *entity.Orangtua) (*entity.Orangtua, error) {
	// lama := new(entity.Orangtua)
	// err := tx.NewSelect().Model(lama).WhereDeleted().Where("nik = ?", baru.NIK).Limit(1).Scan(ctx)
	// if errors.Is(err, sql.ErrNoRows) {
	// 	return nil, nil
	// }
	// if err != nil {
	// 	return nil, err
	// }

	lama.Override(*baru, false)
	now := time.Now()
	lama.DeletedAt, lama.DeletedBy, lama.UpdatedAt = nil, nil, &now

	if _, err := tx.NewUpdate().Model(lama).WherePK().WhereAllWithDeleted().Exec(ctx); err != nil {
		return nil, err
	}
	return lama, nil
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

func (d *StuntingRepository) FindOrangTuaByNik(nik string, loadCache bool) (*types.Orangtua, error) {
	memkey := keyOrangtuaNik(nik)

	if loadCache {
		if cached := new(types.Orangtua); d.Memory.Get(memkey, cached) {
			return cached, nil
		}
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	orangtua := new(entity.Orangtua)

	var res *types.Orangtua

	filter := []port.DbExpression{
		{
			Expr: "nik = ?",
			Args: []any{nik},
		},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, orangtua, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindOrangTuaByNik (%v): ", nik) + helper.ToLogJSON(err))
		return nil, err
	}

	res = res.FromEntity(orangtua)
	if loadCache {
		d.Memory.Set(memkey, res, ttlPanjang)
	}

	return res, nil
}

// entity, bukan types: baris terhapus dipulihkan lewat UPDATE seluruh kolom, jadi tidak boleh ada kolom yang hilang
func (d *StuntingRepository) FindOrangTuaByNikWithDeleted(ctx context.Context, nik string) (*entity.Orangtua, error) {
	temp := new(entity.Orangtua)
	filter := []port.DbExpression{{Expr: "nik = ?", Args: []any{nik}}}

	// FindOne dengan nama tabel tidak memakai tag soft_delete, jadi baris terhapus ikut terbaca
	if err := d.Connection.FindOne(ctx, temp, entity.Orangtua{}.TableName(), []string{"*"}, filter, map[string]int{}); err != nil {
		return nil, err
	}
	return temp, nil
}

func (d *StuntingRepository) FindOrangTuaById(id string, loadCache bool) (*types.Orangtua, error) {
	if !utils.IsStrFilled(id) {
		return nil, exceptions.KolomWajib.Messagef("Id orangtua tidak boleh kosong")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	orangtua := new(entity.Orangtua)
	res := new(types.Orangtua)

	memkey := keyOrangtuaId(id)

	if loadCache {
		// Cari di cache
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	err := d.Connection.FindOne(ctx, orangtua, "[me]", []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindOrangTuaById (%v): ", id) + err.Error())
		return nil, err
	}

	if loadCache {
		// simpan ke cache
		d.Memory.Set(memkey, res.FromEntity(orangtua), ttlPanjang)
	}

	return res.FromEntity(orangtua), nil
}

func (d *StuntingRepository) FindDeletedOrangtua(id string, loadCache bool) (*types.Orangtua, error) {
	if !utils.IsStrFilled(id) {
		return nil, exceptions.KolomWajib.Messagef("Id orangtua tidak boleh kosong")
	}
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	orangtua := new(entity.Orangtua)
	res := new(types.Orangtua)

	memkey := keyOrangtuaId(id)

	if loadCache {
		// Cari di cache
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}

	filter := []port.DbExpression{
		{Expr: "id", Args: []any{id}},
	}

	err := d.Connection.FindOneDeleted(ctx, orangtua, "[me]", []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindDeletedOrangtua (%v): ", id) + err.Error())
		return nil, err
	}

	if loadCache {
		// simpan ke cache
		d.Memory.Set(memkey, res.FromEntity(orangtua), ttlPanjang)
	}

	return res.FromEntity(orangtua), nil
}

func (d *StuntingRepository) FindAnak(id *string, nik *string, orangtua *string, urutan *int16, loadCache bool) (*types.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	res := new(types.Anak)
	tmp := new(entity.Anak)
	var memkey string

	filter := make([]port.DbExpression, 0)

	if utils.IsFilled(id) {
		filter = append(filter, port.DbExpression{
			Expr: "id",
			Args: []any{id},
		})
		memkey = keyAnakId(*id)
	} else if utils.IsFilled(nik) {
		filter = append(filter, port.DbExpression{
			Expr: "nik",
			Args: []any{nik},
		})
		memkey = keyAnakNik(*nik)
	} else {
		if !utils.IsFilled(orangtua) || urutan == nil {
			return nil, exceptions.Validasi.Messagef("untuk mencari anak sebutkan id anak, nik anak, atau id orangtua beserta urutan anak")
		}

		filter = append(filter, port.DbExpression{
			Expr: "id_orangtua = ? AND anak_ke = ?",
			Args: []any{orangtua, urutan},
		})
		memkey = keyAnakUrutan(*orangtua, *urutan)
	}
	sort := map[string]int{
		"anak_ke": 1,
	}

	// Cari dari cache
	if loadCache {
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Anak{}.TableName(), []string{}, filter, sort)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
		}

		logger.Error(fmt.Sprintf("Find Anak (id=%v, nik=%v, id_orangtua=%v, anak_ke=%v): ", utils.Nilai(id), utils.Nilai(nik), utils.Nilai(orangtua), urutan) + helper.ToLogJSON(err))
		return nil, err
	}
	res = res.FromEntity(tmp)

	// simpan di cache
	if loadCache {
		if err := d.Memory.Set(memkey, res, ttlPanjang); err != nil {
			logger.Error("FindAnak:Memory.Set: "+memkey, "detail", err)
		}
	}

	return res, nil
}

func (d *StuntingRepository) FindDeletedAnak(id *string, nik *string, orangtua *string, urutan *int16, loadCache bool) (*types.Anak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	res := new(types.Anak)
	tmp := new(entity.Anak)
	var memkey string

	filter := make([]port.DbExpression, 0)
	sort := map[string]int{
		"anak_ke": 1,
	}

	if utils.IsFilled(id) {
		filter = append(filter, port.DbExpression{
			Expr: "id",
			Args: []any{id},
		})
		memkey = keyAnakId(*id)
	} else if utils.IsFilled(nik) {
		filter = append(filter, port.DbExpression{
			Expr: "nik",
			Args: []any{nik},
		})
		memkey = keyAnakNik(*nik)
	} else {
		if !utils.IsFilled(orangtua) || urutan == nil {
			return nil, exceptions.Validasi.Messagef("untuk mencari anak sebutkan id anak, nik anak, atau id orangtua beserta urutan anak")
		}

		filter = append(filter, port.DbExpression{
			Expr: "id_orangtua = ? AND anak_ke = ?",
			Args: []any{orangtua, urutan},
		})
		memkey = keyAnakUrutan(*orangtua, *urutan)
	}

	// Cari dari cache
	if loadCache {
		if ok := d.Memory.Get(memkey, res); ok {
			return res, nil
		}
	}
	err := d.Connection.FindOneDeleted(ctx, tmp, "[me]", []string{}, filter, sort)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err)
		}

		logger.Error(fmt.Sprintf("Find Anak (id=%v, nik=%v, id_orangtua=%v, anak_ke=%v): ", utils.Nilai(id), utils.Nilai(nik), utils.Nilai(orangtua), urutan) + helper.ToLogJSON(err))
		return nil, err
	}
	res = res.FromEntity(tmp)

	// simpan di cache
	if loadCache {
		if err := d.Memory.Set(memkey, res, ttlPanjang); err != nil {
			logger.Error("FindAnak:Memory.Set: "+memkey, "detail", err)
		}
	}

	return res, nil
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

	// hanya pencarian lewat id orangtua yang di-cache; nik dan no_kk jarang
	// dipakai dan tidak punya jalur invalidasi sendiri
	var memkey string
	if utils.IsFilled(orangtua) && !utils.IsFilled(nikorangtua) && !utils.IsFilled(nokk) {
		memkey = keyListAnak(*orangtua)
		if cached := new(types.ListAnak); d.Memory.Get(memkey, cached) {
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
	var a *types.ListAnak

	res := a.FromEntity(tmp)
	if memkey != "" {
		d.Memory.Set(memkey, res, ttlPanjang)
	}
	return res, nil
}

func (d *StuntingRepository) ListKunjunganAnak(id string) (*types.KunjunganAnakArray, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyListKunjungan(id)
	if cached := new(types.KunjunganAnakArray); d.Memory.Get(memkey, cached) {
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
	var res *types.KunjunganAnakArray
	hasil := res.FromEntity(anak)
	d.Memory.Set(memkey, hasil, ttlSedang)
	return hasil, nil
}

func (d *StuntingRepository) UpdateKunjunganById(id string, input *entity.Kunjungan, cariDulu bool) (*types.Kunjungan, error) {
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

	res := new(types.Kunjungan)
	return res.FromEntity(input), nil
}

func (d *StuntingRepository) UpdateKesehatanById(id string, input *entity.Kesehatan, cariDulu bool) (*types.Kesehatan, error) {
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

	res := new(types.Kesehatan)
	return res.FromEntity(input), nil
}

// pakai bun langsung, lib-sql tidak bisa memenuhi relasi has-many
func (d *StuntingRepository) ListKesehatanAnak(id string) (*types.KesehatanAnakArray, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyListKesehatan(id)
	if cached := new(types.KesehatanAnakArray); d.Memory.Get(memkey, cached) {
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

	var res *types.KesehatanAnakArray
	hasil := res.FromEntity(anak)
	d.Memory.Set(memkey, hasil, ttlPanjang)
	return hasil, nil
}

func (d *StuntingRepository) SummaryAnak(id string) (*types.SummaryAnak, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keySummaryAnak(id)
	if cached := new(types.SummaryAnak); d.Memory.Get(memkey, cached) {
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
	bagiTurnunanKunjungan(anak)

	var res *types.SummaryAnak
	hasil := res.FromEntity(anak)
	d.Memory.Set(memkey, hasil, ttlPendek)
	return hasil, nil
}

func bagiTurnunanKunjungan(anak *entity.Anak) {
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
			// masukkan observasi ke kunjungan
			k.Observasi = append(k.Observasi, o)
		} else {
			// observasi yang tidak memiliki kunjungan (mungkin belum dikirim) masukkan ke anak.observasi
			orphanObs = append(orphanObs, o)
		}
	}

	// Observasi.component, disimpan di tabel observasi sebagai child dari observation
	for _, o := range anak.Observasi {
		// jika o.IDInduk == nil berarti ini resource observation bukan observation.component
		if o.IDInduk == nil {
			continue
		}
		// cari induknya dari daftar, jika ada, masukkan o (observation.component ini) ke induknya (observation.component)
		if p, ada := induk[*o.IDInduk]; ada {
			p.Component = append(p.Component, o)
		}
	}

	// Urutkan observation
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
func (d *StuntingRepository) FindKunjunganById(id string) (*types.Kunjungan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyKunjunganId(id)
	if cached := new(types.Kunjungan); d.Memory.Get(memkey, cached) {
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

	var res *types.Kunjungan
	hasil := res.FromEntity(tmp)
	d.Memory.Set(memkey, hasil, ttlSedang)
	return hasil, nil
}

func (d *StuntingRepository) FindKunjunganByIdAnak(idanak string) (*types.Kunjungan, error) {
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

	var res *types.Kunjungan
	return res.FromEntity(tmp), nil
}

// stunting NULL berarti belum ditentukan, dilewati
func (d *StuntingRepository) FindStatusStuntingTerakhir(idanak string) (*types.Kunjungan, error) {
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

	var res *types.Kunjungan
	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) FindKesehatanById(id string) (*types.Kesehatan, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, time.Second*10)
	defer cancel()

	memkey := keyKesehatanId(id)
	if cached := new(types.Kesehatan); d.Memory.Get(memkey, cached) {
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

	var res *types.Kesehatan
	hasil := res.FromEntity(tmp)
	d.Memory.Set(memkey, hasil, ttlPanjang)
	return hasil, nil
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

	d.LupakanRiwayatAnak(kesehatan.IDAnak)
	d.lupakan(keyKesehatanId(kesehatan.ID))

	return res, nil
}

// kembaran KunjunganTransaction, entitas terakhirnya kesehatan
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
		if kesehatan.SourceData != nil {
			kesehatan.SourceData = utils.PtrStr(fmt.Sprintf("jakantro:%s", utils.StrPtr(kesehatan.SourceData)), true)
		}
		resKesehatan, err = CreateKesehatanTx(ctx, tx, kesehatan)
		return err
	})

	if err != nil {
		return nil, nil, nil, err
	}

	d.LupakanOrangtua(orangtua)
	d.LupakanAnak(anak)
	if kesehatan != nil {
		d.LupakanRiwayatAnak(kesehatan.IDAnak)
		d.lupakan(keyKesehatanId(kesehatan.ID))
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

	d.lupakan(keyOrangtuaId(*id), keyListAnak(*id))

	return nil
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

func (d *StuntingRepository) FindPosyanduDetail(id string) (*types.PosyanduDetail, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	memkey := "posyandu::" + id
	posyandu := new(types.PosyanduDetail)
	posyanduEntity := new(entity.Posyandu)

	// cari dari cache dulu
	if ok := d.Memory.Get(memkey, posyandu); ok {
		// logger.Info(fmt.Sprintf("FindPosyandu(%s) Dari Cache: %s", id, helper.ToLogJSON(posyandu)))
		return posyandu, nil
	}

	// Cari dari database
	filter := []port.DbExpression{
		{
			Expr: "p.id",
			Args: []any{id},
		},
	}
	err := d.Connection.FindOne(ctx, posyanduEntity, "[me],Puskesmas.Induk", []string{}, filter, map[string]int{})

	if err != nil {
		logger.Error("FindPosyanduDetail", err)
		return nil, exceptions.Classify(err)
	}

	// save ke cache
	d.Memory.Set(memkey, posyandu.FromEntity(posyanduEntity), ttlPanjang)

	return posyandu.FromEntity(posyanduEntity), nil
}
func (d *StuntingRepository) FindFaskesByOrgid(orgid string) (*types.Faskes, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	memkey := "Faskes::" + orgid
	faskes := new(types.Faskes)
	faskesEntity := new(entity.Faskes)

	// cari dari cache dulu
	if ok := d.Memory.Get(memkey, faskes); ok {
		// logger.Info(fmt.Sprintf("FindPosyandu(%s) Dari Cache: %s", id, helper.ToLogJSON(posyandu)))
		return faskes, nil
	}

	// Cari dari database
	filter := []port.DbExpression{
		{
			Expr: "f.satusehat_id",
			Args: []any{orgid},
		},
	}
	err := d.Connection.FindOne(ctx, faskesEntity, "[me]", []string{}, filter, map[string]int{})

	if err != nil {
		logger.Error("FindFaskesByOrgid", err)
		return nil, exceptions.Classify(err)
	}

	// save ke cache
	d.Memory.Set(memkey, faskes.FromEntity(faskesEntity), ttlPanjang)

	return faskes.FromEntity(faskesEntity), nil
}
func (d *StuntingRepository) StartMigration(DB interface{}, dialect string, service string, command string, dir string, args []string) error {
	bunDB, ok := DB.(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	switch dialect {
	case "sqlite":
		goose.SetDialect("sqlite3")
	case "postgres":
		goose.SetDialect("postgres")
		logger.Info(fmt.Sprintf("Membuat Schema %s Untuk postgres", d.Context.Config.Database.SchemaName))

		_, err := bunDB.ExecContext(d.Context.Context, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", d.Context.Config.Database.SchemaName))
		if err != nil {
			logger.Error(fmt.Sprintf("Gagal membuat schema %s: %v", d.Context.Config.Database.SchemaName, err))
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

// Encounter bisa datang di bundle terpisah, jadi dicari, bukan dari isi bundle
func (d *StuntingRepository) FindKunjunganBySatusehatId(satusehatId string) (*types.Kunjungan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id encounter tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Kunjungan)
	var res *types.Kunjungan

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Kunjungan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Kunjungan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindKunjunganBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return res.FromEntity(tmp), nil
}

func (d *StuntingRepository) FindObservasiBySatusehatId(satusehatId string) (*types.Observasi, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id observation tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Observasi)
	var res *types.Observasi

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Observasi{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Observasi tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindObservasiBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return res.FromEntity(tmp), nil
}

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

func (d *StuntingRepository) FindEpisodeBySatusehatId(satusehatId string, loadFaskes bool) (*entity.Episode, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id episode tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Episode)

	tabel := entity.Episode{}.TableName()
	kolom := "satusehat_id"

	if loadFaskes {
		tabel = "[me],Faskes"
		kolom = "ep.satusehat_id"
	}

	filter := []port.DbExpression{
		{Expr: kolom + " = ?", Args: []any{satusehatId}},
	}
	if !loadFaskes {
		filter = append(filter, belumDihapus)
	}

	err := d.Connection.FindOne(ctx, tmp, tabel, []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Episode tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindEpisodeBySatusehatId (%v): ", satusehatId) + err.Error())
		return nil, err
	}

	return tmp, nil
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
	Ukur            *types.Kunjungan
}

var belumDihapus = port.DbExpression{Expr: `"deletedAt"`, Args: []any{nil}}

var tabelTurunan = []string{"stunting.observasi", "stunting.diagnosa", "stunting.layanan", "stunting.rujukan"}

// keempat tabel disapu dalam satu statement; UPDATE di dalam WITH tetap dijalankan walau hasilnya tidak dibaca
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

		if data.IdKunjunganUkur != nil && data.Ukur != nil {
			if kolom := utils.KolomUkurTerisi(data.Ukur); len(kolom) > 0 {
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

	// sapuan di dalam transaksi bisa menyentuh kunjungan lain, jadi riwayat
	// anak dibuang seluruhnya -- bukan cuma kunjungan yang barusan ditulis
	d.LupakanRiwayatAnak(idAnakDataMedis(data))
	if data.Kunjungan != nil {
		d.lupakan(keyKunjunganId(data.Kunjungan.ID))
	}
	if data.IdKunjunganStatus != nil {
		d.lupakan(keyKunjunganId(*data.IdKunjunganStatus))
	}

	return nil
}

// idAnakDataMedis mencari id anak dari bagian mana pun yang terisi; kunjungan
// bisa saja nihil saat data medis tiba lebih dulu.
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

func (d *StuntingRepository) OrgIdByEncounterSId(satusehat_id string) (*string, error) {
	if !utils.IsStrFilled(satusehat_id) {
		return nil, exceptions.Forbidden.WithMessage("Gagal Verifikasi", nil)
	}
	kunjungan := new(entity.Kunjungan)
	filter := []port.DbExpression{{
		Expr: "kj.satusehat_id",
		Args: []any{satusehat_id},
	},
	}
	if err := d.Connection.FindOne(d.Context.Context, kunjungan, "[me],Faskes", []string{"faskes.id", "faskes.id_induk", "faskes.satusehat_id"}, filter, map[string]int{}); err != nil {
		return nil, exceptions.Classify(err)
	}

	logger.Info(fmt.Sprintf("OrgIdByEncounterSId (%s): ", satusehat_id) + helper.ToLogJSON(kunjungan))
	if kunjungan.Faskes == nil {
		return nil, nil
	}

	return kunjungan.Faskes.SatusehatID, nil
}

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

// hanya dipanggil saat id_orangtua anak berubah atau anak dihapus; jalur yang sering tidak menyentuh query ini
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

func (d *StuntingRepository) OrgIdByKunjunganId(id string) (*string, error) {
	if !utils.IsStrFilled(id) {
		return nil, exceptions.Forbidden.WithMessage("Gagal Verifikasi", nil)
	}

	kunjungan := new(entity.Kunjungan)
	filter := []port.DbExpression{{Expr: "kj.id", Args: []any{id}}}

	if err := d.Connection.FindOne(d.Context.Context, kunjungan, "[me],Faskes",
		[]string{"faskes.id", "faskes.id_induk", "faskes.satusehat_id"}, filter, map[string]int{}); err != nil {
		return nil, exceptions.Classify(err)
	}

	if kunjungan.Faskes == nil {
		return nil, exceptions.Forbidden.WithMessage("Gagal Verifikasi: kunjungan tanpa faskes", nil)
	}

	return kunjungan.Faskes.SatusehatID, nil
}

func (d *StuntingRepository) FindRujukanBySatusehatId(satusehatId string) (*entity.Rujukan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id rujukan tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Rujukan)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Rujukan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Rujukan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindRujukanBySatusehatId (%v): ", satusehatId) + err.Error())
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindRujukanByIdAnak(idanak string) (*types.RujukanDetail, error) {
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
	var res *types.RujukanDetail

	return res.FromEntity(tmp), nil
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

func (d *StuntingRepository) UpdateKunjunganBySatusehatId(data *entity.Kunjungan, stream bool) (*entity.Kunjungan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	lama, err := d.FindKunjunganBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var gabungan *entity.Kunjungan
	var input map[string]any

	if stream {
		input = map[string]any{
			"tanggal_pengukuran": data.TanggalPengukuran,
			"tanggal_selesai":    data.TanggalSelesai,
		}
	} else {
		gabungan = lama.ToPayload().ToEntity()
		gabungan.Override(*data, false)
		input = utils.StructToMap(gabungan)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IdAnak, input); err != nil {
		return nil, err
	}

	d.lupakan(keyKunjunganId(lama.Id))

	return gabungan, nil
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
	d.LupakanRiwayatAnak(lama.IdAnak)

	return nil
}

func (d *StuntingRepository) UpdateObservasiBySatusehatId(data *entity.Observasi, stream bool) (*entity.Observasi, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	lama, err := d.FindObservasiBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var gabungan *entity.Observasi
	var input map[string]any
	if stream {
		input = map[string]any{
			"system":            data.System,
			"kode":              data.Kode,
			"display":           data.Display,
			"kategori":          data.Kategori,
			"nilai_angka":       data.NilaiAngka,
			"satuan":            data.Satuan,
			"nilai_teks":        data.NilaiTeks,
			"nilai_kode":        data.NilaiKode,
			"nilai_kode_system": data.NilaiKodeSystem,
			"nilai_display":     data.NilaiDisplay,
			"interpretasi":      data.Interpretasi,
			"tanggal":           data.Tanggal,
		}
	} else {
		gabungan = lama.ToEntity()
		gabungan.Override(*data, false)
		input = utils.StructToMap(gabungan)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IdAnak, input); err != nil {
		return nil, err
	}

	return gabungan, nil
}
func (d *StuntingRepository) FindDiagnosaBySatusehatId(satusehatId string) (*entity.Diagnosa, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id diagnosa tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Diagnosa)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Diagnosa{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Diagnosa tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindDiagnosaBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindLayananBySatusehatId(satusehatId string) (*entity.Layanan, error) {
	if !utils.IsStrFilled(satusehatId) {
		return nil, exceptions.Validasi.Messagef("satusehat id layanan tidak boleh kosong")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	tmp := new(entity.Layanan)
	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{satusehatId}},
	}

	filter = append(filter, belumDihapus)
	err := d.Connection.FindOne(ctx, tmp, entity.Layanan{}.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.TidakDitemukan.WithMessage("Data Layanan tidak ditemukan", err)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindLayananBySatusehatId (%v): ", satusehatId) + helper.ToLogJSON(err))
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) updateBySatusehatId(tabel string, satusehatId *string, idAnak string, kolom map[string]any) error {
	if !utils.IsFilled(satusehatId) {
		return exceptions.KolomWajib.Messagef("Id satusehat harus sudah terisi")
	}
	if len(kolom) == 0 {
		return exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}

	kolom["updatedAt"] = time.Now()

	filter := []port.DbExpression{
		{Expr: "satusehat_id = ?", Args: []any{*satusehatId}},
	}

	if _, err := d.Connection.Update(d.Context.Context, tabel, filter, kolom); err != nil {
		logger.Error(fmt.Sprintf("updateBySatusehatId %s (%s): ", tabel, *satusehatId) + err.Error())
		return err
	}

	d.LupakanRiwayatAnak(idAnak)

	return nil
}

func (d *StuntingRepository) UpdateDiagnosaBySatusehatId(data *entity.Diagnosa, stream bool) (*entity.Diagnosa, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindDiagnosaBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":               data.Jenis,
			"system":              data.System,
			"kode":                data.Kode,
			"display":             data.Display,
			"kategori":            data.Kategori,
			"kritikalitas":        data.Kritikalitas,
			"clinical_status":     data.ClinicalStatus,
			"verification_status": data.VerificationStatus,
			"onset":               data.Onset,
			"tanggal_catat":       data.TanggalCatat,
		}
	} else {
		lama.Override(*data, false)
		input = utils.StructToMap(lama)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input); err != nil {
		return nil, err
	}

	if stream {
		return nil, nil
	}
	return lama, nil
}

func (d *StuntingRepository) UpdateLayananBySatusehatId(data *entity.Layanan, stream bool) (*entity.Layanan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindLayananBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":    data.Jenis,
			"system":   data.System,
			"kode":     data.Kode,
			"display":  data.Display,
			"kategori": data.Kategori,
			"status":   data.Status,
			"jumlah":   data.Jumlah,
			"satuan":   data.Satuan,
			"tanggal":  data.Tanggal,
			"catatan":  data.Catatan,
		}
	} else {
		lama.Override(*data, false)
		input = utils.StructToMap(lama)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input); err != nil {
		return nil, err
	}

	if stream {
		return nil, nil
	}
	return lama, nil
}

func (d *StuntingRepository) UpdateRujukanBySatusehatId(data *entity.Rujukan, stream bool) (*entity.Rujukan, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindRujukanBySatusehatId(utils.Nilai(data.SatusehatId))
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"jenis":     data.Jenis,
			"system":    data.System,
			"kode":      data.Kode,
			"display":   data.Display,
			"status":    data.Status,
			"prioritas": data.Prioritas,
			"alasan":    data.Alasan,
			"tanggal":   data.Tanggal,
		}
	} else {
		lama.Override(*data, false)
		input = utils.StructToMap(lama)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input); err != nil {
		return nil, err
	}
	// akses klinis hanya bergantung pada jenis dan tanggal; keduanya cuma bisa berubah lewat stream
	if stream && (data.Jenis != lama.Jenis || !samaWaktu(data.Tanggal, lama.Tanggal)) {
		idOrangtua := ""
		if a, err := d.FindAnak(&lama.IDAnak, nil, nil, nil, true); err == nil && a != nil {
			idOrangtua = a.IdOrangtua
		}
		for _, ref := range []*string{lama.RefFaskesAsal, lama.RefFaskesTujuan} {
			if utils.IsFilled(ref) {
				d.LupakanAksesKlinis(*ref, lama.IDAnak, "")
				if idOrangtua != "" {
					d.LupakanAksesKlinis(*ref, "", idOrangtua)
				}
			}
		}
	}

	if stream {
		return nil, nil
	}
	return lama, nil
}

func samaWaktu(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

func (d *StuntingRepository) UpdateEpisodeBySatusehatId(data *entity.Episode, stream bool) (*entity.Episode, error) {
	if data == nil {
		return nil, exceptions.BentukPayload.Messagef("Tidak ada data untuk di update")
	}
	lama, err := d.FindEpisodeBySatusehatId(utils.Nilai(data.SatusehatId), false)
	if err != nil {
		return nil, err
	}

	var input map[string]any

	if stream {
		input = map[string]any{
			"system":  data.System,
			"kode":    data.Kode,
			"display": data.Display,
			"status":  data.Status,
			"mulai":   data.Mulai,
			"selesai": data.Selesai,
		}
	} else {
		lama.Override(*data, false)
		input = utils.StructToMap(lama)
	}

	if err := d.updateBySatusehatId(data.TableName(), data.SatusehatId, lama.IDAnak, input); err != nil {
		return nil, err
	}

	if stream {
		return nil, nil
	}
	return lama, nil
}

// Region Hapus
func (d *StuntingRepository) DeleteAnakByID(id string, orgid string) error {
	// Karena menggunakan soft delete jadi ada pilihan dimana menandai record yang berhubungan dengan anak di tabel lain untuk sebagai deletedAt = now
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
func (d *StuntingRepository) DeleteOrangtuaByID(id string, orgid string) error {
	// Karena menggunakan soft delete jadi ada pilihan dimana menandai record yang berhubungan dengan orangtua di tabel lain (anak) untuk sebagai deletedAt = now
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
