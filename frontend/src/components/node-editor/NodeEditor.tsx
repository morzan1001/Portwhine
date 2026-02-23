import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from 'react'
import ReactFlow, {
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
  BackgroundVariant,
  NodeTypes,
  NodeMouseHandler,
  ReactFlowInstance,
} from 'reactflow'
import 'reactflow/dist/style.css'
import { toast } from 'sonner'

import { TriggerNode } from './nodes/TriggerNode'
import { WorkerNode } from './nodes/WorkerNode'
import { OutputNode } from './nodes/OutputNode'
import { NodeConfigDialog } from './NodeConfigDialog'
import { NodeCatalogSidebar } from './NodeCatalogSidebar'
import { PipelineNodeType } from '@/gen/portwhine/v1/pipeline_pb'
import {
  type ReactFlowNodeData,
  enrichNodesWithCatalog,
} from '@/lib/node-editor/conversions'
import { useNodeCatalog } from '@/hooks/useNodeCatalog'
import { usePipelineRunStatus } from '@/hooks/usePipelineExecution'

const nodeTypes: NodeTypes = {
  trigger: TriggerNode,
  worker: WorkerNode,
  output: OutputNode,
}

export interface NodeEditorHandle {
  save: () => Promise<void>
}

interface NodeEditorProps {
  initialNodes: Node<ReactFlowNodeData>[]
  initialEdges: Edge[]
  onSave: (nodes: Node<ReactFlowNodeData>[], edges: Edge[]) => Promise<void>
  onDirtyChange?: (dirty: boolean) => void
  runId?: string
}

