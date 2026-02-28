import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'

import { operatorClient } from '@/lib/api/client'
import { useAuthStore } from '@/stores/auth'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
})

type LoginForm = z.infer<typeof loginSchema>

export function LoginPage() {
  const navigate = useNavigate()
  const setTokens = useAuthStore((state) => state.setTokens)

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  const loginMutation = useMutation({
    mutationFn: async (data: LoginForm) => {
      const response = await operatorClient.login({
        username: data.username,
        password: data.password,
      })
      return response
    },
    onSuccess: (data) => {
      setTokens({
        accessToken: data.accessToken,
        refreshToken: data.refreshToken,
        expiresAt: data.expiresAt,
        userId: data.userId,
        username: data.username,
        email: data.email,
        role: data.role,
      })
      toast.success('Login successful!')
      navigate('/dashboard')
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Login failed. Please check your credentials.')
    },
  })

  const onSubmit = (data: LoginForm) => {
    loginMutation.mutate(data)
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-sm">
        <Card className="border-border/50">
          <CardHeader className="space-y-4 pb-4">
            <div className="flex justify-center">
              <img src="/logo.svg" alt="Portwhine Logo" className="h-12 w-auto" />
            </div>
            <div className="space-y-1 text-center">
              <CardTitle className="text-xl font-semibold">Portwhine</CardTitle>
              <CardDescription className="text-xs">
                Sign in to your account
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs">Username</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="Enter your username"
                          autoComplete="username"
                          {...field}
                        />
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
                        <Input
                          type="password"
                          placeholder="Enter your password"
                          autoComplete="current-password"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button
                  type="submit"
                  className="w-full"
                  disabled={loginMutation.isPending}
                >
                  {loginMutation.isPending ? 'Signing in...' : 'Sign in'}
                </Button>
              </form>
            </Form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
