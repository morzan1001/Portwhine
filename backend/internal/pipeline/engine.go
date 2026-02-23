package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	"github.com/portwhine/portwhine/internal/runtime"
	"github.com/portwhine/portwhine/internal/store"
	"gorm.io/datatypes"
)

// HeartbeatFunc is called when a worker/trigger sends a heartbeat update.
// It reports the worker's self-reported metrics.
type HeartbeatFunc func(itemsProcessed, itemsProduced, errorsCount uint64)

// WorkerClient defines the interface for communicating with a worker container
// via gRPC.
type WorkerClient interface {
	Initialize(ctx context.Context, config *portwhinev1.StageConfig) error
	Process(ctx context.Context, input <-chan *portwhinev1.DataItem, output chan<- *portwhinev1.DataItem, onHeartbeat HeartbeatFunc) error
	Shutdown(ctx context.Context) error
}

// TriggerClient defines the interface for communicating with a trigger container
// via gRPC.
type TriggerClient interface {
	Initialize(ctx context.Context, config *portwhinev1.StageConfig) error
	Start(ctx context.Context, output chan<- *portwhinev1.DataItem, onHeartbeat HeartbeatFunc) error
	Stop(ctx context.Context) error
}

// WorkerClientFactory creates a WorkerClient for a given operator address.
type WorkerClientFactory func(address string) (WorkerClient, error)

// TriggerClientFactory creates a TriggerClient for a given operator address.
type TriggerClientFactory func(address string) (TriggerClient, error)

// EngineMetrics is an optional interface for recording pipeline execution
// metrics (e.g. Prometheus counters/histograms). The pipeline package does not
// depend on any metrics library directly; the operator provides the concrete
// implementation.
type EngineMetrics interface {
	RunStarted(pipelineID string)
	RunFinished(pipelineID, status string)
	DataItemPersisted(itemType string)
	StageFinished(nodeID, status string, duration time.Duration)
}

// StageRunner manages the lifecycle of a single pipeline stage: its container,
// gRPC client connection, and the goroutine processing data items.
type StageRunner struct {
	NodeID      string
	Node        *portwhinev1.PipelineNode
	ContainerID runtime.ContainerID
	Status      string // "pending", "running", "completed", "failed"
	Error       error
	StartedAt   *time.Time
	FinishedAt  *time.Time

	// Live metrics updated atomically during execution.
	ItemsIn    atomic.Uint64
	ItemsOut   atomic.Uint64
	ErrorCount atomic.Uint64
	LastError  string

	cancel context.CancelFunc
}

// NodeStatusInfo holds a snapshot of a node's live status, combining
// pipeline-level metrics with container-level health from Docker/Kubernetes.
type NodeStatusInfo struct {
	NodeID     string
	WorkerType string
	Status     string // pipeline stage status: pending, running, completed, failed
	ItemsIn    uint64
	ItemsOut   uint64
	Errors     uint64
	Error      string

	// Container-level info from Docker/Kubernetes runtime.
	ContainerID         string
	ContainerStatus     runtime.ContainerStatus // running, succeeded, failed, pending, unknown
	ContainerExitCode   int
	ContainerStartedAt  *time.Time
	ContainerFinishedAt *time.Time
	ContainerMessage    string
}

// runState holds the mutable state for a single pipeline execution.
type runState struct {
	mu           sync.Mutex
	run          *store.PipelineRun
	graph        *PipelineGraph
	router       *Router
	stages       map[string]*StageRunner
	cancel       context.CancelFunc
	done         chan struct{} // closed when the run completes

	// Pause/resume support.
	pauseCh  chan struct{} // nil = running, non-nil open channel = paused
	pauseMu  sync.Mutex

	// Pub/sub for live result streaming.
	subscribersMu sync.Mutex
	subscribers   []chan *portwhinev1.DataItem
}

// publish sends a DataItem to all active subscribers (non-blocking).
func (s *runState) publish(item *portwhinev1.DataItem) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- item:
		default:
			// Subscriber is slow; drop the item to avoid blocking the pipeline.
		}
	}
}

// waitIfPaused blocks until the run is resumed or the context is cancelled.
func (s *runState) waitIfPaused(ctx context.Context) error {
	s.pauseMu.Lock()
	ch := s.pauseCh
	s.pauseMu.Unlock()

	if ch == nil {
		return nil // not paused
	}

	select {
	case <-ch:
		return nil // resumed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Engine orchestrates pipeline runs. It manages container lifecycles, data
// routing between stages, and persistence of run status and results.
type Engine struct {
	runtime              runtime.Runtime
	store                *store.Store
	logger               *slog.Logger
	workerClientFactory  WorkerClientFactory
	triggerClientFactory TriggerClientFactory
	operatorAddress      string
	metrics              EngineMetrics

	mu   sync.Mutex
	runs map[string]*runState

	gcStop chan struct{} // closed to stop periodic GC
}

// NewEngine creates a new pipeline execution engine.
func NewEngine(rt runtime.Runtime, st *store.Store, logger *slog.Logger) *Engine {
	return &Engine{
		runtime: rt,
		store:   st,
		logger:  logger,
		runs:    make(map[string]*runState),
		gcStop:  make(chan struct{}),
	}
}

// StartBackgroundTasks runs startup container GC and starts periodic GC.
// Call this once after the engine is fully configured.
func (e *Engine) StartBackgroundTasks(ctx context.Context) {
	e.cleanupOrphanedContainers(ctx)
	e.recoverStaleRuns(ctx)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.cleanupOrphanedContainers(context.Background())
			case <-e.gcStop:
				return
			}
		}
	}()
}

// StopBackgroundTasks stops the periodic GC goroutine.
func (e *Engine) StopBackgroundTasks() {
	close(e.gcStop)
}

