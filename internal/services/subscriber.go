package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/conversion"
)

type IdentifySubscriberInput struct {
	WorkspaceID    string
	EnvironmentID  string
	ExternalUserID string
	Channel        string
	ContactValue   string
	Metadata       map[string]any
}

type UpsertPreferenceInput struct {
	WorkspaceID    string
	EnvironmentID  string
	ExternalUserID string
	Channel        string
	EventType      string
	IsEnabled      bool
}

type SubscriberService interface {
	Identify(ctx context.Context, input IdentifySubscriberInput) (*domain.Subscriber, error)
	UpsertPreference(ctx context.Context, input UpsertPreferenceInput) (*domain.UserPreference, error)
	Delete(ctx context.Context, id, workspaceID string) error
	List(ctx context.Context, workspaceID, environmentID string) ([]*domain.Subscriber, error)
}

type subscriberService struct {
	repo repositories.SubscriberRepo
}

func NewSubscriberService(repo repositories.SubscriberRepo) SubscriberService {
	return &subscriberService{repo: repo}
}

func (s *subscriberService) Identify(ctx context.Context, input IdentifySubscriberInput) (*domain.Subscriber, error) {
	if !consts.ValidChannels[input.Channel] {
		return nil, fmt.Errorf("invalid channel: %s", input.Channel)
	}

	metadataBytes, err := conversion.JSONBFromMap(input.Metadata) 
	if err != nil {
		return nil, err
	}

	params := sqlc.UpsertUserContactInfoParams{
		WorkspaceID:    utils.MustStringToUUID(input.WorkspaceID),
		EnvironmentID:  utils.MustStringToUUID(input.EnvironmentID),
		ExternalUserID: input.ExternalUserID,
		Channel:        input.Channel,
		ContactValue:   input.ContactValue,
		Verified:       helpers.Bool(false), 
		Metadata:       metadataBytes,
	}

	subscriber, err := s.repo.UpsertSubscriber(ctx, params)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *subscriberService) UpsertPreference(ctx context.Context, input UpsertPreferenceInput) (*domain.UserPreference, error) {
	if !consts.ValidChannels[input.Channel] {
		return nil, fmt.Errorf("invalid channel: %s", input.Channel)
	}

	subscriber, err := s.repo.GetSubscriberByExternalIDAndChannel(ctx, input.WorkspaceID, input.EnvironmentID, input.ExternalUserID, input.Channel)
	if err != nil {
		return nil, fmt.Errorf("subscriber not found for channel %s. Please identify them first.", input.Channel)
	}

	params := sqlc.UpsertUserPreferenceParams{
		WorkspaceID:   utils.MustStringToUUID(input.WorkspaceID),
		EnvironmentID: utils.MustStringToUUID(input.EnvironmentID),
		SubscriberID:  utils.MustStringToUUID(subscriber.ID),
		Channel:       input.Channel,
		EventType:     helpers.Text(input.EventType),
		IsEnabled:     input.IsEnabled,
	}

	pref, err := s.repo.UpsertPreference(ctx, params)
	if err != nil {
		return nil, err
	}

	pref.ExternalUserID = input.ExternalUserID
	return pref, nil
}

func (s *subscriberService) Delete(ctx context.Context, id, workspaceID string) error {
	idUUID, err := utils.StringToUUID(id)
	if err != nil {
		return fmt.Errorf("%w: subscriber id", apperrors.ErrInvalidInput)
	}

	workspaceUUID, err := utils.StringToUUID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace id", apperrors.ErrInvalidInput)
	}

	return s.repo.DeleteSubscriber(ctx, idUUID, workspaceUUID)
}

func (s *subscriberService) List(ctx context.Context, workspaceID, environmentID string) ([]*domain.Subscriber, error) {
	workspaceUUID, err := utils.StringToUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace id", apperrors.ErrInvalidInput)
	}

	envUUID, err := utils.StringToUUID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: environment id", apperrors.ErrInvalidInput)
	}

	return s.repo.ListSubscribers(ctx, workspaceUUID, envUUID)
}
