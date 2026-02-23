package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) store.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *store.APIKey) error {
	if err := r.db.WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*store.APIKey, error) {
	var key store.APIKey
	if err := r.db.WithContext(ctx).Preload("User").Preload("User.Role").First(&key, "key_prefix = ? AND revoked_at IS NULL", prefix).Error; err != nil {
		return nil, fmt.Errorf("get api key by prefix: %w", err)
	}
	return &key, nil
}

func (r *apiKeyRepository) ListByUser(ctx context.Context, userID string) ([]store.APIKey, error) {
	var keys []store.APIKey
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&store.APIKey{}).Where("id = ?", id).Update("last_used", &now).Error; err != nil {
		return fmt.Errorf("update api key last used: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&store.APIKey{}).Where("id = ?", id).Update("revoked_at", &now).Error; err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}
