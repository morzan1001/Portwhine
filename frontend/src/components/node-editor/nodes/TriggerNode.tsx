import React, { memo } from 'react'
import { Handle, Position, type NodeProps } from 'reactflow'
import type { ReactFlowNodeData } from '@/lib/node-editor/conversions'
import { getNodeIcon } from '@/lib/node-editor/icons'
import { NodeStatusOverlay } from './NodeStatusOverlay'

const ROW_HEIGHT = 24
const HANDLE_SIZE = 10
// Header: py-2.5 (10px) + content (~34px) = 44px from node top
const BODY_TOP = 44

function handleTop(index: number) {
  return BODY_TOP + index * ROW_HEIGHT + (ROW_HEIGHT - HANDLE_SIZE) / 2
}

export const TriggerNode = memo(({ data, selected }: NodeProps<ReactFlowNodeData>) => {
  const color = data.color || 'hsl(var(--primary))'
  const outputs = data.outputTypes || []
  const isRunning = data.nodeStatus?.containerStatus === 'running'

  return (
    <div
      className={`relative px-4 py-2.5 shadow-md rounded-lg border-2 min-w-[180px] transition-all ${
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
          {React.createElement(getNodeIcon(data.icon), { className: "h-3.5 w-3.5" })}
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-sm font-semibold text-foreground truncate">{data.label}</div>
          <div className="text-[10px] font-medium uppercase tracking-wider" style={{ color }}>
            Trigger
          </div>
        </div>
      </div>

      {/* Output labels — each row matches handle position */}
      {outputs.length > 0 && (
        <div>
          {outputs.map((type) => (
            <div key={type} className="flex items-center justify-end" style={{ height: ROW_HEIGHT }}>
              <span className="text-[10px] text-muted-foreground font-mono">{type}</span>
            </div>
          ))}
        </div>
      )}

      {/* Output handles — centered in each row */}
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

      {/* Fallback single handle when no catalog data */}
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

TriggerNode.displayName = 'TriggerNode'
