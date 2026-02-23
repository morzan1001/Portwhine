package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/portwhine/portwhine/internal/store"
)

// contextKey is an unexported type used for context value keys to avoid
// collisions with keys defined in other packages.
type contextKey int

const (
	claimsKey contextKey = iota
)

// ClaimsFromContext extracts the authenticated Claims from the context. It
// returns nil if no claims are present (e.g. for public RPCs).
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}

// withClaims returns a new context carrying the given Claims.
func withClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// APIKeyResolver is the interface for looking up API key credentials. The
// implementation is expected to find the key by prefix, compare hashes,
// and return the associated user.
type APIKeyResolver interface {
	Resolve(ctx context.Context, rawKey string) (*Claims, error)
}

// StoreAPIKeyResolver implements APIKeyResolver using the store repositories.
type StoreAPIKeyResolver struct {
	APIKeys     store.APIKeyRepository
	Users       store.UserRepository
	TeamMembers store.TeamMemberRepository
}

// Resolve validates a raw API key against the store. It looks up the key by
// its prefix, verifies the hash matches, checks expiry/revocation, and
// returns synthesized Claims for the owning user.
func (r *StoreAPIKeyResolver) Resolve(ctx context.Context, rawKey string) (*Claims, error) {
	if len(rawKey) < apiKeyPrefixLen {
		return nil, errors.New("invalid API key format")
	}

	prefix := rawKey[:apiKeyPrefixLen]
	apiKey, err := r.APIKeys.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, errors.New("invalid API key")
	}

	expectedHash := HashAPIKey(rawKey)
	if subtle.ConstantTimeCompare([]byte(apiKey.KeyHash), []byte(expectedHash)) != 1 {
		return nil, errors.New("invalid API key")
	}

	if apiKey.RevokedAt != nil {
		return nil, errors.New("API key has been revoked")
	}
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	user, err := r.Users.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, errors.New("user not found for API key")
	}

	// Load team IDs for the claims.
	var teamIDs []string
	if r.TeamMembers != nil {
		members, err := r.TeamMembers.ListByUser(ctx, user.ID)
		if err == nil {
			teamIDs = make([]string, len(members))
			for i, m := range members {
				teamIDs[i] = m.TeamID
			}
		}
	}

	return &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role.Name,
		TeamIDs:  teamIDs,
		Scopes:   apiKey.Scopes,
	}, nil
}

// publicProcedures lists RPC procedure names that do not require
// authentication.
var publicProcedures = map[string]bool{
	"/portwhine.v1.OperatorService/Login":       true,
	"/portwhine.v1.OperatorService/RefreshToken": true,
	"/portwhine.v1.OperatorService/CreateUser":   true,
}

// rpcPermission maps an RPC procedure to a resource/action pair for
// authorization.
type rpcPermission struct {
	resource string
	action   string
}

