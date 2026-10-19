package repository

import (
	"context"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

// posyandu beserta puskesmas dan induknya
func (d *StuntingRepository) FindPosyanduDetail(id string) (*entity.Posyandu, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	memkey := "posyandu::" + id
	posyandu := new(entity.Posyandu)

	if ok := d.Memory.Get(memkey, posyandu); ok {
		return posyandu, nil
	}

	filter := []port.DbExpression{
		{
			Expr: "p.id",
			Args: []any{id},
		},
	}
	err := d.Connection.FindOne(ctx, posyandu, "[me],Puskesmas.Induk", []string{}, filter, map[string]int{})

	if err != nil {
		logger.Error("FindPosyanduDetail", err)
		return nil, exceptions.Classify(err)
	}

	d.Memory.Set(memkey, posyandu, ttlPanjang)
	return posyandu, nil
}
func (d *StuntingRepository) FindFaskesByOrgid(orgid string) (*entity.Faskes, error) {
	ctx, cancel := context.WithTimeout(d.Context.Context, 10*time.Second)
	defer cancel()

	memkey := "Faskes::" + orgid
	faskes := new(entity.Faskes)

	if ok := d.Memory.Get(memkey, faskes); ok {
		return faskes, nil
	}

	filter := []port.DbExpression{
		{
			Expr: "f.satusehat_id",
			Args: []any{orgid},
		},
	}
	err := d.Connection.FindOne(ctx, faskes, "[me]", []string{}, filter, map[string]int{})

	if err != nil {
		logger.Error("FindFaskesByOrgid", err)
		return nil, exceptions.Classify(err)
	}

	d.Memory.Set(memkey, faskes, ttlPanjang)
	return faskes, nil
}
