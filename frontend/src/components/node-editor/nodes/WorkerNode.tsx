import { memo } from 'react'
import { Handle, Position, type NodeProps } from 'reactflow'
import type { ReactFlowNodeData } from '@/lib/node-editor/conversions'
import { getNodeIcon } from '@/lib/node-editor/icons'
import { Badge } from '@/components/ui/badge'
import { NodeStatusOverlay } from './NodeStatusOverlay'

const ROW_HEIGHT = 24
const HANDLE_SIZE = 10
const BODY_TOP = 44

function handleTop(index: number) {
  return BODY_TOP + index * ROW_HEIGHT + (ROW_HEIGHT - HANDLE_SIZE) / 2
}

export const WorkerNode = memo(({ data, selected }: NodeProps<ReactFlowNodeData>) => {
  const Icon = getNodeIcon(data.icon)
  const color = data.color || 'hsl(var(--secondary))'
  const inputs = data.acceptedInputTypes || []
  const outputs = data.outputTypes || []
  const isRunning = data.nodeStatus?.containerStatus === 'running'

  const handleCount = Math.max(inputs.length, outputs.length)

  return (
    <div
      className={`relative px-4 py-2.5 shadow-md rounded-lg border-2 min-w-[200px] transition-all ${
        selected ? 'ring-2 ring-offset-2' : ''
      } ${isRunning ? 'animate-pulse' : ''}`}
      style={{
        borderColor: color,
        backgroundColor: 'hsl(var(--card))',
        ...(selected ? { ringColor: color } : {}),
      }}
    >
      <NodeStatusOverlay status={data.nodeStatus} />

      {/* Header */}
      <div className="flex items-center gap-2">
        <div
          className="flex h-6 w-6 items-center justify-center rounded"
          style={{ backgroundColor: `${color}20`, color }}
        >
          <Icon className="h-3.5 w-3.5" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="text-sm font-semibold text-foreground truncate">{data.label}</span>
            {data.replicas && data.replicas > 1 && (
              <Badge variant="secondary" className="text-[10px] px-1 py-0 h-4">
                {data.replicas}x
              </Badge>
            )}
          </div>
          <div className="text-[10px] font-medium uppercase tracking-wider" style={{ color }}>
            Worker
          </div>
        </div>
      </div>

      {/* Input/Output type labels — rows aligned with handles */}
      {handleCount > 0 && (
        <div>
          {Array.from({ length: handleCount }).map((_, i) => (
            <div key={i} className="flex items-center justify-between" style={{ height: ROW_HEIGHT }}>
              <span className="text-[10px] text-muted-foreground font-mono">
                {inputs[i] || ''}
              </span>
              <span className="text-[10px] text-muted-foreground font-mono">
                {outputs[i] || ''}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Input handles */}
      {inputs.map((type, index) => (
        <Handle
          key={`input-${type}`}
          type="target"
          position={Position.Left}
          id={`input-${type}`}
          style={{
            top: handleTop(index),
            background: color,
            width: HANDLE_SIZE,
            height: HANDLE_SIZE,
          }}
        />
      ))}

      {/* Output handles */}
      {outputs.map((type, index) => (
        <Handle
          key={`output-${type}`}
          type="source"
          position={Position.Right}
          id={`output-${type}`}
          style={{
            top: handleTop(index),
            background: color,
            width: HANDLE_SIZE,
            height: HANDLE_SIZE,
          }}
        />
      ))}

      {/* Fallback handles when no catalog data */}
      {inputs.length === 0 && (
        <Handle
          type="target"
          position={Position.Left}
          style={{ background: color, width: HANDLE_SIZE, height: HANDLE_SIZE }}
        />
      )}
      {outputs.length === 0 && (
        <Handle
          type="source"
          position={Position.Right}
          style={{ background: color, width: HANDLE_SIZE, height: HANDLE_SIZE }}
        />
      )}
    </div>
  )
})

WorkerNode.displayName = 'WorkerNode'
