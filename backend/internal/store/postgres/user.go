package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) store.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *store.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*store.User, error) {
	var user store.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*store.User, error) {
	var user store.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*store.User, error) {
	var user store.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]store.User, int64, error) {
	var users []store.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&store.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if err := r.db.WithContext(ctx).Preload("Role").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *store.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&store.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
