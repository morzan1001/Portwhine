import { useState } from 'react'
import {
  useUsers,
  useCreateUser,
  useDeleteUser,
  useUpdateUser,
  useRoles,
  useUserPermissions,
} from '@/hooks/useUsers'
import { useGrantPermission, useRevokePermission } from '@/hooks/usePermissions'
import { timestampToDate } from '@/lib/utils'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  UserPlus,
  Trash,
  Shield,
  Search,
  UserX,
  Plus,
  X,
} from 'lucide-react'
import type { UserInfo, PermissionInfo } from '@/gen/portwhine/v1/operator_pb'

// --- Schemas ---

const createUserSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  role: z.string().min(1, 'Role is required'),
})
type CreateUserForm = z.infer<typeof createUserSchema>

// --- Component ---

export function UsersPage() {
  const { data: users, isLoading } = useUsers()
  const { data: roles } = useRoles()
  const createMutation = useCreateUser()
  const deleteMutation = useDeleteUser()

  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<string | null>(null)
  const [selectedUser, setSelectedUser] = useState<UserInfo | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const form = useForm<CreateUserForm>({
    resolver: zodResolver(createUserSchema),
    defaultValues: { username: '', email: '', password: '', role: 'user' },
  })

  const onSubmit = (data: CreateUserForm) => {
    createMutation.mutate(data, {
      onSuccess: () => {
        setCreateDialogOpen(false)
        form.reset()
      },
    })
  }

  const handleDelete = (userId: string) => {
    deleteMutation.mutate(userId)
    setUserToDelete(null)
    if (selectedUser?.id === userId) setSelectedUser(null)
  }

  const filteredUsers = users?.filter(
    (user: UserInfo) =>
      user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const roleOptions = roles?.map((r) => r.name) || ['user', 'admin']

  return (
    <div className="flex flex-col gap-6 p-8 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage user accounts, roles, and permissions
          </p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <UserPlus className="h-4 w-4 mr-2" />
              Create User
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New User</DialogTitle>
              <DialogDescription>Add a new user to the system</DialogDescription>
            </DialogHeader>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Username</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Email</FormLabel>
                      <FormControl>
                        <Input type="email" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Password</FormLabel>
                      <FormControl>
                        <Input type="password" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="role"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Role</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {roleOptions.map((r) => (
                            <SelectItem key={r} value={r}>
                              {r.charAt(0).toUpperCase() + r.slice(1)}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
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
          placeholder="Search users..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Content */}
      <div className="flex gap-6 min-h-0">
        {/* User Table */}
        <div className="flex-1 min-w-0">
          {isLoading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-14 w-full" />
              ))}
            </div>
          ) : filteredUsers && filteredUsers.length > 0 ? (
            <div className="rounded-xl border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-10" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredUsers.map((user: UserInfo) => (
                    <TableRow
                      key={user.id}
                      className={`cursor-pointer ${selectedUser?.id === user.id ? 'bg-accent' : ''}`}
                      onClick={() => setSelectedUser(user)}
                    >
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar className="h-8 w-8">
                            <AvatarFallback className="bg-primary/10 text-primary text-xs">
                              {user.username
                                .split(' ')
                                .map((n) => n[0])
                                .join('')
                                .toUpperCase()
                                .slice(0, 2)}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="text-sm font-medium">{user.username}</p>
                            <p className="text-xs text-muted-foreground">{user.email}</p>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={user.role === 'admin' ? 'default' : 'secondary'}
                        >
                          {user.role}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {user.isActive ? (
                          <Badge className="bg-[hsl(var(--status-completed))]/10 text-[hsl(var(--status-completed))]">
                            Active
                          </Badge>
                        ) : (
                          <Badge variant="secondary">Inactive</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {timestampToDate(user.createdAt)?.toLocaleDateString() || 'N/A'}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                          onClick={(e) => {
                            e.stopPropagation()
                            setUserToDelete(user.id)
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
              <UserX className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
              <p className="text-sm text-muted-foreground">
                {searchQuery ? 'No users match your search' : 'No users found'}
              </p>
            </div>
          )}
        </div>

        {/* Detail Panel */}
        {selectedUser && (
          <UserDetailPanel
            user={selectedUser}
            roleOptions={roleOptions}
            onClose={() => setSelectedUser(null)}
            onDelete={() => setUserToDelete(selectedUser.id)}
          />
        )}
      </div>

      {/* Delete Confirmation */}
      <AlertDialog open={!!userToDelete} onOpenChange={() => setUserToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the user
              account and remove all associated data.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => userToDelete && handleDelete(userToDelete)}
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

// --- User Detail Panel ---

function UserDetailPanel({
  user,
  roleOptions,
  onClose,
  onDelete,
}: {
  user: UserInfo
  roleOptions: string[]
  onClose: () => void
  onDelete: () => void
}) {
  const updateMutation = useUpdateUser()
  const { data: permissions } = useUserPermissions(user.id)
  const grantPermission = useGrantPermission()
  const revokePermission = useRevokePermission()

  const [editRole, setEditRole] = useState(user.role)
  const [editEmail, setEditEmail] = useState(user.email)
  const [editActive, setEditActive] = useState(user.isActive)
  const [hasChanges, setHasChanges] = useState(false)

  // Permission grant form state
  const [grantResourceType, setGrantResourceType] = useState('pipelines')
  const [grantResourceId, setGrantResourceId] = useState('*')
  const [grantAction, setGrantAction] = useState('*')

  // Reset form when user changes
  const [trackedUserId, setTrackedUserId] = useState(user.id)
  if (trackedUserId !== user.id) {
    setTrackedUserId(user.id)
    setEditRole(user.role)
    setEditEmail(user.email)
    setEditActive(user.isActive)
    setHasChanges(false)
  }

  const handleFieldChange = <T,>(setter: (v: T) => void, value: T) => {
    setter(value)
    setHasChanges(true)
  }

  const handleSave = () => {
    updateMutation.mutate(
      {
        userId: user.id,
        email: editEmail,
        role: editRole,
        isActive: editActive,
      },
      { onSuccess: () => setHasChanges(false) }
    )
  }

  const handleGrantPermission = () => {
    grantPermission.mutate(
      {
        subjectType: 'user',
        subjectId: user.id,
        resourceType: grantResourceType,
        resourceId: grantResourceId,
        action: grantAction,
        effect: 'allow',
      },
      {
        onSuccess: () => {
          setGrantResourceId('*')
          setGrantAction('*')
        },
      }
    )
  }

  const userInitials = user.username
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <div className="w-[400px] shrink-0 border rounded-xl bg-card overflow-hidden flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <Avatar className="h-10 w-10">
              <AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
                {userInitials}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-semibold">{user.username}</p>
              <p className="text-xs text-muted-foreground">{user.email}</p>
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
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>

        {/* General Tab */}
        <TabsContent value="general" className="flex-1 overflow-auto p-4 space-y-4 mt-0">
          <div className="space-y-3">
            <div>
              <label className="text-xs font-medium text-muted-foreground">Email</label>
              <Input
                value={editEmail}
                onChange={(e) =>
                  handleFieldChange(setEditEmail, e.target.value)
                }
                className="mt-1"
              />
            </div>

            <div>
              <label className="text-xs font-medium text-muted-foreground">Role</label>
              <Select
                value={editRole}
                onValueChange={(v) => handleFieldChange(setEditRole, v)}
              >
                <SelectTrigger className="mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {roleOptions.map((r) => (
                    <SelectItem key={r} value={r}>
                      {r.charAt(0).toUpperCase() + r.slice(1)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-muted-foreground">Active</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Inactive users cannot log in
                </p>
              </div>
              <Switch
                checked={editActive}
                onCheckedChange={(v) =>
                  handleFieldChange(setEditActive, v)
                }
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
                  setEditEmail(user.email)
                  setEditRole(user.role)
                  setEditActive(user.isActive)
                  setHasChanges(false)
                }}
              >
                Cancel
              </Button>
            </div>
          )}

          <Separator />

          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">Account Details</p>
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div>
                <p className="text-muted-foreground">User ID</p>
                <p className="font-mono truncate" title={user.id}>
                  {user.id}
                </p>
              </div>
              <div>
                <p className="text-muted-foreground">Created</p>
                <p>
                  {timestampToDate(user.createdAt)?.toLocaleDateString() || 'N/A'}
                </p>
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
            Delete User
          </Button>
        </TabsContent>

        {/* Permissions Tab */}
        <TabsContent value="permissions" className="flex-1 overflow-auto p-4 space-y-4 mt-0">
          {/* Grant new permission */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">Grant Permission</p>
            <div className="flex gap-2">
              <Select value={grantResourceType} onValueChange={setGrantResourceType}>
                <SelectTrigger className="flex-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pipelines">Pipelines</SelectItem>
                  <SelectItem value="runs">Runs</SelectItem>
                  <SelectItem value="workers">Workers</SelectItem>
                  <SelectItem value="users">Users</SelectItem>
                  <SelectItem value="teams">Teams</SelectItem>
                </SelectContent>
              </Select>
              <Select value={grantAction} onValueChange={setGrantAction}>
                <SelectTrigger className="flex-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="*">Full Access</SelectItem>
                  <SelectItem value="read">Read</SelectItem>
                  <SelectItem value="update">Update</SelectItem>
                  <SelectItem value="delete">Delete</SelectItem>
                  <SelectItem value="execute">Execute</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex gap-2">
              <Input
                placeholder="Resource ID (* for all)"
                value={grantResourceId}
                onChange={(e) => setGrantResourceId(e.target.value)}
                className="flex-1"
              />
              <Button
                size="sm"
                onClick={handleGrantPermission}
                disabled={grantPermission.isPending}
              >
                <Plus className="h-4 w-4 mr-1" />
                Grant
              </Button>
            </div>
          </div>

          <Separator />

          {/* Existing permissions */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">
              Active Permissions ({permissions?.length || 0})
            </p>
            {permissions && permissions.length > 0 ? (
              <div className="space-y-2">
                {permissions.map((perm: PermissionInfo) => (
                  <div
                    key={perm.id}
                    className="flex items-center justify-between p-2 rounded-lg border text-xs"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Shield className="h-3.5 w-3.5 text-primary shrink-0" />
                      <div className="min-w-0">
                        <p className="font-medium">
                          <span className="capitalize">{perm.resourceType}</span>
                          {perm.resourceId !== '*' && (
                            <span className="text-muted-foreground ml-1 font-mono">
                              {perm.resourceId.slice(0, 8)}...
                            </span>
                          )}
                        </p>
                        <p className="text-muted-foreground">
                          {perm.action === '*' ? 'Full Access' : perm.action}
                          {' '}
                          <Badge
                            variant={perm.effect === 'allow' ? 'default' : 'destructive'}
                            className="text-[10px] px-1 py-0"
                          >
                            {perm.effect}
                          </Badge>
                        </p>
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 w-7 p-0 shrink-0"
                      onClick={() => revokePermission.mutate(perm.id)}
                    >
                      <X className="h-3.5 w-3.5 text-muted-foreground" />
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground py-4 text-center">
                No permissions assigned
              </p>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
