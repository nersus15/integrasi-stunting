package service

import (
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/webcore-go/webcore/app/core"
)

type StuntingService struct {
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Repository *repository.StuntingRepository
	Background *backgroundworker.BackgroundWorker
}

func NewStuntingService(wctx *core.AppContext, cfg *config.ModuleConfig, repository *repository.StuntingRepository, background *backgroundworker.BackgroundWorker) *StuntingService {
	return &StuntingService{
		Context:    wctx,
		Config:     cfg,
		Repository: repository,
		Background: background,
	}
}

// entity nil tanpa galat dianggap tidak ditemukan
func keDTO[E any, T any](ubah func(*T, *E) *T, e *E, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, exceptions.ErrTidakDitemukan
	}
	return ubah(nil, e), nil
}