// cleanupOrphanedContainers finds and removes containers managed by
// Portwhine that are not associated with an active in-memory run.
func (e *Engine) cleanupOrphanedContainers(ctx context.Context) {
	containers, err := e.runtime.List(ctx, map[string]string{
		"portwhine.managed": "true",
	})
	if err != nil {
		e.logger.Warn("failed to list managed containers for GC", slog.Any("error", err))
		return
	}
	if len(containers) == 0 {
		return
	}

	e.mu.Lock()
	activeContainers := make(map[runtime.ContainerID]struct{})
	for _, state := range e.runs {
		state.mu.Lock()
		for _, stage := range state.stages {
			if stage.ContainerID != "" {
				activeContainers[stage.ContainerID] = struct{}{}
			}
		}
		state.mu.Unlock()
	}
	e.mu.Unlock()

	removed := 0
	for _, c := range containers {
		if _, active := activeContainers[c.ID]; active {
			continue
		}
		e.logger.Info("removing orphaned container",
			slog.String("container_id", string(c.ID)),
			slog.String("name", c.Name),
		)
		if err := e.runtime.Stop(ctx, c.ID, 10*time.Second); err != nil {
			e.logger.Debug("stop orphaned container (may already be stopped)",
				slog.String("container_id", string(c.ID)),
				slog.Any("error", err),
			)
		}
		if err := e.runtime.Remove(ctx, c.ID); err != nil {
			e.logger.Warn("failed to remove orphaned container",
				slog.String("container_id", string(c.ID)),
				slog.Any("error", err),
			)
		} else {
			removed++
		}
	}
	if removed > 0 {
		e.logger.Info("container GC completed", slog.Int("removed", removed))
	}
}

// recoverStaleRuns marks pipeline runs stuck in "running" or "pending" as
// "failed". These are runs from a previous operator instance.
func (e *Engine) recoverStaleRuns(ctx context.Context) {
	for _, status := range []string{"running", "pending"} {
		runs, _, err := e.store.PipelineRuns.ListByStatus(ctx, status, 0, 1000)
		if err != nil {
			e.logger.Warn("failed to list stale runs", slog.String("status", status), slog.Any("error", err))
			continue
		}
		for _, run := range runs {
			e.mu.Lock()
			_, active := e.runs[run.ID]
			e.mu.Unlock()
			if !active {
				e.logger.Info("recovering stale run",
					slog.String("run_id", run.ID),
					slog.String("old_status", run.Status),
				)
				if err := e.store.PipelineRuns.UpdateStatus(ctx, run.ID, "failed"); err != nil {
					e.logger.Warn("failed to mark stale run as failed",
						slog.String("run_id", run.ID),
						slog.Any("error", err),
					)
				}
			}
		}
	}
}

// SetWorkerClientFactory sets the factory used to create gRPC WorkerClients.
func (e *Engine) SetWorkerClientFactory(f WorkerClientFactory) {
	e.workerClientFactory = f
}

// SetTriggerClientFactory sets the factory used to create gRPC TriggerClients.
func (e *Engine) SetTriggerClientFactory(f TriggerClientFactory) {
	e.triggerClientFactory = f
}

// SetOperatorAddress sets the address that workers/triggers use to reach the
// operator's gRPC server.
func (e *Engine) SetOperatorAddress(addr string) {
	e.operatorAddress = addr
}

// SetMetrics sets an optional EngineMetrics implementation that receives
// pipeline execution events for monitoring (e.g. Prometheus).
func (e *Engine) SetMetrics(m EngineMetrics) {
	e.metrics = m
}

// StartRun begins executing a pipeline run. It parses the pipeline definition
// from the run's DefinitionSnapshot, validates it, creates step result records,
// starts containers for each node, and launches goroutines to route data.
func (e *Engine) StartRun(ctx context.Context, run *store.PipelineRun) error {
	e.logger.Info("starting pipeline run",
		slog.String("run_id", run.ID),
		slog.String("pipeline_id", run.PipelineID),
	)

	// Parse the pipeline definition from the snapshot.
	var def portwhinev1.PipelineDefinition
	if err := protojson.Unmarshal(run.DefinitionSnapshot, &def); err != nil {
		return e.failRun(ctx, run.ID, fmt.Errorf("failed to parse pipeline definition: %w", err))
	}

	// Build and validate the graph.
	graph, err := FromProto(&def)
	if err != nil {
		return e.failRun(ctx, run.ID, fmt.Errorf("invalid pipeline graph: %w", err))
	}

	// Update run status to "running".
	now := time.Now()
	run.Status = "running"
	run.StartedAt = &now
	if err := e.store.PipelineRuns.Update(ctx, run); err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	// Create a cancellable context for this run, independent of the request
	// context so the pipeline continues after the gRPC response is sent.
	runCtx, cancel := context.WithCancel(context.Background())

	// Build the router from the graph.
	router := NewRouter(graph)

	// Initialize run state.
	state := &runState{
		run:    run,
		graph:  graph,
		router: router,
		stages: make(map[string]*StageRunner),
		cancel: cancel,
		done:   make(chan struct{}),
	}

	e.mu.Lock()
	e.runs[run.ID] = state
	e.mu.Unlock()

	if e.metrics != nil {
		e.metrics.RunStarted(run.PipelineID)
	}

	// Create StepResult records for each node in topological order.
	sortedIDs := graph.TopologicalSort()
	for _, nodeID := range sortedIDs {
		stepResult := &store.StepResult{
			RunID:  run.ID,
			NodeID: nodeID,
			Status: "pending",
		}
		if err := e.store.PipelineRuns.CreateStepResult(ctx, stepResult); err != nil {
			cancel()
			return e.failRun(ctx, run.ID, fmt.Errorf("failed to create step result for node %q: %w", nodeID, err))
		}

		state.stages[nodeID] = &StageRunner{
			NodeID: nodeID,
			Node:   graph.GetNode(nodeID),
			Status: "pending",
		}
	}

	// Launch a coordination goroutine that starts all stages and waits for completion.
	go e.runPipeline(runCtx, state, sortedIDs)

	return nil
}

