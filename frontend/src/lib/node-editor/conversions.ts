import { Node, Edge } from 'reactflow'
import {
  PipelineDefinition,
  PipelineDefinitionSchema,
  PipelineNodeSchema,
  PipelineEdgeSchema,
  PipelineNodeType,
} from '@/gen/portwhine/v1/pipeline_pb'
import type { NodeCatalogEntry } from '@/gen/portwhine/v1/operator_pb'
import type { NodeStatus } from '@/gen/portwhine/v1/operator_pb'
import { create } from '@bufbuild/protobuf'
import type { JsonObject } from '@bufbuild/protobuf'

export type ReactFlowNodeData = {
  id: string
  label: string
  image: string
  type: PipelineNodeType
  config?: JsonObject
  replicas?: number
  inputFilter?: {
    type?: string
    condition?: string
  }
  retryPolicy?: {
    maxRetries: number
    initialBackoffSeconds: number
    maxBackoffSeconds: number
  }
  // Catalog-derived display fields (not persisted in pipeline definition)
  catalogId?: string
  acceptedInputTypes?: string[]
  outputTypes?: string[]
  configSchema?: string
  color?: string
  icon?: string
  // Live run status (set when a pipeline is running)
  nodeStatus?: NodeStatus
}

/**
 * Enriches ReactFlow nodes with catalog metadata (color, icon, input/output types, etc.)
 * by matching each node's image field to a catalog entry.
 */
export function enrichNodesWithCatalog(
  nodes: Node<ReactFlowNodeData>[],
  catalog: NodeCatalogEntry[]
): Node<ReactFlowNodeData>[] {
  return nodes.map((node) => {
    const entry = catalog.find((e) => e.image === node.data.image)
    if (!entry) return node
    return {
      ...node,
      data: {
        ...node.data,
        catalogId: entry.id,
        acceptedInputTypes: entry.acceptedInputTypes,
        outputTypes: entry.outputTypes,
        configSchema: entry.configSchema,
        color: entry.color,
        icon: entry.icon,
      },
    }
  })
}

export function protoToReactFlow(definition: PipelineDefinition): {
  nodes: Node<ReactFlowNodeData>[]
  edges: Edge[]
} {
  const nodes: Node<ReactFlowNodeData>[] = (definition.nodes || []).map((node: any) => {
    const nodeType =
      node.type === PipelineNodeType.TRIGGER
        ? 'trigger'
        : node.type === PipelineNodeType.WORKER
        ? 'worker'
        : 'output'

    return {
      id: node.id,
      type: nodeType,
      position: node.position
        ? { x: node.position.x, y: node.position.y }
        : { x: 0, y: 0 },
      data: {
        id: node.id,
        label: node.label || node.id,
        image: node.image || '',
        type: node.type,
        config: node.config || {},
        replicas: node.replicas || 1,
        inputFilter: node.inputFilter
          ? {
              type: node.inputFilter.type || undefined,
              condition: node.inputFilter.condition || undefined,
            }
          : undefined,
        retryPolicy: node.retryPolicy
          ? {
              maxRetries: node.retryPolicy.maxRetries || 0,
              initialBackoffSeconds: node.retryPolicy.initialBackoffSeconds || 1,
              maxBackoffSeconds: node.retryPolicy.maxBackoffSeconds || 60,
            }
          : undefined,
      },
    }
  })

  const edges: Edge[] = (definition.edges || []).map((edge: any) => ({
    id: edge.id || `${edge.sourceNodeId}-${edge.targetNodeId}`,
    source: edge.sourceNodeId,
    target: edge.targetNodeId,
    sourceHandle: edge.sourceHandle || undefined,
    targetHandle: edge.targetHandle || undefined,
    type: 'smoothstep',
  }))

  return { nodes, edges }
}

export function reactFlowToProto(
  nodes: Node<ReactFlowNodeData>[],
  edges: Edge[]
): PipelineDefinition {
  const pipelineNodes: any[] = nodes.map((node) => {
    const nodeData: any = {
      id: node.id,
      type: node.data.type,
      label: node.data.label,
      image: node.data.image,
      position: {
        x: node.position.x,
        y: node.position.y,
      },
      config: node.data.config || {},
      replicas: node.data.replicas || 1,
    }

    if (node.data.inputFilter) {
      nodeData.inputFilter = {
        type: node.data.inputFilter.type || '',
        condition: node.data.inputFilter.condition || '',
      }
    }

    if (node.data.retryPolicy) {
      nodeData.retryPolicy = {
        maxRetries: node.data.retryPolicy.maxRetries,
        initialBackoffSeconds: node.data.retryPolicy.initialBackoffSeconds,
        maxBackoffSeconds: node.data.retryPolicy.maxBackoffSeconds,
      }
    }

    return create(PipelineNodeSchema, nodeData)
  })

  const pipelineEdges: any[] = edges.map((edge) =>
    create(PipelineEdgeSchema, {
      id: edge.id,
      sourceNodeId: edge.source,
      targetNodeId: edge.target,
      sourceHandle: edge.sourceHandle || '',
      targetHandle: edge.targetHandle || '',
    })
  )

  return create(PipelineDefinitionSchema, {
    nodes: pipelineNodes,
    edges: pipelineEdges,
  })
}