// procedurePermissions maps ConnectRPC procedure names to the resource and
// action required by the ABAC enforcer for type-level checks.
var procedurePermissions = map[string]rpcPermission{
	// Pipelines
	"/portwhine.v1.OperatorService/CreatePipeline": {"pipelines", "create"},
	"/portwhine.v1.OperatorService/GetPipeline":    {"pipelines", "read"},
	"/portwhine.v1.OperatorService/UpdatePipeline": {"pipelines", "update"},
	"/portwhine.v1.OperatorService/ListPipelines":  {"pipelines", "read"},
	"/portwhine.v1.OperatorService/DeletePipeline": {"pipelines", "delete"},
	"/portwhine.v1.OperatorService/StartPipeline":  {"pipelines", "execute"},

	// Runs
	"/portwhine.v1.OperatorService/StopPipelineRun":       {"runs", "execute"},
	"/portwhine.v1.OperatorService/PausePipelineRun":     {"runs", "execute"},
	"/portwhine.v1.OperatorService/ResumePipelineRun":    {"runs", "execute"},
	"/portwhine.v1.OperatorService/GetPipelineRunStatus":  {"runs", "read"},
	"/portwhine.v1.OperatorService/ListPipelineRuns":      {"runs", "read"},
	"/portwhine.v1.OperatorService/StreamPipelineResults": {"runs", "read"},

	// Data Items
	"/portwhine.v1.OperatorService/GetDataItem":      {"runs", "read"},
	"/portwhine.v1.OperatorService/ListDataItems":     {"runs", "read"},
	"/portwhine.v1.OperatorService/SearchDataItems":   {"runs", "read"},
	"/portwhine.v1.OperatorService/ExportDataItems":   {"runs", "read"},

	// Workers
	"/portwhine.v1.OperatorService/RegisterWorkerImage": {"workers", "create"},
	"/portwhine.v1.OperatorService/ListWorkerImages":    {"workers", "read"},
	"/portwhine.v1.OperatorService/DeleteWorkerImage":   {"workers", "delete"},

	// API Keys (treated as user-scoped; require user permissions)
	"/portwhine.v1.OperatorService/CreateAPIKey": {"users", "update"},
	"/portwhine.v1.OperatorService/ListAPIKeys":  {"users", "read"},
	"/portwhine.v1.OperatorService/RevokeAPIKey": {"users", "update"},

	// Users
	"/portwhine.v1.OperatorService/GetUser":    {"users", "read"},
	"/portwhine.v1.OperatorService/ListUsers":  {"users", "read"},
	"/portwhine.v1.OperatorService/UpdateUser": {"users", "update"},
	"/portwhine.v1.OperatorService/DeleteUser": {"users", "delete"},

	// Teams
	"/portwhine.v1.OperatorService/CreateTeam":          {"teams", "create"},
	"/portwhine.v1.OperatorService/GetTeam":             {"teams", "read"},
	"/portwhine.v1.OperatorService/ListTeams":           {"teams", "read"},
	"/portwhine.v1.OperatorService/UpdateTeam":          {"teams", "update"},
	"/portwhine.v1.OperatorService/DeleteTeam":          {"teams", "delete"},
	"/portwhine.v1.OperatorService/AddTeamMember":       {"teams", "update"},
	"/portwhine.v1.OperatorService/RemoveTeamMember":    {"teams", "update"},
	"/portwhine.v1.OperatorService/ListTeamMembers":     {"teams", "read"},
	"/portwhine.v1.OperatorService/UpdateTeamMemberRole": {"teams", "update"},

	// Permissions
	"/portwhine.v1.OperatorService/GrantPermission":  {"permissions", "create"},
	"/portwhine.v1.OperatorService/RevokePermission": {"permissions", "delete"},
	"/portwhine.v1.OperatorService/ListPermissions": {"permissions", "read"},
	// ListMyPermissions is intentionally unmapped — any authenticated user
	// can view their own permissions. The handler scopes results to the caller.

	// Roles
	"/portwhine.v1.OperatorService/CreateRole": {"users", "create"},
	"/portwhine.v1.OperatorService/ListRoles":  {"users", "read"},
	"/portwhine.v1.OperatorService/UpdateRole": {"users", "update"},
	"/portwhine.v1.OperatorService/DeleteRole": {"users", "delete"},
}

// InterceptorOptions holds the dependencies needed by the auth interceptor.
type InterceptorOptions struct {
	JWTService     *JWTService
	Authorizer     *Authorizer
	APIKeyResolver APIKeyResolver
}

// NewAuthInterceptor returns a ConnectRPC unary interceptor that handles
// authentication (JWT or API key) and ABAC type-level authorization.
func NewAuthInterceptor(opts InterceptorOptions) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure

			// Allow public RPCs without any authentication.
			if publicProcedures[procedure] {
				return next(ctx, req)
			}

			// Authenticate: try Bearer token first, then API key.
			claims, err := authenticate(ctx, req.Header(), opts)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			ctx = withClaims(ctx, claims)

			// Type-level authorize using Authorizer.
			if err := authorizeType(ctx, opts.Authorizer, claims, procedure); err != nil {
				return nil, err
			}

			return next(ctx, req)
		})
	})
}

// NewAuthStreamInterceptor returns a ConnectRPC interceptor that handles
// authentication and ABAC authorization for streaming RPCs.
func NewAuthStreamInterceptor(opts InterceptorOptions) *authStreamInterceptor {
	return &authStreamInterceptor{opts: opts}
}