// StopRun cancels a running pipeline and stops all associated containers.
func (e *Engine) StopRun(ctx context.Context, runID string) error {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("run %q not found", runID)
	}

	e.logger.Info("stopping pipeline run", slog.String("run_id", runID))

	// Cancel the run context, which signals all stage goroutines to stop.
	state.cancel()

	// Stop all containers.
	state.mu.Lock()
	stages := make(map[string]*StageRunner, len(state.stages))
	for k, v := range state.stages {
		stages[k] = v
	}
	state.mu.Unlock()

	var errs []error
	for nodeID, stage := range stages {
		if stage.ContainerID != "" {
			if err := e.runtime.Stop(ctx, stage.ContainerID, 30*time.Second); err != nil {
				e.logger.Error("failed to stop container",
					slog.String("node_id", nodeID),
					slog.String("container_id", string(stage.ContainerID)),
					slog.Any("error", err),
				)
				errs = append(errs, err)
			}
		}
	}

	// Update run status.
	if err := e.store.PipelineRuns.UpdateStatus(ctx, runID, "cancelled"); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping run %q: %v", runID, errs)
	}
	return nil
}

// PauseRun pauses a running pipeline. New items in the output stage will block
// until the run is resumed.
func (e *Engine) PauseRun(ctx context.Context, runID string) error {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()
	if !ok {
		return fmt.Errorf("run %q not found", runID)
	}

	state.pauseMu.Lock()
	defer state.pauseMu.Unlock()

	if state.pauseCh != nil {
		return nil // already paused
	}
	state.pauseCh = make(chan struct{})

	e.logger.Info("pipeline run paused", slog.String("run_id", runID))

	if err := e.store.PipelineRuns.UpdateStatus(ctx, runID, "paused"); err != nil {
		e.logger.Warn("failed to update paused status", slog.Any("error", err))
	}

	return nil
}

// ResumeRun resumes a paused pipeline run.
func (e *Engine) ResumeRun(ctx context.Context, runID string) error {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()
	if !ok {
		return fmt.Errorf("run %q not found", runID)
	}

	state.pauseMu.Lock()
	defer state.pauseMu.Unlock()

	if state.pauseCh == nil {
		return nil // not paused
	}
	close(state.pauseCh)
	state.pauseCh = nil

	e.logger.Info("pipeline run resumed", slog.String("run_id", runID))

	if err := e.store.PipelineRuns.UpdateStatus(ctx, runID, "running"); err != nil {
		e.logger.Warn("failed to update running status", slog.Any("error", err))
	}

	return nil
}

