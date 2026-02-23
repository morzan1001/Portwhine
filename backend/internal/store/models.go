package store

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// Role represents a user role (admin, user, viewer, or custom).
type Role struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	IsSystem    bool      `gorm:"default:false;not null"`
	IsCustom    bool      `gorm:"default:false;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

// User represents an authenticated user of the system.
type User struct {
	ID               string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username         string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email            string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash     string         `gorm:"type:varchar(255);not null"`
	RoleID           string         `gorm:"type:uuid;not null"`
	Role             Role           `gorm:"foreignKey:RoleID"`
	IsActive         bool           `gorm:"default:true;not null"`
	CustomAttributes datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
}

// APIKey represents a programmatic access key for a user.
type APIKey struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string         `gorm:"type:uuid;not null;index"`
	User      User           `gorm:"foreignKey:UserID"`
	Name      string         `gorm:"type:varchar(100);not null"`
	KeyHash   string         `gorm:"type:varchar(255);not null"`
	KeyPrefix string         `gorm:"type:varchar(8);not null;index"`
	Scopes    pq.StringArray `gorm:"type:text[]"`
	ExpiresAt *time.Time
	LastUsed  *time.Time
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	RevokedAt *time.Time
}

// WorkerImage represents a registered worker container image.
type WorkerImage struct {
	ID                 string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name               string    `gorm:"type:varchar(200);uniqueIndex;not null"`
	Image              string    `gorm:"type:varchar(500);not null"`
	Description        string    `gorm:"type:text"`
	InterfaceVersion   string    `gorm:"type:varchar(20);default:'v1';not null"`
	DefaultCPULimit    string    `gorm:"type:varchar(20);default:'500m'"`
	DefaultMemoryLimit string    `gorm:"type:varchar(20);default:'256Mi'"`
	IsActive           bool      `gorm:"default:true;not null"`
	CreatedByID        *string   `gorm:"type:uuid"`
	CreatedBy          *User     `gorm:"foreignKey:CreatedByID"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// Pipeline represents a pipeline definition stored as a graph.
type Pipeline struct {
	ID          string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string         `gorm:"type:varchar(200);not null"`
	Description string         `gorm:"type:text"`
	Definition  datatypes.JSON `gorm:"type:jsonb;not null"`
	Labels      datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	Schedule    string         `gorm:"type:varchar(100)"`
	Version     int            `gorm:"default:1;not null"`
	IsActive    bool           `gorm:"default:true;not null"`
	CreatedByID string         `gorm:"type:uuid;not null;index"`
	CreatedBy   User           `gorm:"foreignKey:CreatedByID"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
}

// PipelineRun represents a single execution of a pipeline.
type PipelineRun struct {
	ID                 string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PipelineID         string         `gorm:"type:uuid;not null;index"`
	Pipeline           Pipeline       `gorm:"foreignKey:PipelineID"`
	DefinitionSnapshot datatypes.JSON `gorm:"type:jsonb;not null"`
	Status             string         `gorm:"type:varchar(20);default:'pending';not null;index"`
	StartedAt          *time.Time
	FinishedAt         *time.Time
	CreatedByID        string    `gorm:"type:uuid;not null"`
	CreatedBy          User      `gorm:"foreignKey:CreatedByID"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	ErrorMessage       string    `gorm:"type:text"`
}

// StepResult represents the execution result of a single pipeline node.
type StepResult struct {
	ID            string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RunID         string         `gorm:"type:uuid;not null;index"`
	Run           PipelineRun    `gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE"`
	NodeID        string         `gorm:"type:varchar(200);not null"`
	ContainerID   string         `gorm:"type:varchar(200)"`
	WorkerImageID *string        `gorm:"type:uuid"`
	WorkerImage   *WorkerImage   `gorm:"foreignKey:WorkerImageID"`
	Status        string         `gorm:"type:varchar(20);default:'pending';not null;index"`
	ExitCode      *int
	Output        datatypes.JSON `gorm:"type:jsonb"`
	StartedAt     *time.Time
	FinishedAt    *time.Time
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	ErrorMessage  string    `gorm:"type:text"`
}

// DataItemRecord persists DataItems produced during pipeline runs.
type DataItemRecord struct {
	ID         string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RunID      string         `gorm:"type:uuid;not null;index"`
	Run        PipelineRun    `gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE"`
	Type       string         `gorm:"type:varchar(100);not null;index"`
	Data       datatypes.JSON `gorm:"type:jsonb"`
	Metadata   datatypes.JSON `gorm:"type:jsonb"`
	RawPayload []byte         `gorm:"type:bytea"`
	ParentIDs  pq.StringArray `gorm:"type:text[]"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
}

// AuditLog records user actions for auditing purposes.
type AuditLog struct {
	ID         uint           `gorm:"primaryKey;autoIncrement"`
	UserID     *string        `gorm:"type:uuid;index"`
	User       *User          `gorm:"foreignKey:UserID"`
	Action     string         `gorm:"type:varchar(100);not null"`
	Resource   string         `gorm:"type:varchar(100);not null"`
	ResourceID *string        `gorm:"type:uuid"`
	Details    datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;index"`
}

// Team represents a group of users that can share resources.
type Team struct {
	ID          string       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string       `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string       `gorm:"type:text"`
	CreatedByID string       `gorm:"type:uuid;not null"`
	CreatedBy   User         `gorm:"foreignKey:CreatedByID"`
	Members     []TeamMember `gorm:"foreignKey:TeamID"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
}

// TeamMember represents a user's membership in a team with a team-level role.
type TeamMember struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_team_user"`
	Team      Team      `gorm:"foreignKey:TeamID"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_team_user"`
	User      User      `gorm:"foreignKey:UserID"`
	Role      string    `gorm:"type:varchar(20);not null;default:'member'"` // owner, admin, member
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// ResourcePermission represents an explicit per-resource grant or deny.
type ResourcePermission struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubjectType  string    `gorm:"type:varchar(10);not null;index"` // "user" or "team"
	SubjectID    string    `gorm:"type:uuid;not null;index"`
	ResourceType string    `gorm:"type:varchar(50);not null;index"` // "pipelines", "runs", "workers"
	ResourceID   string    `gorm:"type:uuid;not null;index"`
	Action       string    `gorm:"type:varchar(20);not null"` // "read", "update", "delete", "execute", "*"
	Effect       string    `gorm:"type:varchar(10);not null;default:'allow'"` // "allow" or "deny"
	Condition    string    `gorm:"type:text"`
	GrantedByID  string    `gorm:"type:uuid;not null"`
	GrantedBy    User      `gorm:"foreignKey:GrantedByID"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
