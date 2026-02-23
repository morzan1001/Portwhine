package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) store.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, entry *store.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, offset, limit int) ([]store.AuditLog, error) {
	var entries []store.AuditLog
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	return entries, nil
}

func (r *auditLogRepository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]store.AuditLog, error) {
	var entries []store.AuditLog
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list audit log by user: %w", err)
	}
	return entries, nil
}
