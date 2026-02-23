package operator

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/v1/portwhinev1connect"
	"github.com/portwhine/portwhine/internal/auth"
	"github.com/portwhine/portwhine/internal/pipeline"
	"github.com/portwhine/portwhine/internal/runtime"
	"github.com/portwhine/portwhine/internal/store"
)

// Compile-time check that Handler implements the generated interface.
var _ portwhinev1connect.OperatorServiceHandler = (*Handler)(nil)

var errEngineUnavailable = errors.New("pipeline engine not available")

// Handler implements portwhinev1connect.OperatorServiceHandler.
type Handler struct {
	store      *store.Store
	jwt        *auth.JWTService
	engine     *pipeline.Engine
	authorizer *auth.Authorizer
	scheduler  *Scheduler
	logger     *slog.Logger
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(s *store.Store, jwt *auth.JWTService, engine *pipeline.Engine, authorizer *auth.Authorizer, scheduler *Scheduler, logger *slog.Logger) *Handler {
	return &Handler{
		store:      s,
		jwt:        jwt,
		engine:     engine,
		authorizer: authorizer,
		scheduler:  scheduler,
		logger:     logger,
	}
}

// audit records an action in the audit log. Errors are logged but not propagated.
func (h *Handler) audit(ctx context.Context, action, resource string, resourceID *string, details map[string]any) {
	userID := userIDFromContext(ctx)
	var uid *string
	if userID != "" {
		uid = &userID
	}

	var detailsJSON datatypes.JSON
	if details != nil {
		raw, err := json.Marshal(details)
		if err == nil {
			detailsJSON = datatypes.JSON(raw)
		}
	}

	entry := &store.AuditLog{
		UserID:     uid,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    detailsJSON,
	}

	if err := h.store.AuditLog.Create(ctx, entry); err != nil {
		h.logger.Warn("failed to write audit log",
			slog.String("action", action),
			slog.String("resource", resource),
			slog.Any("error", err),
		)
	}
}

// ---------------------------------------------------------------------------
// Pipeline CRUD
// ---------------------------------------------------------------------------

func (h *Handler) CreatePipeline(
	ctx context.Context,
	req *connect.Request[portwhinev1.CreatePipelineRequest],
) (*connect.Response[portwhinev1.CreatePipelineResponse], error) {
	def := req.Msg.GetDefinition()
	if def == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("definition is required"))
	}

	// Allow creating empty pipelines - they can be configured in the editor.
	// Validation will happen before execution (StartPipeline).
	// if err := pipeline.ValidatePipeline(def); err != nil {
	// 	return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid pipeline: %w", err))
	// }

	defJSON, err := protojson.Marshal(def)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("marshal definition: %w", err))
	}

	p := &store.Pipeline{
		Name:        def.GetName(),
		Description: def.GetDescription(),
		Schedule:    def.GetSchedule(),
		Definition:  datatypes.JSON(defJSON),
		Version:     1,
		IsActive:    true,
		CreatedByID: userIDFromContext(ctx),
	}

	if err := h.store.Pipelines.Create(ctx, p); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create pipeline: %w", err))
	}

	h.scheduler.UpdateSchedule(p.ID, p.Schedule, p.IsActive)
	h.audit(ctx, "create", "pipeline", &p.ID, map[string]any{"name": p.Name})

	return connect.NewResponse(&portwhinev1.CreatePipelineResponse{
		PipelineId: p.ID,
	}), nil
}

func (h *Handler) GetPipeline(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetPipelineRequest],
) (*connect.Response[portwhinev1.GetPipelineResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "pipelines", req.Msg.GetPipelineId(), "read"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	p, err := h.store.Pipelines.GetByID(ctx, req.Msg.GetPipelineId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pipeline not found: %w", err))
	}

	def, err := unmarshalPipelineDefinition(p.Definition)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unmarshal definition: %w", err))
	}

	return connect.NewResponse(&portwhinev1.GetPipelineResponse{
		PipelineId: p.ID,
		Definition: def,
		Version:    int32(p.Version),
		CreatedBy:  p.CreatedByID,
		CreatedAt:  timestamppb.New(p.CreatedAt),
		UpdatedAt:  timestamppb.New(p.UpdatedAt),
	}), nil
}

func (h *Handler) UpdatePipeline(
	ctx context.Context,
	req *connect.Request[portwhinev1.UpdatePipelineRequest],
) (*connect.Response[portwhinev1.UpdatePipelineResponse], error) {
	def := req.Msg.GetDefinition()
	if def == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("definition is required"))
	}

	// Validate the pipeline definition before persisting.
	if err := pipeline.ValidatePipeline(def); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid pipeline: %w", err))
	}

	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "pipelines", req.Msg.GetPipelineId(), "update"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	p, err := h.store.Pipelines.GetByID(ctx, req.Msg.GetPipelineId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pipeline not found: %w", err))
	}

	defJSON, err := protojson.Marshal(def)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("marshal definition: %w", err))
	}

	p.Name = def.GetName()
	p.Description = def.GetDescription()
	p.Schedule = def.GetSchedule()
	p.Definition = datatypes.JSON(defJSON)
	p.Version++

	if err := h.store.Pipelines.Update(ctx, p); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update pipeline: %w", err))
	}

	h.scheduler.UpdateSchedule(p.ID, p.Schedule, p.IsActive)

	return connect.NewResponse(&portwhinev1.UpdatePipelineResponse{
		Version: int32(p.Version),
	}), nil
}

func (h *Handler) ListPipelines(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListPipelinesRequest],
) (*connect.Response[portwhinev1.ListPipelinesResponse], error) {
	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	claims := claimsFromContext(ctx)
	allowedIDs, unrestricted, err := h.authorizer.FilterAccessibleIDs(ctx, claims.UserID, claims.Role, "pipelines")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("filter accessible: %w", err))
	}

	var pipelines []store.Pipeline
	var total int64
	if unrestricted {
		pipelines, total, err = h.store.Pipelines.List(ctx, offset, pageSize)
	} else {
		pipelines, total, err = h.store.Pipelines.ListByIDs(ctx, allowedIDs, offset, pageSize)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list pipelines: %w", err))
	}

	summaries := make([]*portwhinev1.PipelineSummary, 0, len(pipelines))
	for _, p := range pipelines {
		summaries = append(summaries, &portwhinev1.PipelineSummary{
			PipelineId:  p.ID,
			Name:        p.Name,
			Description: p.Description,
			Version:     int32(p.Version),
			CreatedAt:   timestamppb.New(p.CreatedAt),
			UpdatedAt:   timestamppb.New(p.UpdatedAt),
		})
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.ListPipelinesResponse{
		Pipelines:     summaries,
		NextPageToken: nextToken,
	}), nil
}

