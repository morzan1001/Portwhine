package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type workerImageRepository struct {
	db *gorm.DB
}

func NewWorkerImageRepository(db *gorm.DB) store.WorkerImageRepository {
	return &workerImageRepository{db: db}
}

func (r *workerImageRepository) Create(ctx context.Context, image *store.WorkerImage) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return fmt.Errorf("create worker image: %w", err)
	}
	return nil
}

func (r *workerImageRepository) GetByID(ctx context.Context, id string) (*store.WorkerImage, error) {
	var image store.WorkerImage
	if err := r.db.WithContext(ctx).First(&image, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get worker image by id: %w", err)
	}
	return &image, nil
}

func (r *workerImageRepository) GetByName(ctx context.Context, name string) (*store.WorkerImage, error) {
	var image store.WorkerImage
	if err := r.db.WithContext(ctx).First(&image, "name = ?", name).Error; err != nil {
		return nil, fmt.Errorf("get worker image by name: %w", err)
	}
	return &image, nil
}

func (r *workerImageRepository) List(ctx context.Context) ([]store.WorkerImage, error) {
	var images []store.WorkerImage
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name ASC").Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list worker images: %w", err)
	}
	return images, nil
}

func (r *workerImageRepository) Update(ctx context.Context, image *store.WorkerImage) error {
	if err := r.db.WithContext(ctx).Save(image).Error; err != nil {
		return fmt.Errorf("update worker image: %w", err)
	}
	return nil
}

func (r *workerImageRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Model(&store.WorkerImage{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("delete worker image: %w", err)
	}
	return nil
}
