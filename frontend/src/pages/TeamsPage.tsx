import { useState } from 'react'
import {
  useTeams,
  useCreateTeam,
  useUpdateTeam,
  useDeleteTeam,
  useTeamMembers,
  useAddTeamMember,
  useRemoveTeamMember,
  useUpdateTeamMemberRole,
  useUsers,
} from '@/hooks/useUsers'
import { timestampToDate } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  UsersRound,
  Plus,
  Trash,
  Search,
  X,
  UserPlus,
  Crown,
  ShieldCheck,
  User,
} from 'lucide-react'
import type { TeamInfo, TeamMemberInfo } from '@/gen/portwhine/v1/operator_pb'

// --- Schemas ---

const createTeamSchema = z.object({
  name: z.string().min(3, 'Team name must be at least 3 characters'),
  description: z.string().optional(),
})
type CreateTeamForm = z.infer<typeof createTeamSchema>

// --- Helpers ---

function teamRoleIcon(role: string) {
  switch (role) {
    case 'owner':
      return <Crown className="h-3.5 w-3.5" />
    case 'admin':
      return <ShieldCheck className="h-3.5 w-3.5" />
    default:
      return <User className="h-3.5 w-3.5" />
  }
}

// --- Component ---

export function TeamsPage() {
  const { data: teams, isLoading } = useTeams()
  const createMutation = useCreateTeam()
  const deleteMutation = useDeleteTeam()

  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [teamToDelete, setTeamToDelete] = useState<string | null>(null)
  const [selectedTeam, setSelectedTeam] = useState<TeamInfo | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const form = useForm<CreateTeamForm>({
    resolver: zodResolver(createTeamSchema),
    defaultValues: { name: '', description: '' },
  })

  const onSubmit = (data: CreateTeamForm) => {
    createMutation.mutate(data, {
      onSuccess: () => {
        setCreateDialogOpen(false)
        form.reset()
      },
    })
  }

  const handleDelete = (teamId: string) => {
    deleteMutation.mutate(teamId)
    setTeamToDelete(null)
    if (selectedTeam?.id === teamId) setSelectedTeam(null)
  }

  const filteredTeams = teams?.filter(
    (team: TeamInfo) =>
      team.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (team.description || '').toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="flex flex-col gap-6 p-8 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Teams</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage teams, members, and collaborate on pipelines
          </p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Create Team
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Team</DialogTitle>
              <DialogDescription>
                Create a team to collaborate with other users
              </DialogDescription>
            </DialogHeader>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Team Name</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Description</FormLabel>
                      <FormControl>
                        <Textarea {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="flex justify-end gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setCreateDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createMutation.isPending}>
                    {createMutation.isPending ? 'Creating...' : 'Create'}
                  </Button>
                </div>
              </form>
            </Form>
          </DialogContent>
        </Dialog>
      </div>

      {/* Search */}
      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search teams..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Content */}
      <div className="flex gap-6 min-h-0">
        {/* Team Table */}
        <div className="flex-1 min-w-0">
          {isLoading ? (
            <div className="space-y-2">
              {[...Array(4)].map((_, i) => (
                <Skeleton key={i} className="h-14 w-full" />
              ))}
            </div>
          ) : filteredTeams && filteredTeams.length > 0 ? (
            <div className="rounded-xl border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Team</TableHead>
                    <TableHead>Members</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-10" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredTeams.map((team: TeamInfo) => (
                    <TableRow
                      key={team.id}
                      className={`cursor-pointer ${selectedTeam?.id === team.id ? 'bg-accent' : ''}`}
                      onClick={() => setSelectedTeam(team)}
                    >
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <div className="rounded-lg bg-primary/5 p-2">
                            <UsersRound className="h-4 w-4 text-primary" />
                          </div>
                          <div>
                            <p className="text-sm font-medium">{team.name}</p>
                            {team.description && (
                              <p className="text-xs text-muted-foreground truncate max-w-[300px]">
                                {team.description}
                              </p>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">
                          {team.memberCount || 0} members
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {timestampToDate(team.createdAt)?.toLocaleDateString() || 'N/A'}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                          onClick={(e) => {
                            e.stopPropagation()
                            setTeamToDelete(team.id)
                          }}
                        >
                          <Trash className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-16">
              <UsersRound className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
              <p className="text-sm text-muted-foreground">
                {searchQuery ? 'No teams match your search' : 'No teams found'}
              </p>
            </div>
          )}
        </div>

        {/* Detail Panel */}
        {selectedTeam && (
          <TeamDetailPanel
            team={selectedTeam}
            onClose={() => setSelectedTeam(null)}
            onDelete={() => setTeamToDelete(selectedTeam.id)}
          />
        )}
      </div>

      {/* Delete Confirmation */}
      <AlertDialog open={!!teamToDelete} onOpenChange={() => setTeamToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Team?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. All team members will lose access granted
              through this team.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => teamToDelete && handleDelete(teamToDelete)}
              className="bg-destructive text-destructive-foreground"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

// --- Team Detail Panel ---

function TeamDetailPanel({
  team,
  onClose,
  onDelete,
}: {
  team: TeamInfo
  onClose: () => void
  onDelete: () => void
}) {
  const updateMutation = useUpdateTeam()
  const { data: members, isLoading: membersLoading } = useTeamMembers(team.id)
  const { data: allUsers } = useUsers()
  const addMember = useAddTeamMember()
  const removeMember = useRemoveTeamMember()
  const updateMemberRole = useUpdateTeamMemberRole()

  const [editName, setEditName] = useState(team.name)
  const [editDescription, setEditDescription] = useState(team.description)
  const [hasChanges, setHasChanges] = useState(false)

  // Add member form
  const [addUserId, setAddUserId] = useState('')
  const [addRole, setAddRole] = useState('member')

  // Member to remove
  const [memberToRemove, setMemberToRemove] = useState<string | null>(null)

  // Reset when team changes
  const [trackedTeamId, setTrackedTeamId] = useState(team.id)
  if (trackedTeamId !== team.id) {
    setTrackedTeamId(team.id)
    setEditName(team.name)
    setEditDescription(team.description)
    setHasChanges(false)
  }

  const handleSave = () => {
    updateMutation.mutate(
      { teamId: team.id, name: editName, description: editDescription },
      { onSuccess: () => setHasChanges(false) }
    )
  }

  const handleAddMember = () => {
    if (!addUserId) return
    addMember.mutate(
      { teamId: team.id, userId: addUserId, role: addRole },
      {
        onSuccess: () => {
          setAddUserId('')
          setAddRole('member')
        },
      }
    )
  }

  const handleRemoveMember = (userId: string) => {
    removeMember.mutate({ teamId: team.id, userId })
    setMemberToRemove(null)
  }

  // Filter out users already in the team
  const memberUserIds = new Set(members?.map((m: TeamMemberInfo) => m.userId) || [])
  const availableUsers = allUsers?.filter((u) => !memberUserIds.has(u.id)) || []

  return (
    <div className="w-[420px] shrink-0 border rounded-xl bg-card overflow-hidden flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-primary/5 p-2.5">
              <UsersRound className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="text-sm font-semibold">{team.name}</p>
              <p className="text-xs text-muted-foreground">
                {team.memberCount || 0} members
              </p>
            </div>
          </div>
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="general" className="flex-1 flex flex-col min-h-0">
        <TabsList className="mx-4 mt-3">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="members">Members</TabsTrigger>
        </TabsList>

        {/* General Tab */}
        <TabsContent value="general" className="flex-1 overflow-auto p-4 space-y-4 mt-0">
          <div className="space-y-3">
            <div>
              <label className="text-xs font-medium text-muted-foreground">Name</label>
              <Input
                value={editName}
                onChange={(e) => {
                  setEditName(e.target.value)
                  setHasChanges(true)
                }}
                className="mt-1"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground">
                Description
              </label>
              <Textarea
                value={editDescription}
                onChange={(e) => {
                  setEditDescription(e.target.value)
                  setHasChanges(true)
                }}
                className="mt-1"
                rows={3}
              />
            </div>
          </div>

          {hasChanges && (
            <div className="flex gap-2">
              <Button size="sm" onClick={handleSave} disabled={updateMutation.isPending}>
                {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  setEditName(team.name)
                  setEditDescription(team.description)
                  setHasChanges(false)
                }}
              >
                Cancel
              </Button>
            </div>
          )}

          <Separator />

          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">Details</p>
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div>
                <p className="text-muted-foreground">Team ID</p>
                <p className="font-mono truncate" title={team.id}>
                  {team.id}
                </p>
              </div>
              <div>
                <p className="text-muted-foreground">Created</p>
                <p>{timestampToDate(team.createdAt)?.toLocaleDateString() || 'N/A'}</p>
              </div>
            </div>
          </div>

          <Separator />

          <Button
            variant="outline"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={onDelete}
          >
            <Trash className="h-4 w-4 mr-2" />
            Delete Team
          </Button>
        </TabsContent>

        {/* Members Tab */}
        <TabsContent value="members" className="flex-1 overflow-auto p-4 space-y-4 mt-0">
          {/* Add member */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">Add Member</p>
            <div className="flex gap-2">
              <Select value={addUserId} onValueChange={setAddUserId}>
                <SelectTrigger className="flex-1">
                  <SelectValue placeholder="Select user..." />
                </SelectTrigger>
                <SelectContent>
                  {availableUsers.map((u) => (
                    <SelectItem key={u.id} value={u.id}>
                      {u.username}
                    </SelectItem>
                  ))}
                  {availableUsers.length === 0 && (
                    <div className="px-2 py-1.5 text-xs text-muted-foreground">
                      No users available
                    </div>
                  )}
                </SelectContent>
              </Select>
              <Select value={addRole} onValueChange={setAddRole}>
                <SelectTrigger className="w-[110px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="member">Member</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="owner">Owner</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Button
              size="sm"
              onClick={handleAddMember}
              disabled={!addUserId || addMember.isPending}
              className="w-full"
            >
              <UserPlus className="h-4 w-4 mr-1.5" />
              {addMember.isPending ? 'Adding...' : 'Add Member'}
            </Button>
          </div>

          <Separator />

          {/* Member list */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">
              Members ({members?.length || 0})
            </p>
            {membersLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : members && members.length > 0 ? (
              <div className="space-y-1">
                {members.map((member: TeamMemberInfo) => (
                  <div
                    key={member.userId}
                    className="flex items-center justify-between p-2 rounded-lg border"
                  >
                    <div className="flex items-center gap-2.5 min-w-0">
                      <Avatar className="h-7 w-7">
                        <AvatarFallback className="bg-primary/10 text-primary text-[10px]">
                          {member.username
                            .split(' ')
                            .map((n) => n[0])
                            .join('')
                            .toUpperCase()
                            .slice(0, 2)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="min-w-0">
                        <p className="text-xs font-medium truncate">{member.username}</p>
                        <p className="text-[10px] text-muted-foreground truncate">
                          {member.email}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <Select
                        value={member.role}
                        onValueChange={(newRole) =>
                          updateMemberRole.mutate({
                            teamId: team.id,
                            userId: member.userId,
                            role: newRole,
                          })
                        }
                      >
                        <SelectTrigger className="h-7 w-[100px] text-xs">
                          <div className="flex items-center gap-1.5">
                            {teamRoleIcon(member.role)}
                            <SelectValue />
                          </div>
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="member">Member</SelectItem>
                          <SelectItem value="admin">Admin</SelectItem>
                          <SelectItem value="owner">Owner</SelectItem>
                        </SelectContent>
                      </Select>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 w-7 p-0"
                        onClick={() => setMemberToRemove(member.userId)}
                      >
                        <X className="h-3.5 w-3.5 text-muted-foreground" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground py-4 text-center">
                No members yet
              </p>
            )}
          </div>
        </TabsContent>
      </Tabs>

      {/* Remove member confirmation */}
      <AlertDialog
        open={!!memberToRemove}
        onOpenChange={() => setMemberToRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Member?</AlertDialogTitle>
            <AlertDialogDescription>
              This user will lose all access granted through this team.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => memberToRemove && handleRemoveMember(memberToRemove)}
              className="bg-destructive text-destructive-foreground"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