func (h *Handler) DeletePipeline(
	ctx context.Context,
	req *connect.Request[portwhinev1.DeletePipelineRequest],
) (*connect.Response[portwhinev1.DeletePipelineResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "pipelines", req.Msg.GetPipelineId(), "delete"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	pid := req.Msg.GetPipelineId()
	if err := h.store.Pipelines.Delete(ctx, pid); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete pipeline: %w", err))
	}
	h.scheduler.UpdateSchedule(pid, "", false)
	h.audit(ctx, "delete", "pipeline", &pid, nil)
	return connect.NewResponse(&portwhinev1.DeletePipelineResponse{}), nil
}

// ---------------------------------------------------------------------------
// Pipeline Execution
// ---------------------------------------------------------------------------

func (h *Handler) StartPipeline(
	ctx context.Context,
	req *connect.Request[portwhinev1.StartPipelineRequest],
) (*connect.Response[portwhinev1.StartPipelineResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "pipelines", req.Msg.GetPipelineId(), "execute"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	p, err := h.store.Pipelines.GetByID(ctx, req.Msg.GetPipelineId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pipeline not found: %w", err))
	}

	run := &store.PipelineRun{
		PipelineID:         p.ID,
		DefinitionSnapshot: p.Definition,
		Status:             "pending",
		CreatedByID:        userIDFromContext(ctx),
	}

	if err := h.store.PipelineRuns.Create(ctx, run); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create pipeline run: %w", err))
	}

	// Start execution via the pipeline engine.
	if h.engine != nil {
		if err := h.engine.StartRun(ctx, run); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("start pipeline run: %w", err))
		}
	}

	h.audit(ctx, "start", "pipeline_run", &run.ID, map[string]any{"pipeline_id": p.ID})

	return connect.NewResponse(&portwhinev1.StartPipelineResponse{
		RunId: run.ID,
	}), nil
}

func (h *Handler) StopPipelineRun(
	ctx context.Context,
	req *connect.Request[portwhinev1.StopPipelineRunRequest],
) (*connect.Response[portwhinev1.StopPipelineRunResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", req.Msg.GetRunId(), "execute"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	// Stop via the engine if the run is active.
	if h.engine != nil {
		if err := h.engine.StopRun(ctx, req.Msg.GetRunId()); err != nil {
			// If engine doesn't know about it, just update DB status.
			if err2 := h.store.PipelineRuns.UpdateStatus(ctx, req.Msg.GetRunId(), "cancelled"); err2 != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("stop pipeline run: %w", err2))
			}
		}
	} else {
		if err := h.store.PipelineRuns.UpdateStatus(ctx, req.Msg.GetRunId(), "cancelled"); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("stop pipeline run: %w", err))
		}
	}
	return connect.NewResponse(&portwhinev1.StopPipelineRunResponse{}), nil
}

func (h *Handler) PausePipelineRun(
	ctx context.Context,
	req *connect.Request[portwhinev1.PausePipelineRunRequest],
) (*connect.Response[portwhinev1.PausePipelineRunResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", req.Msg.GetRunId(), "execute"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	if h.engine == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errEngineUnavailable)
	}
	if err := h.engine.PauseRun(ctx, req.Msg.GetRunId()); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pause pipeline run: %w", err))
	}
	return connect.NewResponse(&portwhinev1.PausePipelineRunResponse{}), nil
}

func (h *Handler) ResumePipelineRun(
	ctx context.Context,
	req *connect.Request[portwhinev1.ResumePipelineRunRequest],
) (*connect.Response[portwhinev1.ResumePipelineRunResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", req.Msg.GetRunId(), "execute"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	if h.engine == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errEngineUnavailable)
	}
	if err := h.engine.ResumeRun(ctx, req.Msg.GetRunId()); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("resume pipeline run: %w", err))
	}
	return connect.NewResponse(&portwhinev1.ResumePipelineRunResponse{}), nil
}

func (h *Handler) GetPipelineRunStatus(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetPipelineRunStatusRequest],
) (*connect.Response[portwhinev1.GetPipelineRunStatusResponse], error) {
	claims := claimsFromContext(ctx)
	runID := req.Msg.GetRunId()
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", runID, "read"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	run, err := h.store.PipelineRuns.GetByID(ctx, runID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pipeline run not found: %w", err))
	}

	status := pipelineRunToStatus(run)

	// Try to get live node statuses from the engine first.
	status.Nodes = h.resolveNodeStatuses(ctx, runID)

	return connect.NewResponse(&portwhinev1.GetPipelineRunStatusResponse{
		Status: status,
	}), nil
}

// resolveNodeStatuses tries the live engine first; falls back to DB step results.
func (h *Handler) resolveNodeStatuses(ctx context.Context, runID string) []*portwhinev1.NodeStatus {
	if h.engine != nil {
		if liveNodes, ok := h.engine.GetNodeStatuses(ctx, runID); ok {
			return liveNodeStatusesToProto(liveNodes)
		}
	}
	return h.buildNodeStatusesFromDB(ctx, runID)
}

// liveNodeStatusesToProto converts engine NodeStatusInfo values to proto NodeStatus messages.
func liveNodeStatusesToProto(liveNodes []pipeline.NodeStatusInfo) []*portwhinev1.NodeStatus {
	nodes := make([]*portwhinev1.NodeStatus, 0, len(liveNodes))
	for _, n := range liveNodes {
		ns := &portwhinev1.NodeStatus{
			NodeId:            n.NodeID,
			WorkerType:        n.WorkerType,
			WorkerStatus:      mapWorkerStatus(n.Status),
			ItemsIn:           n.ItemsIn,
			ItemsOut:          n.ItemsOut,
			Errors:            n.Errors,
			ErrorMessage:      n.Error,
			ContainerId:       n.ContainerID,
			ContainerStatus:   string(n.ContainerStatus),
			ContainerExitCode: int32(n.ContainerExitCode),
			ContainerMessage:  n.ContainerMessage,
		}
		if n.ContainerStartedAt != nil {
			ns.ContainerStartedAt = timestamppb.New(*n.ContainerStartedAt)
		}
		if n.ContainerFinishedAt != nil {
			ns.ContainerFinishedAt = timestamppb.New(*n.ContainerFinishedAt)
		}
		nodes = append(nodes, ns)
	}
	return nodes
}

// buildNodeStatusesFromDB loads StepResult records from the database and
// converts them to proto NodeStatus messages. The Output JSONB field is parsed
// for items_in, items_out, and errors counters.
func (h *Handler) buildNodeStatusesFromDB(ctx context.Context, runID string) []*portwhinev1.NodeStatus {
	steps, err := h.store.PipelineRuns.ListStepResults(ctx, runID)
	if err != nil {
		return nil
	}

	nodes := make([]*portwhinev1.NodeStatus, 0, len(steps))
	for _, sr := range steps {
		ns := &portwhinev1.NodeStatus{
			NodeId:          sr.NodeID,
			WorkerStatus:    mapWorkerStatus(sr.Status),
			ContainerId:     sr.ContainerID,
			ContainerStatus: sr.Status, // best approximation from DB
			ErrorMessage:    sr.ErrorMessage,
		}
		if sr.ExitCode != nil {
			ns.ContainerExitCode = int32(*sr.ExitCode)
		}
		if sr.StartedAt != nil {
			ns.ContainerStartedAt = timestamppb.New(*sr.StartedAt)
		}
		if sr.FinishedAt != nil {
			ns.ContainerFinishedAt = timestamppb.New(*sr.FinishedAt)
		}

		// Parse Output JSONB for counters.
		if len(sr.Output) > 0 {
			var out struct {
				ItemsIn  uint64 `json:"items_in"`
				ItemsOut uint64 `json:"items_out"`
				Errors   uint64 `json:"errors"`
			}
			if json.Unmarshal(sr.Output, &out) == nil {
				ns.ItemsIn = out.ItemsIn
				ns.ItemsOut = out.ItemsOut
				ns.Errors = out.Errors
			}
		}

		nodes = append(nodes, ns)
	}
	return nodes
}

