package service

import (
	"context"
	"fmt"

	"github.com/webcore-go/webcore/app/core"

	"github.com/nersus15/integrasi/mod-satusehat/entity"
	"github.com/nersus15/integrasi/mod-satusehat/repository"
)

type SatusehatService interface {
	GetSatusehats(ctx context.Context) ([]entity.Satusehat, error)
	GetSatusehat(ctx context.Context, id string) (*entity.Satusehat, error)
	CreateSatusehat(ctx context.Context, input entity.SatusehatInput) (*entity.Satusehat, error)
}

type Service struct {
	repo repository.SatusehatRepository
}

func NewService(ctx *core.AppContext, repo repository.SatusehatRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSatusehats(ctx context.Context) ([]entity.Satusehat, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) GetSatusehat(ctx context.Context, id string) (*entity.Satusehat, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) CreateSatusehat(ctx context.Context, input entity.SatusehatInput) (*entity.Satusehat, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.Create(ctx, input)
}
