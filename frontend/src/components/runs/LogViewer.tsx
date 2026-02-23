import { useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface LogViewerProps {
  logs: string[]
  isStreaming: boolean
  title?: string
}

export function LogViewer({ logs, isStreaming, title = 'Logs' }: LogViewerProps) {
  const logEndRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (logEndRef.current && containerRef.current) {
      const container = containerRef.current
      const isNearBottom =
        container.scrollHeight - container.scrollTop - container.clientHeight < 100

      // Only auto-scroll if user is near the bottom
      if (isNearBottom) {
        logEndRef.current.scrollIntoView({ behavior: 'smooth' })
      }
    }
  }, [logs])

  const getLogColor = (log: string) => {
    const lower = log.toLowerCase()
    if (lower.includes('error') || lower.includes('fail')) {
      return 'text-red-400'
    } else if (lower.includes('warn')) {
      return 'text-yellow-400'
    } else if (lower.includes('success') || lower.includes('complete')) {
      return 'text-green-400'
    }
    return 'text-gray-300'
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">{title}</CardTitle>
          {isStreaming && (
            <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500">
              <span className="animate-pulse mr-2">●</span>
              Streaming
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex-1 p-0">
        <div
          ref={containerRef}
          className="h-full overflow-y-auto bg-gray-950 p-4 font-mono text-sm"
        >
          {logs.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">
              {isStreaming ? 'Waiting for logs...' : 'No logs available'}
            </p>
          ) : (
            <div className="space-y-1">
              {logs.map((log, index) => (
                <div key={index} className={`${getLogColor(log)} whitespace-pre-wrap break-words`}>
                  <span className="text-gray-500 mr-2">[{index + 1}]</span>
                  {log}
                </div>
              ))}
              <div ref={logEndRef} />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
