import { memo } from 'react'
import type { NodeStatus } from '@/gen/portwhine/v1/operator_pb'
import { CheckCircle, XCircle, Loader2, Clock } from 'lucide-react'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface NodeStatusOverlayProps {
  status?: NodeStatus
}

export const NodeStatusOverlay = memo(({ status }: NodeStatusOverlayProps) => {
  if (!status) return null

  const containerStatus = status.containerStatus || ''
  const isRunning = containerStatus === 'running'
  const isSucceeded = containerStatus === 'succeeded'
  const isFailed = containerStatus === 'failed'
  const hasError = (status.errors ?? 0n) > 0n || !!status.errorMessage

  return (
    <>
      {/* Status indicator in top-right corner */}
      <div className="absolute -top-1.5 -right-1.5 z-10">
        {isRunning && (
          <Loader2 className="h-4 w-4 text-green-500 animate-spin" />
        )}
        {isSucceeded && (
          <CheckCircle className="h-4 w-4 text-green-500" />
        )}
        {isFailed && (
          <XCircle className="h-4 w-4 text-red-500" />
        )}
        {!isRunning && !isSucceeded && !isFailed && containerStatus && (
          <Clock className="h-4 w-4 text-muted-foreground" />
        )}
      </div>

      {/* Item counters at bottom */}
      {(status.itemsIn > 0n || status.itemsOut > 0n || (status.errors ?? 0n) > 0n) && (
        <div className="flex items-center gap-2 mt-1.5 text-[10px] font-mono text-muted-foreground">
          <span>IN:{status.itemsIn.toString()}</span>
          <span>OUT:{status.itemsOut.toString()}</span>
          {(status.errors ?? 0n) > 0n && (
            <span className="text-red-500">ERR:{status.errors!.toString()}</span>
          )}
        </div>
      )}

      {/* Error tooltip */}
      {hasError && status.errorMessage && (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="mt-1 text-[10px] text-red-500 truncate max-w-[160px] cursor-help">
                {status.errorMessage}
              </div>
            </TooltipTrigger>
            <TooltipContent side="bottom" className="max-w-xs">
              <p className="text-xs font-mono whitespace-pre-wrap">{status.errorMessage}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </>
  )
})

NodeStatusOverlay.displayName = 'NodeStatusOverlay'