func (h *Handler) ListPipelineRuns(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListPipelineRunsRequest],
) (*connect.Response[portwhinev1.ListPipelineRunsResponse], error) {
	claims := claimsFromContext(ctx)
	pipelineID := req.Msg.GetPipelineId()

	// If a specific pipeline is requested, check access to it.
	if pipelineID != "" {
		if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "pipelines", pipelineID, "read"); err != nil {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
	}

	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	var runs []store.PipelineRun
	var total int64
	var err error

	if pipelineID != "" {
		runs, total, err = h.store.PipelineRuns.ListByPipeline(ctx, pipelineID, offset, pageSize)
	} else {
		runs, total, err = h.store.PipelineRuns.ListAll(ctx, offset, pageSize)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list pipeline runs: %w", err))
	}

	statuses := make([]*portwhinev1.PipelineRunStatus, 0, len(runs))
	for i := range runs {
		statuses = append(statuses, pipelineRunToStatus(&runs[i]))
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.ListPipelineRunsResponse{
		Runs:          statuses,
		NextPageToken: nextToken,
	}), nil
}

func (h *Handler) StreamPipelineResults(
	ctx context.Context,
	req *connect.Request[portwhinev1.StreamPipelineResultsRequest],
	stream *connect.ServerStream[portwhinev1.StreamPipelineResultsResponse],
) error {
	if h.engine == nil {
		return connect.NewError(connect.CodeUnavailable, errEngineUnavailable)
	}

	ch, unsubscribe, err := h.engine.Subscribe(req.Msg.GetRunId())
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("subscribe to run: %w", err))
	}
	defer unsubscribe()

	typeFilter := req.Msg.GetTypeFilter()
	filterSet := make(map[string]struct{}, len(typeFilter))
	for _, t := range typeFilter {
		filterSet[t] = struct{}{}
	}

	for item := range ch {
		if err := ctx.Err(); err != nil {
			return connect.NewError(connect.CodeCanceled, err)
		}

		if len(filterSet) > 0 {
			if _, ok := filterSet[item.GetType()]; !ok {
				continue
			}
		}

		if err := stream.Send(&portwhinev1.StreamPipelineResultsResponse{Item: item}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) GetNodeLogs(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetNodeLogsRequest],
	stream *connect.ServerStream[portwhinev1.GetNodeLogsResponse],
) error {
	if h.engine == nil {
		return connect.NewError(connect.CodeUnavailable, errEngineUnavailable)
	}

	opts := runtime.LogOptions{
		Tail:   int(req.Msg.GetTail()),
		Follow: req.Msg.GetFollow(),
	}

	reader, err := h.engine.GetNodeLogs(ctx, req.Msg.GetRunId(), req.Msg.GetNodeId(), opts)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("get node logs: %w", err))
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return connect.NewError(connect.CodeCanceled, err)
		}
		if err := stream.Send(&portwhinev1.GetNodeLogsResponse{Line: scanner.Text()}); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("read logs: %w", err))
	}

	return nil
}

// ---------------------------------------------------------------------------
// Data Items
// ---------------------------------------------------------------------------

func (h *Handler) GetDataItem(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetDataItemRequest],
) (*connect.Response[portwhinev1.GetDataItemResponse], error) {
	item, err := h.store.DataItems.GetByID(ctx, req.Msg.GetDataItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("data item not found: %w", err))
	}

	// Check the caller has read access to the parent run.
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", item.RunID, "read"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	return connect.NewResponse(&portwhinev1.GetDataItemResponse{
		Item: dataItemRecordToProto(item),
	}), nil
}

func (h *Handler) ListDataItems(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListDataItemsRequest],
) (*connect.Response[portwhinev1.ListDataItemsResponse], error) {
	runID := req.Msg.GetRunId()
	if runID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("run_id is required"))
	}

	// Check the caller has read access to the run.
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", runID, "read"); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	var items []store.DataItemRecord
	var total int64
	var err error

	if typeFilter := req.Msg.GetTypeFilter(); typeFilter != "" {
		items, total, err = h.store.DataItems.ListByRunAndType(ctx, runID, typeFilter, offset, pageSize)
	} else {
		items, total, err = h.store.DataItems.ListByRun(ctx, runID, offset, pageSize)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list data items: %w", err))
	}

	infos := make([]*portwhinev1.DataItemInfo, 0, len(items))
	for i := range items {
		infos = append(infos, dataItemRecordToProto(&items[i]))
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.ListDataItemsResponse{
		Items:         infos,
		NextPageToken: nextToken,
		TotalCount:    total,
	}), nil
}

func (h *Handler) SearchDataItems(
	ctx context.Context,
	req *connect.Request[portwhinev1.SearchDataItemsRequest],
) (*connect.Response[portwhinev1.SearchDataItemsResponse], error) {
	claims := claimsFromContext(ctx)

	// If run_id is provided, check access to that specific run.
	if runID := req.Msg.GetRunId(); runID != "" {
		if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", runID, "read"); err != nil {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
	}

	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	params := buildSearchParams(req.Msg.GetRunId(), req.Msg.GetQuery(), req.Msg.GetTypes(),
		req.Msg.GetCreatedAfter(), req.Msg.GetCreatedBefore())

	items, total, err := h.store.DataItems.Search(ctx, params, offset, pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("search data items: %w", err))
	}

	infos := make([]*portwhinev1.DataItemInfo, 0, len(items))
	for i := range items {
		infos = append(infos, dataItemRecordToProto(&items[i]))
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.SearchDataItemsResponse{
		Items:         infos,
		NextPageToken: nextToken,
		TotalCount:    total,
	}), nil
}

func (h *Handler) ExportDataItems(
	ctx context.Context,
	req *connect.Request[portwhinev1.ExportDataItemsRequest],
	stream *connect.ServerStream[portwhinev1.ExportDataItemsResponse],
) error {
	if err := h.checkExportRunAccess(ctx, req.Msg.GetRunId()); err != nil {
		return err
	}

	params := buildSearchParams(req.Msg.GetRunId(), req.Msg.GetQuery(), req.Msg.GetTypes(),
		req.Msg.GetCreatedAfter(), req.Msg.GetCreatedBefore())

	format := req.Msg.GetFormat()
	if format == "" {
		format = "json"
	}

	if format == "csv" {
		if err := stream.Send(&portwhinev1.ExportDataItemsResponse{Data: []byte("id,run_id,type,data,created_at\n")}); err != nil {
			return err
		}
	}

	return h.streamExportBatches(ctx, stream, params, format)
}

