package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"gorm.io/gorm"

	"github.com/portwhine/portwhine/internal/store"
)

// Authorizer combines Casbin type-level checks with instance-level ownership,
// team membership, and explicit resource permission checks.
type Authorizer struct {
	enforcer *casbin.Enforcer
	store    *store.Store
	logger   *slog.Logger
}

// NewAuthorizer creates a new Authorizer.
func NewAuthorizer(enforcer *casbin.Enforcer, s *store.Store, logger *slog.Logger) *Authorizer {
	return &Authorizer{
		enforcer: enforcer,
		store:    s,
		logger:   logger,
	}
}

// Enforcer returns the underlying Casbin enforcer for direct policy management.
func (a *Authorizer) Enforcer() *casbin.Enforcer {
	return a.enforcer
}

// CheckTypeAccess verifies that the user's role (or team membership) grants
// type-level access to the given resource and action. Called by the interceptor
// for every RPC.
func (a *Authorizer) CheckTypeAccess(ctx context.Context, userID, role, resource, action string) error {
	// Check via role name directly.
	allowed, err := a.enforcer.Enforce(role, resource, action)
	if err != nil {
		return fmt.Errorf("casbin enforce: %w", err)
	}
	if allowed {
		return nil
	}

	// Also check via user ID (covers team-based g2 mappings).
	allowed, err = a.enforcer.Enforce(userID, resource, action)
	if err != nil {
		return fmt.Errorf("casbin enforce user: %w", err)
	}
	if allowed {
		return nil
	}

	return errors.New("permission denied")
}

// CheckInstanceAccess verifies that the user can perform the given action on a
// specific resource instance. Called by handlers after type-level check passes.
//
// Resolution order:
//  1. Explicit deny for user → DENY
//  2. Explicit deny for any user team → DENY
//  3. Admin role → ALLOW
//  4. Ownership (CreatedByID == userID) → ALLOW
//  5. Explicit allow for any user team → ALLOW
//  6. Explicit allow for user → ALLOW
//  7. Default → DENY
func (a *Authorizer) CheckInstanceAccess(ctx context.Context, userID, role, resourceType, resourceID, action string) error {
	if resourceID == "" {
		return errors.New("permission denied: missing resource ID")
	}

	// 1. Check explicit user deny.
	denied, err := a.store.ResourcePermissions.CheckDeny(ctx, "user", userID, resourceType, resourceID, action)
	if err != nil {
		return fmt.Errorf("check user deny: %w", err)
	}
	if denied {
		return errors.New("permission denied: explicit deny")
	}

	// 2. Check explicit team deny.
	teamIDs, err := a.getUserTeamIDs(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user teams: %w", err)
	}
	for _, teamID := range teamIDs {
		denied, err := a.store.ResourcePermissions.CheckDeny(ctx, "team", teamID, resourceType, resourceID, action)
		if err != nil {
			return fmt.Errorf("check team deny: %w", err)
		}
		if denied {
			return errors.New("permission denied: team deny")
		}
	}

	// 3. Admin role bypasses instance-level checks.
	if role == "admin" {
		return nil
	}

	// 4. Check ownership.
	ownerID, err := a.getResourceOwner(ctx, resourceType, resourceID)
	if err != nil {
		a.logger.Warn("could not determine resource owner",
			"resource_type", resourceType,
			"resource_id", resourceID,
			"error", err,
		)
	} else if ownerID == userID {
		return nil
	}

	// 5. Check explicit team allow.
	for _, teamID := range teamIDs {
		_, err := a.store.ResourcePermissions.CheckPermission(ctx, "team", teamID, resourceType, resourceID, action)
		if err == nil {
			return nil
		}
	}

	// 6. Check explicit user allow.
	_, err = a.store.ResourcePermissions.CheckPermission(ctx, "user", userID, resourceType, resourceID, action)
	if err == nil {
		return nil
	}

	// 7. Default deny.
	return errors.New("permission denied")
}

// FilterAccessibleIDs returns the list of resource IDs the user can access for
// the given resource type. If the user is an admin, isUnrestricted is true and
// the caller should return all resources.
func (a *Authorizer) FilterAccessibleIDs(ctx context.Context, userID, role, resourceType string) (ids []string, isUnrestricted bool, err error) {
	// Admin sees everything.
	if role == "admin" {
		return nil, true, nil
	}

	// Collect IDs from ownership.
	ownedIDs, err := a.getOwnedResourceIDs(ctx, userID, resourceType)
	if err != nil {
		return nil, false, fmt.Errorf("get owned ids: %w", err)
	}

	idSet := make(map[string]struct{})
	for _, id := range ownedIDs {
		idSet[id] = struct{}{}
	}

	// Collect IDs from explicit user grants.
	userGrantIDs, err := a.store.ResourcePermissions.ListAccessibleResourceIDs(ctx, "user", userID, resourceType)
	if err != nil {
		return nil, false, fmt.Errorf("get user grant ids: %w", err)
	}
	for _, id := range userGrantIDs {
		idSet[id] = struct{}{}
	}

	// Collect IDs from team grants.
	teamIDs, err := a.getUserTeamIDs(ctx, userID)
	if err != nil {
		return nil, false, fmt.Errorf("get user teams: %w", err)
	}
	for _, teamID := range teamIDs {
		teamGrantIDs, err := a.store.ResourcePermissions.ListAccessibleResourceIDs(ctx, "team", teamID, resourceType)
		if err != nil {
			return nil, false, fmt.Errorf("get team grant ids: %w", err)
		}
		for _, id := range teamGrantIDs {
			idSet[id] = struct{}{}
		}
	}

	result := make([]string, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}

	return result, false, nil
}

// getUserTeamIDs returns all team IDs the user is a member of.
func (a *Authorizer) getUserTeamIDs(ctx context.Context, userID string) ([]string, error) {
	members, err := a.store.TeamMembers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.TeamID
	}
	return ids, nil
}

// getResourceOwner returns the CreatedByID for the given resource.
func (a *Authorizer) getResourceOwner(ctx context.Context, resourceType, resourceID string) (string, error) {
	switch resourceType {
	case "pipelines":
		p, err := a.store.Pipelines.GetByID(ctx, resourceID)
		if err != nil {
			return "", err
		}
		return p.CreatedByID, nil
	case "runs":
		r, err := a.store.PipelineRuns.GetByID(ctx, resourceID)
		if err != nil {
			return "", err
		}
		return r.CreatedByID, nil
	case "workers":
		w, err := a.store.WorkerImages.GetByID(ctx, resourceID)
		if err != nil {
			return "", err
		}
		if w.CreatedByID != nil {
			return *w.CreatedByID, nil
		}
		return "", errors.New("worker image has no owner")
	default:
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// getOwnedResourceIDs returns all resource IDs owned by the user.
func (a *Authorizer) getOwnedResourceIDs(ctx context.Context, userID, resourceType string) ([]string, error) {
	switch resourceType {
	case "pipelines":
		pipelines, _, err := a.store.Pipelines.ListByUser(ctx, userID, 0, 10000)
		if err != nil {
			return nil, err
		}
		ids := make([]string, len(pipelines))
		for i, p := range pipelines {
			ids[i] = p.ID
		}
		return ids, nil
	case "runs":
		// Runs are accessed via pipeline ownership — handled by pipeline check.
		return nil, nil
	case "workers":
		// Workers are generally shared; no per-user ownership query needed.
		return nil, nil
	default:
		return nil, nil
	}
}

// ResourceLookup provides a way to load a specific resource for ownership check
// without exposing the full store. Used by the authorizer internally.
type ResourceLookup struct {
	DB *gorm.DB
}
