package services
import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

type LayoutService interface {
	Create(ctx context.Context, params domain.CreateLayoutParams) (*domain.Layout, error)
	GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Layout, error)
	List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.Layout, error)
	Update(ctx context.Context, params domain.UpdateLayoutParams) (*domain.Layout, error)
	Delete(ctx context.Context, id, workspaceID pgtype.UUID) error
	SetDefault(ctx context.Context, id, workspaceID pgtype.UUID) error
}

type layoutService struct {
	repo repositories.LayoutRepo
}

func NewLayoutService(repo repositories.LayoutRepo) *layoutService {
	return &layoutService{repo: repo}
}

func (s *layoutService) Create(ctx context.Context, params domain.CreateLayoutParams) (*domain.Layout, error) {
	layout, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return layout, nil
}

func (s *layoutService) GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Layout, error) {
	return s.repo.GetLayoutByID(ctx, id, workspaceID)
}

func (s *layoutService) List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.Layout, error) {
	return s.repo.List(ctx, workspaceID)
}

func (s *layoutService) Update(ctx context.Context, params domain.UpdateLayoutParams) (*domain.Layout, error) {

	updatedLayout, err := s.repo.UpdateLayout(ctx, params)
	if err != nil {
		return nil, err
	}

	return updatedLayout, nil
}

func (s *layoutService) Delete(ctx context.Context, id, workspaceID pgtype.UUID) error {
	current, err := s.repo.GetLayoutByID(ctx, id, workspaceID)
	if err != nil {
		return err
	}

	if current.IsDefault {
		return fmt.Errorf("%w: cannot delete a default layout. Set another layout as default first", apperrors.ErrInvalidInput)
	}

	return s.repo.DeleteLayout(ctx, id, workspaceID)
}

func (s *layoutService) SetDefault(ctx context.Context, id, workspaceID pgtype.UUID) error {
	
	return s.repo.SetLayoutDefault(ctx, id, workspaceID)
}
