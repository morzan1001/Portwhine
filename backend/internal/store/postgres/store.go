package postgres

import (
	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

// NewStore creates a Store with all PostgreSQL GORM implementations.
func NewStore(db *gorm.DB) *store.Store {
	return &store.Store{
		Roles:               NewRoleRepository(db),
		Users:               NewUserRepository(db),
		Pipelines:           NewPipelineRepository(db),
		PipelineRuns:        NewPipelineRunRepository(db),
		WorkerImages:        NewWorkerImageRepository(db),
		DataItems:           NewDataItemRepository(db),
		APIKeys:             NewAPIKeyRepository(db),
		AuditLog:            NewAuditLogRepository(db),
		Teams:               NewTeamRepository(db),
		TeamMembers:         NewTeamMemberRepository(db),
		ResourcePermissions: NewResourcePermissionRepository(db),
	}
}
