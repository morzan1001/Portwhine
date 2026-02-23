package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) store.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *store.Team) error {
	if err := r.db.WithContext(ctx).Create(team).Error; err != nil {
		return fmt.Errorf("create team: %w", err)
	}
	return nil
}

func (r *teamRepository) GetByID(ctx context.Context, id string) (*store.Team, error) {
	var team store.Team
	if err := r.db.WithContext(ctx).Preload("CreatedBy").First(&team, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get team by id: %w", err)
	}
	return &team, nil
}

func (r *teamRepository) GetByName(ctx context.Context, name string) (*store.Team, error) {
	var team store.Team
	if err := r.db.WithContext(ctx).First(&team, "name = ?", name).Error; err != nil {
		return nil, fmt.Errorf("get team by name: %w", err)
	}
	return &team, nil
}

func (r *teamRepository) List(ctx context.Context, offset, limit int) ([]store.Team, int64, error) {
	var teams []store.Team
	var total int64

	if err := r.db.WithContext(ctx).Model(&store.Team{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count teams: %w", err)
	}

	if err := r.db.WithContext(ctx).Preload("CreatedBy").Offset(offset).Limit(limit).Order("created_at DESC").Find(&teams).Error; err != nil {
		return nil, 0, fmt.Errorf("list teams: %w", err)
	}

	return teams, total, nil
}

func (r *teamRepository) ListByUser(ctx context.Context, userID string) ([]store.Team, error) {
	var teams []store.Team
	if err := r.db.WithContext(ctx).
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID).
		Preload("CreatedBy").
		Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("list teams by user: %w", err)
	}
	return teams, nil
}

func (r *teamRepository) Update(ctx context.Context, team *store.Team) error {
	if err := r.db.WithContext(ctx).Save(team).Error; err != nil {
		return fmt.Errorf("update team: %w", err)
	}
	return nil
}

func (r *teamRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ?", id).Delete(&store.TeamMember{}).Error; err != nil {
			return fmt.Errorf("delete team members: %w", err)
		}
		if err := tx.Where("subject_type = ? AND subject_id = ?", "team", id).Delete(&store.ResourcePermission{}).Error; err != nil {
			return fmt.Errorf("delete team permissions: %w", err)
		}
		if err := tx.Delete(&store.Team{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("delete team: %w", err)
		}
		return nil
	})
}

type teamMemberRepository struct {
	db *gorm.DB
}

func NewTeamMemberRepository(db *gorm.DB) store.TeamMemberRepository {
	return &teamMemberRepository{db: db}
}

func (r *teamMemberRepository) Add(ctx context.Context, member *store.TeamMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

func (r *teamMemberRepository) Remove(ctx context.Context, teamID, userID string) error {
	result := r.db.WithContext(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&store.TeamMember{})
	if result.Error != nil {
		return fmt.Errorf("remove team member: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("team member not found")
	}
	return nil
}

func (r *teamMemberRepository) GetMembership(ctx context.Context, teamID, userID string) (*store.TeamMember, error) {
	var member store.TeamMember
	if err := r.db.WithContext(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("get membership: %w", err)
	}
	return &member, nil
}

func (r *teamMemberRepository) ListByTeam(ctx context.Context, teamID string) ([]store.TeamMember, error) {
	var members []store.TeamMember
	if err := r.db.WithContext(ctx).Preload("User").Preload("User.Role").Where("team_id = ?", teamID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	return members, nil
}

func (r *teamMemberRepository) ListByUser(ctx context.Context, userID string) ([]store.TeamMember, error) {
	var members []store.TeamMember
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list user memberships: %w", err)
	}
	return members, nil
}

func (r *teamMemberRepository) UpdateRole(ctx context.Context, teamID, userID, role string) error {
	result := r.db.WithContext(ctx).Model(&store.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Update("role", role)
	if result.Error != nil {
		return fmt.Errorf("update team member role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("team member not found")
	}
	return nil
}
