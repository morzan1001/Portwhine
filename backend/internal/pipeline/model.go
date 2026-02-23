package pipeline

import (
	"fmt"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
)

// PipelineGraph is an in-memory, indexed representation of a PipelineDefinition.
// It provides efficient lookups for nodes, edges, and graph traversal operations.
type PipelineGraph struct {
	// Definition is the original protobuf definition.
	Definition *portwhinev1.PipelineDefinition

	// nodes maps node IDs to their protobuf definitions.
	nodes map[string]*portwhinev1.PipelineNode

	// edges is the full list of edges.
	edges []*portwhinev1.PipelineEdge

	// downstream maps a source node ID to its outgoing edge targets.
	downstream map[string][]string

	// upstream maps a target node ID to its incoming edge sources.
	upstream map[string][]string

	// nodeOrder holds all node IDs in the order they appear in the definition.
	nodeOrder []string
}

// FromProto builds a PipelineGraph from a protobuf PipelineDefinition.
// It validates the definition first, then constructs all internal indices.
func FromProto(def *portwhinev1.PipelineDefinition) (*PipelineGraph, error) {
	if err := ValidatePipeline(def); err != nil {
		return nil, fmt.Errorf("invalid pipeline definition: %w", err)
	}

	g := &PipelineGraph{
		Definition: def,
		nodes:      make(map[string]*portwhinev1.PipelineNode, len(def.GetNodes())),
		edges:      def.GetEdges(),
		downstream: make(map[string][]string),
		upstream:   make(map[string][]string),
		nodeOrder:  make([]string, 0, len(def.GetNodes())),
	}

	for _, node := range def.GetNodes() {
		g.nodes[node.GetId()] = node
		g.nodeOrder = append(g.nodeOrder, node.GetId())
	}

	for _, edge := range def.GetEdges() {
		src := edge.GetSourceNodeId()
		tgt := edge.GetTargetNodeId()
		g.downstream[src] = append(g.downstream[src], tgt)
		g.upstream[tgt] = append(g.upstream[tgt], src)
	}

	return g, nil
}

// GetNode returns the PipelineNode with the given ID, or nil if not found.
func (g *PipelineGraph) GetNode(id string) *portwhinev1.PipelineNode {
	return g.nodes[id]
}

// GetDownstream returns the node IDs directly downstream of the given node.
func (g *PipelineGraph) GetDownstream(nodeID string) []string {
	return g.downstream[nodeID]
}

// GetUpstream returns the node IDs directly upstream of the given node.
func (g *PipelineGraph) GetUpstream(nodeID string) []string {
	return g.upstream[nodeID]
}

// TopologicalSort returns all node IDs in a valid topological order using
// Kahn's algorithm. Because the graph was validated at construction time,
// this always succeeds.
func (g *PipelineGraph) TopologicalSort() []string {
	inDegree := make(map[string]int, len(g.nodes))
	for id := range g.nodes {
		inDegree[id] = 0
	}
	for _, edge := range g.edges {
		inDegree[edge.GetTargetNodeId()]++
	}

	// Seed the queue with zero-in-degree nodes, preserving definition order
	// for deterministic output.
	queue := make([]string, 0, len(g.nodes))
	for _, id := range g.nodeOrder {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	sorted := make([]string, 0, len(g.nodes))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, neighbor := range g.downstream[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return sorted
}

// GetTriggerNodes returns all nodes with type PIPELINE_NODE_TYPE_TRIGGER.
func (g *PipelineGraph) GetTriggerNodes() []*portwhinev1.PipelineNode {
	return g.getNodesByType(portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_TRIGGER)
}

// GetWorkerNodes returns all nodes with type PIPELINE_NODE_TYPE_WORKER.
func (g *PipelineGraph) GetWorkerNodes() []*portwhinev1.PipelineNode {
	return g.getNodesByType(portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_WORKER)
}

// GetOutputNodes returns all nodes with type PIPELINE_NODE_TYPE_OUTPUT.
func (g *PipelineGraph) GetOutputNodes() []*portwhinev1.PipelineNode {
	return g.getNodesByType(portwhinev1.PipelineNodeType_PIPELINE_NODE_TYPE_OUTPUT)
}

// NodeIDs returns all node IDs in definition order.
func (g *PipelineGraph) NodeIDs() []string {
	result := make([]string, len(g.nodeOrder))
	copy(result, g.nodeOrder)
	return result
}

// getNodesByType is a helper that filters nodes by their PipelineNodeType.
func (g *PipelineGraph) getNodesByType(t portwhinev1.PipelineNodeType) []*portwhinev1.PipelineNode {
	var result []*portwhinev1.PipelineNode
	for _, id := range g.nodeOrder {
		node := g.nodes[id]
		if node.GetType() == t {
			result = append(result, node)
		}
	}
	return result
}
