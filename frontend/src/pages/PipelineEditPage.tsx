import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { ArrowLeft, Play, Square, Save, Settings, RotateCcw } from 'lucide-react'

import { usePipeline, usePipelineRuns } from '@/hooks/usePipelines'
import {
  useStartPipeline,
  useStopPipeline,
  usePipelineRunStatus,
} from '@/hooks/usePipelineExecution'
import { PipelineRunState } from '@/gen/portwhine/v1/operator_pb'
import { operatorClient } from '@/lib/api/client'
import { NodeEditor, type NodeEditorHandle } from '@/components/node-editor/NodeEditor'
import { PipelineMetadataDialog } from '@/components/node-editor/PipelineMetadataDialog'
import { protoToReactFlow, reactFlowToProto } from '@/lib/node-editor/conversions'
import { Node, Edge } from 'reactflow'
import type { ReactFlowNodeData } from '@/lib/node-editor/conversions'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { NodeEditorSkeleton } from '@/components/LoadingState'

function getRunStateBadge(state?: PipelineRunState) {
  switch (state) {
    case PipelineRunState.RUNNING:
      return (
        <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500">
          <span className="animate-pulse mr-1.5">●</span>
          Running
        </Badge>
      )
    case PipelineRunState.COMPLETED:
      return (
        <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500">
          Completed
        </Badge>
      )
    case PipelineRunState.FAILED:
      return (
        <Badge variant="outline" className="bg-red-500/10 text-red-500 border-red-500">
          Failed
        </Badge>
      )
    case PipelineRunState.CANCELLED:
      return (
        <Badge variant="outline" className="bg-yellow-500/10 text-yellow-500 border-yellow-500">
          Cancelled
        </Badge>
      )
    case PipelineRunState.PAUSED:
      return (
        <Badge variant="outline" className="bg-orange-500/10 text-orange-500 border-orange-500">
          Paused
        </Badge>
      )
    case PipelineRunState.PENDING:
      return (
        <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500">
          Pending
        </Badge>
      )
    default:
      return null
  }
}

