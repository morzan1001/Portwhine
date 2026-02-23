import { Workflow, PlayCircle, CheckCircle, XCircle, Plus } from 'lucide-react'
import { usePipelines } from '@/hooks/usePipelines'
import { useDashboardStats } from '@/hooks/useDashboardStats'
import { StatsCard } from '@/components/dashboard/StatsCard'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { StatsGridSkeleton } from '@/components/LoadingState'
import { ErrorState } from '@/components/ErrorState'
import { useNavigate } from 'react-router-dom'
import { timestampToDate, runStateBadge, runStateLabel } from '@/lib/utils'

export function DashboardPage() {
  const navigate = useNavigate()
  const { data: pipelines, isLoading: pipelinesLoading } = usePipelines()
  const { data: stats, isLoading: statsLoading, isError: statsError, refetch: refetchStats } = useDashboardStats()

  return (
    <div className="flex flex-col gap-8 p-8 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Pipeline overview and recent activity
          </p>
        </div>
        <Button onClick={() => navigate('/pipelines')} leftIcon={<Plus className="h-4 w-4" />}>
          New Pipeline
        </Button>
      </div>

      {statsLoading ? (
        <StatsGridSkeleton />
      ) : statsError ? (
        <ErrorState
          type="server"
          title="Failed to load dashboard stats"
          onRetry={() => refetchStats()}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatsCard title="Total Pipelines" value={stats?.totalPipelines || 0} icon={Workflow} />
          <StatsCard title="Running" value={stats?.runningRuns || 0} icon={PlayCircle} />
          <StatsCard title="Completed" value={stats?.completedRuns || 0} icon={CheckCircle} />
          <StatsCard title="Failed" value={stats?.failedRuns || 0} icon={XCircle} />
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">Recent Pipelines</CardTitle>
          </CardHeader>
          <CardContent>
            {pipelinesLoading ? (
              <div className="space-y-3">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-12 bg-muted animate-pulse rounded-lg" />
                ))}
              </div>
            ) : pipelines && pipelines.length > 0 ? (
              <div className="space-y-1">
                {pipelines.slice(0, 5).map((pipeline) => (
                  <div
                    key={pipeline.pipelineId}
                    className="flex items-center justify-between p-3 rounded-lg hover:bg-accent transition-colors duration-150 cursor-pointer"
                    onClick={() => navigate(`/pipelines/${pipeline.pipelineId}/edit`)}
                  >
                    <div className="flex items-center gap-3">
                      <Workflow className="h-4 w-4 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">{pipeline.name}</p>
                        <p className="text-xs text-muted-foreground">v{pipeline.version}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <Workflow className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
                <p className="text-sm text-muted-foreground">No pipelines yet</p>
                <Button onClick={() => navigate('/pipelines')} variant="outline" className="mt-3" size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  Create Pipeline
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">Recent Runs</CardTitle>
          </CardHeader>
          <CardContent>
            {statsLoading ? (
              <div className="space-y-3">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-12 bg-muted animate-pulse rounded-lg" />
                ))}
              </div>
            ) : stats?.recentRuns && stats.recentRuns.length > 0 ? (
              <div className="space-y-1">
                {stats.recentRuns.slice(0, 5).map((run) => (
                  <div
                    key={run.runId}
                    className="flex items-center justify-between p-3 rounded-lg hover:bg-accent transition-colors duration-150 cursor-pointer"
                    onClick={() => navigate(`/runs/${run.runId}`)}
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{run.pipelineName || run.pipelineId}</p>
                      <p className="text-xs text-muted-foreground">
                        {timestampToDate(run.startedAt)?.toLocaleString() || 'Not started'}
                      </p>
                    </div>
                    <StatusBadge status={runStateBadge(run.state)} withDot>
                      {runStateLabel(run.state)}
                    </StatusBadge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <PlayCircle className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
                <p className="text-sm text-muted-foreground">No runs yet</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
