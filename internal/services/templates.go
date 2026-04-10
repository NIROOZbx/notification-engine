package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

type templateService struct {
	repo repositories.TemplateRepository
}

type TemplateService interface {
	Create(ctx context.Context, params domain.CreateTemplateParams) (*domain.Template, error)
	GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Template, error)
	List(ctx context.Context, workspaceID, envID pgtype.UUID) ([]*domain.Template, error)
	Update(ctx context.Context, params domain.UpdateTemplateParams) (*domain.Template, error)
	Delete(ctx context.Context, id, workspaceID pgtype.UUID) error

	// ---- Template Channels ----
	CreateChannel(ctx context.Context, params domain.CreateTemplateChannelParams) (*domain.TemplateChannel, error)
	ListChannels(ctx context.Context, templateID, workspaceID pgtype.UUID) ([]*domain.TemplateChannel, error)
	UpdateChannel(ctx context.Context, params domain.UpdateTemplateChannelParams) (*domain.TemplateChannel, error)
	DeleteChannel(ctx context.Context, id, templateID, workspaceID pgtype.UUID) error
}

var validChannels = map[string]bool{
	"email": true, "sms": true, "push": true,
	"slack": true, "whatsapp": true, "webhook": true, "in_app": true,
}

var validStatuses = map[string]bool{
	"draft": true, "live": true, "dropped": true,
}

func NewTemplateService(repo repositories.TemplateRepository) *templateService {
	return &templateService{
		repo: repo,
	}
}

func (s *templateService) Create(ctx context.Context, params domain.CreateTemplateParams) (*domain.Template, error) {
	return s.repo.Create(ctx, params)
}

func (s *templateService) GetByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.Template, error) {
	return s.repo.GetByID(ctx, id, workspaceID)
}

func (s *templateService) List(ctx context.Context, workspaceID, envID pgtype.UUID) ([]*domain.Template, error) {
	return s.repo.List(ctx, workspaceID, envID)
}

func (s *templateService) Update(ctx context.Context, params domain.UpdateTemplateParams) (*domain.Template, error) {
	if params.Status != nil && !validStatuses[*params.Status] {
		return nil, fmt.Errorf("%w: invalid status %s", apperrors.ErrInvalidInput, *params.Status)
	}
	
	current, err := s.repo.GetByID(ctx, params.ID, params.WorkspaceID)
    if err != nil {
        return nil, err 
    }

	if params.Status != nil && *params.Status == "live" && current.Status != "live" {
		hasActive, err := s.repo.HasActiveChannels(ctx, params.ID)
		if err != nil {
			return nil, err
		}
		if !hasActive {
			return nil, apperrors.ErrNoActiveChannels
		}
	}

	if current.Status == "dropped" {
		isRecovering := params.Status != nil && (*params.Status == "draft" || *params.Status == "live")
		if !isRecovering {
			return nil, fmt.Errorf("%w: cannot update dropped template unless restoring status", apperrors.ErrInvalidInput)
		}
	}

	return s.repo.Update(ctx, params)
}

func (s *templateService) Delete(ctx context.Context, id, workspaceID pgtype.UUID) error {
	current, err := s.repo.GetByID(ctx, id, workspaceID)
	if err != nil {
		return err
	}
	if current.Status == "live" {
		return fmt.Errorf("%w: cannot delete a live template. switch to draft or dropped first", apperrors.ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id, workspaceID)
}

// ---- Template Channels ----

func (s *templateService) CreateChannel(ctx context.Context, params domain.CreateTemplateChannelParams) (*domain.TemplateChannel, error) {
	if !validChannels[params.Channel] {
		return nil, fmt.Errorf("invalid channel: %s", params.Channel)
	}
	if len(params.Content) == 0 {
		return nil, fmt.Errorf("content is required")
	}
	
	_, err := s.repo.GetByID(ctx, params.TemplateID, params.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("template not found")
	}
	
	return s.repo.CreateChannel(ctx, params)
}

func (s *templateService) ListChannels(ctx context.Context, templateID, workspaceID pgtype.UUID) ([]*domain.TemplateChannel, error) {
	_, err := s.repo.GetByID(ctx, templateID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("template not found")
	}
	return s.repo.ListChannels(ctx, templateID)
}

func (s *templateService) UpdateChannel(ctx context.Context, params domain.UpdateTemplateChannelParams) (*domain.TemplateChannel, error) {
	_, err := s.repo.GetByID(ctx, params.TemplateID, params.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("template not found")
	}
	return s.repo.UpdateChannel(ctx, params)
}

func (s *templateService) DeleteChannel(ctx context.Context, id, templateID, workspaceID pgtype.UUID) error {
	_, err := s.repo.GetByID(ctx, templateID, workspaceID)
	if err != nil {
		return fmt.Errorf("template not found")
	}
	return s.repo.DeleteChannel(ctx, id, templateID)
}