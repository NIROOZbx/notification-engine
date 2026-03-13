package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
    StoreRefreshToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error
	BlackListRefreshToken(ctx context.Context, tokenID string, ttl time.Time) error
	UpgradeTokenVersion(ctx context.Context, userID string) error
	DeleteRefreshToken(ctx context.Context, tokenID string) error
	GetTokenVersion(ctx context.Context, userID string) (int64, error)
	IsRefreshBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

type redisStore struct {
	client *redis.Client
}

func(r *redisStore)StoreRefreshToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error{
	
	key:=fmt.Sprintf("refresh:%s", tokenID)
	
	err:=r.client.Set(ctx,key,userID,expiry).Err()

	 if err != nil {
        return fmt.Errorf("storing refresh token: %w", err)
    }
    return nil

}

func (r *redisStore) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key:=fmt.Sprintf("refresh:%s", tokenID)
    if err := r.client.Del(ctx,key ).Err(); err != nil {
        return fmt.Errorf("deleting refresh token: %w", err)
    }
    return nil
}


func (r *redisStore) BlackListRefreshToken(ctx context.Context, tokenID string, ttl time.Time) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)

	expiresAt := time.Until(ttl)
	err := r.client.Set(ctx, key, true, expiresAt).Err()

	if err != nil {
		return err
	}

	return nil

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
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("key name %s does not exist", key)
	}

	return version, nil

}

func (r *redisStore) IsRefreshBlacklisted(ctx context.Context, tokenID string) (bool, error) {

	key := fmt.Sprintf("blacklist:%s", tokenID)

	count, err := r.client.Exists(ctx, key).Result()

	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}

	return count > 0, nil

}

func NewStore(client *redis.Client) Store {
	return &redisStore{
		client: client,
	}
}
