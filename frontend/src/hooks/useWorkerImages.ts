import { useQuery } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'

export function useWorkerImages() {
  return useQuery({
    queryKey: ['worker-images'],
    queryFn: async () => {
      const response = await operatorClient.listWorkerImages({})
      return response.images
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}
