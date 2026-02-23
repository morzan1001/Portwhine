package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type pipelineRunRepository struct {
	db *gorm.DB
}

func NewPipelineRunRepository(db *gorm.DB) store.PipelineRunRepository {
	return &pipelineRunRepository{db: db}
}

func (r *pipelineRunRepository) Create(ctx context.Context, run *store.PipelineRun) error {
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return fmt.Errorf("create pipeline run: %w", err)
	}
	return nil
}

func (r *pipelineRunRepository) GetByID(ctx context.Context, id string) (*store.PipelineRun, error) {
	var run store.PipelineRun
	if err := r.db.WithContext(ctx).Preload("Pipeline").First(&run, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get pipeline run by id: %w", err)
	}
	return &run, nil
}

func (r *pipelineRunRepository) ListAll(ctx context.Context, offset, limit int) ([]store.PipelineRun, int64, error) {
	var runs []store.PipelineRun
	var total int64

	query := r.db.WithContext(ctx).Model(&store.PipelineRun{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count all pipeline runs: %w", err)
	}

	if err := r.db.WithContext(ctx).Preload("Pipeline").Offset(offset).Limit(limit).Order("created_at DESC").Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("list all pipeline runs: %w", err)
	}

	return runs, total, nil
}

func (r *pipelineRunRepository) ListByPipeline(ctx context.Context, pipelineID string, offset, limit int) ([]store.PipelineRun, int64, error) {
	var runs []store.PipelineRun
	var total int64

	query := r.db.WithContext(ctx).Model(&store.PipelineRun{}).Where("pipeline_id = ?", pipelineID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipeline runs: %w", err)
	}

	if err := query.Preload("Pipeline").Offset(offset).Limit(limit).Order("created_at DESC").Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("list pipeline runs: %w", err)
	}

	return runs, total, nil
}

func (r *pipelineRunRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]store.PipelineRun, int64, error) {
	var runs []store.PipelineRun
	var total int64

	query := r.db.WithContext(ctx).Model(&store.PipelineRun{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipeline runs by status: %w", err)
	}

	if err := query.Preload("Pipeline").Offset(offset).Limit(limit).Order("created_at DESC").Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("list pipeline runs by status: %w", err)
	}

	return runs, total, nil
}

func (r *pipelineRunRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if err := r.db.WithContext(ctx).Model(&store.PipelineRun{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("update pipeline run status: %w", err)
	}
	return nil
}

func (r *pipelineRunRepository) Update(ctx context.Context, run *store.PipelineRun) error {
	if err := r.db.WithContext(ctx).Save(run).Error; err != nil {
		return fmt.Errorf("update pipeline run: %w", err)
	}
	return nil
}

func (r *pipelineRunRepository) CreateStepResult(ctx context.Context, result *store.StepResult) error {
	if err := r.db.WithContext(ctx).Create(result).Error; err != nil {
		return fmt.Errorf("create step result: %w", err)
	}
	return nil
}

func (r *pipelineRunRepository) GetStepResult(ctx context.Context, id string) (*store.StepResult, error) {
	var result store.StepResult
	if err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get step result by id: %w", err)
	}
	return &result, nil
}

func (r *pipelineRunRepository) ListStepResults(ctx context.Context, runID string) ([]store.StepResult, error) {
	var results []store.StepResult
	if err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("created_at ASC").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("list step results: %w", err)
	}
	return results, nil
}

func (r *pipelineRunRepository) UpdateStepResult(ctx context.Context, result *store.StepResult) error {
	if err := r.db.WithContext(ctx).Save(result).Error; err != nil {
		return fmt.Errorf("update step result: %w", err)
	}
	return nil
}

func (r *pipelineRunRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&store.PipelineRun{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count all pipeline runs: %w", err)
	}
	return total, nil
}

func (r *pipelineRunRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&store.PipelineRun{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count pipeline runs by status: %w", err)
	}
	return total, nil
}

func (r *pipelineRunRepository) ListRecent(ctx context.Context, limit int) ([]store.PipelineRun, error) {
	var runs []store.PipelineRun
	if err := r.db.WithContext(ctx).Preload("Pipeline").Order("created_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("list recent pipeline runs: %w", err)
	}
	return runs, nil
}
