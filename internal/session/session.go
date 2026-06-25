package session

import (
	"context"
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/serializer"
	"github.com/redis/go-redis/v9"
)

type Store interface {
	StoreRefreshToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error
	BlackListRefreshToken(ctx context.Context, tokenID string, ttl time.Time) error
	UpgradeTokenVersion(ctx context.Context, userID string) error
	DeleteRefreshToken(ctx context.Context, tokenID string) error
	GetTokenVersion(ctx context.Context, userID string) (int64, error)
	IsRefreshBlacklisted(ctx context.Context, tokenID string) (time.Time, error)
	ClaimRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) (alreadyClaimed bool, err error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SetObject(ctx context.Context, key string, value any, ttl time.Duration) error
	GetObject(ctx context.Context, key string, target any) error
	Delete(ctx context.Context, key string) error
}

type redisStore struct {
	client *redis.Client
}

func NewStore(client *redis.Client) Store {
	return &redisStore{
		client: client,
	}
}

func (r *redisStore) StoreRefreshToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error {

	key := fmt.Sprintf("refresh:%s", tokenID)

	err := r.client.Set(ctx, key, userID, expiry).Err()

	if err != nil {
		return fmt.Errorf("storing refresh token: %w", err)
	}
	return nil

}

func (r *redisStore) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("refresh:%s", tokenID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("deleting refresh token: %w", err)
	}
	return nil
}

func (r *redisStore) BlackListRefreshToken(ctx context.Context, tokenID string, ttl time.Time) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	timestamp := helpers.ToUnixTimestamp(time.Now())
	expiresAt := time.Until(ttl)
	return r.client.Set(ctx, key, timestamp, expiresAt).Err()
}

func (r *redisStore) UpgradeTokenVersion(ctx context.Context, userID string) error {

	key := fmt.Sprintf("auth:user:%s:version", userID)

	err := r.client.Incr(ctx, key).Err()

	if err != nil {
		return fmt.Errorf("redis incr version error: %w", err)
	}

	return nil

}

func (r *redisStore) GetTokenVersion(ctx context.Context, userID string) (int64, error) {

	key := fmt.Sprintf("auth:user:%s:version", userID)

	version, err := r.client.Get(ctx, key).Int64()

	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("getting token version: %w", err)
	}

	return version, nil

}

func (r *redisStore) IsRefreshBlacklisted(ctx context.Context, tokenID string) (time.Time, error) {

	key := fmt.Sprintf("blacklist:%s", tokenID)

	issuedTime, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return time.Time{}, nil
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("redis check failed: %w", err)
	}
	return helpers.FromUnixTimestamp(issuedTime)

}

func (r *redisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisStore) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisStore) SetObject(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := serializer.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisStore) GetObject(ctx context.Context, key string, target any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return serializer.Unmarshal([]byte(val), target)
}

func (r *redisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisStore) ClaimRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("refresh-claimed:%s", tokenID)
	set, err := r.client.SetNX(ctx, key, time.Now().UTC().Format(time.RFC3339), ttl).Result()
	if err != nil {
		return false, fmt.Errorf("claiming refresh token: %w", err)
	}
	// SetNX returns true if WE set it (first caller), false if key already existed
	return !set, nil
}
