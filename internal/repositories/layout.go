package repositories

import (
	"context"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

type layoutRepo struct {
	queries *sqlc.Queries
}

type LayoutRepo interface {
	Create(ctx context.Context, params domain.CreateLayoutParams) (*domain.Layout, error)
	List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.Layout, error)
	GetLayoutByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Layout, error)
	GetDefaultLayout(ctx context.Context, workspaceID pgtype.UUID) (*domain.Layout, error)
	UpdateLayout(ctx context.Context, params domain.UpdateLayoutParams) (*domain.Layout, error)
	DeleteLayout(ctx context.Context, id, workspaceID pgtype.UUID) error
	SetLayoutDefault(ctx context.Context, id, workspaceID pgtype.UUID) error
}

func NewLayoutRepo(queries *sqlc.Queries) *layoutRepo {
	return &layoutRepo{
		queries: queries,
	}
}

func (l *layoutRepo) Create(ctx context.Context, params domain.CreateLayoutParams) (*domain.Layout, error) {
	layout, err := l.queries.CreateLayout(ctx, sqlc.CreateLayoutParams{
		WorkspaceID: params.WorkspaceID,
		Name:        params.Name,
		Html:        params.Html,
		IsDefault:   params.IsDefault,
	})

	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToLayout(layout), nil

}
func (l *layoutRepo) List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.Layout, error) {

	rows, err := l.queries.ListLayouts(ctx, workspaceID)

	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToLayouts(rows), nil

}
func (l *layoutRepo) GetLayoutByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Layout, error) {

	layout, err := l.queries.GetLayoutByID(ctx, sqlc.GetLayoutByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})

	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToLayout(layout), nil

}

func (l *layoutRepo) GetDefaultLayout(ctx context.Context, workspaceID pgtype.UUID) (*domain.Layout, error) {
	layout, err := l.queries.GetDefaultLayout(ctx, workspaceID)
	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToLayout(layout), nil
}

func (l *layoutRepo) DeleteLayout(ctx context.Context, id, workspaceID pgtype.UUID) error {
	err := l.queries.DeleteLayout(ctx, sqlc.DeleteLayoutParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return apperrors.MapDBError(err)
	}
	return nil
}

func (l *layoutRepo) UpdateLayout(ctx context.Context, params domain.UpdateLayoutParams) (*domain.Layout, error) {
	current, err := l.GetLayoutByID(ctx, params.ID, params.WorkspaceID)
	if err != nil {
		return nil, err
	}

	updateParams := sqlc.UpdateLayoutParams{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
		Name:        current.Name,
		Html:        current.Html,
	}

	if params.Name != "" {
		updateParams.Name = params.Name
	}
	if params.Html != "" {
		updateParams.Html = params.Html
	}

	row, err := l.queries.UpdateLayout(ctx, updateParams)
	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToLayout(row), nil
}

func (l *layoutRepo) SetLayoutDefault(ctx context.Context, id, workspaceID pgtype.UUID) error {
	err := l.queries.SetLayoutDefault(ctx, sqlc.SetLayoutDefaultParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return apperrors.MapDBError(err)
	}
	return nil
}

func mapToLayout(layout sqlc.Layout) *domain.Layout {

	return &domain.Layout{
		ID:          layout.ID,
		WorkspaceID: layout.WorkspaceID,
		Name:        layout.Name,
		IsDefault:   layout.IsDefault,
		Html:        layout.Html,
		CreatedAt:   layout.CreatedAt.Time,
		UpdatedAt:   layout.UpdatedAt.Time,
	}

}

func mapToLayouts(layouts []sqlc.Layout) []*domain.Layout {
	result := make([]*domain.Layout, len(layouts))
	for i, layout := range layouts {
		result[i] = mapToLayout(layout)
	}
	return result
}
