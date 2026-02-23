import { useState, useEffect, useCallback, useRef } from 'react'
import { operatorClient } from '@/lib/api/client'
import { toast } from 'sonner'
import type { DataItem } from '@/gen/portwhine/v1/common_pb'

export function useStreamPipelineResults(runId: string, enabled: boolean = true) {
  const [items, setItems] = useState<DataItem[]>([])
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())
  const reconnectAttemptsRef = useRef(0)
  const maxReconnectAttempts = 5

  const clearItems = useCallback(() => {
    setItems([])
    setError(null)
  }, [])

  useEffect(() => {
    if (!enabled || !runId) {
      setIsStreaming(false)
      return
    }

    const abortController = new AbortController()
    setIsStreaming(true)
    setError(null)

    const startStream = async () => {
      try {
        const stream = operatorClient.streamPipelineResults(
          { runId },
          { signal: abortController.signal }
        )

        reconnectAttemptsRef.current = 0

        for await (const response of stream) {
          if (response.item) {
            setItems((prev) => [...prev, response.item!])
            setLastUpdate(new Date())
          }
        }
      } catch (err: unknown) {
        if (!abortController.signal.aborted) {
          setError(err instanceof Error ? err : new Error(String(err)))

          // Auto-reconnect logic for transient errors
          if (reconnectAttemptsRef.current < maxReconnectAttempts) {
            reconnectAttemptsRef.current++
            const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 10000)
            toast.info(`Reconnecting... (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`)
            setTimeout(() => {
              if (!abortController.signal.aborted) {
                startStream()
              }
            }, delay)
          } else {
            toast.error('Stream connection lost. Please refresh.')
          }
        }
      } finally {
        if (!abortController.signal.aborted) {
          setIsStreaming(false)
        }
      }
    }

    startStream()

    return () => {
      abortController.abort()
      setIsStreaming(false)
    }
  }, [runId, enabled])

  return { items, isStreaming, error, lastUpdate, clearItems }
}

export function useStreamNodeLogs(
  runId: string,
  nodeId: string,
  enabled: boolean = true
) {
  const [logs, setLogs] = useState<string[]>([])
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())
  const reconnectAttemptsRef = useRef(0)
  const maxReconnectAttempts = 5

  const clearLogs = useCallback(() => {
    setLogs([])
    setError(null)
  }, [])

  useEffect(() => {
    if (!enabled || !runId || !nodeId) {
      setIsStreaming(false)
      return
    }

    const abortController = new AbortController()
    setIsStreaming(true)
    setError(null)
    setLogs([])

    const startStream = async () => {
      try {
        const stream = operatorClient.getNodeLogs(
          {
            runId,
            nodeId,
            follow: true,
            tail: 100,
          },
          { signal: abortController.signal }
        )

        reconnectAttemptsRef.current = 0

        for await (const response of stream) {
          if (response.line) {
            setLogs((prev) => [...prev, response.line])
            setLastUpdate(new Date())
          }
        }
      } catch (err: unknown) {
        if (!abortController.signal.aborted) {
          setError(err instanceof Error ? err : new Error(String(err)))

          // Auto-reconnect logic
          if (reconnectAttemptsRef.current < maxReconnectAttempts) {
            reconnectAttemptsRef.current++
            const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 10000)
            setTimeout(() => {
              if (!abortController.signal.aborted) {
                startStream()
              }
            }, delay)
          }
        }
      } finally {
        if (!abortController.signal.aborted) {
          setIsStreaming(false)
        }
      }
    }

    startStream()

    return () => {
      abortController.abort()
      setIsStreaming(false)
    }
  }, [runId, nodeId, enabled])

  return { logs, isStreaming, error, lastUpdate, clearLogs }
}
