import { useState } from 'react'
import { toast } from 'sonner'
import { Trash2, Plus, Shield, Users, User, Loader2 } from 'lucide-react'

import {
  usePipelinePermissions,
  useGrantPermission,
  useRevokePermission,
  useUsers,
  useTeams,
} from '@/hooks/usePermissions'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const ACTIONS = [
  { value: '*', label: 'Full Access' },
  { value: 'read', label: 'Read' },
  { value: 'update', label: 'Edit' },
  { value: 'delete', label: 'Delete' },
  { value: 'execute', label: 'Run' },
]

const ACTION_LABELS: Record<string, string> = {
  '*': 'Full Access',
  read: 'Read',
  update: 'Edit',
  delete: 'Delete',
  execute: 'Run',
}

interface PipelinePermissionsProps {
  pipelineId: string
}

export function PipelinePermissions({ pipelineId }: PipelinePermissionsProps) {
  const { data: permissions, isLoading } = usePipelinePermissions(pipelineId)
  const { data: users } = useUsers()
  const { data: teams } = useTeams()
  const grantPermission = useGrantPermission()
  const revokePermission = useRevokePermission()

  const [showAdd, setShowAdd] = useState(false)
  const [subjectType, setSubjectType] = useState<string>('team')
  const [subjectId, setSubjectId] = useState<string>('')
  const [action, setAction] = useState<string>('read')

  const handleGrant = async () => {
    if (!subjectId || !action) return
    try {
      await grantPermission.mutateAsync({
        subjectType,
        subjectId,
        resourceType: 'pipelines',
        resourceId: pipelineId,
        action,
        effect: 'allow',
      })
      toast.success('Permission granted')
      setShowAdd(false)
      setSubjectId('')
      setAction('read')
    } catch (error: Error) {
      toast.error(error.message || 'Failed to grant permission')
    }
  }

  const handleRevoke = async (permissionId: string) => {
    try {
      await revokePermission.mutateAsync(permissionId)
      toast.success('Permission revoked')
    } catch (error: Error) {
      toast.error(error.message || 'Failed to revoke permission')
    }
  }

  const getSubjectName = (type: string, id: string) => {
    if (type === 'user') {
      const user = users?.find((u) => u.id === id)
      return user?.username || id.slice(0, 8)
    }
    const team = teams?.find((t) => t.id === id)
    return team?.name || id.slice(0, 8)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Existing permissions */}
      {(!permissions || permissions.length === 0) && !showAdd && (
        <p className="text-sm text-muted-foreground text-center py-4">
          No additional permissions. Only the owner and admins can access this pipeline.
        </p>
      )}

      {permissions && permissions.length > 0 && (
        <div className="space-y-2">
          {permissions.map((perm) => (
            <div
              key={perm.id}
              className="flex items-center justify-between rounded-md border px-3 py-2"
            >
              <div className="flex items-center gap-2 min-w-0">
                {perm.subjectType === 'team' ? (
                  <Users className="h-4 w-4 text-muted-foreground shrink-0" />
                ) : (
                  <User className="h-4 w-4 text-muted-foreground shrink-0" />
                )}
                <span className="text-sm font-medium truncate">
                  {getSubjectName(perm.subjectType, perm.subjectId)}
                </span>
                <Badge
                  variant={perm.effect === 'deny' ? 'destructive' : 'secondary'}
                  className="text-xs"
                >
                  {perm.effect === 'deny' ? 'Deny' : ''}{' '}
                  {ACTION_LABELS[perm.action] || perm.action}
                </Badge>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0"
                onClick={() => handleRevoke(perm.id)}
                disabled={revokePermission.isPending}
              >
                <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
              </Button>
            </div>
          ))}
        </div>
      )}

      {/* Add permission form */}
      {showAdd ? (
        <div className="rounded-md border p-3 space-y-3">
          <div className="flex items-center gap-2">
            <Shield className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Grant Access</span>
          </div>

          <div className="grid grid-cols-3 gap-2">
            <Select value={subjectType} onValueChange={setSubjectType}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="team">Team</SelectItem>
                <SelectItem value="user">User</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={subjectId}
              onValueChange={setSubjectId}
            >
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="Select..." />
              </SelectTrigger>
              <SelectContent>
                {subjectType === 'team'
                  ? teams?.map((t) => (
                      <SelectItem key={t.id} value={t.id}>
                        {t.name}
                      </SelectItem>
                    ))
                  : users?.map((u) => (
                      <SelectItem key={u.id} value={u.id}>
                        {u.username}
                      </SelectItem>
                    ))}
              </SelectContent>
            </Select>

            <Select value={action} onValueChange={setAction}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ACTIONS.map((a) => (
                  <SelectItem key={a.value} value={a.value}>
                    {a.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowAdd(false)}
            >
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={handleGrant}
              disabled={!subjectId || grantPermission.isPending}
            >
              {grantPermission.isPending ? 'Granting...' : 'Grant'}
            </Button>
          </div>
        </div>
      ) : (
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={() => setShowAdd(true)}
        >
          <Plus className="h-4 w-4 mr-1.5" />
          Add Permission
        </Button>
      )}
    </div>
  )
}
