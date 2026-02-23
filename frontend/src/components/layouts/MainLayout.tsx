import { Outlet, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { Sidebar } from '@/components/layout/Sidebar'

export function MainLayout() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated())

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="h-screen overflow-hidden bg-background">
      <Sidebar />
      <main className="ml-16 h-full overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
