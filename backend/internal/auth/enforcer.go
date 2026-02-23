package auth

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// casbinModel defines the ABAC model with deny-override, role hierarchy, and
// team groupings. The 4-field policy (sub, obj, act, eft) supports explicit
// allow/deny effects. g defines user→role hierarchy, g2 defines user→team
// groupings for team-based type-level access.
const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (g(r.sub, p.sub) || g2(r.sub, p.sub)) && r.obj == p.obj && (r.act == p.act || p.act == "*")
`

// NewEnforcer creates a Casbin enforcer backed by the provided GORM database.
// It loads the ABAC model from an inline string and seeds the default policies
// if they are not already present.
func NewEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		return nil, fmt.Errorf("loading casbin model: %w", err)
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("creating casbin gorm adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("creating casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("loading casbin policies: %w", err)
	}

	if err := SeedPolicies(enforcer); err != nil {
		return nil, fmt.Errorf("seeding default policies: %w", err)
	}

	return enforcer, nil
}

// SeedPolicies adds the default ABAC policies for admin, user, and viewer
// roles, including role hierarchy. It is idempotent; existing policies will
// not be duplicated.
func SeedPolicies(enforcer *casbin.Enforcer) error {
	allResources := []string{"pipelines", "runs", "workers", "users", "teams", "permissions"}

	// admin: full access to everything.
	for _, res := range allResources {
		if _, err := enforcer.AddPolicy("admin", res, "*", "allow"); err != nil {
			return fmt.Errorf("adding admin policy (%s): %w", res, err)
		}
	}

	// user: CRUD + execute on pipelines and runs, read-only on workers, read teams.
	userPolicies := []struct{ resource, action string }{
		{"pipelines", "create"},
		{"pipelines", "read"},
		{"pipelines", "update"},
		{"pipelines", "delete"},
		{"pipelines", "execute"},
		{"runs", "create"},
		{"runs", "read"},
		{"runs", "update"},
		{"runs", "delete"},
		{"runs", "execute"},
		{"workers", "read"},
		{"teams", "read"},
		{"teams", "create"},
	}
	for _, p := range userPolicies {
		if _, err := enforcer.AddPolicy("user", p.resource, p.action, "allow"); err != nil {
			return fmt.Errorf("adding user policy (%s, %s): %w", p.resource, p.action, err)
		}
	}

	// viewer: read-only on pipelines, runs, workers.
	viewerResources := []string{"pipelines", "runs", "workers"}
	for _, res := range viewerResources {
		if _, err := enforcer.AddPolicy("viewer", res, "read", "allow"); err != nil {
			return fmt.Errorf("adding viewer policy (%s, read): %w", res, err)
		}
	}

	// Role hierarchy: admin inherits user, user inherits viewer.
	if _, err := enforcer.AddGroupingPolicy("admin", "user"); err != nil {
		return fmt.Errorf("adding admin→user hierarchy: %w", err)
	}
	if _, err := enforcer.AddGroupingPolicy("user", "viewer"); err != nil {
		return fmt.Errorf("adding user→viewer hierarchy: %w", err)
	}

	return nil
}

// AddUserRoleMapping adds a user→role mapping in Casbin (g grouping).
func AddUserRoleMapping(enforcer *casbin.Enforcer, userID, role string) error {
	if _, err := enforcer.AddGroupingPolicy(userID, role); err != nil {
		return fmt.Errorf("add user role mapping: %w", err)
	}
	return nil
}

// RemoveUserRoleMapping removes a user→role mapping in Casbin (g grouping).
func RemoveUserRoleMapping(enforcer *casbin.Enforcer, userID, role string) error {
	if _, err := enforcer.RemoveGroupingPolicy(userID, role); err != nil {
		return fmt.Errorf("remove user role mapping: %w", err)
	}
	return nil
}

// AddUserTeamMapping adds a user→team mapping in Casbin (g2 grouping).
func AddUserTeamMapping(enforcer *casbin.Enforcer, userID, teamID string) error {
	if _, err := enforcer.AddNamedGroupingPolicy("g2", userID, teamID); err != nil {
		return fmt.Errorf("add user team mapping: %w", err)
	}
	return nil
}

// RemoveUserTeamMapping removes a user→team mapping in Casbin (g2 grouping).
func RemoveUserTeamMapping(enforcer *casbin.Enforcer, userID, teamID string) error {
	if _, err := enforcer.RemoveNamedGroupingPolicy("g2", userID, teamID); err != nil {
		return fmt.Errorf("remove user team mapping: %w", err)
	}
	return nil
}

// AddTeamTypePolicy grants a team type-level access via Casbin policy.
func AddTeamTypePolicy(enforcer *casbin.Enforcer, teamID, resource, action, effect string) error {
	if _, err := enforcer.AddPolicy(teamID, resource, action, effect); err != nil {
		return fmt.Errorf("add team type policy: %w", err)
	}
	return nil
}
