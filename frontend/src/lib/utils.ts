import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import type { Timestamp } from "@bufbuild/protobuf/wkt"
import { PipelineRunState } from "@/gen/portwhine/v1/operator_pb"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function timestampToDate(timestamp: Timestamp | undefined): Date | undefined {
  if (!timestamp) return undefined
  const seconds = typeof timestamp.seconds === 'bigint' ? Number(timestamp.seconds) : timestamp.seconds
  const nanos = timestamp.nanos || 0
  return new Date(seconds * 1000 + Math.round(nanos / 1e6))
}

export type BadgeStatus = "running" | "completed" | "failed" | "paused" | "pending"

const stateToLabel: Record<PipelineRunState, string> = {
  [PipelineRunState.RUNNING]: 'Running',
  [PipelineRunState.COMPLETED]: 'Completed',
  [PipelineRunState.FAILED]: 'Failed',
  [PipelineRunState.CANCELLED]: 'Cancelled',
  [PipelineRunState.PAUSED]: 'Paused',
  [PipelineRunState.PENDING]: 'Pending',
  [PipelineRunState.UNSPECIFIED]: 'Unknown',
}

const stateToBadge: Record<PipelineRunState, BadgeStatus> = {
  [PipelineRunState.RUNNING]: 'running',
  [PipelineRunState.COMPLETED]: 'completed',
  [PipelineRunState.FAILED]: 'failed',
  [PipelineRunState.CANCELLED]: 'failed',
  [PipelineRunState.PAUSED]: 'paused',
  [PipelineRunState.PENDING]: 'pending',
  [PipelineRunState.UNSPECIFIED]: 'pending',
}

export function runStateLabel(state: PipelineRunState): string {
  return stateToLabel[state] ?? 'Unknown'
}

export function runStateBadge(state: PipelineRunState): BadgeStatus {
  return stateToBadge[state] ?? 'pending'
}