// checkExportRunAccess verifies run read access if a run_id is specified.
func (h *Handler) checkExportRunAccess(ctx context.Context, runID string) error {
	if runID == "" {
		return nil
	}
	claims := claimsFromContext(ctx)
	if err := h.authorizer.CheckInstanceAccess(ctx, claims.UserID, claims.Role, "runs", runID, "read"); err != nil {
		return connect.NewError(connect.CodePermissionDenied, err)
	}
	return nil
}

// streamExportBatches fetches data items in batches and streams them.
func (h *Handler) streamExportBatches(
	ctx context.Context,
	stream *connect.ServerStream[portwhinev1.ExportDataItemsResponse],
	params store.DataItemSearchParams,
	format string,
) error {
	const batchSize = 100
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return connect.NewError(connect.CodeCanceled, err)
		}

		items, _, err := h.store.DataItems.Search(ctx, params, offset, batchSize)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("export data items: %w", err))
		}
		if len(items) == 0 {
			break
		}

		if err := sendExportBatch(stream, items, format); err != nil {
			return err
		}

		if len(items) < batchSize {
			break
		}
		offset += batchSize
	}
	return nil
}

// sendExportBatch formats and sends a batch of data items to the export stream.
func sendExportBatch(stream *connect.ServerStream[portwhinev1.ExportDataItemsResponse], items []store.DataItemRecord, format string) error {
	for i := range items {
		row, err := formatDataItemRow(&items[i], format)
		if err != nil {
			slog.Warn("skipping data item during export",
				slog.String("item_id", items[i].ID),
				slog.String("format", format),
				slog.Any("error", err),
			)
			continue
		}
		if err := stream.Send(&portwhinev1.ExportDataItemsResponse{Data: row}); err != nil {
			return err
		}
	}
	return nil
}

// buildSearchParams constructs DataItemSearchParams from RPC fields.
func buildSearchParams(runID, query string, types []string, after, before *timestamppb.Timestamp) store.DataItemSearchParams {
	params := store.DataItemSearchParams{
		RunID: runID,
		Query: query,
		Types: types,
	}
	if after != nil {
		t := after.AsTime()
		params.CreatedAfter = &t
	}
	if before != nil {
		t := before.AsTime()
		params.CreatedBefore = &t
	}
	return params
}

// formatDataItemRow serializes a DataItemRecord as a JSON line or CSV row.
func formatDataItemRow(item *store.DataItemRecord, format string) ([]byte, error) {
	if format == "csv" {
		dataStr := strings.ReplaceAll(string(item.Data), "\"", "\"\"")
		line := fmt.Sprintf("%s,%s,%s,\"%s\",%s\n",
			item.ID, item.RunID, item.Type, dataStr, item.CreatedAt.Format(time.RFC3339))
		return []byte(line), nil
	}
	// JSON lines format.
	row := map[string]any{
		"id":         item.ID,
		"run_id":     item.RunID,
		"type":       item.Type,
		"data":       json.RawMessage(item.Data),
		"parent_ids": item.ParentIDs,
		"created_at": item.CreatedAt.Format(time.RFC3339),
	}
	if len(item.Metadata) > 0 {
		row["metadata"] = json.RawMessage(item.Metadata)
	}
	b, err := json.Marshal(row)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// ---------------------------------------------------------------------------
// Worker Image Registry
// ---------------------------------------------------------------------------

func (h *Handler) RegisterWorkerImage(
	ctx context.Context,
	req *connect.Request[portwhinev1.RegisterWorkerImageRequest],
) (*connect.Response[portwhinev1.RegisterWorkerImageResponse], error) {
	img := &store.WorkerImage{
		Name:        req.Msg.GetName(),
		Image:       req.Msg.GetImage(),
		Description: req.Msg.GetDescription(),
		IsActive:    true,
	}

	if err := h.store.WorkerImages.Create(ctx, img); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("register worker image: %w", err))
	}

	return connect.NewResponse(&portwhinev1.RegisterWorkerImageResponse{
		WorkerImageId: img.ID,
	}), nil
}

func (h *Handler) ListWorkerImages(
	ctx context.Context,
	_ *connect.Request[portwhinev1.ListWorkerImagesRequest],
) (*connect.Response[portwhinev1.ListWorkerImagesResponse], error) {
	images, err := h.store.WorkerImages.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list worker images: %w", err))
	}

	infos := make([]*portwhinev1.WorkerImageInfo, 0, len(images))
	for _, img := range images {
		infos = append(infos, &portwhinev1.WorkerImageInfo{
			Id:          img.ID,
			Name:        img.Name,
			Image:       img.Image,
			Description: img.Description,
			CreatedAt:   timestamppb.New(img.CreatedAt),
		})
	}

	return connect.NewResponse(&portwhinev1.ListWorkerImagesResponse{
		Images: infos,
	}), nil
}

func (h *Handler) DeleteWorkerImage(
	ctx context.Context,
	req *connect.Request[portwhinev1.DeleteWorkerImageRequest],
) (*connect.Response[portwhinev1.DeleteWorkerImageResponse], error) {
	if err := h.store.WorkerImages.Delete(ctx, req.Msg.GetWorkerImageId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete worker image: %w", err))
	}
	return connect.NewResponse(&portwhinev1.DeleteWorkerImageResponse{}), nil
}

// ---------------------------------------------------------------------------
// Auth RPCs
// ---------------------------------------------------------------------------

func (h *Handler) Login(
	ctx context.Context,
	req *connect.Request[portwhinev1.LoginRequest],
) (*connect.Response[portwhinev1.LoginResponse], error) {
	user, err := h.store.Users.GetByUsername(ctx, req.Msg.GetUsername())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("user account is disabled"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Msg.GetPassword())); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	// Load team IDs for JWT claims.
	teamIDs, err := h.getUserTeamIDs(ctx, user.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("load team ids: %w", err))
	}

	accessToken, refreshToken, expiresAt, err := h.jwt.GenerateTokenPair(user.ID, user.Username, user.Role.Name, teamIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate tokens: %w", err))
	}

	return connect.NewResponse(&portwhinev1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    timestamppb.New(expiresAt),
		UserId:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role.Name,
	}), nil
}

