package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
)

func (d *StuntingRepository) KunjunganTransaction(orangtua *entity.Orangtua, anak *entity.Anak, kunjungan *entity.Kunjungan) (*entity.Orangtua, *entity.Anak, *entity.Kunjungan, error) {
	resOrangtua, resAnak, err := d.registrasi(orangtua, anak, func(ctx context.Context, tx bun.Tx, idAnak *string) error {
		if idAnak != nil {
			kunjungan.IDAnak = *idAnak
		}
		_, err := CreateKunjunganTx(ctx, tx, kunjungan)
		return err
	})
	if err != nil {
		return nil, nil, nil, err
	}

	d.LupakanRiwayatAnak(kunjungan.IDAnak)
	d.lupakan(keyKunjunganId(kunjungan.ID))
	return resOrangtua, resAnak, kunjungan, nil
}

func (d *StuntingRepository) KesehatanTransaction(orangtua *entity.Orangtua, anak *entity.Anak, kesehatan *entity.Kesehatan) (*entity.Orangtua, *entity.Anak, *entity.Kesehatan, error) {
	resOrangtua, resAnak, err := d.registrasi(orangtua, anak, func(ctx context.Context, tx bun.Tx, idAnak *string) error {
		if idAnak != nil {
			kesehatan.IDAnak = *idAnak
		}
		_, err := CreateKesehatanTx(ctx, tx, kesehatan)
		return err
	})
	if err != nil {
		return nil, nil, nil, err
	}

	d.LupakanRiwayatAnak(kesehatan.IDAnak)
	d.lupakan(keyKesehatanId(kesehatan.ID))
	return resOrangtua, resAnak, kesehatan, nil
}

// orangtua dan anak disimpan bersama entitas terakhirnya dalam satu transaksi
func (d *StuntingRepository) registrasi(orangtua *entity.Orangtua, anak *entity.Anak, simpan func(ctx context.Context, tx bun.Tx, idAnak *string) error) (*entity.Orangtua, *entity.Anak, error) {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return nil, nil, exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	var resOrangtua *entity.Orangtua
	var resAnak *entity.Anak
	err := bunDB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var err error
		if orangtua != nil {
			if resOrangtua, err = CreateOrangTuaTx(ctx, tx, orangtua); err != nil {
				return err
			}
			anak.IDOrangtua = resOrangtua.ID
		}

		var idAnak *string
		if anak != nil {
			if resAnak, err = CreateAnakTx(ctx, tx, anak); err != nil {
				return err
			}
			idAnak = &resAnak.ID
		}
		return simpan(ctx, tx, idAnak)
	})
	if err != nil {
		return nil, nil, err
	}

	d.LupakanOrangtua(orangtua)
	// NIK yang sudah ada dipakai ulang atau dipulihkan: barisnya bisa beda id dari payload
	if orangtua != nil && resOrangtua != nil && resOrangtua.ID != orangtua.ID {
		d.LupakanOrangtua(resOrangtua)
	}
	d.LupakanAnak(anak)
	return resOrangtua, resAnak, nil
}
