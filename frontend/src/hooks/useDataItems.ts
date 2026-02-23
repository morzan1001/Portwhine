import { useQuery } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'

export function useDataItems(runId: string, typeFilter?: string, pageSize = 50, pageToken?: string) {
  return useQuery({
    queryKey: ['data-items', runId, typeFilter ?? '', pageSize, pageToken ?? ''],
    queryFn: async () => {
      const response = await operatorClient.listDataItems({
        runId,
        typeFilter: typeFilter ?? '',
        pageSize,
        pageToken: pageToken ?? '',
      })
      return {
        items: response.items,
        nextPageToken: response.nextPageToken,
        totalCount: Number(response.totalCount),
      }
    },
    enabled: !!runId,
  })
}
