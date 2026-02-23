package postgres

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

// AutoMigrate creates or updates all database tables based on GORM models.
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&store.Role{},
		&store.User{},
		&store.APIKey{},
		&store.WorkerImage{},
		&store.Pipeline{},
		&store.PipelineRun{},
		&store.StepResult{},
		&store.DataItemRecord{},
		&store.AuditLog{},
		&store.Team{},
		&store.TeamMember{},
		&store.ResourcePermission{},
	)
	if err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}

// Seed inserts default data (roles) if they don't already exist.
func Seed(db *gorm.DB) error {
	roles := []store.Role{
		{Name: "admin", Description: "Full system access including user management", IsSystem: true},
		{Name: "user", Description: "Can create and manage own pipelines and runs", IsSystem: true},
		{Name: "viewer", Description: "Read-only access to pipelines and run results", IsSystem: true},
	}

	for _, role := range roles {
		result := db.Where("name = ?", role.Name).FirstOrCreate(&role)
		if result.Error != nil {
			return fmt.Errorf("seed role %s: %w", role.Name, result.Error)
		}
	}

	return nil
}
