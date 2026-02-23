import { usePipelineRuns } from '@/hooks/usePipelines'
import { useNavigate } from 'react-router-dom'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/ui/status-badge'
import { RunTableSkeleton } from '@/components/LoadingState'
import { ErrorState } from '@/components/ErrorState'
import { PlayCircle } from 'lucide-react'
import { timestampToDate, runStateBadge, runStateLabel } from '@/lib/utils'
import type { PipelineRunStatus } from '@/gen/portwhine/v1/operator_pb'

export function RunsPage() {
  const navigate = useNavigate()
  const { data: runs, isLoading, isError, refetch } = usePipelineRuns()

  return (
    <div className="flex flex-col gap-8 p-8 animate-fade-in">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Pipeline Runs</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Monitor all pipeline executions
        </p>
      </div>

      {isLoading ? (
        <RunTableSkeleton />
      ) : isError ? (
        <ErrorState
          type="server"
          title="Failed to load runs"
          message="Unable to fetch pipeline runs. Please try again."
          onRetry={refetch}
        />
      ) : runs && runs.length > 0 ? (
        <div className="rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Run ID</TableHead>
                <TableHead>Pipeline</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Started</TableHead>
                <TableHead>Duration</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {runs.map((run: PipelineRunStatus) => {
                const startDate = timestampToDate(run.startedAt)
                const endDate = timestampToDate(run.finishedAt)
                const duration = startDate && endDate
                  ? Math.floor((endDate.getTime() - startDate.getTime()) / 1000)
                  : null

                return (
                  <TableRow
                    key={run.runId}
                    className="cursor-pointer"
                    onClick={() => navigate(`/runs/${run.runId}`)}
                  >
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {run.runId.substring(0, 12)}...
                    </TableCell>
                    <TableCell className="font-medium">
                      {run.pipelineName || run.pipelineId}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={runStateBadge(run.state)} withDot>
                        {runStateLabel(run.state)}
                      </StatusBadge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {startDate?.toLocaleString() || 'N/A'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {duration ? `${duration}s` : '-'}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className="text-center py-16">
          <PlayCircle className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">No runs yet</p>
        </div>
      )}
    </div>
  )
}
