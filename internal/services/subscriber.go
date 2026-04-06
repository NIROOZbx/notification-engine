package services

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
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
	SubscriberID   string
	Channel        string
	EventType      string
	IsEnabled      bool
}

type SubscriberService interface {
	Identify(ctx context.Context, input IdentifySubscriberInput) (*sqlc.UserInfo, error)
	UpsertPreference(ctx context.Context, input UpsertPreferenceInput) (*sqlc.UserPreference, error)
}

type subscriberService struct {
	repo repositories.SubscriberRepo
}

func NewSubscriberService(repo repositories.SubscriberRepo) SubscriberService {
	return &subscriberService{repo: repo}
}

func (s *subscriberService) Identify(ctx context.Context, input IdentifySubscriberInput) (*sqlc.UserInfo, error) {
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

	userInfo, err := s.repo.UpsertSubscriber(ctx, params)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *subscriberService) UpsertPreference(ctx context.Context, input UpsertPreferenceInput) (*sqlc.UserPreference, error) {
	params := sqlc.UpsertUserPreferenceParams{
		WorkspaceID:   utils.MustStringToUUID(input.WorkspaceID),
		EnvironmentID: utils.MustStringToUUID(input.EnvironmentID),
		SubscriberID:  utils.MustStringToUUID(input.SubscriberID),
		Channel:       input.Channel,
		EventType:     helpers.Text(input.EventType),
		IsEnabled:     input.IsEnabled,
	}

	pref, err := s.repo.UpsertPreference(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pref, nil
}