// GetRunStatus returns a summary of the current status of all stages in a run.
func (e *Engine) GetRunStatus(ctx context.Context, runID string) (map[string]string, error) {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()

	if !ok {
		// Fall back to database.
		steps, err := e.store.PipelineRuns.ListStepResults(ctx, runID)
		if err != nil {
			return nil, fmt.Errorf("failed to list step results: %w", err)
		}
		result := make(map[string]string, len(steps))
		for _, step := range steps {
			result[step.NodeID] = step.Status
		}
		return result, nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	result := make(map[string]string, len(state.stages))
	for nodeID, stage := range state.stages {
		result[nodeID] = stage.Status
	}
	return result, nil
}

// runPipeline is the main coordination goroutine for a pipeline run.
// It starts all stage goroutines and waits for them to complete.
func (e *Engine) runPipeline(ctx context.Context, state *runState, sortedIDs []string) {
	defer close(state.done)

	var wg sync.WaitGroup
	stageErrors := make(chan error, len(sortedIDs))

	for _, nodeID := range sortedIDs {
		node := state.graph.GetNode(nodeID)
		if node == nil {
			continue
		}

		wg.Add(1)
		go func(nodeID string, node *portwhinev1.PipelineNode) {
			defer wg.Done()

			var stageFunc func() error
			switch node.GetType() {
			case portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_TRIGGER:
				stageFunc = func() error { return e.runTriggerStage(ctx, state, nodeID, node) }
			case portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_WORKER:
				stageFunc = func() error { return e.runWorkerStage(ctx, state, nodeID, node) }
			case portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_OUTPUT:
				stageFunc = func() error { return e.runOutputStage(ctx, state, nodeID, node) }
			default:
				stageFunc = func() error {
					return fmt.Errorf("unknown node type %v for node %q", node.GetType(), nodeID)
				}
			}

			err := e.runWithRetry(ctx, nodeID, node, stageFunc)

			if err != nil {
				e.logger.Error("stage failed",
					slog.String("node_id", nodeID),
					slog.Any("error", err),
				)
				stageErrors <- fmt.Errorf("node %q: %w", nodeID, err)
			}
		}(nodeID, node)
	}

	// Wait for all stages to finish.
	wg.Wait()
	close(stageErrors)

	// Collect errors.
	var runErr error
	for err := range stageErrors {
		if runErr == nil {
			runErr = err
		} else {
			runErr = fmt.Errorf("%v; %w", runErr, err)
		}
	}

	// Update final run status.
	finalStatus := "completed"
	errMsg := ""
	if runErr != nil {
		finalStatus = "failed"
		errMsg = runErr.Error()
	}

	// Check if the run was cancelled.
	if ctx.Err() != nil {
		finalStatus = "cancelled"
	}

	if e.metrics != nil {
		e.metrics.RunFinished(state.run.PipelineID, finalStatus)
	}

	now := time.Now()
	state.mu.Lock()
	state.run.Status = finalStatus
	state.run.FinishedAt = &now
	state.run.ErrorMessage = errMsg
	state.mu.Unlock()

	updateCtx := context.Background()
	if err := e.store.PipelineRuns.UpdateStatus(updateCtx, state.run.ID, finalStatus); err != nil {
		e.logger.Error("failed to update final run status",
			slog.String("run_id", state.run.ID),
			slog.Any("error", err),
		)
	}

	e.logger.Info("pipeline run finished",
		slog.String("run_id", state.run.ID),
		slog.String("status", finalStatus),
	)

	// Clean up any remaining containers (covers failure/cancel paths where
	// individual stage cleanup may not have run).
	e.cleanupRunContainers(state)

	// Clean up run state.
	e.mu.Lock()
	delete(e.runs, state.run.ID)
	e.mu.Unlock()
}

// cleanupRunContainers stops and removes all containers for a run.
func (e *Engine) cleanupRunContainers(state *runState) {
	state.mu.Lock()
	stages := make(map[string]runtime.ContainerID, len(state.stages))
	for nodeID, stage := range state.stages {
		if stage.ContainerID != "" {
			stages[nodeID] = stage.ContainerID
		}
	}
	state.mu.Unlock()

	ctx := context.Background()
	for nodeID, containerID := range stages {
		e.cleanupContainer(ctx, containerID, nodeID)
	}
}

// runWithRetry executes a stage function with optional retries based on the
// node's RetryPolicy. If the policy is nil or max_retries is 0, the function
// runs exactly once. Between retries, it waits with exponential backoff.
func (e *Engine) runWithRetry(ctx context.Context, nodeID string, node *portwhinev1.PipelineNode, fn func() error) error {
	policy := node.GetRetryPolicy()
	maxRetries := 0
	if policy != nil {
		maxRetries = int(policy.GetMaxRetries())
	}

	initialBackoff := 5 * time.Second
	maxBackoff := 5 * time.Minute
	if policy != nil && policy.GetInitialBackoffSeconds() > 0 {
		initialBackoff = time.Duration(policy.GetInitialBackoffSeconds()) * time.Second
	}
	if policy != nil && policy.GetMaxBackoffSeconds() > 0 {
		maxBackoff = time.Duration(policy.GetMaxBackoffSeconds()) * time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff: initialBackoff * 2^(attempt-1), capped at maxBackoff.
			backoff := initialBackoff * (1 << (attempt - 1))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			e.logger.Info("retrying stage",
				slog.String("node_id", nodeID),
				slog.Int("attempt", attempt+1),
				slog.Int("max_attempts", maxRetries+1),
				slog.Duration("backoff", backoff),
				slog.Any("last_error", lastErr),
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't retry if the context was cancelled.
		if ctx.Err() != nil {
			return lastErr
		}
	}

	return lastErr
}

// runTriggerStage starts a trigger container, initializes it via gRPC, and
// routes its output to downstream stages.
func (e *Engine) runTriggerStage(ctx context.Context, state *runState, nodeID string, node *portwhinev1.PipelineNode) error {
	// Ensure downstream channels are closed even if this stage fails early.
	defer state.router.SignalSourceDone(nodeID)

	e.updateStageStatus(state, nodeID, "running")
	if err := e.updateStepResult(ctx, state.run.ID, nodeID, "running"); err != nil {
		return err
	}

	// Start the container.
	containerID, err := e.startContainer(ctx, state.run.ID, nodeID, node)
	if err != nil {
		e.updateStageStatus(state, nodeID, "failed")
		_ = e.updateStepResult(ctx, state.run.ID, nodeID, "failed")
		return fmt.Errorf("failed to start trigger container: %w", err)
	}

	state.mu.Lock()
	state.stages[nodeID].ContainerID = containerID
	state.mu.Unlock()

	// Create a channel for the trigger to produce items into.
	output := make(chan *portwhinev1.DataItem, channelCapacity)

	// Connect to the trigger container via gRPC with readiness polling.
	// The Initialize call is included in the backoff loop so that transient
	// "connection refused" errors (container still starting) are retried.
	if e.triggerClientFactory != nil {
		containerAddr := e.containerAddress(state.run.ID, nodeID)
		stageConfig := e.buildStageConfig(state.run.ID, nodeID, node)

		triggerClient, err := connectWithBackoff(ctx, nodeID, e.logger, func() (TriggerClient, error) {
			return e.triggerClientFactory(containerAddr)
		}, func(c TriggerClient) error {
			return c.Initialize(ctx, stageConfig)
		})
		if err != nil {
			e.updateStageStatus(state, nodeID, "failed")
			_ = e.updateStepResult(ctx, state.run.ID, nodeID, "failed")
			return fmt.Errorf("failed to initialize trigger: %w", err)
		}

		// Build heartbeat callback that updates StageRunner metrics.
		stage := state.stages[nodeID]
		onHeartbeat := func(_, itemsProduced, errorsCount uint64) {
			stage.ItemsOut.Store(itemsProduced)
			stage.ErrorCount.Store(errorsCount)
		}

		// Start the trigger; it writes items to the output channel.
		if err := triggerClient.Start(ctx, output, onHeartbeat); err != nil {
			e.updateStageStatus(state, nodeID, "failed")
			_ = e.updateStepResult(ctx, state.run.ID, nodeID, "failed")
			return fmt.Errorf("trigger start failed: %w", err)
		}
	} else {
		// No trigger client factory configured: close output immediately.
		// This path is used in testing when no real gRPC clients exist.
		close(output)
	}

	// Persist, publish, and route all trigger output to downstream nodes.
	stage := state.stages[nodeID]
	for item := range output {
		stage.ItemsOut.Add(1)
		e.persistAndPublish(ctx, state, nodeID, item)
		state.router.Route(nodeID, item)
	}

	// Stop and remove the container.
	e.cleanupContainer(ctx, containerID, nodeID)

	e.persistStageMetrics(ctx, state, nodeID)
	e.updateStageStatus(state, nodeID, "completed")
	return e.updateStepResult(ctx, state.run.ID, nodeID, "completed")
}

// runWorkerStage starts one or more worker containers (based on the replicas
// setting), connects via gRPC, and processes data items from upstream. Multiple
// replicas share the same input channel for work-stealing parallelism.
func (e *Engine) runWorkerStage(ctx context.Context, state *runState, nodeID string, node *portwhinev1.PipelineNode) error {
	// Ensure downstream channels are closed even if this stage fails early.
	defer state.router.SignalSourceDone(nodeID)

	e.updateStageStatus(state, nodeID, "running")
	if err := e.updateStepResult(ctx, state.run.ID, nodeID, "running"); err != nil {
		return err
	}

	replicas := int(node.GetReplicas())
	if replicas <= 0 {
		replicas = 1
	}

	// Get this node's input channel from the router (shared across replicas).
	inputCh := state.router.GetInputChannel(nodeID)

	// Merged output channel for all replicas.
	outputCh := make(chan *portwhinev1.DataItem, channelCapacity)

	stage := state.stages[nodeID]

	if e.workerClientFactory != nil {
		var replicaWg sync.WaitGroup
		var replicaErrors []error
		var replicaErrMu sync.Mutex

		for r := 0; r < replicas; r++ {
			replicaID := fmt.Sprintf("%s-r%d", nodeID, r)
			containerName := nodeID
			if replicas > 1 {
				containerName = replicaID
			}

			containerID, err := e.startContainer(ctx, state.run.ID, containerName, node)
			if err != nil {
				e.updateStageStatus(state, nodeID, "failed")
				_ = e.updateStepResult(ctx, state.run.ID, nodeID, "failed")
				return fmt.Errorf("failed to start worker container %s: %w", replicaID, err)
			}

			// Store the first replica's container ID in the stage for status reporting.
			if r == 0 {
				state.mu.Lock()
				stage.ContainerID = containerID
				state.mu.Unlock()
			}

			replicaWg.Add(1)
			go func(replicaID string, containerID runtime.ContainerID) {
				defer replicaWg.Done()
				defer e.cleanupContainer(ctx, containerID, replicaID)

				containerAddr := e.containerAddress(state.run.ID, replicaID)
				if replicas == 1 {
					containerAddr = e.containerAddress(state.run.ID, nodeID)
				}

				if err := e.runWorkerReplica(ctx, replicaParams{
					replicaID:     replicaID,
					nodeID:        nodeID,
					runID:         state.run.ID,
					node:          node,
					containerAddr: containerAddr,
					stage:         stage,
					inputCh:       inputCh,
					outputCh:      outputCh,
				}); err != nil {
					replicaErrMu.Lock()
					replicaErrors = append(replicaErrors, err)
					replicaErrMu.Unlock()
				}
			}(replicaID, containerID)
		}

		// Wait for all replicas to finish, then close the merged output channel.
		go func() {
			replicaWg.Wait()
			close(outputCh)
		}()
	} else {
		// No worker client factory: pass-through mode for testing.
		go func() {
			defer close(outputCh)
			for item := range inputCh {
				select {
				case outputCh <- item:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Persist, publish, and route all worker output to downstream nodes.
	for item := range outputCh {
		stage.ItemsOut.Add(1)
		e.persistAndPublish(ctx, state, nodeID, item)
		state.router.Route(nodeID, item)
	}

	e.persistStageMetrics(ctx, state, nodeID)
	e.updateStageStatus(state, nodeID, "completed")
	return e.updateStepResult(ctx, state.run.ID, nodeID, "completed")
}

// runWorkerReplica runs a single worker replica: connects via gRPC, initializes,
// processes items from inputCh writing to outputCh, and shuts down.
// replicaParams bundles the arguments for runWorkerReplica to keep the
// function signature compact.
type replicaParams struct {
	replicaID     string
	nodeID        string
	runID         string
	node          *portwhinev1.PipelineNode
	containerAddr string
	stage         *StageRunner
	inputCh       <-chan *portwhinev1.DataItem
	outputCh      chan<- *portwhinev1.DataItem
}

func (e *Engine) runWorkerReplica(ctx context.Context, p replicaParams) error {
	stageConfig := e.buildStageConfig(p.runID, p.nodeID, p.node)

	workerClient, err := connectWithBackoff(ctx, p.replicaID, e.logger, func() (WorkerClient, error) {
		return e.workerClientFactory(p.containerAddr)
	}, func(c WorkerClient) error {
		return c.Initialize(ctx, stageConfig)
	})
	if err != nil {
		return fmt.Errorf("replica %s init: %w", p.replicaID, err)
	}

	onHeartbeat := func(itemsProcessed, itemsProduced, errorsCount uint64) {
		p.stage.ItemsIn.Add(itemsProcessed)
		p.stage.ItemsOut.Add(itemsProduced)
		p.stage.ErrorCount.Add(errorsCount)
	}

	if err := workerClient.Process(ctx, p.inputCh, p.outputCh, onHeartbeat); err != nil {
		return fmt.Errorf("replica %s process: %w", p.replicaID, err)
	}

	if err := workerClient.Shutdown(ctx); err != nil {
		e.logger.Warn("worker shutdown returned error",
			slog.String("replica_id", p.replicaID),
			slog.Any("error", err),
		)
	}

	return nil
}

// persistAndPublish converts a DataItem to a store record, persists it to the
// database, and publishes it to live-stream subscribers. Called from trigger
// and worker output loops to ensure all produced items are automatically
// persisted regardless of pipeline topology.
func (e *Engine) persistAndPublish(ctx context.Context, state *runState, nodeID string, item *portwhinev1.DataItem) {
	record, err := dataItemToRecord(state.run.ID, item)
	if err != nil {
		e.logger.Error("failed to convert data item to record",
			slog.String("node_id", nodeID),
			slog.String("item_id", item.GetId()),
			slog.Any("error", err),
		)
		return
	}

	if err := e.store.DataItems.Create(ctx, record); err != nil {
		e.logger.Error("failed to persist data item",
			slog.String("node_id", nodeID),
			slog.String("item_id", item.GetId()),
			slog.Any("error", err),
		)
	} else if e.metrics != nil {
		e.metrics.DataItemPersisted(item.GetType())
	}

	state.publish(item)
}

// runOutputStage handles output nodes. If the node specifies a container image
// (e.g. webhook-output, email-output), it launches the container and streams
// items through the standard Worker gRPC interface. Otherwise it drains the
// input channel (legacy drain-only behaviour).
func (e *Engine) runOutputStage(ctx context.Context, state *runState, nodeID string, node *portwhinev1.PipelineNode) error {
	e.updateStageStatus(state, nodeID, "running")
	if err := e.updateStepResult(ctx, state.run.ID, nodeID, "running"); err != nil {
		return err
	}

	inputCh := state.router.GetInputChannel(nodeID)
	stage := state.stages[nodeID]

	if node.GetImage() != "" && e.workerClientFactory != nil {
		// Container-based output node: full worker lifecycle.
		containerID, err := e.startContainer(ctx, state.run.ID, nodeID, node)
		if err != nil {
			e.updateStageStatus(state, nodeID, "failed")
			_ = e.updateStepResult(ctx, state.run.ID, nodeID, "failed")
			return fmt.Errorf("start output container %s: %w", nodeID, err)
		}

		state.mu.Lock()
		stage.ContainerID = containerID
		state.mu.Unlock()

		defer e.cleanupContainer(ctx, containerID, nodeID)

		outputCh := make(chan *portwhinev1.DataItem, channelCapacity)

		go func() {
			defer close(outputCh)
			containerAddr := e.containerAddress(state.run.ID, nodeID)
			if err := e.runWorkerReplica(ctx, replicaParams{
				replicaID:     nodeID,
				nodeID:        nodeID,
				runID:         state.run.ID,
				node:          node,
				containerAddr: containerAddr,
				stage:         stage,
				inputCh:       inputCh,
				outputCh:      outputCh,
			}); err != nil {
				e.logger.Error("output container failed",
					slog.String("node_id", nodeID),
					slog.Any("error", err),
				)
			}
		}()

		// Drain optional confirmation items (webhook_delivery, email_delivery).
		for item := range outputCh {
			stage.ItemsOut.Add(1)
			e.persistAndPublish(ctx, state, nodeID, item)
		}
	} else {
		// Legacy drain-only behaviour for output nodes without an image.
		for range inputCh {
			stage.ItemsIn.Add(1)
		}
	}

	e.persistStageMetrics(ctx, state, nodeID)
	e.updateStageStatus(state, nodeID, "completed")
	return e.updateStepResult(ctx, state.run.ID, nodeID, "completed")
}

// resolveImage looks up the Docker image reference for a registered worker/trigger name.
// If the name contains a '/' or ':', it is assumed to be a full image reference already.
func (e *Engine) resolveImage(ctx context.Context, imageName string) (string, error) {
	// If it looks like a full Docker image reference, use it directly.
	if strings.Contains(imageName, "/") || strings.Contains(imageName, ":") {
		return imageName, nil
	}

	// Look up in the worker_images table by name.
	wi, err := e.store.WorkerImages.GetByName(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("resolve image %q: %w", imageName, err)
	}
	return wi.Image, nil
}

// startContainer starts a container for the given pipeline node.
func (e *Engine) startContainer(ctx context.Context, runID, nodeID string, node *portwhinev1.PipelineNode) (runtime.ContainerID, error) {
	containerName := fmt.Sprintf("portwhine-%s-%s", runID[:8], nodeID)

	// Resolve the image name to a full Docker image reference.
	dockerImage, err := e.resolveImage(ctx, node.GetImage())
	if err != nil {
		return "", fmt.Errorf("resolve image for node %q: %w", nodeID, err)
	}

	spec := runtime.ContainerSpec{
		Image: dockerImage,
		Name:  containerName,
		Env: map[string]string{
			"OPERATOR_ADDRESS": e.operatorAddress,
			"PORTWHINE_RUN_ID": runID,
			"PORTWHINE_NODE_ID": nodeID,
		},
		Labels: map[string]string{
			"portwhine.run_id":  runID,
			"portwhine.node_id": nodeID,
			"portwhine.managed": "true",
		},
		Network: runtime.NetworkConfig{
			OperatorAddress: e.operatorAddress,
		},
	}

	// Add NET_ADMIN capability for workers that need raw socket access (nmap, etc.)
	if strings.Contains(strings.ToLower(dockerImage), "nmap") {
		spec.Capabilities = []string{"NET_ADMIN", "NET_RAW"}
		e.logger.Info("adding network capabilities for container",
			slog.String("node_id", nodeID),
			slog.String("image", dockerImage),
			slog.Any("capabilities", spec.Capabilities),
		)
	}

	id, err := e.runtime.Start(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("failed to start container %q (image %s): %w",
			containerName, node.GetImage(), err)
	}

	e.logger.Info("started container",
		slog.String("node_id", nodeID),
		slog.String("container_id", string(id)),
		slog.String("image", node.GetImage()),
	)

	return id, nil
}

// cleanupContainer stops and removes a container, logging any errors.
func (e *Engine) cleanupContainer(ctx context.Context, id runtime.ContainerID, nodeID string) {
	if id == "" {
		return
	}

	// Use a background context for cleanup so it succeeds even if the run
	// context was cancelled.
	cleanupCtx := context.Background()

	if err := e.runtime.Stop(cleanupCtx, id, 30*time.Second); err != nil {
		// Don't warn if container is already gone (common when containers exit quickly)
		if !isContainerNotFoundError(err) {
			e.logger.Warn("failed to stop container during cleanup",
				slog.String("node_id", nodeID),
				slog.String("container_id", string(id)),
				slog.Any("error", err),
			)
		}
	}

	if err := e.runtime.Remove(cleanupCtx, id); err != nil {
		// Don't warn if container is already gone (common when containers exit quickly)
		if !isContainerNotFoundError(err) {
			e.logger.Warn("failed to remove container during cleanup",
				slog.String("node_id", nodeID),
				slog.String("container_id", string(id)),
				slog.Any("error", err),
			)
		}
	}
}

// isContainerNotFoundError returns true if the error indicates the container doesn't exist.
func isContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "No such container") ||
		strings.Contains(msg, "no such container") ||
		strings.Contains(msg, "not found")
}

// connectWithBackoff retries the given factory function and readiness probe with
// exponential backoff until both succeed, the context is cancelled, or the
// maximum timeout (60s) is reached. The ready function is called after the
// factory succeeds to verify the container's gRPC server is actually reachable
// (e.g. by calling Initialize). If ready returns a non-transient error, the
// function fails immediately without further retries.
func connectWithBackoff[T any](ctx context.Context, nodeID string, logger *slog.Logger, factory func() (T, error), ready func(T) error) (T, error) {
	var zero T
	const maxWait = 60 * time.Second
	deadline := time.Now().Add(maxWait)
	delay := 500 * time.Millisecond

	for {
		client, err := factory()
		if err == nil && ready != nil {
			if readyErr := ready(client); readyErr != nil {
				if !isTransientError(readyErr) {
					return zero, readyErr
				}
				err = readyErr
			} else {
				return client, nil
			}
		} else if err == nil {
			return client, nil
		}

		if time.Now().After(deadline) {
			return zero, fmt.Errorf("timed out waiting for node %q gRPC server after %s: %w", nodeID, maxWait, err)
		}
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		logger.Debug("waiting for container gRPC server",
			slog.String("node_id", nodeID),
			slog.Duration("retry_in", delay),
			slog.Any("error", err),
		)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return zero, ctx.Err()
		}

		// Exponential backoff: 500ms, 1s, 2s, 4s, capped at 5s.
		delay = delay * 2
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}
	}
}

// isTransientError returns true if the error is likely a transient connection
// issue (container not ready yet) rather than a permanent failure.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "unavailable") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection reset")
}

// containerAddress returns the gRPC address for a container running within the
// same Docker network. Containers are named portwhine-<runID[:8]>-<nodeID> and
// listen on port 50051.
func (e *Engine) containerAddress(runID, nodeID string) string {
	containerName := fmt.Sprintf("portwhine-%s-%s", runID[:8], nodeID)
	return fmt.Sprintf("http://%s:50051", containerName)
}

// buildStageConfig creates a StageConfig protobuf message for a pipeline node.
func (e *Engine) buildStageConfig(runID, nodeID string, node *portwhinev1.PipelineNode) *portwhinev1.StageConfig {
	return &portwhinev1.StageConfig{
		NodeId:          nodeID,
		PipelineRunId:   runID,
		Parameters:      node.GetConfig(),
		OperatorAddress: e.operatorAddress,
	}
}

// updateStageStatus updates the in-memory status of a stage runner.
func (e *Engine) updateStageStatus(state *runState, nodeID, status string) {
	state.mu.Lock()
	defer state.mu.Unlock()

	stage, ok := state.stages[nodeID]
	if !ok {
		return
	}

	stage.Status = status
	now := time.Now()
	switch status {
	case "running":
		stage.StartedAt = &now
	case "completed", "failed":
		stage.FinishedAt = &now
	}
}

// updateStepResult persists the step result status to the database.
func (e *Engine) updateStepResult(ctx context.Context, runID, nodeID, status string) error {
	steps, err := e.store.PipelineRuns.ListStepResults(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to list step results: %w", err)
	}

	for i := range steps {
		if steps[i].NodeID == nodeID {
			steps[i].Status = status
			now := time.Now()
			switch status {
			case "running":
				steps[i].StartedAt = &now
			case "completed", "failed":
				steps[i].FinishedAt = &now
			}
			if err := e.store.PipelineRuns.UpdateStepResult(ctx, &steps[i]); err != nil {
				return fmt.Errorf("failed to update step result for node %q: %w", nodeID, err)
			}
			return nil
		}
	}

	return fmt.Errorf("step result not found for node %q in run %q", nodeID, runID)
}

// dataItemToRecord converts a protobuf DataItem to a store DataItemRecord.
func dataItemToRecord(runID string, item *portwhinev1.DataItem) (*store.DataItemRecord, error) {
	var dataJSON, metaJSON datatypes.JSON

	if item.GetData() != nil {
		raw, err := json.Marshal(item.GetData().AsMap())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data item data: %w", err)
		}
		dataJSON = datatypes.JSON(raw)
	}

	if item.GetMetadata() != nil {
		metaMap := map[string]any{
			"source":  item.GetMetadata().GetSource(),
			"node_id": item.GetMetadata().GetNodeId(),
			"labels":  item.GetMetadata().GetLabels(),
		}
		if item.GetMetadata().GetCreatedAt() != nil {
			metaMap["created_at"] = item.GetMetadata().GetCreatedAt().AsTime().Format(time.RFC3339Nano)
		}
		raw, err := json.Marshal(metaMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data item metadata: %w", err)
		}
		metaJSON = datatypes.JSON(raw)
	}

	record := &store.DataItemRecord{
		RunID:      runID,
		Type:       item.GetType(),
		Data:       dataJSON,
		Metadata:   metaJSON,
		RawPayload: item.GetRawPayload(),
		ParentIDs:  item.GetParentIds(),
	}

	// Use the item's own ID if provided.
	if item.GetId() != "" {
		record.ID = item.GetId()
	}

	return record, nil
}

