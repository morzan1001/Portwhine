package operator

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/portwhine/portwhine/internal/pipeline"
	"github.com/portwhine/portwhine/internal/store"
)

// Scheduler manages cron-based pipeline execution schedules.
type Scheduler struct {
	cron   *cron.Cron
	store  *store.Store
	engine *pipeline.Engine
	logger *slog.Logger

	mu      sync.Mutex
	// maps pipeline ID to cron entry ID for removal/update
	entries map[string]cron.EntryID
}

// NewScheduler creates a new Scheduler.
func NewScheduler(s *store.Store, engine *pipeline.Engine, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()), // support 6-field cron (sec min hour dom mon dow)
		store:   s,
		engine:  engine,
		logger:  logger,
		entries: make(map[string]cron.EntryID),
	}
}

// Start loads all pipelines with schedules from the store and starts the cron scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	// Load all pipelines and register those with schedules.
	pipelines, _, err := s.store.Pipelines.List(ctx, 0, 10000)
	if err != nil {
		return err
	}

	for _, p := range pipelines {
		if p.Schedule != "" && p.IsActive {
			s.addSchedule(p.ID, p.Schedule)
		}
	}

	s.cron.Start()
	s.logger.Info("scheduler started", slog.Int("scheduled_pipelines", len(s.entries)))
	return nil
}

// Stop stops the cron scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// UpdateSchedule updates (or removes) the schedule for a pipeline.
// Called when a pipeline is created or updated.
func (s *Scheduler) UpdateSchedule(pipelineID, schedule string, active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing entry if any.
	if entryID, ok := s.entries[pipelineID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, pipelineID)
	}

	// Add new entry if schedule is set and pipeline is active.
	if schedule != "" && active {
		s.addScheduleLocked(pipelineID, schedule)
	}
}

// addSchedule adds a cron schedule for a pipeline (acquires lock).
func (s *Scheduler) addSchedule(pipelineID, schedule string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addScheduleLocked(pipelineID, schedule)
}

// addScheduleLocked adds a cron schedule (caller must hold s.mu).
func (s *Scheduler) addScheduleLocked(pipelineID, schedule string) {
	entryID, err := s.cron.AddFunc(schedule, func() {
		s.triggerRun(pipelineID)
	})
	if err != nil {
		s.logger.Warn("failed to add cron schedule",
			slog.String("pipeline_id", pipelineID),
			slog.String("schedule", schedule),
			slog.Any("error", err),
		)
		return
	}
	s.entries[pipelineID] = entryID
	s.logger.Info("registered cron schedule",
		slog.String("pipeline_id", pipelineID),
		slog.String("schedule", schedule),
	)
}

// triggerRun creates and starts a pipeline run for the given pipeline.
func (s *Scheduler) triggerRun(pipelineID string) {
	ctx := context.Background()

	p, err := s.store.Pipelines.GetByID(ctx, pipelineID)
	if err != nil {
		s.logger.Error("scheduled run: pipeline not found",
			slog.String("pipeline_id", pipelineID),
			slog.Any("error", err),
		)
		return
	}

	if !p.IsActive {
		s.logger.Info("skipping scheduled run for inactive pipeline",
			slog.String("pipeline_id", pipelineID),
		)
		return
	}

	run := &store.PipelineRun{
		PipelineID:         p.ID,
		DefinitionSnapshot: p.Definition,
		Status:             "pending",
		CreatedByID:        p.CreatedByID, // attribute to pipeline creator
	}

	if err := s.store.PipelineRuns.Create(ctx, run); err != nil {
		s.logger.Error("scheduled run: failed to create run",
			slog.String("pipeline_id", pipelineID),
			slog.Any("error", err),
		)
		return
	}

	if err := s.engine.StartRun(ctx, run); err != nil {
		s.logger.Error("scheduled run: failed to start run",
			slog.String("pipeline_id", pipelineID),
			slog.String("run_id", run.ID),
			slog.Any("error", err),
		)
		if updateErr := s.store.PipelineRuns.UpdateStatus(ctx, run.ID, "failed"); updateErr != nil {
			s.logger.Error("scheduled run: failed to mark run as failed",
				slog.String("run_id", run.ID),
				slog.Any("error", updateErr),
			)
		}
		return
	}

	s.logger.Info("scheduled run started",
		slog.String("pipeline_id", pipelineID),
		slog.String("run_id", run.ID),
	)
}
