import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'
import { toast } from 'sonner'
import { PipelineRunState } from '@/gen/portwhine/v1/operator_pb'

export function useStartPipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (pipelineId: string) => {
      const response = await operatorClient.startPipeline({
        pipelineId,
      })
      return response.runId
    },
    onSuccess: (runId) => {
      queryClient.invalidateQueries({ queryKey: ['pipeline-runs'] })
      toast.success(`Pipeline started. Run ID: ${runId}`)
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to start pipeline')
    },
  })
}

export function useStopPipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (runId: string) => {
      await operatorClient.stopPipelineRun({
        runId,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline-runs'] })
      toast.success('Pipeline stopped')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to stop pipeline')
    },
  })
}

export function usePipelineRunStatus(runId: string) {
  return useQuery({
    queryKey: ['pipeline-run-status', runId],
    queryFn: async () => {
      const response = await operatorClient.getPipelineRunStatus({
        runId,
      })
      return response.status
    },
    refetchInterval: (query) => {
      const state = query.state.data?.state
      if (state === PipelineRunState.COMPLETED || state === PipelineRunState.FAILED || state === PipelineRunState.CANCELLED) {
        return false
      }
      return 3000
    },
    enabled: !!runId,
  })
}
