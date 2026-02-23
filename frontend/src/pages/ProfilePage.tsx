import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { useAuthStore } from '@/stores/auth'
import { useProfile, useUpdateProfile, useChangePassword } from '@/hooks/useProfile'
import { useApiKeys, useCreateApiKey, useRevokeApiKey } from '@/hooks/useApiKeys'
import { timestampToDate } from '@/lib/utils'
import type { APIKeyInfo } from '@/gen/portwhine/v1/operator_pb'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ErrorState } from '@/components/ErrorState'
import { Plus, Copy, Trash, Key, LogOut, AlertCircle, Sun, Moon, Monitor } from 'lucide-react'
import { toast } from 'sonner'
import { useTheme } from '@/hooks/useTheme'

// --- Schemas ---

const profileSchema = z.object({
  email: z.string().email('Invalid email address'),
})
type ProfileForm = z.infer<typeof profileSchema>

const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, 'Current password is required'),
    newPassword: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })
type ChangePasswordForm = z.infer<typeof changePasswordSchema>

const createApiKeySchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  scopes: z.string().optional(),
})
type CreateApiKeyForm = z.infer<typeof createApiKeySchema>

// --- Component ---

export function ProfilePage() {
  const navigate = useNavigate()
  const username = useAuthStore((state) => state.username)
  const expiresAt = useAuthStore((state) => state.expiresAt)
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const clearAuth = useAuthStore((state) => state.clearAuth)

  // Theme
  const { theme, setTheme } = useTheme()

  // Profile data
  const { data: profile, isLoading, isError, refetch } = useProfile()
  const updateMutation = useUpdateProfile()
  const changePasswordMutation = useChangePassword()

  // API Keys
  const { data: apiKeys, isLoading: keysLoading } = useApiKeys()
  const createKeyMutation = useCreateApiKey()
  const revokeKeyMutation = useRevokeApiKey()

  // Local state
  const [createKeyDialogOpen, setCreateKeyDialogOpen] = useState(false)
  const [keyToRevoke, setKeyToRevoke] = useState<string | null>(null)
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null)

  // Forms
  const profileForm = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: { email: '' },
  })

  const passwordForm = useForm<ChangePasswordForm>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { currentPassword: '', newPassword: '', confirmPassword: '' },
  })

  const apiKeyForm = useForm<CreateApiKeyForm>({
    resolver: zodResolver(createApiKeySchema),
    defaultValues: { name: '', scopes: '' },
  })

  // Populate profile form when data loads
  useEffect(() => {
    if (profile) {
      profileForm.reset({ email: profile.email })
    }
  }, [profile, profileForm])

  // Derived
  const userInitials =
    username
      ?.split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase() || 'U'

  // Handlers
  const onProfileSubmit = (data: ProfileForm) => {
    updateMutation.mutate(data, {
      onSuccess: () => profileForm.reset({ email: data.email }),
    })
  }

  const onPasswordSubmit = (data: ChangePasswordForm) => {
    changePasswordMutation.mutate(
      { currentPassword: data.currentPassword, newPassword: data.newPassword },
      { onSuccess: () => passwordForm.reset() }
    )
  }

  const onCreateKey = async (data: CreateApiKeyForm) => {
    const scopes = data.scopes
      ? data.scopes
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
      : []
    createKeyMutation.mutate(
      { name: data.name, scopes },
      {
        onSuccess: (response) => {
          setNewlyCreatedKey(response.apiKey)
          setCreateKeyDialogOpen(false)
          apiKeyForm.reset()
        },
      }
    )
  }

  const handleRevokeKey = (keyId: string) => {
    revokeKeyMutation.mutate(keyId)
    setKeyToRevoke(null)
  }

  const handleCopyKey = () => {
    if (newlyCreatedKey) {
      navigator.clipboard.writeText(newlyCreatedKey)
      toast.success('API key copied to clipboard')
    }
  }

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  return (
    <div className="flex flex-col gap-8 p-8 animate-fade-in">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Profile</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your account settings, security, and API keys
        </p>
      </div>

      <div className="max-w-3xl">
        <Tabs defaultValue="profile">
          <TabsList>
            <TabsTrigger value="profile">Profile</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
            <TabsTrigger value="api-keys">API Keys</TabsTrigger>
          </TabsList>

          {/* ===== PROFILE TAB ===== */}
          <TabsContent value="profile" className="space-y-6 mt-6">
            {isLoading ? (
              <ProfileSkeleton />
            ) : isError ? (
              <ErrorState type="server" title="Failed to load profile" onRetry={refetch} />
            ) : profile ? (
              <>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm font-medium">Account Information</CardTitle>
                    <CardDescription className="text-xs">Your personal details</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    <div className="flex items-center gap-4">
                      <Avatar className="h-16 w-16">
                        <AvatarFallback className="bg-primary/10 text-primary text-lg font-medium">
                          {userInitials}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="text-base font-semibold">{profile.username}</p>
                        <Badge
                          variant={profile.role === 'admin' ? 'default' : 'secondary'}
                          className="mt-1"
                        >
                          {profile.role}
                        </Badge>
                      </div>
                    </div>

                    <Separator />

                    <Form {...profileForm}>
                      <form
                        onSubmit={profileForm.handleSubmit(onProfileSubmit)}
                        className="space-y-4"
                      >
                        <FormField
                          control={profileForm.control}
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
                        {profileForm.formState.isDirty && (
                          <div className="flex gap-2">
                            <Button
                              type="submit"
                              size="sm"
                              disabled={updateMutation.isPending}
                            >
                              {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                            </Button>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              onClick={() => profileForm.reset()}
                            >
                              Cancel
                            </Button>
                          </div>
                        )}
                      </form>
                    </Form>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm font-medium">Account Details</CardTitle>
                    <CardDescription className="text-xs">Account metadata</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <div>
                        <p className="text-xs font-medium text-muted-foreground">User ID</p>
                        <p className="font-mono text-xs mt-0.5">{profile.id}</p>
                      </div>
                      <div>
                        <p className="text-xs font-medium text-muted-foreground">Status</p>
                        <div className="mt-0.5">
                          {profile.isActive ? (
                            <Badge className="bg-[hsl(var(--status-completed))]/10 text-[hsl(var(--status-completed))]">
                              Active
                            </Badge>
                          ) : (
                            <Badge variant="secondary">Inactive</Badge>
                          )}
                        </div>
                      </div>
                      <div>
                        <p className="text-xs font-medium text-muted-foreground">Member Since</p>
                        <p className="text-sm mt-0.5">
                          {timestampToDate(profile.createdAt)?.toLocaleDateString() || 'N/A'}
                        </p>
                      </div>
                      <div>
                        <p className="text-xs font-medium text-muted-foreground">Role</p>
                        <p className="text-sm mt-0.5 capitalize">{profile.role}</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm font-medium">Appearance</CardTitle>
                    <CardDescription className="text-xs">
                      Choose how Portwhine looks for you
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="inline-flex items-center rounded-lg border bg-muted p-1 gap-1">
                      {([
                        { value: 'light' as const, label: 'Light', icon: Sun },
                        { value: 'system' as const, label: 'System', icon: Monitor },
                        { value: 'dark' as const, label: 'Dark', icon: Moon },
                      ]).map(({ value, label, icon: Icon }) => (
                        <button
                          key={value}
                          onClick={() => setTheme(value)}
                          className={`
                            flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all
                            ${theme === value
                              ? 'bg-background text-foreground shadow-sm'
                              : 'text-muted-foreground hover:text-foreground'
                            }
                          `}
                        >
                          <Icon className="h-4 w-4" />
                          {label}
                        </button>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </>
            ) : null}
          </TabsContent>

          {/* ===== SECURITY TAB ===== */}
          <TabsContent value="security" className="space-y-6 mt-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">Change Password</CardTitle>
                <CardDescription className="text-xs">
                  Update your password to keep your account secure
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Form {...passwordForm}>
                  <form
                    onSubmit={passwordForm.handleSubmit(onPasswordSubmit)}
                    className="space-y-4"
                  >
                    <FormField
                      control={passwordForm.control}
                      name="currentPassword"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel className="text-xs">Current Password</FormLabel>
                          <FormControl>
                            <Input type="password" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={passwordForm.control}
                      name="newPassword"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel className="text-xs">New Password</FormLabel>
                          <FormControl>
                            <Input type="password" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={passwordForm.control}
                      name="confirmPassword"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel className="text-xs">Confirm New Password</FormLabel>
                          <FormControl>
                            <Input type="password" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button
                      type="submit"
                      size="sm"
                      disabled={changePasswordMutation.isPending}
                    >
                      {changePasswordMutation.isPending ? 'Updating...' : 'Update Password'}
                    </Button>
                  </form>
                </Form>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">Session</CardTitle>
                <CardDescription className="text-xs">Current session information</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">Session Expires</p>
                    <p className="text-sm mt-0.5">
                      {expiresAt ? expiresAt.toLocaleString() : 'Unknown'}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">Authentication</p>
                    <div className="mt-0.5">
                      {isAuthenticated() ? (
                        <Badge className="bg-[hsl(var(--status-completed))]/10 text-[hsl(var(--status-completed))]">
                          Active
                        </Badge>
                      ) : (
                        <Badge variant="destructive">Expired</Badge>
                      )}
                    </div>
                  </div>
                </div>
                <Separator />
                <Button variant="outline" size="sm" onClick={handleLogout}>
                  <LogOut className="h-4 w-4 mr-2" />
                  Sign Out
                </Button>
              </CardContent>
            </Card>
          </TabsContent>

          {/* ===== API KEYS TAB ===== */}
          <TabsContent value="api-keys" className="space-y-6 mt-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-medium">API Keys</h3>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Create and manage API keys for programmatic access
                </p>
              </div>
              <Dialog open={createKeyDialogOpen} onOpenChange={setCreateKeyDialogOpen}>
                <DialogTrigger asChild>
                  <Button size="sm">
                    <Plus className="h-4 w-4 mr-2" />
                    Create Key
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Create API Key</DialogTitle>
                    <DialogDescription>
                      Create a new API key for programmatic access
                    </DialogDescription>
                  </DialogHeader>
                  <Form {...apiKeyForm}>
                    <form
                      onSubmit={apiKeyForm.handleSubmit(onCreateKey)}
                      className="space-y-4"
                    >
                      <FormField
                        control={apiKeyForm.control}
                        name="name"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel className="text-xs">Key Name</FormLabel>
                            <FormControl>
                              <Input placeholder="e.g. CI/CD Pipeline" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={apiKeyForm.control}
                        name="scopes"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel className="text-xs">Scopes (optional)</FormLabel>
                            <FormControl>
                              <Input placeholder="e.g. read,write,execute" {...field} />
                            </FormControl>
                            <FormDescription className="text-xs">
                              Comma-separated list of scopes. Leave empty for full access.
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <div className="flex justify-end gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => setCreateKeyDialogOpen(false)}
                        >
                          Cancel
                        </Button>
                        <Button type="submit" disabled={createKeyMutation.isPending}>
                          {createKeyMutation.isPending ? 'Creating...' : 'Create'}
                        </Button>
                      </div>
                    </form>
                  </Form>
                </DialogContent>
              </Dialog>
            </div>

            {newlyCreatedKey && (
              <Card className="border-primary/20 bg-primary/5">
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <div className="space-y-2 flex-1 min-w-0">
                      <p className="text-sm font-medium">API Key Created</p>
                      <p className="text-xs text-muted-foreground">
                        Copy this key now. You will not be able to see it again.
                      </p>
                      <div className="flex items-center gap-2">
                        <code className="text-xs bg-background px-2 py-1 rounded border flex-1 font-mono break-all">
                          {newlyCreatedKey}
                        </code>
                        <Button size="sm" variant="outline" onClick={handleCopyKey}>
                          <Copy className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}

            {keysLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : apiKeys && apiKeys.length > 0 ? (
              <div className="rounded-xl border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Key Prefix</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead>Last Used</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {apiKeys.map((key: APIKeyInfo) => (
                      <TableRow key={key.id} className={key.revoked ? 'opacity-50' : ''}>
                        <TableCell className="font-medium text-sm">{key.name}</TableCell>
                        <TableCell>
                          <code className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">
                            {key.keyPrefix}...
                          </code>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {timestampToDate(key.createdAt)?.toLocaleDateString() || 'N/A'}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {key.lastUsed
                            ? timestampToDate(key.lastUsed)?.toLocaleDateString()
                            : 'Never'}
                        </TableCell>
                        <TableCell>
                          {key.revoked ? (
                            <Badge variant="destructive">Revoked</Badge>
                          ) : (
                            <Badge className="bg-[hsl(var(--status-completed))]/10 text-[hsl(var(--status-completed))]">
                              Active
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          {!key.revoked && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setKeyToRevoke(key.id)}
                            >
                              <Trash className="h-4 w-4 text-destructive" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center">
                  <Key className="h-8 w-8 text-muted-foreground/50 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">No API keys yet</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Create an API key to access the Portwhine API programmatically
                  </p>
                </CardContent>
              </Card>
            )}

            <AlertDialog open={!!keyToRevoke} onOpenChange={() => setKeyToRevoke(null)}>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Revoke API Key?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. Any applications using this key will lose access
                    immediately.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={() => keyToRevoke && handleRevokeKey(keyToRevoke)}
                    className="bg-destructive text-destructive-foreground"
                  >
                    Revoke
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

// --- Loading Skeleton ---

function ProfileSkeleton() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-4 w-40" />
          <Skeleton className="h-3 w-28" />
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-4">
            <Skeleton className="h-16 w-16 rounded-full" />
            <div className="space-y-2">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-5 w-16" />
            </div>
          </div>
          <Separator />
          <Skeleton className="h-10 w-full" />
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-24" />
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            {[...Array(4)].map((_, i) => (
              <div key={i}>
                <Skeleton className="h-3 w-20" />
                <Skeleton className="h-4 w-28 mt-1" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
