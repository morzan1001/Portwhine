import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const statusBadgeVariants = cva(
  "inline-flex items-center rounded-md px-2.5 py-0.5 text-xs font-medium transition-colors focus:outline-none",
  {
    variants: {
      status: {
        running: "bg-[hsl(var(--status-running))]/8 text-[hsl(var(--status-running))]",
        completed: "bg-[hsl(var(--status-completed))]/8 text-[hsl(var(--status-completed))]",
        failed: "bg-[hsl(var(--status-failed))]/8 text-[hsl(var(--status-failed))]",
        paused: "bg-[hsl(var(--status-paused))]/8 text-[hsl(var(--status-paused))]",
        pending: "bg-[hsl(var(--status-pending))]/8 text-[hsl(var(--status-pending))]",
      },
      withDot: {
        true: "pl-1.5",
        false: "",
      },
    },
    defaultVariants: {
      status: "pending",
      withDot: false,
    },
  }
)

export interface StatusBadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof statusBadgeVariants> {}

function StatusBadge({ className, status, withDot, children, ...props }: StatusBadgeProps) {
  return (
    <div className={cn(statusBadgeVariants({ status, withDot }), className)} {...props}>
      {withDot && (
        <span
          className={cn(
            "mr-1.5 h-1.5 w-1.5 rounded-full",
            status === "running" && "bg-[hsl(var(--status-running))] animate-pulse",
            status === "completed" && "bg-[hsl(var(--status-completed))]",
            status === "failed" && "bg-[hsl(var(--status-failed))]",
            status === "paused" && "bg-[hsl(var(--status-paused))]",
            status === "pending" && "bg-[hsl(var(--status-pending))]"
          )}
        />
      )}
      {children}
    </div>
  )
}

export { StatusBadge, statusBadgeVariants }