// failRun marks a pipeline run as failed with the given error and returns the
// error for convenient chaining in callers.
func (e *Engine) failRun(ctx context.Context, runID string, err error) error {
	e.logger.Error("pipeline run failed",
		slog.String("run_id", runID),
		slog.Any("error", err),
	)

	if updateErr := e.store.PipelineRuns.UpdateStatus(ctx, runID, "failed"); updateErr != nil {
		e.logger.Error("failed to update run status to failed",
			slog.String("run_id", runID),
			slog.Any("error", updateErr),
		)
	}

	return err
}

// ---------------------------------------------------------------------------
// Live Monitoring
// ---------------------------------------------------------------------------

// persistStageMetrics writes the final item counts into the StepResult.Output
// JSONB column so metrics survive after the run is cleaned up from memory.
func (e *Engine) persistStageMetrics(ctx context.Context, state *runState, nodeID string) {
	state.mu.Lock()
	stage, ok := state.stages[nodeID]
	state.mu.Unlock()
	if !ok {
		return
	}

	metrics := map[string]uint64{
		"items_in":  stage.ItemsIn.Load(),
		"items_out": stage.ItemsOut.Load(),
		"errors":    stage.ErrorCount.Load(),
	}
	raw, err := json.Marshal(metrics)
	if err != nil {
		e.logger.Warn("failed to marshal stage metrics", slog.Any("error", err))
		return
	}

	steps, err := e.store.PipelineRuns.ListStepResults(ctx, state.run.ID)
	if err != nil {
		e.logger.Warn("failed to list step results for metrics persist", slog.Any("error", err))
		return
	}
	for i := range steps {
		if steps[i].NodeID == nodeID {
			steps[i].Output = datatypes.JSON(raw)
			if updateErr := e.store.PipelineRuns.UpdateStepResult(ctx, &steps[i]); updateErr != nil {
				e.logger.Warn("failed to persist stage metrics",
					slog.String("node_id", nodeID),
					slog.Any("error", updateErr),
				)
			}
			return
		}
	}
}

