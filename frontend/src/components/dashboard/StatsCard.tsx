import { Card, CardContent } from '@/components/ui/card'
import { LucideIcon } from 'lucide-react'

interface StatsCardProps {
  title: string
  value: string | number
  description?: string
  icon?: LucideIcon
  trend?: {
    value: number
    isPositive: boolean
  }
}

export function StatsCard({
  title,
  value,
  description,
  icon: Icon,
  trend,
}: StatsCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <p className="text-xs font-medium text-muted-foreground">{title}</p>
            <h3 className="text-2xl font-semibold mt-2">{value}</h3>
            {description && (
              <p className="text-xs text-muted-foreground mt-1">{description}</p>
            )}
            {trend && (
              <p
                className={`text-xs mt-1 font-medium ${
                  trend.isPositive ? 'text-[hsl(var(--status-completed))]' : 'text-[hsl(var(--status-failed))]'
                }`}
              >
                {trend.isPositive ? '+' : ''}
                {trend.value}% from last month
              </p>
            )}
          </div>
          {Icon && (
            <div className="rounded-lg bg-primary/5 p-3">
              <Icon className="h-5 w-5 text-primary" />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
