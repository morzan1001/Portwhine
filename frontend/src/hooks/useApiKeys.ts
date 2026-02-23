import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'
import { toast } from 'sonner'

export function useApiKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const response = await operatorClient.listAPIKeys({})
      return response.keys
    },
  })
}

export function useCreateApiKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: {
      name: string
      scopes?: string[]
    }) => {
      const response = await operatorClient.createAPIKey({
        name: data.name,
        scopes: data.scopes || [],
      })
      return response
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      toast.success('API key created successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create API key')
    },
  })
}

export function useRevokeApiKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (apiKeyId: string) => {
      await operatorClient.revokeAPIKey({ apiKeyId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      toast.success('API key revoked')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to revoke API key')
    },
  })
}
