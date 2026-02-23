import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { Button } from '@/components/ui/button'
import { PlayCircle, Square, Clock, CheckCircle, XCircle } from 'lucide-react'

interface RunStatusCardProps {
  status: string
  startedAt?: Date
  completedAt?: Date
  pipelineName: string
  onStop?: () => void
  isStopPending?: boolean
}

export function RunStatusCard({
  status,
  startedAt,
  completedAt,
  pipelineName,
  onStop,
  isStopPending,
}: RunStatusCardProps) {
  const getStatusIcon = () => {
    switch (status) {
      case 'RUNNING':
        return <PlayCircle className="h-5 w-5 text-[hsl(var(--status-running))] animate-pulse" />
      case 'COMPLETED':
        return <CheckCircle className="h-5 w-5 text-[hsl(var(--status-completed))]" />
      case 'FAILED':
        return <XCircle className="h-5 w-5 text-[hsl(var(--status-failed))]" />
      default:
        return <Clock className="h-5 w-5 text-muted-foreground" />
    }
  }

  const getStatusBadgeStatus = (): "running" | "completed" | "failed" | "paused" | "pending" => {
    const statusMap: Record<string, "running" | "completed" | "failed" | "paused" | "pending"> = {
      'RUNNING': 'running',
      'COMPLETED': 'completed',
      'FAILED': 'failed',
      'PAUSED': 'paused',
    }
    return statusMap[status] || 'pending'
  }

  const getDuration = () => {
    if (!startedAt) return 'Not started'
    const end = completedAt || new Date()
    const diff = end.getTime() - startedAt.getTime()
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`
    } else {
      return `${seconds}s`
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {getStatusIcon()}
            <div>
              <CardTitle className="text-base">{pipelineName}</CardTitle>
              <p className="text-xs text-muted-foreground mt-0.5">
                Pipeline Execution
              </p>
            </div>
          </div>
          <StatusBadge status={getStatusBadgeStatus()} withDot>
            {status}
          </StatusBadge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-muted-foreground">Started</p>
            <p className="text-sm font-medium mt-0.5">
              {startedAt ? startedAt.toLocaleString() : 'N/A'}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Duration</p>
            <p className="text-sm font-medium mt-0.5">{getDuration()}</p>
          </div>
        </div>

        {status === 'RUNNING' && onStop && (
          <div className="mt-4">
            <Button
              variant="destructive"
              size="sm"
              onClick={onStop}
              disabled={isStopPending}
            >
              <Square className="h-3.5 w-3.5 mr-2" />
              {isStopPending ? 'Stopping...' : 'Stop Pipeline'}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
