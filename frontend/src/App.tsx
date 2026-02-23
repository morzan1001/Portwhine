import { Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/hooks/useTheme'

// Layouts
import { AuthLayout } from '@/components/layouts/AuthLayout'
import { MainLayout } from '@/components/layouts/MainLayout'

// Pages
import { LoginPage } from '@/pages/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { PipelinesPage } from '@/pages/PipelinesPage'
import { PipelineEditPage } from '@/pages/PipelineEditPage'
import { RunsPage } from '@/pages/RunsPage'
import { RunDetailPage } from '@/pages/RunDetailPage'
import { UsersPage } from '@/pages/UsersPage'
import { TeamsPage } from '@/pages/TeamsPage'
import { ProfilePage } from '@/pages/ProfilePage'

function App() {
  const hydrate = useAuthStore((state) => state.hydrate)
  useTheme() // Initialize theme

  useEffect(() => {
    hydrate()
  }, [hydrate])

  return (
    <Routes>
      {/* Public routes */}
      <Route element={<AuthLayout />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>

      {/* Protected routes */}
      <Route element={<MainLayout />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/pipelines" element={<PipelinesPage />} />
        <Route path="/pipelines/:id/edit" element={<PipelineEditPage />} />
        <Route path="/runs" element={<RunsPage />} />
        <Route path="/runs/:id" element={<RunDetailPage />} />
        <Route path="/users" element={<UsersPage />} />
        <Route path="/teams" element={<TeamsPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>
    </Routes>
  )
}

export default App
