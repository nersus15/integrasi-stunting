package repository

import (
	"context"

	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/port"

	"github.com/nersus15/integrasi/mod-satusehat/config"
	"github.com/nersus15/integrasi/mod-satusehat/entity"
)

type SatusehatRepository interface {
	GetAll(ctx context.Context) ([]entity.Satusehat, error)
	GetByID(ctx context.Context, id string) (*entity.Satusehat, error)
	Create(ctx context.Context, input entity.SatusehatInput) (*entity.Satusehat, error)
}

type Repository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
}

func NewRepository(ctx *core.AppContext, cfg *config.ModuleConfig, db port.IDatabase) *Repository {
	return &Repository{Connection: db, Context: ctx, Config: cfg}
}

func (r *Repository) GetAll(ctx context.Context) ([]entity.Satusehat, error) {
	var items []entity.Satusehat

	filter := []port.DbExpression{
		{Expr: "deleted_at", Args: []any{nil}}, // soft-delete filter
	}
	sort := map[string]int{"created_at": -1}

	err := r.Connection.Find(ctx, &items, "satusehats", []string{}, filter, sort, 0, 0)
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Satusehat, error) {
	var item entity.Satusehat

	filter := []port.DbExpression{
		{Expr: "id", Args: []any{id}},
		{Expr: "deleted_at", Args: []any{nil}},
	}

	if err := r.Connection.FindOne(ctx, &item, "satusehats", []string{}, filter, nil); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, input entity.SatusehatInput) (*entity.Satusehat, error) {
	item := &entity.Satusehat{Name: input.Name}

	if _, err := r.Connection.InsertOne(ctx, "satusehats", item); err != nil {
		return nil, err
	}
	return item, nil
}