export function PipelineEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data: pipeline, isLoading } = usePipeline(id || '')
  const { data: runs } = usePipelineRuns(id || '')
  const startPipeline = useStartPipeline()
  const stopPipeline = useStopPipeline()
  const editorRef = useRef<NodeEditorHandle>(null)

  const [initialNodes, setInitialNodes] = useState<Node<ReactFlowNodeData>[]>([])
  const [initialEdges, setInitialEdges] = useState<Edge[]>([])
  const [runId, setRunId] = useState<string | undefined>()
  const [isDirty, setIsDirty] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [metadataDialogOpen, setMetadataDialogOpen] = useState(false)
  const [pipelineName, setPipelineName] = useState('')
  const [pipelineDescription, setPipelineDescription] = useState('')

  // Get the current run status (polls while running, stops when terminal)
  const { data: currentRunStatus } = usePipelineRunStatus(runId || '')
  const runState = currentRunStatus?.state
  const isRunning =
    runState === PipelineRunState.RUNNING || runState === PipelineRunState.PENDING

  // Auto-load the most recent run when opening the page
  useEffect(() => {
    if (!runs || runs.length === 0 || runId) return
    const mostRecent = runs[0]
    if (mostRecent?.runId) {
      setRunId(mostRecent.runId)
    }
  }, [runs]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (pipeline?.definition) {
      const { nodes, edges } = protoToReactFlow(pipeline.definition)
      setInitialNodes(nodes)
      setInitialEdges(edges)
      setPipelineName(pipeline.definition.name || '')
      setPipelineDescription(pipeline.definition.description || '')
    }
  }, [pipeline])

  const updateMutation = useMutation({
    mutationFn: async ({
      nodes,
      edges,
    }: {
      nodes: Node<ReactFlowNodeData>[]
      edges: Edge[]
    }) => {
      const definition = reactFlowToProto(nodes, edges)
      definition.name = pipelineName
      definition.description = pipelineDescription
      await operatorClient.updatePipeline({
        pipelineId: id || '',
        definition,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline', id] })
      queryClient.invalidateQueries({ queryKey: ['pipelines'] })
      toast.success('Pipeline saved')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to save pipeline')
    },
  })

  const handleSave = async (nodes: Node<ReactFlowNodeData>[], edges: Edge[]) => {
    await updateMutation.mutateAsync({ nodes, edges })
  }

  const handleSaveClick = async () => {
    if (!editorRef.current) return
    setIsSaving(true)
    try {
      await editorRef.current.save()
    } finally {
      setIsSaving(false)
    }
  }

  const handleStart = async () => {
    if (!id) return
    const result = await startPipeline.mutateAsync(id)
    setRunId(result)
  }

  const handleStop = async () => {
    if (!runId) return
    await stopPipeline.mutateAsync(runId)
  }

  const handleClearRun = () => {
    setRunId(undefined)
  }

  const handleMetadataSave = (name: string, description: string) => {
    setPipelineName(name)
    setPipelineDescription(description)
    setIsDirty(true)
  }

  if (isLoading) {
    return (
      <div className="h-screen flex flex-col">
        <div className="border-b bg-card p-4">
          <NodeEditorSkeleton />
        </div>
      </div>
    )
  }

  if (!pipeline) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground">Pipeline not found</p>
      </div>
    )
  }

  return (
    <div className="h-screen flex flex-col">
      {/* Unified Toolbar */}
      <div className="border-b bg-card px-4 py-2">
        <div className="flex items-center gap-3">
          {/* Left: Back + Name + Badge */}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 shrink-0"
            onClick={() => navigate('/pipelines')}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>

          <h1 className="text-sm font-semibold truncate">{pipelineName || 'Unnamed Pipeline'}</h1>

          {runId && getRunStateBadge(runState)}

          {/* Spacer */}
          <div className="flex-1" />

          {/* Right: Actions */}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => setMetadataDialogOpen(true)}
            title="Pipeline settings"
          >
            <Settings className="h-4 w-4" />
          </Button>

          {runId && !isRunning && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClearRun}
              title="Clear run status overlay"
            >
              <RotateCcw className="h-4 w-4 mr-1.5" />
              Clear
            </Button>
          )}

          {isDirty && (
            <Button
              size="sm"
              onClick={handleSaveClick}
              disabled={isSaving}
            >
              <Save className="h-4 w-4 mr-1.5" />
              {isSaving ? 'Saving...' : 'Save'}
            </Button>
          )}

          {isRunning ? (
            <Button
              variant="destructive"
              size="sm"
              onClick={handleStop}
              disabled={stopPipeline.isPending}
            >
              <Square className="h-4 w-4 mr-1.5" />
              {stopPipeline.isPending ? 'Stopping...' : 'Stop'}
            </Button>
          ) : (
            <Button
              size="sm"
              onClick={handleStart}
              disabled={startPipeline.isPending || isDirty}
              title={isDirty ? 'Save changes before starting' : 'Start pipeline'}
            >
              <Play className="h-4 w-4 mr-1.5" />
              {startPipeline.isPending ? 'Starting...' : 'Start'}
            </Button>
          )}
        </div>
      </div>

      {/* Editor */}
      <div className="flex-1 min-h-0">
        <NodeEditor
          ref={editorRef}
          key={`${id}-${initialNodes.length}-${initialEdges.length}`}
          initialNodes={initialNodes}
          initialEdges={initialEdges}
          onSave={handleSave}
          onDirtyChange={setIsDirty}
          runId={runId}
        />
      </div>

      {/* Metadata Dialog */}
      <PipelineMetadataDialog
        open={metadataDialogOpen}
        onOpenChange={setMetadataDialogOpen}
        pipelineId={id || ''}
        name={pipelineName}
        description={pipelineDescription}
        onSave={handleMetadataSave}
      />
    </div>
  )
}
