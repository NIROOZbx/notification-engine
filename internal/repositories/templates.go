package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type templateRepository struct {
	queries *sqlc.Queries
}

type TemplateRepository interface {
	Create(ctx context.Context, params domain.CreateTemplateParams) (*domain.Template, error)
	GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Template, error)
	List(ctx context.Context, workspaceID, envID pgtype.UUID) ([]*domain.Template, error)
	Update(ctx context.Context, params domain.UpdateTemplateParams) (*domain.Template, error)
	Delete(ctx context.Context, id, workspaceID pgtype.UUID) error
	HasActiveChannels(ctx context.Context, templateID pgtype.UUID) (bool, error)


	CreateChannel(ctx context.Context, params domain.CreateTemplateChannelParams) (*domain.TemplateChannel, error)
	ListChannels(ctx context.Context, templateID pgtype.UUID) ([]*domain.TemplateChannel, error)
	UpdateChannel(ctx context.Context, params domain.UpdateTemplateChannelParams) (*domain.TemplateChannel, error)
	DeleteChannel(ctx context.Context, id, templateID pgtype.UUID) error
}

func NewTemplateRepository(queries *sqlc.Queries) *templateRepository {
	return &templateRepository{queries: queries}
}

func (r *templateRepository) Create(ctx context.Context, params domain.CreateTemplateParams) (*domain.Template, error) {
	row, err := r.queries.CreateTemplate(ctx, sqlc.CreateTemplateParams{
		WorkspaceID:   params.WorkspaceID,
		EnvironmentID: params.EnvironmentID,
		LayoutID:      params.LayoutID,
		CreatedBy:     params.CreatedBy,
		Name:          params.Name,
		Description:   helpers.Text(params.Description),
		EventType:     params.EventType,
	})
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
            return nil, apperrors.ErrAlreadyExists
        }
		if apperrors.IsForeignKeyViolation(err) {
             return nil, fmt.Errorf("%w: invalid layout or user reference", apperrors.ErrInvalidInput)
        }
		return nil, fmt.Errorf("create template: %w", err)
	}
	return toTemplate(row), nil
}

func (r *templateRepository) GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Template, error) {
	row, err := r.queries.GetTemplateByID(ctx, sqlc.GetTemplateByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperrors.ErrTemplateNotFound
        }
        return nil, fmt.Errorf("get template: %w", err)
    }
	return toTemplate(row), nil
}

func (r *templateRepository) List(ctx context.Context, workspaceID, envID pgtype.UUID) ([]*domain.Template, error) {
	rows, err := r.queries.ListTemplates(ctx, sqlc.ListTemplatesParams{
		WorkspaceID:   workspaceID,
		EnvironmentID: envID,
	})
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	templates := make([]*domain.Template, len(rows))
	for i, row := range rows {
		templates[i] = toTemplate(row)
	}
	return templates, nil
}

func (r *templateRepository) Update(ctx context.Context, params domain.UpdateTemplateParams) (*domain.Template, error) {
	current, err := r.GetByID(ctx, params.ID, params.WorkspaceID)
	if err != nil {
		return nil, err
	}

	updateParams := sqlc.UpdateTemplateParams{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
		Name:        current.Name,
		Description: helpers.Text(current.Description),
		Status:      current.Status,
		LayoutID:    current.LayoutID,
	}

	if params.Name != nil {
		updateParams.Name = *params.Name
	}
	if params.Description != nil {
		updateParams.Description = helpers.Text(*params.Description)
	}
	if params.Status != nil {
		updateParams.Status = *params.Status
	}
	if params.LayoutID != nil {
		updateParams.LayoutID = *params.LayoutID
	}

	row, err := r.queries.UpdateTemplate(ctx, updateParams)
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return nil, apperrors.ErrAlreadyExists
		}
		return nil, fmt.Errorf("update template: %w", err)
	}
	return toTemplate(row), nil
}

