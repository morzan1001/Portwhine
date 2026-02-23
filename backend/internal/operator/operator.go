package operator

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/portwhine/portwhine/gen/go/portwhine/v1/portwhinev1connect"
	"github.com/portwhine/portwhine/internal/auth"
	"github.com/portwhine/portwhine/internal/pipeline"
	"github.com/portwhine/portwhine/internal/runtime"
	"github.com/portwhine/portwhine/internal/store"
	storepostgres "github.com/portwhine/portwhine/internal/store/postgres"
	"github.com/portwhine/portwhine/internal/trigger"
	"github.com/portwhine/portwhine/internal/worker"
)

// Operator is the main application struct that wires all components together.
type Operator struct {
	cfg       *Config
	logger    *slog.Logger
	db        *gorm.DB
	store     *store.Store
	handler   *Handler
	engine    *pipeline.Engine
	scheduler *Scheduler
	rt        runtime.Runtime
	mux       *http.ServeMux
}

// New creates a new Operator instance with all dependencies wired.
func New(cfg *Config, logger *slog.Logger) (*Operator, error) {
	// Initialize database
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Run auto-migrations
	if err := storepostgres.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto-migrate database: %w", err)
	}

	// Seed default data
	if err := storepostgres.Seed(db); err != nil {
		return nil, fmt.Errorf("seed database: %w", err)
	}

	// Initialize store
	s := storepostgres.NewStore(db)

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		[]byte(cfg.Auth.JWTSecret),
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)

	// Initialize Casbin enforcer
	enforcer, err := auth.NewEnforcer(db)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	// Initialize Authorizer (combines Casbin type-level + instance-level checks)
	authorizer := auth.NewAuthorizer(enforcer, s, logger)

	// Initialize API key resolver
	apiKeyResolver := &auth.StoreAPIKeyResolver{
		APIKeys:     s.APIKeys,
		Users:       s.Users,
		TeamMembers: s.TeamMembers,
	}

	// Initialize auth interceptors
	authInterceptors := auth.NewAuthInterceptors(auth.InterceptorOptions{
		JWTService:     jwtService,
		Authorizer:     authorizer,
		APIKeyResolver: apiKeyResolver,
	})

	// Initialize container runtime
	rt, err := runtime.New(runtime.Config{
		Type: runtime.RuntimeType(cfg.Runtime.Type),
		Docker: runtime.DockerConfig{
			NetworkName: cfg.Runtime.Docker.Network,
		},
		Kubernetes: runtime.KubernetesConfig{
			Namespace:       cfg.Runtime.Kubernetes.Namespace,
			WorkerNamespace: cfg.Runtime.Kubernetes.WorkerNamespace,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create runtime: %w", err)
	}

	// Initialize Prometheus metrics
	metrics := NewMetrics()

	// Initialize pipeline engine with worker/trigger client factories
	engine := pipeline.NewEngine(rt, s, logger)
	engine.SetOperatorAddress(cfg.Server.OperatorAddress)
	engine.SetWorkerClientFactory(worker.NewWorkerClientFactory(logger))
	engine.SetTriggerClientFactory(trigger.NewTriggerClientFactory(logger))
	engine.SetMetrics(metrics)

	// Initialize cron scheduler for pipeline schedules.
	scheduler := NewScheduler(s, engine, logger)

	// Initialize handler
	handler := NewHandler(s, jwtService, engine, authorizer, scheduler, logger)

	// Create HTTP mux and register ConnectRPC handler with auth interceptors
	mux := http.NewServeMux()
	path, svcHandler := portwhinev1connect.NewOperatorServiceHandler(
		handler,
		connect.WithInterceptors(authInterceptors...),
	)
	mux.Handle(path, svcHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Start engine background tasks (orphaned container GC, stale run recovery).
	engine.StartBackgroundTasks(context.Background())

	return &Operator{
		cfg:       cfg,
		logger:    logger,
		db:        db,
		store:     s,
		handler:   handler,
		engine:    engine,
		scheduler: scheduler,
		rt:        rt,
		mux:       mux,
	}, nil
}

// corsHandler wraps an HTTP handler with CORS support for web browsers
func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Run starts the operator and blocks until shutdown.
func (o *Operator) Run() error {
	// Create HTTP/2 server (h2c for plaintext gRPC support)
	srv := &http.Server{
		Addr:              o.cfg.Server.GRPCAddr,
		Handler:           corsHandler(h2c.NewHandler(o.mux, &http2.Server{})),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		o.logger.Info("ConnectRPC server listening",
			"addr", o.cfg.Server.GRPCAddr,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http serve: %w", err)
		}
	}()

	// Start cron scheduler for scheduled pipelines.
	if err := o.scheduler.Start(context.Background()); err != nil {
		o.logger.Warn("failed to start scheduler", slog.Any("error", err))
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		o.logger.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		return err
	}

	// Graceful shutdown
	o.logger.Info("shutting down operator")

	o.engine.StopBackgroundTasks()
	o.scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	return nil
}