export const NodeEditor = forwardRef<NodeEditorHandle, NodeEditorProps>(
  function NodeEditor(
    { initialNodes, initialEdges, onSave, onDirtyChange, runId },
    ref
  ) {
    const { data: catalog } = useNodeCatalog()
    const { data: runStatus } = usePipelineRunStatus(runId || '')

    // Enrich initial nodes with catalog metadata
    const enrichedInitial = useMemo(() => {
      if (!catalog) return initialNodes
      return enrichNodesWithCatalog(initialNodes, catalog)
    }, [initialNodes, catalog])

    const [nodes, setNodes, onNodesChange] = useNodesState(enrichedInitial)
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
    const [isDirty, setIsDirty] = useState(false)
    const [selectedNode, setSelectedNode] = useState<ReactFlowNodeData | null>(
      null
    )
    const [configDialogOpen, setConfigDialogOpen] = useState(false)
    const [reactFlowInstance, setReactFlowInstance] =
      useState<ReactFlowInstance | null>(null)
    const reactFlowWrapper = useRef<HTMLDivElement>(null)

    // Track dirty state
    const markDirty = useCallback(() => {
      if (!isDirty) {
        setIsDirty(true)
        onDirtyChange?.(true)
      }
    }, [isDirty, onDirtyChange])

    // Expose save to parent via ref
    useImperativeHandle(
      ref,
      () => ({
        save: async () => {
          await onSave(nodes, edges)
          setIsDirty(false)
          onDirtyChange?.(false)
        },
      }),
      [nodes, edges, onSave, onDirtyChange]
    )

    // Re-enrich when catalog loads after nodes are already set
    useEffect(() => {
      if (catalog && nodes.length > 0) {
        setNodes((nds) => enrichNodesWithCatalog(nds, catalog))
      }
    }, [catalog]) // eslint-disable-line react-hooks/exhaustive-deps

    // Update nodes with live run status
    useEffect(() => {
      if (!runStatus?.nodes) return
      const statusMap = new Map(
        runStatus.nodes.map((ns) => [ns.nodeId, ns])
      )
      setNodes((nds) =>
        nds.map((node) => {
          const ns = statusMap.get(node.id)
          if (!ns && !node.data.nodeStatus) return node
          return {
            ...node,
            data: { ...node.data, nodeStatus: ns },
          }
        })
      )
    }, [runStatus]) // eslint-disable-line react-hooks/exhaustive-deps

    // Wrap onNodesChange to track dirty
    const handleNodesChange: typeof onNodesChange = useCallback(
      (changes) => {
        // Only mark dirty for position changes (drag) and removes, not selection
        const hasMutation = changes.some(
          (c) => c.type === 'position' || c.type === 'remove'
        )
        if (hasMutation) markDirty()
        onNodesChange(changes)
      },
      [onNodesChange, markDirty]
    )

    // Wrap onEdgesChange to track dirty
    const handleEdgesChange: typeof onEdgesChange = useCallback(
      (changes) => {
        const hasMutation = changes.some(
          (c) => c.type === 'remove'
        )
        if (hasMutation) markDirty()
        onEdgesChange(changes)
      },
      [onEdgesChange, markDirty]
    )

    // Connection validation — check type compatibility per handle
    const onConnect = useCallback(
      (connection: Connection) => {
        const sourceHandle = connection.sourceHandle
        const targetHandle = connection.targetHandle

        const sourceType = sourceHandle?.replace(/^output-/, '')
        const targetType = targetHandle?.replace(/^input-/, '')

        if (sourceType && targetType && sourceType !== targetType) {
          const sourceNode = nodes.find((n) => n.id === connection.source)
          const targetNode = nodes.find((n) => n.id === connection.target)
          toast.warning(
            `Incompatible: ${sourceNode?.data.label ?? 'Source'} output [${sourceType}] → ${targetNode?.data.label ?? 'Target'} input [${targetType}]`
          )
          return
        }

        markDirty()
        setEdges((eds) => addEdge({ ...connection, type: 'smoothstep' }, eds))
      },
      [setEdges, nodes, markDirty]
    )

    // Drag-and-drop from sidebar
    const onDragOver = useCallback((event: React.DragEvent) => {
      event.preventDefault()
      event.dataTransfer.dropEffect = 'move'
    }, [])

    const onDrop = useCallback(
      (event: React.DragEvent) => {
        event.preventDefault()
        const raw = event.dataTransfer.getData('application/portwhine-node')
        if (!raw || !reactFlowInstance) return

        let entry: { id?: string; displayName?: string; image?: string; nodeType?: string; acceptedInputTypes?: string[]; outputTypes?: string[]; configSchema?: string; color?: string; icon?: string }
        try {
          entry = JSON.parse(raw)
        } catch {
          return
        }
        const position = reactFlowInstance.screenToFlowPosition({
          x: event.clientX,
          y: event.clientY,
        })

        const nodeType =
          entry.nodeType === 'trigger'
            ? 'trigger'
            : entry.nodeType === 'output'
              ? 'output'
              : 'worker'

        const pipelineType =
          nodeType === 'trigger'
            ? PipelineNodeType.TRIGGER
            : nodeType === 'worker'
              ? PipelineNodeType.WORKER
              : PipelineNodeType.OUTPUT

        const id = `node-${Date.now()}`
        const newNode: Node<ReactFlowNodeData> = {
          id,
          type: nodeType,
          position,
          data: {
            id,
            label: entry.displayName ?? '',
            image: entry.image ?? '',
            type: pipelineType,
            catalogId: entry.id,
            acceptedInputTypes: entry.acceptedInputTypes,
            outputTypes: entry.outputTypes,
            configSchema: entry.configSchema,
            color: entry.color,
            icon: entry.icon,
          },
        }

        markDirty()
        setNodes((nds) => [...nds, newNode])
      },
      [reactFlowInstance, setNodes, markDirty]
    )

    const onNodeClick: NodeMouseHandler = useCallback((_event, node) => {
      setSelectedNode(node.data as ReactFlowNodeData)
      setConfigDialogOpen(true)
    }, [])

    const handleNodeConfigSave = (updates: Partial<ReactFlowNodeData>) => {
      if (!selectedNode) return
      markDirty()
      setNodes((nds) =>
        nds.map((node) =>
          node.id === selectedNode.id
            ? { ...node, data: { ...node.data, ...updates } }
            : node
        )
      )
    }

    const handleDeleteNode = useCallback(
      (nodeId: string) => {
        markDirty()
        setEdges((eds) =>
          eds.filter((e) => e.source !== nodeId && e.target !== nodeId)
        )
        setNodes((nds) => nds.filter((n) => n.id !== nodeId))
        setConfigDialogOpen(false)
        setSelectedNode(null)
      },
      [setNodes, setEdges, markDirty]
    )

    return (
      <div className="h-full flex">
        <NodeCatalogSidebar />

        <div className="flex-1 h-full" ref={reactFlowWrapper}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={handleNodesChange}
            onEdgesChange={handleEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onInit={setReactFlowInstance}
            onDragOver={onDragOver}
            onDrop={onDrop}
            nodeTypes={nodeTypes}
            deleteKeyCode={['Backspace', 'Delete']}
            fitView
          >
            <Controls />
            <MiniMap />
            <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
          </ReactFlow>
        </div>

        <NodeConfigDialog
          open={configDialogOpen}
          onOpenChange={setConfigDialogOpen}
          nodeData={selectedNode}
          onSave={handleNodeConfigSave}
          onDelete={handleDeleteNode}
          runId={runId}
        />
      </div>
    )
  }
)
