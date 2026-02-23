import { useQuery } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'

export function useNodeCatalog() {
  return useQuery({
    queryKey: ['node-catalog'],
    queryFn: async () => {
      const response = await operatorClient.listNodeCatalog({})
      return response.entries
    },
    staleTime: 10 * 60 * 1000, // 10 minutes — catalog changes rarely
  })
}
