package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

const orderCreatedAtASC = "created_at ASC"

type dataItemRepository struct {
	db *gorm.DB
}

func NewDataItemRepository(db *gorm.DB) store.DataItemRepository {
	return &dataItemRepository{db: db}
}

func (r *dataItemRepository) Create(ctx context.Context, item *store.DataItemRecord) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("create data item: %w", err)
	}
	return nil
}

func (r *dataItemRepository) CreateBatch(ctx context.Context, items []store.DataItemRecord) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(items, 100).Error; err != nil {
		return fmt.Errorf("create data items batch: %w", err)
	}
	return nil
}

func (r *dataItemRepository) GetByID(ctx context.Context, id string) (*store.DataItemRecord, error) {
	var item store.DataItemRecord
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get data item by id: %w", err)
	}
	return &item, nil
}

func (r *dataItemRepository) ListByRun(ctx context.Context, runID string, offset, limit int) ([]store.DataItemRecord, int64, error) {
	var items []store.DataItemRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&store.DataItemRecord{}).Where("run_id = ?", runID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count data items: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Order(orderCreatedAtASC).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list data items: %w", err)
	}

	return items, total, nil
}

func (r *dataItemRepository) ListByRunAndType(ctx context.Context, runID, itemType string, offset, limit int) ([]store.DataItemRecord, int64, error) {
	var items []store.DataItemRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&store.DataItemRecord{}).Where("run_id = ? AND type = ?", runID, itemType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count data items by type: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Order(orderCreatedAtASC).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list data items by type: %w", err)
	}

	return items, total, nil
}

func (r *dataItemRepository) Search(ctx context.Context, params store.DataItemSearchParams, offset, limit int) ([]store.DataItemRecord, int64, error) {
	var items []store.DataItemRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&store.DataItemRecord{})

	if params.RunID != "" {
		query = query.Where("run_id = ?", params.RunID)
	}
	if params.Query != "" {
		query = query.Where("data::text ILIKE ?", "%"+params.Query+"%")
	}
	if len(params.Types) == 1 {
		query = query.Where("type = ?", params.Types[0])
	} else if len(params.Types) > 1 {
		query = query.Where("type IN ?", params.Types)
	}
	if params.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *params.CreatedAfter)
	}
	if params.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *params.CreatedBefore)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count search results: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Order(orderCreatedAtASC).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("search data items: %w", err)
	}

	return items, total, nil
}