func (h *Handler) RefreshToken(
	ctx context.Context,
	req *connect.Request[portwhinev1.RefreshTokenRequest],
) (*connect.Response[portwhinev1.RefreshTokenResponse], error) {
	userID, err := h.jwt.ValidateRefreshToken(req.Msg.GetRefreshToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
	}

	user, err := h.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not found"))
	}

	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("user account is disabled"))
	}

	teamIDs, err := h.getUserTeamIDs(ctx, user.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("load team ids: %w", err))
	}

	accessToken, refreshToken, expiresAt, err := h.jwt.GenerateTokenPair(user.ID, user.Username, user.Role.Name, teamIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate tokens: %w", err))
	}

	return connect.NewResponse(&portwhinev1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    timestamppb.New(expiresAt),
	}), nil
}

func (h *Handler) CreateAPIKey(
	ctx context.Context,
	req *connect.Request[portwhinev1.CreateAPIKeyRequest],
) (*connect.Response[portwhinev1.CreateAPIKeyResponse], error) {
	rawKey, keyHash, prefix, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate api key: %w", err))
	}

	var expiresAt *time.Time
	if req.Msg.GetExpiresAt() != nil {
		t := req.Msg.GetExpiresAt().AsTime()
		expiresAt = &t
	}

	apiKey := &store.APIKey{
		UserID:    userIDFromContext(ctx),
		Name:      req.Msg.GetName(),
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Scopes:    req.Msg.GetScopes(),
		ExpiresAt: expiresAt,
	}

	if err := h.store.APIKeys.Create(ctx, apiKey); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create api key: %w", err))
	}

	return connect.NewResponse(&portwhinev1.CreateAPIKeyResponse{
		ApiKey:    rawKey,
		KeyPrefix: prefix,
	}), nil
}

func (h *Handler) ListAPIKeys(
	ctx context.Context,
	_ *connect.Request[portwhinev1.ListAPIKeysRequest],
) (*connect.Response[portwhinev1.ListAPIKeysResponse], error) {
	keys, err := h.store.APIKeys.ListByUser(ctx, userIDFromContext(ctx))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list api keys: %w", err))
	}

	infos := make([]*portwhinev1.APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		info := &portwhinev1.APIKeyInfo{
			Id:        k.ID,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			Scopes:    k.Scopes,
			CreatedAt: timestamppb.New(k.CreatedAt),
			Revoked:   k.RevokedAt != nil,
		}
		if k.ExpiresAt != nil {
			info.ExpiresAt = timestamppb.New(*k.ExpiresAt)
		}
		if k.LastUsed != nil {
			info.LastUsed = timestamppb.New(*k.LastUsed)
		}
		infos = append(infos, info)
	}

	return connect.NewResponse(&portwhinev1.ListAPIKeysResponse{
		Keys: infos,
	}), nil
}

func (h *Handler) RevokeAPIKey(
	ctx context.Context,
	req *connect.Request[portwhinev1.RevokeAPIKeyRequest],
) (*connect.Response[portwhinev1.RevokeAPIKeyResponse], error) {
	if err := h.store.APIKeys.Revoke(ctx, req.Msg.GetApiKeyId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("revoke api key: %w", err))
	}
	return connect.NewResponse(&portwhinev1.RevokeAPIKeyResponse{}), nil
}

// ---------------------------------------------------------------------------
// User Management
// ---------------------------------------------------------------------------

func (h *Handler) CreateUser(
	ctx context.Context,
	req *connect.Request[portwhinev1.CreateUserRequest],
) (*connect.Response[portwhinev1.CreateUserResponse], error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("hash password: %w", err))
	}

	roleName := req.Msg.GetRole()
	if roleName == "" {
		roleName = "user"
	}

	role, err := h.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid role %q: %w", roleName, err))
	}

	user := &store.User{
		Username:     req.Msg.GetUsername(),
		Email:        req.Msg.GetEmail(),
		PasswordHash: string(hash),
		RoleID:       role.ID,
		IsActive:     true,
	}

	if err := h.store.Users.Create(ctx, user); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create user: %w", err))
	}

	h.audit(ctx, "create", "user", &user.ID, map[string]any{"username": user.Username, "role": roleName})

	return connect.NewResponse(&portwhinev1.CreateUserResponse{
		UserId: user.ID,
	}), nil
}

func (h *Handler) GetUser(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetUserRequest],
) (*connect.Response[portwhinev1.GetUserResponse], error) {
	user, err := h.store.Users.GetByID(ctx, req.Msg.GetUserId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}

	return connect.NewResponse(&portwhinev1.GetUserResponse{
		User: userToProto(user),
	}), nil
}

func (h *Handler) ListUsers(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListUsersRequest],
) (*connect.Response[portwhinev1.ListUsersResponse], error) {
	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	users, total, err := h.store.Users.List(ctx, offset, pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list users: %w", err))
	}

	infos := make([]*portwhinev1.UserInfo, 0, len(users))
	for i := range users {
		infos = append(infos, userToProto(&users[i]))
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.ListUsersResponse{
		Users:         infos,
		NextPageToken: nextToken,
	}), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context,
	req *connect.Request[portwhinev1.UpdateUserRequest],
) (*connect.Response[portwhinev1.UpdateUserResponse], error) {
	user, err := h.store.Users.GetByID(ctx, req.Msg.GetUserId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}

	if email := req.Msg.GetEmail(); email != "" {
		user.Email = email
	}
	if roleName := req.Msg.GetRole(); roleName != "" {
		role, err := h.store.Roles.GetByName(ctx, roleName)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid role %q: %w", roleName, err))
		}
		user.RoleID = role.ID
	}
	user.IsActive = req.Msg.GetIsActive()

	if err := h.store.Users.Update(ctx, user); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update user: %w", err))
	}

	return connect.NewResponse(&portwhinev1.UpdateUserResponse{}), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context,
	req *connect.Request[portwhinev1.DeleteUserRequest],
) (*connect.Response[portwhinev1.DeleteUserResponse], error) {
	uid := req.Msg.GetUserId()
	if err := h.store.Users.Delete(ctx, uid); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete user: %w", err))
	}
	h.audit(ctx, "delete", "user", &uid, nil)
	return connect.NewResponse(&portwhinev1.DeleteUserResponse{}), nil
}

// ---------------------------------------------------------------------------
// Team Management
// ---------------------------------------------------------------------------

func (h *Handler) CreateTeam(
	ctx context.Context,
	req *connect.Request[portwhinev1.CreateTeamRequest],
) (*connect.Response[portwhinev1.CreateTeamResponse], error) {
	userID := userIDFromContext(ctx)

	team := &store.Team{
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		CreatedByID: userID,
	}

	if err := h.store.Teams.Create(ctx, team); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create team: %w", err))
	}

	// Add creator as team owner.
	member := &store.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   "owner",
	}
	if err := h.store.TeamMembers.Add(ctx, member); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("add team owner: %w", err))
	}

	// Add Casbin g2 mapping for the creator.
	if err := auth.AddUserTeamMapping(h.authorizer.Enforcer(), userID, team.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("add casbin team mapping: %w", err))
	}

	return connect.NewResponse(&portwhinev1.CreateTeamResponse{
		TeamId: team.ID,
	}), nil
}

