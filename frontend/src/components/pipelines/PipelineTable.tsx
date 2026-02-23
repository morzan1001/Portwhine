import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Trash, Workflow } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { timestampToDate } from '@/lib/utils'
import type { PipelineSummary } from '@/gen/portwhine/v1/operator_pb'

interface PipelineTableProps {
  pipelines: PipelineSummary[]
  onDelete: (pipelineId: string) => void
  searchQuery?: string
}

export function PipelineTable({ pipelines, onDelete, searchQuery }: PipelineTableProps) {
  const navigate = useNavigate()

  if (pipelines.length === 0) {
    return (
      <div className="text-center py-16">
        <Workflow className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
        <p className="text-sm text-muted-foreground">
          {searchQuery ? 'No pipelines match your search' : 'No pipelines found'}
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-xl border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Version</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="w-10" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {pipelines.map((pipeline) => (
            <TableRow
              key={pipeline.pipelineId}
              className="cursor-pointer"
              onClick={() => navigate(`/pipelines/${pipeline.pipelineId}/edit`)}
            >
              <TableCell>
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-primary/5 p-2">
                    <Workflow className="h-4 w-4 text-primary" />
                  </div>
                  <span className="text-sm font-medium">{pipeline.name}</span>
                </div>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground max-w-[300px] truncate">
                {pipeline.description || 'No description'}
              </TableCell>
              <TableCell>
                <Badge variant="secondary">v{pipeline.version}</Badge>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {timestampToDate(pipeline.createdAt)?.toLocaleDateString() || 'N/A'}
              </TableCell>
              <TableCell>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(pipeline.pipelineId)
                  }}
                >
                  <Trash className="h-4 w-4 text-destructive" />
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
