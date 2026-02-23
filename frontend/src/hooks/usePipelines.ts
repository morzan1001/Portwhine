import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'
import { toast } from 'sonner'

export function usePipelines() {
  return useQuery({
    queryKey: ['pipelines'],
    queryFn: async () => {
      const response = await operatorClient.listPipelines({
        pageSize: 100,
      })
      return response.pipelines
    },
  })
}

export function usePipeline(pipelineId: string) {
  return useQuery({
    queryKey: ['pipeline', pipelineId],
    queryFn: async () => {
      return await operatorClient.getPipeline({ pipelineId })
    },
    enabled: !!pipelineId,
  })
}

export function useDeletePipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (pipelineId: string) => {
      await operatorClient.deletePipeline({ pipelineId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipelines'] })
      toast.success('Pipeline deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete pipeline')
    },
  })
}

export function usePipelineRuns(pipelineId?: string) {
  return useQuery({
    queryKey: ['pipeline-runs', pipelineId ?? 'all'],
    queryFn: async () => {
      const response = await operatorClient.listPipelineRuns({
        pipelineId: pipelineId ?? '',
        pageSize: 100,
      })
      return response.runs
    },
    refetchInterval: 5000,
  })
}