func (h *Handler) GetTeam(
	ctx context.Context,
	req *connect.Request[portwhinev1.GetTeamRequest],
) (*connect.Response[portwhinev1.GetTeamResponse], error) {
	team, err := h.store.Teams.GetByID(ctx, req.Msg.GetTeamId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("team not found: %w", err))
	}

	members, err := h.store.TeamMembers.ListByTeam(ctx, team.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count members: %w", err))
	}

	return connect.NewResponse(&portwhinev1.GetTeamResponse{
		Team: teamToProto(team, len(members)),
	}), nil
}

func (h *Handler) ListTeams(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListTeamsRequest],
) (*connect.Response[portwhinev1.ListTeamsResponse], error) {
	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		parsed, err := strconv.Atoi(tok)
		if err == nil {
			offset = parsed
		}
	}

	claims := claimsFromContext(ctx)
	var teams []store.Team
	var total int64
	var err error

	if claims.Role == "admin" {
		teams, total, err = h.store.Teams.List(ctx, offset, pageSize)
	} else {
		// Non-admin: only teams the user is a member of.
		allTeams, listErr := h.store.Teams.ListByUser(ctx, claims.UserID)
		if listErr != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list teams: %w", listErr))
		}
		total = int64(len(allTeams))
		end := offset + pageSize
		if end > len(allTeams) {
			end = len(allTeams)
		}
		if offset < len(allTeams) {
			teams = allTeams[offset:end]
		}
		err = nil
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list teams: %w", err))
	}

	infos := make([]*portwhinev1.TeamInfo, 0, len(teams))
	for _, t := range teams {
		infos = append(infos, teamToProto(&t, 0))
	}

	var nextToken string
	if nextOff := offset + pageSize; int64(nextOff) < total {
		nextToken = strconv.Itoa(nextOff)
	}

	return connect.NewResponse(&portwhinev1.ListTeamsResponse{
		Teams:         infos,
		NextPageToken: nextToken,
	}), nil
}

func (h *Handler) UpdateTeam(
	ctx context.Context,
	req *connect.Request[portwhinev1.UpdateTeamRequest],
) (*connect.Response[portwhinev1.UpdateTeamResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.requireTeamAdmin(ctx, claims, req.Msg.GetTeamId()); err != nil {
		return nil, err
	}

	team, err := h.store.Teams.GetByID(ctx, req.Msg.GetTeamId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("team not found: %w", err))
	}

	if name := req.Msg.GetName(); name != "" {
		team.Name = name
	}
	if desc := req.Msg.GetDescription(); desc != "" {
		team.Description = desc
	}

	if err := h.store.Teams.Update(ctx, team); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update team: %w", err))
	}

	return connect.NewResponse(&portwhinev1.UpdateTeamResponse{}), nil
}

func (h *Handler) DeleteTeam(
	ctx context.Context,
	req *connect.Request[portwhinev1.DeleteTeamRequest],
) (*connect.Response[portwhinev1.DeleteTeamResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.requireTeamAdmin(ctx, claims, req.Msg.GetTeamId()); err != nil {
		return nil, err
	}

	if err := h.store.Teams.Delete(ctx, req.Msg.GetTeamId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete team: %w", err))
	}

	return connect.NewResponse(&portwhinev1.DeleteTeamResponse{}), nil
}

func (h *Handler) AddTeamMember(
	ctx context.Context,
	req *connect.Request[portwhinev1.AddTeamMemberRequest],
) (*connect.Response[portwhinev1.AddTeamMemberResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.requireTeamAdmin(ctx, claims, req.Msg.GetTeamId()); err != nil {
		return nil, err
	}

	role := req.Msg.GetRole()
	if role == "" {
		role = "member"
	}

	member := &store.TeamMember{
		TeamID: req.Msg.GetTeamId(),
		UserID: req.Msg.GetUserId(),
		Role:   role,
	}

	if err := h.store.TeamMembers.Add(ctx, member); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("add team member: %w", err))
	}

	// Add Casbin g2 mapping.
	if err := auth.AddUserTeamMapping(h.authorizer.Enforcer(), req.Msg.GetUserId(), req.Msg.GetTeamId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("add casbin team mapping: %w", err))
	}

	return connect.NewResponse(&portwhinev1.AddTeamMemberResponse{}), nil
}

func (h *Handler) RemoveTeamMember(
	ctx context.Context,
	req *connect.Request[portwhinev1.RemoveTeamMemberRequest],
) (*connect.Response[portwhinev1.RemoveTeamMemberResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.requireTeamAdmin(ctx, claims, req.Msg.GetTeamId()); err != nil {
		return nil, err
	}

	if err := h.store.TeamMembers.Remove(ctx, req.Msg.GetTeamId(), req.Msg.GetUserId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("remove team member: %w", err))
	}

	// Remove Casbin g2 mapping.
	if err := auth.RemoveUserTeamMapping(h.authorizer.Enforcer(), req.Msg.GetUserId(), req.Msg.GetTeamId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("remove casbin team mapping: %w", err))
	}

	return connect.NewResponse(&portwhinev1.RemoveTeamMemberResponse{}), nil
}

func (h *Handler) ListTeamMembers(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListTeamMembersRequest],
) (*connect.Response[portwhinev1.ListTeamMembersResponse], error) {
	members, err := h.store.TeamMembers.ListByTeam(ctx, req.Msg.GetTeamId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list team members: %w", err))
	}

	infos := make([]*portwhinev1.TeamMemberInfo, 0, len(members))
	for _, m := range members {
		infos = append(infos, &portwhinev1.TeamMemberInfo{
			UserId:   m.UserID,
			Username: m.User.Username,
			Email:    m.User.Email,
			Role:     m.Role,
			JoinedAt: timestamppb.New(m.CreatedAt),
		})
	}

	return connect.NewResponse(&portwhinev1.ListTeamMembersResponse{
		Members: infos,
	}), nil
}

func (h *Handler) UpdateTeamMemberRole(
	ctx context.Context,
	req *connect.Request[portwhinev1.UpdateTeamMemberRoleRequest],
) (*connect.Response[portwhinev1.UpdateTeamMemberRoleResponse], error) {
	claims := claimsFromContext(ctx)
	if err := h.requireTeamAdmin(ctx, claims, req.Msg.GetTeamId()); err != nil {
		return nil, err
	}

	if err := h.store.TeamMembers.UpdateRole(ctx, req.Msg.GetTeamId(), req.Msg.GetUserId(), req.Msg.GetRole()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update team member role: %w", err))
	}

	return connect.NewResponse(&portwhinev1.UpdateTeamMemberRoleResponse{}), nil
}

// ---------------------------------------------------------------------------
// Permission Management
// ---------------------------------------------------------------------------

