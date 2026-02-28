import React, { memo } from 'react'
import type { NodeCatalogEntry } from '@/gen/portwhine/v1/operator_pb'
import { getNodeIcon } from '@/lib/node-editor/icons'
import { Badge } from '@/components/ui/badge'

interface CatalogEntryCardProps {
  entry: NodeCatalogEntry
}

export const CatalogEntryCard = memo(({ entry }: CatalogEntryCardProps) => {
  const onDragStart = (event: React.DragEvent) => {
    event.dataTransfer.setData('application/portwhine-node', JSON.stringify(entry))
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <div
      draggable
      onDragStart={onDragStart}
      className="flex items-start gap-3 rounded-md border p-2.5 cursor-grab active:cursor-grabbing hover:bg-muted/50 transition-colors"
      style={{ borderLeftWidth: 3, borderLeftColor: entry.color }}
    >
      <div
        className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md"
        style={{ backgroundColor: `${entry.color}20`, color: entry.color }}
      >
        {React.createElement(getNodeIcon(entry.icon), { className: "h-4 w-4" })}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="text-sm font-medium truncate">{entry.displayName}</span>
          <Badge
            variant="outline"
            className="text-[10px] px-1 py-0 h-4 shrink-0"
          >
            {entry.nodeType}
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
          {entry.description}
        </p>
      </div>
    </div>
  )
})

CatalogEntryCard.displayName = 'CatalogEntryCard'
