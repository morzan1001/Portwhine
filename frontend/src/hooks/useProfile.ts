import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { operatorClient } from '@/lib/api/client'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'sonner'

export function useProfile() {
  const userId = useAuthStore((state) => state.userId)

  return useQuery({
    queryKey: ['profile', userId],
    queryFn: async () => {
      if (!userId) throw new Error('Not authenticated')
      const response = await operatorClient.getUser({ userId })
      return response.user
    },
    enabled: !!userId,
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  const userId = useAuthStore((state) => state.userId)

  return useMutation({
    mutationFn: async (data: { email: string }) => {
      if (!userId) throw new Error('Not authenticated')
      const cached = queryClient.getQueryData(['profile', userId]) as
        | { role: string; isActive: boolean }
        | undefined
      await operatorClient.updateUser({
        userId,
        email: data.email,
        role: cached?.role || '',
        isActive: cached?.isActive ?? true,
      })
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['profile', userId] })
      useAuthStore.setState({ email: variables.email })
      toast.success('Profile updated successfully')
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : 'Failed to update profile'
      toast.error(message)
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (_data: { currentPassword: string; newPassword: string }) => {
      // TODO: Implement when ChangePassword RPC is added to the backend
      throw new Error('Password change is not yet supported by the server')
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : 'Failed to change password'
      toast.error(message)
    },
  })
}