func (r *templateRepository) Delete(ctx context.Context, id, workspaceID pgtype.UUID) error {
	err := r.queries.DeleteTemplate(ctx, sqlc.DeleteTemplateParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
        if apperrors.IsForeignKeyViolation(err) {
            return fmt.Errorf("%w: template is in use by channels", apperrors.ErrDependencyFailure)
        }
        return fmt.Errorf("db delete: %w", err)
    }
	return nil
}

func (r *templateRepository) HasActiveChannels(ctx context.Context, templateID pgtype.UUID) (bool, error) {
	hasActive, err := r.queries.HasActiveChannels(ctx, templateID)
	if err != nil {
		return false, fmt.Errorf("check active channels: %w", err)
	}
	return hasActive, nil
}

// ---- Template Channels ----

func (r *templateRepository) CreateChannel(ctx context.Context, params domain.CreateTemplateChannelParams) (*domain.TemplateChannel, error) {
	content, err := sonic.Marshal(params.Content)
	if err != nil {
		return nil, fmt.Errorf("marshal content: %w", err)
	}
	
	row, err := r.queries.CreateTemplateChannel(ctx, sqlc.CreateTemplateChannelParams{
		TemplateID:      params.TemplateID,
		ChannelConfigID: params.ChannelConfigID,
		Channel:         params.Channel,
		Content:         content,
		IsActive:        helpers.Bool(true), // Default to active
	})
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return nil, apperrors.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create template channel: %w", err)
	}
	return toTemplateChannel(row), nil
}

func (r *templateRepository) ListChannels(ctx context.Context, templateID pgtype.UUID) ([]*domain.TemplateChannel, error) {
	rows, err := r.queries.ListTemplateChannels(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("list template channels: %w", err)
	}
	channels := make([]*domain.TemplateChannel, len(rows))
	for i, row := range rows {
		channels[i] = toTemplateChannel(row)
	}
	return channels, nil
}

func (r *templateRepository) UpdateChannel(ctx context.Context, params domain.UpdateTemplateChannelParams) (*domain.TemplateChannel, error) {
	current, err := r.queries.GetTemplateChannelByID(ctx, params.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("template channel not found: %w", err)
	}

	content, err := sonic.Marshal(params.Content)
	if err != nil {
		return nil, fmt.Errorf("marshal content: %w", err)
	}

	updateParams := sqlc.UpdateTemplateChannelParams{
		ID:              params.ID,
		ChannelConfigID: current.ChannelConfigID,
		Content:         content,
		IsActive:        current.IsActive,
	}

	if params.IsActive != nil {
		updateParams.IsActive = helpers.Bool(*params.IsActive)
	}

	row, err := r.queries.UpdateTemplateChannel(ctx, updateParams)
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return nil, apperrors.ErrAlreadyExists
		}
		return nil, fmt.Errorf("update template channel: %w", err)
	}
	return toTemplateChannel(row), nil
}

func (r *templateRepository) DeleteChannel(ctx context.Context, id, templateID pgtype.UUID) error {
	err := r.queries.DeleteTemplateChannel(ctx, sqlc.DeleteTemplateChannelParams{
		ID:         id,
		TemplateID: templateID,
	})
	if err != nil {
		return fmt.Errorf("delete template channel: %w", err)
	}
	return nil
}

func toTemplate(row sqlc.Template) *domain.Template {
	return &domain.Template{
		ID:            row.ID,
		WorkspaceID:   row.WorkspaceID,
		EnvironmentID: row.EnvironmentID,
		LayoutID:      row.LayoutID,
		CreatedBy:     row.CreatedBy,
		Name:          row.Name,
		Description:   row.Description.String,
		EventType:     row.EventType,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func toTemplateChannel(row sqlc.TemplateChannel) *domain.TemplateChannel {
	return &domain.TemplateChannel{
		ID:              row.ID,
		TemplateID:      row.TemplateID,
		ChannelConfigID: row.ChannelConfigID,
		Channel:         row.Channel,
		Content:         row.Content,
		IsActive:        row.IsActive.Bool,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}