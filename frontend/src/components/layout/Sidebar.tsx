import { Link, useLocation, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard,
  Workflow,
  PlayCircle,
  Users,
  UsersRound,
  UserCircle,
  LogOut,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { cn } from '@/lib/utils'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const navigation = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: LayoutDashboard,
  },
  {
    name: 'Pipelines',
    href: '/pipelines',
    icon: Workflow,
  },
  {
    name: 'Runs',
    href: '/runs',
    icon: PlayCircle,
  },
  {
    name: 'Users',
    href: '/users',
    icon: Users,
  },
  {
    name: 'Teams',
    href: '/teams',
    icon: UsersRound,
  },
]

export function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const username = useAuthStore((state) => state.username)
  const role = useAuthStore((state) => state.role)
  const clearAuth = useAuthStore((state) => state.clearAuth)

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  const userInitials = username
    ?.split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase() || 'U'

  return (
    <div className="group/sidebar fixed inset-y-0 left-0 z-30 flex w-16 flex-col border-r bg-card transition-all duration-300 ease-in-out hover:w-56">
      {/* Logo */}
      <div className="flex h-14 items-center border-b px-4">
        <img src="/logo.svg" alt="Portwhine" className="h-7 w-7 shrink-0" />
        <span className="ml-3 text-sm font-semibold tracking-tight text-foreground opacity-0 transition-opacity duration-300 group-hover/sidebar:opacity-100 truncate">
          Portwhine
        </span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-2 py-3">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href || location.pathname.startsWith(item.href + '/')
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors duration-150',
                isActive
                  ? 'text-primary bg-primary/5'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent'
              )}
            >
              {isActive && (
                <div className="absolute left-0 top-1/2 -translate-y-1/2 h-5 w-0.5 rounded-full bg-primary" />
              )}
              <item.icon className="h-[18px] w-[18px] shrink-0" />
              <span className="opacity-0 transition-opacity duration-300 group-hover/sidebar:opacity-100 truncate">
                {item.name}
              </span>
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="border-t px-2 py-2 space-y-1">
        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150 hover:bg-accent">
              <Avatar className="h-7 w-7 shrink-0">
                <AvatarFallback className="bg-primary/10 text-primary text-xs font-medium">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col items-start opacity-0 transition-opacity duration-300 group-hover/sidebar:opacity-100 truncate">
                <span className="text-sm font-medium text-foreground truncate">{username || 'User'}</span>
                <span className="text-xs text-muted-foreground">{role || 'user'}</span>
              </div>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" side="right" className="w-48">
            <DropdownMenuLabel className="text-xs text-muted-foreground">My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => navigate('/profile')} className="cursor-pointer text-sm">
              <UserCircle className="mr-2 h-4 w-4" />
              Profile
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-sm text-destructive focus:text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
