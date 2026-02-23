import { usePipelineRunStatus, useStopPipeline } from '@/hooks/usePipelineExecution'
import { useStreamPipelineResults } from '@/hooks/useStreaming'
import { RunStatusCard } from '@/components/runs/RunStatusCard'
import { DataItemsTable } from '@/components/runs/DataItemsTable'
import { useNavigate, useParams } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { DetailPageSkeleton } from '@/components/LoadingState'
import { ArrowLeft, Package, AlertCircle, Activity } from 'lucide-react'
import { timestampToDate, runStateLabel } from '@/lib/utils'
import { PipelineRunState, type NodeStatus } from '@/gen/portwhine/v1/operator_pb'

export function RunDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: runStatus, isLoading } = usePipelineRunStatus(id || '')
  const stopMutation = useStopPipeline()

  const isRunning = runStatus?.state === PipelineRunState.RUNNING
  const { items: streamItems, isStreaming } = useStreamPipelineResults(id || '', isRunning)

  const handleStop = () => {
    if (id) {
      stopMutation.mutate(id)
    }
  }

  if (isLoading) {
    return (
      <div className="flex flex-col gap-8 p-8">
        <DetailPageSkeleton />
      </div>
    )
  }

  if (!runStatus) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-muted-foreground">Run not found</p>
      </div>
    )
  }

  const nodes = runStatus.nodes || []
  const totalItems = nodes.reduce((sum: number, n: NodeStatus) => sum + Number(n.itemsOut), 0)
  const errorCount = nodes.reduce((sum: number, n: NodeStatus) => sum + Number(n.errors), 0)
  const successRate = totalItems > 0 ? ((totalItems - errorCount) / totalItems) * 100 : 0

  return (
    <div className="flex flex-col gap-8 p-8 animate-fade-in">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate('/runs')}
          leftIcon={<ArrowLeft className="h-4 w-4" />}
        >
          Back
        </Button>
      </div>

      <RunStatusCard
        status={runStateLabel(runStatus.state)}
        startedAt={timestampToDate(runStatus.startedAt)}
        completedAt={timestampToDate(runStatus.finishedAt)}
        pipelineName={runStatus.pipelineName || runStatus.pipelineId}
        onStop={handleStop}
        isStopPending={stopMutation.isPending}
      />

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-muted-foreground">Total Items</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{totalItems}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-muted-foreground">Errors</CardTitle>
            <AlertCircle className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-destructive">{errorCount}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-muted-foreground">Success Rate</CardTitle>
            <Activity className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{successRate.toFixed(1)}%</div>
            <Progress value={successRate} className="mt-2" />
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="data" className="w-full">
        <TabsList>
          <TabsTrigger value="data">Data Items</TabsTrigger>
          <TabsTrigger value="nodes">Node Status</TabsTrigger>
        </TabsList>

        <TabsContent value="data">
          <DataItemsTable
            runId={id || ''}
            isStreaming={isStreaming}
            streamItemCount={streamItems.length}
          />
        </TabsContent>

        <TabsContent value="nodes">
          {nodes.length > 0 ? (
            <div className="rounded-xl border bg-card p-6">
              <div className="space-y-2">
                {nodes.map((node: NodeStatus) => (
                  <div
                    key={node.nodeId}
                    className="flex items-center justify-between p-3 rounded-lg bg-accent/50"
                  >
                    <div>
                      <p className="text-sm font-medium">{node.nodeId}</p>
                      <p className="text-xs text-muted-foreground">
                        {node.workerType} &middot; {node.containerStatus || 'Unknown'}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-muted-foreground">
                        In: {Number(node.itemsIn)} | Out: {Number(node.itemsOut)}
                      </p>
                      {Number(node.errors) > 0 && (
                        <p className="text-xs text-destructive">
                          {Number(node.errors)} errors
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="text-center py-12 rounded-xl border">
              <p className="text-sm text-muted-foreground">No node status available</p>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
