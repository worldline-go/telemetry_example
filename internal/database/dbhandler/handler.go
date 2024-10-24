package dbhandler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/model"
)

var (
	ErrDuplicate = errors.New("duplicate record")
	ErrNotFound  = errors.New("not found")
)

type Handler struct {
	db *goqu.Database
}

func New(db *sqlx.DB) *Handler {
	dbGoqu := goqu.New("posgres", db)

	return &Handler{db: dbGoqu}
}

func (h *Handler) GetProduct(ctx context.Context, name string) (*model.Product, error) {
	var product model.Product

	found, err := h.db.From("products").Where(goqu.C("name").Eq(name)).Executor().ScanStructContext(ctx, &product)
	if err != nil {
		return nil, err
	}

	if found {
		return &product, nil
	}

	return nil, fmt.Errorf("product [%s] %w", name, ErrNotFound)
}

func (h *Handler) AddNewProduct(ctx context.Context, name, description string) (int64, error) {
	var id int64

	_, err := h.db.Insert("products").Rows(
		goqu.Record{
			"name":        name,
			"description": description,
			"last_user":   config.ServiceName,
			"updated_at":  time.Now(),
		},
	).Returning("id").Executor().ScanValContext(ctx, &id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("name [%s] %w: %w", name, ErrDuplicate, err)
			}
		}

		return 0, err
	}

	return id, nil
}