func (h *Handler) GrantPermission(
	ctx context.Context,
	req *connect.Request[portwhinev1.GrantPermissionRequest],
) (*connect.Response[portwhinev1.GrantPermissionResponse], error) {
	effect := req.Msg.GetEffect()
	if effect == "" {
		effect = "allow"
	}
	if effect != "allow" && effect != "deny" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("effect must be 'allow' or 'deny'"))
	}

	subjectType := req.Msg.GetSubjectType()
	if subjectType != "user" && subjectType != "team" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("subject_type must be 'user' or 'team'"))
	}

	perm := &store.ResourcePermission{
		SubjectType:  subjectType,
		SubjectID:    req.Msg.GetSubjectId(),
		ResourceType: req.Msg.GetResourceType(),
		ResourceID:   req.Msg.GetResourceId(),
		Action:       req.Msg.GetAction(),
		Effect:       effect,
		GrantedByID:  userIDFromContext(ctx),
	}

	if err := h.store.ResourcePermissions.Grant(ctx, perm); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("grant permission: %w", err))
	}

	h.audit(ctx, "grant_permission", perm.ResourceType, &perm.ResourceID, map[string]any{
		"subject_type": perm.SubjectType, "subject_id": perm.SubjectID,
		"action": perm.Action, "effect": perm.Effect,
	})

	return connect.NewResponse(&portwhinev1.GrantPermissionResponse{
		PermissionId: perm.ID,
	}), nil
}

func (h *Handler) RevokePermission(
	ctx context.Context,
	req *connect.Request[portwhinev1.RevokePermissionRequest],
) (*connect.Response[portwhinev1.RevokePermissionResponse], error) {
	permID := req.Msg.GetPermissionId()
	if err := h.store.ResourcePermissions.Revoke(ctx, permID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("revoke permission: %w", err))
	}
	h.audit(ctx, "revoke_permission", "permission", &permID, nil)

	return connect.NewResponse(&portwhinev1.RevokePermissionResponse{}), nil
}

func (h *Handler) ListPermissions(
	ctx context.Context,
	req *connect.Request[portwhinev1.ListPermissionsRequest],
) (*connect.Response[portwhinev1.ListPermissionsResponse], error) {
	var perms []store.ResourcePermission
	var err error

	if req.Msg.GetResourceType() != "" && req.Msg.GetResourceId() != "" {
		perms, err = h.store.ResourcePermissions.ListForResource(ctx, req.Msg.GetResourceType(), req.Msg.GetResourceId())
	} else if req.Msg.GetSubjectType() != "" && req.Msg.GetSubjectId() != "" {
		perms, err = h.store.ResourcePermissions.ListForSubject(ctx, req.Msg.GetSubjectType(), req.Msg.GetSubjectId())
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("provide either subject or resource filter"))
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list permissions: %w", err))
	}

	return connect.NewResponse(&portwhinev1.ListPermissionsResponse{
		Permissions: permissionsToProto(perms),
	}), nil
}

func (h *Handler) ListMyPermissions(
	ctx context.Context,
	_ *connect.Request[portwhinev1.ListMyPermissionsRequest],
) (*connect.Response[portwhinev1.ListMyPermissionsResponse], error) {
	userID := userIDFromContext(ctx)

	// Get direct user permissions.
	userPerms, err := h.store.ResourcePermissions.ListForSubject(ctx, "user", userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list user permissions: %w", err))
	}

	// Get team permissions.
	teamIDs, err := h.getUserTeamIDs(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get user teams: %w", err))
	}

	var allPerms []store.ResourcePermission
	allPerms = append(allPerms, userPerms...)
	for _, teamID := range teamIDs {
		teamPerms, err := h.store.ResourcePermissions.ListForSubject(ctx, "team", teamID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list team permissions: %w", err))
		}
		allPerms = append(allPerms, teamPerms...)
	}

	return connect.NewResponse(&portwhinev1.ListMyPermissionsResponse{
		Permissions: permissionsToProto(allPerms),
	}), nil
}

// ---------------------------------------------------------------------------
// Role Management
// ---------------------------------------------------------------------------

func (h *Handler) CreateRole(
	ctx context.Context,
	req *connect.Request[portwhinev1.CreateRoleRequest],
) (*connect.Response[portwhinev1.CreateRoleResponse], error) {
	role := &store.Role{
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		IsCustom:    true,
	}

	// Use raw GORM via user repo's DB (roles don't have a Create in the interface yet).
	// Add a Create to RoleRepository or use the existing store.
	if err := h.store.Roles.Create(ctx, role); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create role: %w", err))
	}

	return connect.NewResponse(&portwhinev1.CreateRoleResponse{
		RoleId: role.ID,
	}), nil
}

func (h *Handler) ListRoles(
	ctx context.Context,
	_ *connect.Request[portwhinev1.ListRolesRequest],
) (*connect.Response[portwhinev1.ListRolesResponse], error) {
	roles, err := h.store.Roles.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list roles: %w", err))
	}

	infos := make([]*portwhinev1.RoleInfo, 0, len(roles))
	for _, r := range roles {
		infos = append(infos, &portwhinev1.RoleInfo{
			Id:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
			IsCustom:    r.IsCustom,
			CreatedAt:   timestamppb.New(r.CreatedAt),
		})
	}

	return connect.NewResponse(&portwhinev1.ListRolesResponse{
		Roles: infos,
	}), nil
}

func (h *Handler) UpdateRole(
	ctx context.Context,
	req *connect.Request[portwhinev1.UpdateRoleRequest],
) (*connect.Response[portwhinev1.UpdateRoleResponse], error) {
	role, err := h.store.Roles.GetByID(ctx, req.Msg.GetRoleId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("role not found: %w", err))
	}

	if role.IsSystem {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot modify system roles"))
	}

	if name := req.Msg.GetName(); name != "" {
		role.Name = name
	}
	if desc := req.Msg.GetDescription(); desc != "" {
		role.Description = desc
	}

	if err := h.store.Roles.Update(ctx, role); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update role: %w", err))
	}

	return connect.NewResponse(&portwhinev1.UpdateRoleResponse{}), nil
}

func (h *Handler) DeleteRole(
	ctx context.Context,
	req *connect.Request[portwhinev1.DeleteRoleRequest],
) (*connect.Response[portwhinev1.DeleteRoleResponse], error) {
	role, err := h.store.Roles.GetByID(ctx, req.Msg.GetRoleId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("role not found: %w", err))
	}

	if role.IsSystem {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot delete system roles"))
	}

	if err := h.store.Roles.Delete(ctx, role.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete role: %w", err))
	}

	return connect.NewResponse(&portwhinev1.DeleteRoleResponse{}), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func unmarshalPipelineDefinition(data datatypes.JSON) (*portwhinev1.PipelineDefinition, error) {
	def := &portwhinev1.PipelineDefinition{}
	if err := protojson.Unmarshal(data, def); err != nil {
		return nil, err
	}
	return def, nil
}

func pipelineRunToStatus(run *store.PipelineRun) *portwhinev1.PipelineRunStatus {
	s := &portwhinev1.PipelineRunStatus{
		RunId:        run.ID,
		PipelineId:   run.PipelineID,
		PipelineName: run.Pipeline.Name,
		State:        mapRunState(run.Status),
		CreatedBy:    run.CreatedByID,
	}
	if run.StartedAt != nil {
		s.StartedAt = timestamppb.New(*run.StartedAt)
	}
	if run.FinishedAt != nil {
		s.FinishedAt = timestamppb.New(*run.FinishedAt)
	}
	return s
}

