// Package pipeline implements the pipeline execution engine for Portwhine.
// It handles DAG validation, graph construction, data routing between stages,
// and orchestrating container-based pipeline runs.
package pipeline

import (
	"errors"
	"fmt"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
)

// Validation errors returned by ValidatePipeline.
var (
	ErrNilDefinition     = errors.New("pipeline definition is nil")
	ErrNoNodes           = errors.New("pipeline has no nodes")
	ErrNoTriggerNode     = errors.New("pipeline must have at least one trigger node")
	ErrCycleDetected     = errors.New("pipeline contains a cycle")
	ErrOrphanNode        = errors.New("pipeline contains orphan nodes")
	ErrInvalidEdgeSource = errors.New("edge references a non-existent source node")
	ErrInvalidEdgeTarget = errors.New("edge references a non-existent target node")
	ErrEmptyNodeID       = errors.New("node has an empty ID")
	ErrDuplicateNodeID   = errors.New("duplicate node ID")
	ErrEmptyEdgeID       = errors.New("edge has an empty ID")
	ErrDuplicateEdgeID   = errors.New("duplicate edge ID")
	ErrSelfLoop          = errors.New("edge creates a self-loop")
	ErrNodeMissingImage  = errors.New("trigger/worker node is missing a container image")
	ErrUnspecifiedType   = errors.New("node has unspecified type")
)

// ValidatePipeline checks a PipelineDefinition for structural correctness.
// It verifies that:
//   - The definition is non-nil and contains nodes
//   - At least one trigger node exists
//   - All node IDs are non-empty and unique
//   - All edge IDs are non-empty and unique
//   - All edge source/target references point to existing nodes
//   - No self-loops exist
//   - The graph is acyclic (via topological sort)
//   - No orphan nodes exist (every non-trigger node is reachable from a trigger)
func ValidatePipeline(def *portwhinev1.PipelineDefinition) error {
	if def == nil {
		return ErrNilDefinition
	}

	nodes := def.GetNodes()
	edges := def.GetEdges()

	if len(nodes) == 0 {
		return ErrNoNodes
	}

	// Build a set of node IDs and validate each node.
	nodeIDs := make(map[string]*portwhinev1.PipelineNode, len(nodes))
	for _, node := range nodes {
		if node.GetId() == "" {
			return ErrEmptyNodeID
		}
		if _, exists := nodeIDs[node.GetId()]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateNodeID, node.GetId())
		}
		if node.GetType() == portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_UNSPECIFIED {
			return fmt.Errorf("%w: node %q", ErrUnspecifiedType, node.GetId())
		}
		if node.GetType() == portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_TRIGGER ||
			node.GetType() == portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_WORKER {
			if node.GetImage() == "" {
				return fmt.Errorf("%w: node %q", ErrNodeMissingImage, node.GetId())
			}
		}
		nodeIDs[node.GetId()] = node
	}

	// Ensure at least one trigger node.
	hasTrigger := false
	for _, node := range nodes {
		if node.GetType() == portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_TRIGGER {
			hasTrigger = true
			break
		}
	}
	if !hasTrigger {
		return ErrNoTriggerNode
	}

	// Validate edges.
	edgeIDs := make(map[string]struct{}, len(edges))
	for _, edge := range edges {
		if edge.GetId() == "" {
			return ErrEmptyEdgeID
		}
		if _, exists := edgeIDs[edge.GetId()]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateEdgeID, edge.GetId())
		}
		edgeIDs[edge.GetId()] = struct{}{}

		if edge.GetSourceNodeId() == edge.GetTargetNodeId() {
			return fmt.Errorf("%w: edge %q (%s -> %s)",
				ErrSelfLoop, edge.GetId(), edge.GetSourceNodeId(), edge.GetTargetNodeId())
		}
		if _, ok := nodeIDs[edge.GetSourceNodeId()]; !ok {
			return fmt.Errorf("%w: edge %q references source %q",
				ErrInvalidEdgeSource, edge.GetId(), edge.GetSourceNodeId())
		}
		if _, ok := nodeIDs[edge.GetTargetNodeId()]; !ok {
			return fmt.Errorf("%w: edge %q references target %q",
				ErrInvalidEdgeTarget, edge.GetId(), edge.GetTargetNodeId())
		}
	}

	// Cycle detection via Kahn's algorithm (topological sort).
	inDegree := make(map[string]int, len(nodes))
	adjacency := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		inDegree[node.GetId()] = 0
	}
	for _, edge := range edges {
		adjacency[edge.GetSourceNodeId()] = append(adjacency[edge.GetSourceNodeId()], edge.GetTargetNodeId())
		inDegree[edge.GetTargetNodeId()]++
	}

	queue := make([]string, 0, len(nodes))
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visited++

		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visited != len(nodes) {
		return ErrCycleDetected
	}

	// Orphan detection: every non-trigger node must be reachable via edges.
	// A node is an orphan if it has no incoming edges and is not a trigger.
	// Additionally, build reachability from trigger nodes to detect isolated subgraphs.
	reachable := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.GetType() == portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_TRIGGER {
			reachable[node.GetId()] = true
		}
	}

	// BFS from all trigger nodes.
	bfsQueue := make([]string, 0, len(nodes))
	for id := range reachable {
		bfsQueue = append(bfsQueue, id)
	}
	for len(bfsQueue) > 0 {
		current := bfsQueue[0]
		bfsQueue = bfsQueue[1:]
		for _, neighbor := range adjacency[current] {
			if !reachable[neighbor] {
				reachable[neighbor] = true
				bfsQueue = append(bfsQueue, neighbor)
			}
		}
	}

	for _, node := range nodes {
		if !reachable[node.GetId()] {
			return fmt.Errorf("%w: node %q is not reachable from any trigger node",
				ErrOrphanNode, node.GetId())
		}
	}

	return nil
}
