package pipeline

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/cel-go/cel"
	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
)

// channelCapacity is the bounded buffer size for inter-stage channels.
// This provides backpressure: if a consumer falls behind by more than
// this many items, the producer blocks until space is available.
const channelCapacity = 10000

// Router connects pipeline stages via Go channels, supporting fan-out
// (one source to multiple targets) and fan-in (multiple sources to one target).
// It also applies InputFilter-based type matching and CEL condition evaluation
// before forwarding items.
type Router struct {
	mu sync.RWMutex

	// routes maps a source node ID to the set of target node IDs it feeds.
	routes map[string][]string

	// channels maps a target node ID to its input channel.
	// Multiple sources can write to the same channel (fan-in).
	channels map[string]chan *portwhinev1.DataItem

	// filters maps a target node ID to its InputFilter.
	// If nil, all items are accepted.
	filters map[string]*portwhinev1.InputFilter

	// writers tracks how many active writers exist per channel.
	// When the count drops to zero, the channel is closed.
	writers map[string]int

	// celPrograms caches compiled CEL programs per target node ID.
	celPrograms map[string]cel.Program
}

// NewRouter creates a Router for the given pipeline graph.
// It pre-creates channels for every node and registers all edge routes.
func NewRouter(graph *PipelineGraph) *Router {
	r := &Router{
		routes:      make(map[string][]string),
		channels:    make(map[string]chan *portwhinev1.DataItem),
		filters:     make(map[string]*portwhinev1.InputFilter),
		writers:     make(map[string]int),
		celPrograms: make(map[string]cel.Program),
	}

	// Create a channel for every node.
	for _, id := range graph.NodeIDs() {
		r.channels[id] = make(chan *portwhinev1.DataItem, channelCapacity)

		node := graph.GetNode(id)
		if node.GetInputFilter() != nil {
			r.filters[id] = node.GetInputFilter()

			// Pre-compile CEL condition if present.
			if cond := node.GetInputFilter().GetCondition(); cond != "" {
				prog, err := compileCEL(cond)
				if err != nil {
					slog.Warn("failed to compile CEL condition, filter will be ignored",
						slog.String("node_id", id),
						slog.String("condition", cond),
						slog.Any("error", err),
					)
				} else {
					r.celPrograms[id] = prog
				}
			}
		}
	}

	// Register routes from edges and count writers per target.
	for _, id := range graph.NodeIDs() {
		downstream := graph.GetDownstream(id)
		if len(downstream) > 0 {
			r.routes[id] = downstream
			for _, tgt := range downstream {
				r.writers[tgt]++
			}
		}
	}

	return r
}

// AddRoute registers an additional route from fromNodeID to toNodeID.
// If the target channel does not yet exist, one is created.
func (r *Router) AddRoute(fromNodeID, toNodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.channels[toNodeID]; !ok {
		r.channels[toNodeID] = make(chan *portwhinev1.DataItem, channelCapacity)
	}

	r.routes[fromNodeID] = append(r.routes[fromNodeID], toNodeID)
	r.writers[toNodeID]++
}

// Route sends a DataItem from the source node to all downstream targets.
// Items are only forwarded to a target if they pass the target's InputFilter.
// This method blocks if any downstream channel is full (backpressure).
func (r *Router) Route(fromNodeID string, item *portwhinev1.DataItem) {
	r.mu.RLock()
	targets := r.routes[fromNodeID]
	r.mu.RUnlock()

	for _, tgt := range targets {
		if r.matchesFilter(tgt, item) {
			r.mu.RLock()
			ch := r.channels[tgt]
			r.mu.RUnlock()

			if ch != nil {
				ch <- item
			}
		}
	}
}

// GetInputChannel returns the input channel for the given node.
// Stages read from this channel to receive their input DataItems.
func (r *Router) GetInputChannel(nodeID string) <-chan *portwhinev1.DataItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channels[nodeID]
}

// GetOutputChannel returns the output channel for the given node.
// Stages write to this channel; the Router's Route method is the preferred
// way to send items, but this is useful for trigger nodes that produce
// items without an upstream source.
func (r *Router) GetOutputChannel(nodeID string) chan<- *portwhinev1.DataItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channels[nodeID]
}

// CloseWriter signals that one writer for the target node has finished.
// When all writers for a channel have finished, the channel is closed,
// signaling to downstream readers that no more data will arrive.
func (r *Router) CloseWriter(targetNodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.writers[targetNodeID]--
	if r.writers[targetNodeID] <= 0 {
		if ch, ok := r.channels[targetNodeID]; ok {
			close(ch)
			// Remove channel from map to prevent double-close.
			delete(r.channels, targetNodeID)
		}
	}
}

// SignalSourceDone is called when a source node has finished producing items.
// It decrements the writer count for all downstream targets of that source.
func (r *Router) SignalSourceDone(sourceNodeID string) {
	r.mu.RLock()
	targets := r.routes[sourceNodeID]
	r.mu.RUnlock()

	for _, tgt := range targets {
		r.CloseWriter(tgt)
	}
}

// matchesFilter checks whether a DataItem passes the InputFilter for a target
// node. If no filter is set, the item always matches.
func (r *Router) matchesFilter(targetNodeID string, item *portwhinev1.DataItem) bool {
	r.mu.RLock()
	filter := r.filters[targetNodeID]
	prog := r.celPrograms[targetNodeID]
	r.mu.RUnlock()

	if filter == nil {
		return true
	}

	// Type filter: only accept items whose type matches.
	if filter.GetType() != "" && item.GetType() != filter.GetType() {
		return false
	}

	// CEL condition evaluation.
	if prog != nil {
		return evalCEL(prog, item)
	}

	return true
}

// compileCEL compiles a CEL expression string into a reusable Program.
// The expression has access to: type (string), data (map), metadata (map).
func compileCEL(expr string) (cel.Program, error) {
	env, err := cel.NewEnv(
		cel.Variable("type", cel.StringType),
		cel.Variable("data", cel.DynType),
	)
	if err != nil {
		return nil, fmt.Errorf("create CEL env: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile CEL: %w", issues.Err())
	}

	prog, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program CEL: %w", err)
	}

	return prog, nil
}

// evalCEL evaluates a compiled CEL program against a DataItem.
// Returns true if the expression evaluates to true, false otherwise
// (including on error).
func evalCEL(prog cel.Program, item *portwhinev1.DataItem) bool {
	vars := map[string]any{
		"type": item.GetType(),
		"data": map[string]any{},
	}

	if item.GetData() != nil {
		vars["data"] = item.GetData().AsMap()
	}

	out, _, err := prog.Eval(vars)
	if err != nil {
		return false
	}

	result, ok := out.Value().(bool)
	return ok && result
}

// String returns a human-readable summary of all routes for debugging.
func (r *Router) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := "Router routes:\n"
	for src, targets := range r.routes {
		for _, tgt := range targets {
			result += fmt.Sprintf("  %s -> %s\n", src, tgt)
		}
	}
	return result
}
