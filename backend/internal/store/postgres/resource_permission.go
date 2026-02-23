package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type resourcePermissionRepository struct {
	db *gorm.DB
}

func NewResourcePermissionRepository(db *gorm.DB) store.ResourcePermissionRepository {
	return &resourcePermissionRepository{db: db}
}

func (r *resourcePermissionRepository) Grant(ctx context.Context, perm *store.ResourcePermission) error {
	if err := r.db.WithContext(ctx).Create(perm).Error; err != nil {
		return fmt.Errorf("grant permission: %w", err)
	}
	return nil
}

func (r *resourcePermissionRepository) Revoke(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&store.ResourcePermission{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("revoke permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}

func (r *resourcePermissionRepository) ListForSubject(ctx context.Context, subjectType, subjectID string) ([]store.ResourcePermission, error) {
	var perms []store.ResourcePermission
	if err := r.db.WithContext(ctx).
		Where("subject_type = ? AND subject_id = ?", subjectType, subjectID).
		Order("created_at DESC").
		Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("list permissions for subject: %w", err)
	}
	return perms, nil
}

func (r *resourcePermissionRepository) ListForResource(ctx context.Context, resourceType, resourceID string) ([]store.ResourcePermission, error) {
	var perms []store.ResourcePermission
	if err := r.db.WithContext(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Order("created_at DESC").
		Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("list permissions for resource: %w", err)
	}
	return perms, nil
}

func (r *resourcePermissionRepository) CheckPermission(ctx context.Context, subjectType, subjectID, resourceType, resourceID, action string) (*store.ResourcePermission, error) {
	var perm store.ResourcePermission
	err := r.db.WithContext(ctx).
		Where("subject_type = ? AND subject_id = ? AND resource_type = ? AND resource_id = ? AND effect = ? AND (action = ? OR action = ?)",
			subjectType, subjectID, resourceType, resourceID, "allow", action, "*").
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *resourcePermissionRepository) CheckDeny(ctx context.Context, subjectType, subjectID, resourceType, resourceID, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&store.ResourcePermission{}).
		Where("subject_type = ? AND subject_id = ? AND resource_type = ? AND resource_id = ? AND effect = ? AND (action = ? OR action = ?)",
			subjectType, subjectID, resourceType, resourceID, "deny", action, "*").
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check deny: %w", err)
	}
	return count > 0, nil
}

func (r *resourcePermissionRepository) ListAccessibleResourceIDs(ctx context.Context, subjectType, subjectID, resourceType string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&store.ResourcePermission{}).
		Select("DISTINCT resource_id").
		Where("subject_type = ? AND subject_id = ? AND resource_type = ? AND effect = ?",
			subjectType, subjectID, resourceType, "allow").
		Pluck("resource_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("list accessible resource ids: %w", err)
	}
	return ids, nil
}
