import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'

export function usePipelinePermissions(pipelineId: string) {
  return useQuery({
    queryKey: ['permissions', 'pipelines', pipelineId],
    queryFn: async () => {
      const response = await operatorClient.listPermissions({
        resourceType: 'pipelines',
        resourceId: pipelineId,
      })
      return response.permissions
    },
    enabled: !!pipelineId,
  })
}

export function useGrantPermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      subjectType: string
      subjectId: string
      resourceType: string
      resourceId: string
      action: string
      effect: string
    }) => {
      const response = await operatorClient.grantPermission(params)
      return response.permissionId
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['permissions', variables.resourceType, variables.resourceId],
      })
    },
  })
}

export function useRevokePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (permissionId: string) => {
      await operatorClient.revokePermission({ permissionId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
    },
  })
}

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await operatorClient.listUsers({})
      return response.users
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useTeams() {
  return useQuery({
    queryKey: ['teams'],
    queryFn: async () => {
      const response = await operatorClient.listTeams({})
      return response.teams
    },
    staleTime: 5 * 60 * 1000,
  })
}
