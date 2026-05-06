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
	"github.com/NIROOZbx/notification-engine/pkg/parallel"
)

type ContactInput struct {
	Channel      string
	ContactValue string
}

type IdentifySubscriberInput struct {
	WorkspaceID    string
	EnvironmentID  string
	ExternalUserID string
	Contacts       []ContactInput
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
	Identify(ctx context.Context, input IdentifySubscriberInput) ([]*domain.Subscriber, error)
	UpsertPreference(ctx context.Context, input UpsertPreferenceInput) (*domain.UserPreference, error)
	Delete(ctx context.Context, id, workspaceID string) error
	List(ctx context.Context, workspaceID, environmentID string, page, pageSize int32) (*domain.SubscriberList, error)
	GetPreferencesByExternalID(ctx context.Context, workspaceID, environmentID, externalUserID string) ([]*domain.UserPreference, error)
}

type subscriberService struct {
	repo repositories.SubscriberRepo
}

func NewSubscriberService(repo repositories.SubscriberRepo) *subscriberService {
	return &subscriberService{repo: repo}
}

func (s *subscriberService) Identify(ctx context.Context, input IdentifySubscriberInput) ([]*domain.Subscriber, error) {
	metadataBytes, err := conversion.JSONBFromMap(input.Metadata) 
	if err != nil {
		return nil, err
	}

	var subscribers []*domain.Subscriber

	for _, contact := range input.Contacts {
		if !consts.ValidChannels[contact.Channel] {
			return nil, fmt.Errorf("invalid channel: %s", contact.Channel)
		}

		params := sqlc.UpsertUserContactInfoParams{
			WorkspaceID:    utils.MustStringToUUID(input.WorkspaceID),
			EnvironmentID:  utils.MustStringToUUID(input.EnvironmentID),
			ExternalUserID: input.ExternalUserID,
			Channel:        contact.Channel,
			ContactValue:   contact.ContactValue,
			Verified:       helpers.Bool(false), 
			Metadata:       metadataBytes,
		}

		subscriber, err := s.repo.UpsertSubscriber(ctx, params)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
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

func (s *subscriberService) List(ctx context.Context, workspaceID, environmentID string, page, pageSize int32) (*domain.SubscriberList, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	workspaceUUID, err := utils.StringToUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace id", apperrors.ErrInvalidInput)
	}

	envUUID, err := utils.StringToUUID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: environment id", apperrors.ErrInvalidInput)
	}

	offset := (page - 1) * pageSize
	subscribers, totalCount, err := parallel.Query2(ctx,
		func(c context.Context) ([]*domain.Subscriber, error) {
			return s.repo.ListSubscribers(c, workspaceUUID, envUUID, pageSize, offset)
		},
		func(c context.Context) (int64, error) {
			return s.repo.CountSubscribers(c, workspaceUUID, envUUID)
		},
	)

	if err != nil {
		return nil, err
	}

	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &domain.SubscriberList{
		Subscribers: subscribers,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *subscriberService) GetPreferencesByExternalID(ctx context.Context, workspaceID, environmentID, externalUserID string) ([]*domain.UserPreference, error) {
	workspaceUUID, err := utils.StringToUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace id", apperrors.ErrInvalidInput)
	}

	envUUID, err := utils.StringToUUID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: environment id", apperrors.ErrInvalidInput)
	}

	return s.repo.ListPreferences(ctx, workspaceUUID, envUUID, externalUserID)
}