// GetNodeStatuses returns live per-node status for an active pipeline run.
// The second return value is false if the run is not in memory (finished or unknown).
func (e *Engine) GetNodeStatuses(ctx context.Context, runID string) ([]NodeStatusInfo, bool) {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()
	if !ok {
		return nil, false
	}

	state.mu.Lock()
	stages := make(map[string]*StageRunner, len(state.stages))
	for k, v := range state.stages {
		stages[k] = v
	}
	state.mu.Unlock()

	infos := make([]NodeStatusInfo, 0, len(stages))
	for _, stage := range stages {
		info := NodeStatusInfo{
			NodeID:     stage.NodeID,
			Status:     stage.Status,
			ItemsIn:    stage.ItemsIn.Load(),
			ItemsOut:   stage.ItemsOut.Load(),
			Errors:     stage.ErrorCount.Load(),
			Error:      stage.LastError,
		}

		if stage.Node != nil {
			info.WorkerType = stage.Node.GetImage()
		}

		// Enrich with live container status from Docker/Kubernetes.
		if stage.ContainerID != "" {
			info.ContainerID = string(stage.ContainerID)
			cInfo, err := e.runtime.Status(ctx, stage.ContainerID)
			if err == nil {
				info.ContainerStatus = cInfo.Status
				info.ContainerExitCode = cInfo.ExitCode
				if !cInfo.StartedAt.IsZero() {
					t := cInfo.StartedAt
					info.ContainerStartedAt = &t
				}
				if !cInfo.FinishedAt.IsZero() {
					t := cInfo.FinishedAt
					info.ContainerFinishedAt = &t
				}
				info.ContainerMessage = cInfo.Message
			}
		}

		infos = append(infos, info)
	}

	return infos, true
}

