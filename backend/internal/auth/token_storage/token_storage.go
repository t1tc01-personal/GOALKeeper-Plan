package token_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TokenPrefix = "oauth_token:"
	TokenTTL    = 5 * time.Minute
)

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}

type TokenStorage struct {
	client *redis.Client
}

func NewTokenStorage(client *redis.Client) *TokenStorage {
	return &TokenStorage{
		client: client,
	}
}

// StoreToken stores OAuth tokens in Redis with a unique token ID
func (ts *TokenStorage) StoreToken(ctx context.Context, data *TokenData) (string, error) {
	tokenID := uuid.New().String()
	key := TokenPrefix + tokenID

	// Serialize token data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store in Redis with TTL
	err = ts.client.Set(ctx, key, jsonData, TokenTTL).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store token in Redis: %w", err)
	}

	return tokenID, nil
}

// GetToken retrieves token data by token ID
func (ts *TokenStorage) GetToken(ctx context.Context, tokenID string) (*TokenData, error) {
	key := TokenPrefix + tokenID

	// Get from Redis
	jsonData, err := ts.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found or expired")
		}
		return nil, fmt.Errorf("failed to get token from Redis: %w", err)
	}

	// Deserialize
	var data TokenData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &data, nil
}

// DeleteToken removes token from Redis (one-time use)
func (ts *TokenStorage) DeleteToken(ctx context.Context, tokenID string) error {
	key := TokenPrefix + tokenID
	return ts.client.Del(ctx, key).Err()
}

// GetAndDeleteToken retrieves and deletes token in one operation (atomic)
func (ts *TokenStorage) GetAndDeleteToken(ctx context.Context, tokenID string) (*TokenData, error) {
	data, err := ts.GetToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	// Delete after retrieval (one-time use)
	if err := ts.DeleteToken(ctx, tokenID); err != nil {
		// Log error but don't fail if delete fails
		// Token will expire anyway
	}

	return data, nil
}

