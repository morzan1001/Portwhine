import { cn } from '@/lib/utils'

interface LiveIndicatorProps {
  isLive: boolean
  lastUpdate?: Date
  className?: string
}

export function LiveIndicator({ isLive, lastUpdate, className }: LiveIndicatorProps) {
  const getTimeSince = (date: Date) => {
    const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000)
    if (seconds < 60) return `${seconds}s ago`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    return `${hours}h ago`
  }

  if (!isLive) return null

  return (
    <div className={cn('flex items-center gap-2 text-xs', className)}>
      <div className="flex items-center gap-1.5">
        <div className="relative flex h-1.5 w-1.5">
          <div className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary opacity-75" />
          <div className="relative inline-flex h-1.5 w-1.5 rounded-full bg-primary" />
        </div>
        <span className="font-medium text-primary tracking-wide">LIVE</span>
      </div>
      {lastUpdate && (
        <span className="text-muted-foreground">
          Updated {getTimeSince(lastUpdate)}
        </span>
      )}
    </div>
  )
}
