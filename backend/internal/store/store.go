package store

import (
	"context"
	"time"
)

// Store aggregates all repository interfaces.
type Store struct {
	Roles               RoleRepository
	Users               UserRepository
	Pipelines           PipelineRepository
	PipelineRuns        PipelineRunRepository
	WorkerImages        WorkerImageRepository
	DataItems           DataItemRepository
	APIKeys             APIKeyRepository
	AuditLog            AuditLogRepository
	Teams               TeamRepository
	TeamMembers         TeamMemberRepository
	ResourcePermissions ResourcePermissionRepository
}

// RoleRepository handles role persistence.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
}

// UserRepository handles user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, offset, limit int) ([]User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// PipelineRepository handles pipeline definition persistence.
type PipelineRepository interface {
	Create(ctx context.Context, pipeline *Pipeline) error
	GetByID(ctx context.Context, id string) (*Pipeline, error)
	List(ctx context.Context, offset, limit int) ([]Pipeline, int64, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]Pipeline, int64, error)
	ListByIDs(ctx context.Context, ids []string, offset, limit int) ([]Pipeline, int64, error)
	Update(ctx context.Context, pipeline *Pipeline) error
	Delete(ctx context.Context, id string) error
	CountAll(ctx context.Context) (int64, error)
}

// PipelineRunRepository handles pipeline execution records.
type PipelineRunRepository interface {
	Create(ctx context.Context, run *PipelineRun) error
	GetByID(ctx context.Context, id string) (*PipelineRun, error)
	ListAll(ctx context.Context, offset, limit int) ([]PipelineRun, int64, error)
	ListByPipeline(ctx context.Context, pipelineID string, offset, limit int) ([]PipelineRun, int64, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]PipelineRun, int64, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Update(ctx context.Context, run *PipelineRun) error
	CountAll(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	ListRecent(ctx context.Context, limit int) ([]PipelineRun, error)

	// StepResult operations
	CreateStepResult(ctx context.Context, result *StepResult) error
	GetStepResult(ctx context.Context, id string) (*StepResult, error)
	ListStepResults(ctx context.Context, runID string) ([]StepResult, error)
	UpdateStepResult(ctx context.Context, result *StepResult) error
}

// WorkerImageRepository handles worker image registry.
type WorkerImageRepository interface {
	Create(ctx context.Context, image *WorkerImage) error
	GetByID(ctx context.Context, id string) (*WorkerImage, error)
	GetByName(ctx context.Context, name string) (*WorkerImage, error)
	List(ctx context.Context) ([]WorkerImage, error)
	Update(ctx context.Context, image *WorkerImage) error
	Delete(ctx context.Context, id string) error
}

// DataItemSearchParams holds the filter criteria for searching data items.
type DataItemSearchParams struct {
	RunID         string
	Query         string   // free-text search against JSONB data
	Types         []string // match any of these types
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// DataItemRepository handles persisted pipeline results.
type DataItemRepository interface {
	Create(ctx context.Context, item *DataItemRecord) error
	CreateBatch(ctx context.Context, items []DataItemRecord) error
	GetByID(ctx context.Context, id string) (*DataItemRecord, error)
	ListByRun(ctx context.Context, runID string, offset, limit int) ([]DataItemRecord, int64, error)
	ListByRunAndType(ctx context.Context, runID, itemType string, offset, limit int) ([]DataItemRecord, int64, error)
	Search(ctx context.Context, params DataItemSearchParams, offset, limit int) ([]DataItemRecord, int64, error)
}

// APIKeyRepository handles API key persistence.
type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByPrefix(ctx context.Context, prefix string) (*APIKey, error)
	ListByUser(ctx context.Context, userID string) ([]APIKey, error)
	UpdateLastUsed(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
}

// AuditLogRepository handles audit log persistence.
type AuditLogRepository interface {
	Create(ctx context.Context, entry *AuditLog) error
	List(ctx context.Context, offset, limit int) ([]AuditLog, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]AuditLog, error)
}

// TeamRepository handles team persistence.
type TeamRepository interface {
	Create(ctx context.Context, team *Team) error
	GetByID(ctx context.Context, id string) (*Team, error)
	GetByName(ctx context.Context, name string) (*Team, error)
	List(ctx context.Context, offset, limit int) ([]Team, int64, error)
	ListByUser(ctx context.Context, userID string) ([]Team, error)
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id string) error
}

// TeamMemberRepository handles team membership persistence.
type TeamMemberRepository interface {
	Add(ctx context.Context, member *TeamMember) error
	Remove(ctx context.Context, teamID, userID string) error
	GetMembership(ctx context.Context, teamID, userID string) (*TeamMember, error)
	ListByTeam(ctx context.Context, teamID string) ([]TeamMember, error)
	ListByUser(ctx context.Context, userID string) ([]TeamMember, error)
	UpdateRole(ctx context.Context, teamID, userID, role string) error
}

// ResourcePermissionRepository handles explicit resource-level permission grants/denies.
type ResourcePermissionRepository interface {
	Grant(ctx context.Context, perm *ResourcePermission) error
	Revoke(ctx context.Context, id string) error
	ListForSubject(ctx context.Context, subjectType, subjectID string) ([]ResourcePermission, error)
	ListForResource(ctx context.Context, resourceType, resourceID string) ([]ResourcePermission, error)
	CheckPermission(ctx context.Context, subjectType, subjectID, resourceType, resourceID, action string) (*ResourcePermission, error)
	CheckDeny(ctx context.Context, subjectType, subjectID, resourceType, resourceID, action string) (bool, error)
	ListAccessibleResourceIDs(ctx context.Context, subjectType, subjectID, resourceType string) ([]string, error)
}
