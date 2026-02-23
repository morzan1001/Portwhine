import { Outlet, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'

export function AuthLayout() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated())

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  return <Outlet />
}
