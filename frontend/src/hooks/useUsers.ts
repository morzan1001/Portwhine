import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'
import { toast } from 'sonner'

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await operatorClient.listUsers({})
      return response.users
    },
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: {
      username: string
      email: string
      password: string
      role: string
    }) => {
      await operatorClient.createUser(data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success('User created successfully')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create user')
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (userId: string) => {
      await operatorClient.deleteUser({ userId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success('User deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete user')
    },
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: {
      userId: string
      email: string
      role: string
      isActive: boolean
    }) => {
      await operatorClient.updateUser(data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success('User updated')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update user')
    },
  })
}

export function useRoles() {
  return useQuery({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await operatorClient.listRoles({})
      return response.roles
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useUserPermissions(userId: string) {
  return useQuery({
    queryKey: ['permissions', 'user', userId],
    queryFn: async () => {
      const response = await operatorClient.listPermissions({
        subjectType: 'user',
        subjectId: userId,
      })
      return response.permissions
    },
    enabled: !!userId,
  })
}

export function useTeams() {
  return useQuery({
    queryKey: ['teams'],
    queryFn: async () => {
      const response = await operatorClient.listTeams({})
      return response.teams
    },
  })
}

export function useUpdateTeam() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { teamId: string; name: string; description: string }) => {
      await operatorClient.updateTeam(data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      toast.success('Team updated')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update team')
    },
  })
}

export function useDeleteTeam() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (teamId: string) => {
      await operatorClient.deleteTeam({ teamId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      toast.success('Team deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete team')
    },
  })
}

export function useTeamMembers(teamId: string) {
  return useQuery({
    queryKey: ['team-members', teamId],
    queryFn: async () => {
      const response = await operatorClient.listTeamMembers({ teamId })
      return response.members
    },
    enabled: !!teamId,
  })
}

export function useAddTeamMember() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { teamId: string; userId: string; role: string }) => {
      await operatorClient.addTeamMember(data)
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', variables.teamId] })
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      toast.success('Member added')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add member')
    },
  })
}

export function useRemoveTeamMember() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { teamId: string; userId: string }) => {
      await operatorClient.removeTeamMember(data)
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', variables.teamId] })
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      toast.success('Member removed')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove member')
    },
  })
}

export function useUpdateTeamMemberRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { teamId: string; userId: string; role: string }) => {
      await operatorClient.updateTeamMemberRole(data)
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['team-members', variables.teamId] })
      toast.success('Member role updated')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update member role')
    },
  })
}

export function useCreateTeam() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { name: string; description?: string }) => {
      await operatorClient.createTeam(data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      toast.success('Team created successfully')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create team')
    },
  })
}