func mapWorkerStatus(status string) portwhinev1.WorkerStatus {
	switch status {
	case "running":
		return portwhinev1.WorkerStatus_WORKER_STATUS_PROCESSING
	case "completed":
		return portwhinev1.WorkerStatus_WORKER_STATUS_STOPPED
	case "failed":
		return portwhinev1.WorkerStatus_WORKER_STATUS_ERROR
	case "pending":
		return portwhinev1.WorkerStatus_WORKER_STATUS_INITIALIZING
	case "draining":
		return portwhinev1.WorkerStatus_WORKER_STATUS_DRAINING
	case "ready":
		return portwhinev1.WorkerStatus_WORKER_STATUS_READY
	default:
		return portwhinev1.WorkerStatus_WORKER_STATUS_UNSPECIFIED
	}
}

func mapRunState(status string) portwhinev1.PipelineRunState {
	switch status {
	case "pending":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_PENDING
	case "running":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_RUNNING
	case "completed":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_COMPLETED
	case "failed":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_FAILED
	case "cancelled":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_CANCELLED
	case "paused":
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_PAUSED
	default:
		return portwhinev1.PipelineRunState_PIPELINE_RUN_STATE_UNSPECIFIED
	}
}

func userIDFromContext(ctx context.Context) string {
	claims := auth.ClaimsFromContext(ctx)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

func claimsFromContext(ctx context.Context) *auth.Claims {
	claims := auth.ClaimsFromContext(ctx)
	if claims == nil {
		return &auth.Claims{}
	}
	return claims
}

func userToProto(u *store.User) *portwhinev1.UserInfo {
	return &portwhinev1.UserInfo{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role.Name,
		IsActive:  u.IsActive,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}

func teamToProto(t *store.Team, memberCount int) *portwhinev1.TeamInfo {
	return &portwhinev1.TeamInfo{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		CreatedBy:   t.CreatedByID,
		MemberCount: int32(memberCount),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func permissionsToProto(perms []store.ResourcePermission) []*portwhinev1.PermissionInfo {
	infos := make([]*portwhinev1.PermissionInfo, 0, len(perms))
	for _, p := range perms {
		infos = append(infos, &portwhinev1.PermissionInfo{
			Id:           p.ID,
			SubjectType:  p.SubjectType,
			SubjectId:    p.SubjectID,
			ResourceType: p.ResourceType,
			ResourceId:   p.ResourceID,
			Action:       p.Action,
			Effect:       p.Effect,
			GrantedBy:    p.GrantedByID,
			CreatedAt:    timestamppb.New(p.CreatedAt),
		})
	}
	return infos
}

func dataItemRecordToProto(item *store.DataItemRecord) *portwhinev1.DataItemInfo {
	info := &portwhinev1.DataItemInfo{
		Id:        item.ID,
		RunId:     item.RunID,
		Type:      item.Type,
		ParentIds: item.ParentIDs,
		CreatedAt: timestamppb.New(item.CreatedAt),
	}

	// Parse Data JSONB into protobuf Struct.
	if len(item.Data) > 0 {
		s := &structpb.Struct{}
		if protojson.Unmarshal(item.Data, s) == nil {
			info.Data = s
		}
	}

	// Parse Metadata JSONB into map.
	if len(item.Metadata) > 0 {
		var meta map[string]string
		if json.Unmarshal(item.Metadata, &meta) == nil {
			info.Metadata = meta
		}
	}

	return info
}

// requireTeamAdmin checks that the caller is a team owner/admin or system admin.
func (h *Handler) requireTeamAdmin(ctx context.Context, claims *auth.Claims, teamID string) error {
	if claims.Role == "admin" {
		return nil
	}

	membership, err := h.store.TeamMembers.GetMembership(ctx, teamID, claims.UserID)
	if err != nil {
		return connect.NewError(connect.CodePermissionDenied, errors.New("not a team member"))
	}
	if membership.Role != "owner" && membership.Role != "admin" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("requires team owner or admin role"))
	}
	return nil
}

// ──────────────────────── Node Catalog ──────────────────────────

// ListNodeCatalog returns the static catalog of all available node types.
func (h *Handler) ListNodeCatalog(
	ctx context.Context,
	_ *connect.Request[portwhinev1.ListNodeCatalogRequest],
) (*connect.Response[portwhinev1.ListNodeCatalogResponse], error) {
	return connect.NewResponse(&portwhinev1.ListNodeCatalogResponse{
		Entries: getNodeCatalog(),
	}), nil
}

// ──────────────────────── Dashboard ──────────────────────────

func (h *Handler) GetDashboardStats(
	ctx context.Context,
	_ *connect.Request[portwhinev1.GetDashboardStatsRequest],
) (*connect.Response[portwhinev1.GetDashboardStatsResponse], error) {
	totalPipelines, err := h.store.Pipelines.CountAll(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count pipelines: %w", err))
	}

	totalRuns, err := h.store.PipelineRuns.CountAll(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count runs: %w", err))
	}

	runningRuns, err := h.store.PipelineRuns.CountByStatus(ctx, "running")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count running: %w", err))
	}

	completedRuns, err := h.store.PipelineRuns.CountByStatus(ctx, "completed")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count completed: %w", err))
	}

	failedRuns, err := h.store.PipelineRuns.CountByStatus(ctx, "failed")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count failed: %w", err))
	}

	recentDBRuns, err := h.store.PipelineRuns.ListRecent(ctx, 10)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list recent runs: %w", err))
	}

	recentRuns := make([]*portwhinev1.RecentRun, 0, len(recentDBRuns))
	for _, r := range recentDBRuns {
		rr := &portwhinev1.RecentRun{
			RunId:        r.ID,
			PipelineId:   r.PipelineID,
			PipelineName: r.Pipeline.Name,
			State:        mapRunState(r.Status),
		}
		if r.StartedAt != nil {
			rr.StartedAt = timestamppb.New(*r.StartedAt)
		}
		if r.FinishedAt != nil {
			rr.FinishedAt = timestamppb.New(*r.FinishedAt)
		}
		recentRuns = append(recentRuns, rr)
	}

	return connect.NewResponse(&portwhinev1.GetDashboardStatsResponse{
		TotalPipelines: int32(totalPipelines),
		TotalRuns:      int32(totalRuns),
		RunningRuns:    int32(runningRuns),
		CompletedRuns:  int32(completedRuns),
		FailedRuns:     int32(failedRuns),
		RecentRuns:     recentRuns,
	}), nil
}

// getUserTeamIDs returns all team IDs the user belongs to.
func (h *Handler) getUserTeamIDs(ctx context.Context, userID string) ([]string, error) {
	members, err := h.store.TeamMembers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.TeamID
	}
	return ids, nil
}