// GetNodeLogs returns a log reader for a specific node's container in an active run.
func (e *Engine) GetNodeLogs(ctx context.Context, runID, nodeID string, opts runtime.LogOptions) (io.ReadCloser, error) {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("run %q not found or already finished", runID)
	}

	state.mu.Lock()
	stage, ok := state.stages[nodeID]
	state.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("node %q not found in run %q", nodeID, runID)
	}

	if stage.ContainerID == "" {
		return nil, fmt.Errorf("node %q has no container (status: %s)", nodeID, stage.Status)
	}

	return e.runtime.Logs(ctx, stage.ContainerID, opts)
}

// Subscribe registers a subscriber for live DataItems from a running pipeline.
// Returns a channel that receives items, an unsubscribe function, and an error.
// The channel is closed when the run completes.
func (e *Engine) Subscribe(runID string) (<-chan *portwhinev1.DataItem, func(), error) {
	e.mu.Lock()
	state, ok := e.runs[runID]
	e.mu.Unlock()
	if !ok {
		return nil, nil, fmt.Errorf("run %q not found or already finished", runID)
	}

	ch := make(chan *portwhinev1.DataItem, 1000)

	state.subscribersMu.Lock()
	state.subscribers = append(state.subscribers, ch)
	state.subscribersMu.Unlock()

	unsubscribe := func() {
		state.subscribersMu.Lock()
		defer state.subscribersMu.Unlock()
		for i, sub := range state.subscribers {
			if sub == ch {
				state.subscribers = append(state.subscribers[:i], state.subscribers[i+1:]...)
				break
			}
		}
	}

	// Close the subscriber channel when the run finishes.
	go func() {
		<-state.done
		close(ch)
	}()

	return ch, unsubscribe, nil
}
