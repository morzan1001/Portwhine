package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) store.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *store.Role) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*store.Role, error) {
	var role store.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get role by id: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*store.Role, error) {
	var role store.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, fmt.Errorf("get role by name: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]store.Role, error) {
	var roles []store.Role
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *store.Role) error {
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&store.Role{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}
