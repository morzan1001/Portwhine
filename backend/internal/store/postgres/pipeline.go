package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type pipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) store.PipelineRepository {
	return &pipelineRepository{db: db}
}

func (r *pipelineRepository) Create(ctx context.Context, pipeline *store.Pipeline) error {
	if err := r.db.WithContext(ctx).Create(pipeline).Error; err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}
	return nil
}

func (r *pipelineRepository) GetByID(ctx context.Context, id string) (*store.Pipeline, error) {
	var pipeline store.Pipeline
	if err := r.db.WithContext(ctx).Preload("CreatedBy").First(&pipeline, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get pipeline by id: %w", err)
	}
	return &pipeline, nil
}

func (r *pipelineRepository) List(ctx context.Context, offset, limit int) ([]store.Pipeline, int64, error) {
	var pipelines []store.Pipeline
	var total int64

	if err := r.db.WithContext(ctx).Model(&store.Pipeline{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipelines: %w", err)
	}

	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Offset(offset).Limit(limit).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, 0, fmt.Errorf("list pipelines: %w", err)
	}

	return pipelines, total, nil
}

func (r *pipelineRepository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]store.Pipeline, int64, error) {
	var pipelines []store.Pipeline
	var total int64

	query := r.db.WithContext(ctx).Model(&store.Pipeline{}).Where("is_active = ? AND created_by_id = ?", true, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user pipelines: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, 0, fmt.Errorf("list user pipelines: %w", err)
	}

	return pipelines, total, nil
}

func (r *pipelineRepository) ListByIDs(ctx context.Context, ids []string, offset, limit int) ([]store.Pipeline, int64, error) {
	var pipelines []store.Pipeline
	var total int64

	query := r.db.WithContext(ctx).Model(&store.Pipeline{}).Where("is_active = ? AND id IN ?", true, ids)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipelines by ids: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, 0, fmt.Errorf("list pipelines by ids: %w", err)
	}

	return pipelines, total, nil
}

func (r *pipelineRepository) Update(ctx context.Context, pipeline *store.Pipeline) error {
	if err := r.db.WithContext(ctx).Save(pipeline).Error; err != nil {
		return fmt.Errorf("update pipeline: %w", err)
	}
	return nil
}

func (r *pipelineRepository) Delete(ctx context.Context, id string) error {
	// Soft delete: set is_active = false
	if err := r.db.WithContext(ctx).Model(&store.Pipeline{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("delete pipeline: %w", err)
	}
	return nil
}

func (r *pipelineRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&store.Pipeline{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count pipelines: %w", err)
	}
	return total, nil
}
