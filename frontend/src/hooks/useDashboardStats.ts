import { useQuery } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      return await operatorClient.getDashboardStats({})
    },
    refetchInterval: 10000,
  })
}
