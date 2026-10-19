package repository

import (
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/port"
)

type StuntingRepository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Memory     port.ICacheMemory
}

func NewStuntingRepository(ctx *core.AppContext, cfg *config.ModuleConfig, conn port.IDatabase, mem port.ICacheMemory) *StuntingRepository {
	return &StuntingRepository{
		Connection: conn,
		Context:    ctx,
		Config:     cfg,
		Memory:     mem,
	}
}

var belumDihapus = port.DbExpression{Expr: `"deletedAt"`, Args: []any{nil}}
