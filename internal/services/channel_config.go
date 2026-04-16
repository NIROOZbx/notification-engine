package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/encryptor"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgtype"
)

type channelConfigService struct {
	repo      repositories.ChannelConfigRepo
	secretKey string
}

type ChannelConfigService interface {
	Create(ctx context.Context, params domain.CreateChannelConfigParams) (*domain.ChannelConfig, error)
	List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.ChannelConfig, error)
	GetChannelConfigByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.ChannelConfig, error)
	GetDefaultChannelConfig(ctx context.Context, workspaceID pgtype.UUID, channel string) (*domain.ChannelConfig, error)
	UpdateChannelConfig(ctx context.Context, params domain.UpdateChannelConfigParams) (*domain.ChannelConfig, error)
	DeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error
	SetChannelConfigDefault(ctx context.Context, id, workspaceID pgtype.UUID) error
}

func NewChannelConfigService(repo repositories.ChannelConfigRepo, secretKey string) *channelConfigService {
	return &channelConfigService{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (s *channelConfigService) Create(ctx context.Context, params domain.CreateChannelConfigParams) (*domain.ChannelConfig, error) {
	if err := s.validate(params); err != nil {
		return nil, err
	}

	credBytes, err := sonic.Marshal(params.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to format credentials: %w", err)
	}

	encryptedString, err := encryptor.Encrypt(credBytes, s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	result, err := s.repo.Create(ctx, encryptedString, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *channelConfigService) List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.ChannelConfig, error) {
	return s.repo.List(ctx, workspaceID)
}

func (s *channelConfigService) GetChannelConfigByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.ChannelConfig, error) {
	return s.repo.GetChannelConfigByID(ctx, id, workspaceID)
}

func (s *channelConfigService) GetDefaultChannelConfig(ctx context.Context, workspaceID pgtype.UUID, channel string) (*domain.ChannelConfig, error) {
	if !consts.ValidChannels[channel] {
		return nil, fmt.Errorf("invalid channel: %s", channel)
	}
	return s.repo.GetDefaultChannelConfig(ctx, workspaceID, channel)
}

func (s *channelConfigService) UpdateChannelConfig(ctx context.Context, params domain.UpdateChannelConfigParams) (*domain.ChannelConfig, error) {
	var encrypted *string
	if params.Credentials != nil {
		credBytes, err := sonic.Marshal(params.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to format credentials: %w", err)
		}

		encryptedData, err := encryptor.Encrypt(credBytes, s.secretKey)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		encrypted = &encryptedData
	}

	cfg, err := s.repo.UpdateChannelConfig(ctx, encrypted, params)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *channelConfigService) DeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error {
	config, err := s.repo.GetChannelConfigByID(ctx, id, workspaceID)
	if err != nil {
		return err
	}

	if config.IsDefault {
		count, err := s.repo.CountActiveProvidersForChannel(ctx, workspaceID, config.Channel)
		if err != nil {
			return err
		}
		if count > 1 {
			return fmt.Errorf("cannot delete the default %s provider. Please set another provider as default first", config.Channel)
		}
	}else {
        if err := s.repo.RemoveProviderOverride(ctx, id, workspaceID); err != nil {
            return fmt.Errorf("failed to detach template overrides: %w", err)
        }
    }
	return s.repo.DeleteChannelConfig(ctx, id, workspaceID)
}

func (s *channelConfigService) SetChannelConfigDefault(ctx context.Context, id, workspaceID pgtype.UUID) error {
	result, err := s.repo.GetChannelConfigByID(ctx, id, workspaceID)
	if err != nil {
		return err
	}

	if !result.IsActive {
		return apperrors.ErrInactiveProvider
	}
	return s.repo.SetChannelConfigDefault(ctx, id, workspaceID)
}

func (s *channelConfigService) decryptCredentials(cfg *domain.ChannelConfig) error {
	if cfg == nil || cfg.Encrypted == "" {
		return nil
	}

	decryptedBytes, err := encryptor.Decrypt(cfg.Encrypted, s.secretKey)
	if err != nil {
		return apperrors.ErrDecryptionFailed
	}

	var originalCreds map[string]any
	if err := sonic.Unmarshal(decryptedBytes, &originalCreds); err != nil {
		return fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	cfg.Credentials = originalCreds
	return nil
}

func (s *channelConfigService) validate(params domain.CreateChannelConfigParams) error {

	validProviders, ok := consts.ValidProvidersByChannel[params.Channel]
	if !ok {
		return fmt.Errorf("unsupported channel: %s", params.Channel)
	}

	if !validProviders[params.Provider] {
		return fmt.Errorf("provider %q is not valid for channel %q", params.Provider, params.Channel)
	}

	if len(params.Credentials) == 0 {
		return fmt.Errorf("credentials cannot be empty")
	}

	return nil
}
