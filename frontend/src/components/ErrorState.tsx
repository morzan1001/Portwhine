import { AlertCircle, WifiOff, Lock, ServerCrash, FileQuestion, RefreshCw } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

type ErrorType = 'network' | 'permission' | 'server' | 'not-found' | 'generic'

interface ErrorStateProps {
  type?: ErrorType
  title?: string
  message?: string
  onRetry?: () => void
  className?: string
}

const errorConfig = {
  network: {
    icon: WifiOff,
    defaultTitle: 'Connection Error',
    defaultMessage: 'Unable to connect to the server. Please check your internet connection.',
  },
  permission: {
    icon: Lock,
    defaultTitle: 'Access Denied',
    defaultMessage: 'You do not have permission to access this resource.',
  },
  server: {
    icon: ServerCrash,
    defaultTitle: 'Server Error',
    defaultMessage: 'Something went wrong on our end. Please try again later.',
  },
  'not-found': {
    icon: FileQuestion,
    defaultTitle: 'Not Found',
    defaultMessage: 'The resource you are looking for could not be found.',
  },
  generic: {
    icon: AlertCircle,
    defaultTitle: 'Something Went Wrong',
    defaultMessage: 'An unexpected error occurred. Please try again.',
  },
}

export function ErrorState({
  type = 'generic',
  title,
  message,
  onRetry,
  className,
}: ErrorStateProps) {
  const config = errorConfig[type]
  const Icon = config.icon
  const displayTitle = title || config.defaultTitle
  const displayMessage = message || config.defaultMessage

  return (
    <Card className={`max-w-md mx-auto mt-8 ${className || ''}`}>
      <CardContent className="p-8 text-center">
        <div className="rounded-xl bg-destructive/5 p-4 w-14 h-14 mx-auto mb-4 flex items-center justify-center">
          <Icon className="h-6 w-6 text-destructive" />
        </div>
        <h3 className="text-sm font-semibold mb-1">{displayTitle}</h3>
        <p className="text-xs text-muted-foreground mb-4">{displayMessage}</p>
        {onRetry && (
          <Button onClick={onRetry} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-3.5 w-3.5" />
            Try Again
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