// authStreamInterceptor implements connect.Interceptor for streaming RPCs.
type authStreamInterceptor struct {
	opts InterceptorOptions
}

// WrapUnary passes unary calls through unchanged (use NewAuthInterceptor
// for unary auth).
func (i *authStreamInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return next
}

// WrapStreamingClient passes client-side streaming calls through unchanged.
func (i *authStreamInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler applies authentication and authorization to
// server-side streaming RPCs.
func (i *authStreamInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		procedure := conn.Spec().Procedure

		// Allow public RPCs without authentication.
		if publicProcedures[procedure] {
			return next(ctx, conn)
		}

		// Authenticate: try Bearer token first, then API key.
		claims, err := authenticate(ctx, conn.RequestHeader(), i.opts)
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, err)
		}

		ctx = withClaims(ctx, claims)

		// Type-level authorize using Authorizer.
		if err := authorizeType(ctx, i.opts.Authorizer, claims, procedure); err != nil {
			return err
		}

		return next(ctx, conn)
	})
}

// NewAuthInterceptors returns both the unary and stream auth interceptors
// as a slice of connect.Interceptor, suitable for passing directly to
// connect.WithInterceptors.
func NewAuthInterceptors(opts InterceptorOptions) []connect.Interceptor {
	return []connect.Interceptor{
		NewAuthInterceptor(opts),
		NewAuthStreamInterceptor(opts),
	}
}

// authorizeType checks the ABAC type-level policy for the given claims and
// procedure. It returns a connect error if access is denied. For API key
// authentication, it also enforces the key's scopes.
func authorizeType(ctx context.Context, authorizer *Authorizer, claims *Claims, procedure string) error {
	perm, ok := procedurePermissions[procedure]
	if !ok {
		// If no permission mapping exists, allow (it may be a public or
		// unmapped RPC).
		return nil
	}

	// Enforce API key scopes if present. Scopes use "resource:action" format
	// (e.g. "pipelines:read", "pipelines:*", "*"). An empty scopes list
	// means unrestricted (backwards compatible).
	if len(claims.Scopes) > 0 {
		if !scopeAllows(claims.Scopes, perm.resource, perm.action) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("API key scope insufficient"))
		}
	}

	if err := authorizer.CheckTypeAccess(ctx, claims.UserID, claims.Role, perm.resource, perm.action); err != nil {
		return connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
	}

	return nil
}

// scopeAllows checks whether any of the given scopes permit the requested
// resource and action. Scopes follow the format "resource:action". The
// wildcard "*" matches everything. "resource:*" matches all actions on a
// resource.
func scopeAllows(scopes []string, resource, action string) bool {
	required := resource + ":" + action
	resourceWild := resource + ":*"
	for _, s := range scopes {
		if s == "*" || s == required || s == resourceWild {
			return true
		}
	}
	return false
}

// headerGetter is a minimal interface for reading request headers. Both
// http.Header and connect request types satisfy this through their Get method.
type headerGetter interface {
	Get(string) string
}

// authenticate extracts credentials from the request headers and validates
// them. It supports Bearer JWT tokens (via the Authorization header) and
// API keys (via the X-Api-Key header).
func authenticate(ctx context.Context, headers headerGetter, opts InterceptorOptions) (*Claims, error) {
	// Try Authorization: Bearer <token> first.
	if authHeader := headers.Get("Authorization"); authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			tokenStr := strings.TrimPrefix(authHeader, bearerPrefix)
			claims, err := opts.JWTService.ValidateAccessToken(tokenStr)
			if err != nil {
				return nil, errors.New("invalid bearer token")
			}
			return claims, nil
		}
	}

	// Try X-Api-Key header.
	if apiKey := headers.Get("X-Api-Key"); apiKey != "" {
		if opts.APIKeyResolver == nil {
			return nil, errors.New("API key authentication not configured")
		}
		claims, err := opts.APIKeyResolver.Resolve(ctx, apiKey)
		if err != nil {
			return nil, err
		}
		return claims, nil
	}

	return nil, errors.New("missing authentication credentials")
}
