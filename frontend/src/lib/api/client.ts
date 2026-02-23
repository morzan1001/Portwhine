import { createClient, type Interceptor } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { OperatorService } from '@/gen/portwhine/v1/operator_pb'
import { useAuthStore } from '@/stores/auth'

// Auth interceptor that adds Bearer token to requests
const authInterceptor: Interceptor = (next) => async (req) => {
  const { accessToken, isAuthenticated, clearAuth } = useAuthStore.getState()

  if (isAuthenticated() && accessToken) {
    req.header.set('Authorization', `Bearer ${accessToken}`)
  }

  try {
    return await next(req)
  } catch (error: any) {
    // If we get an unauthenticated error, clear auth and rethrow
    if (error.code === 'unauthenticated') {
      clearAuth()
    }
    throw error
  }
}

// Create transport with base URL and interceptors
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:50051',
  useBinaryFormat: true,
  interceptors: [authInterceptor],
})

// Create and export the client
export const operatorClient = createClient(OperatorService, transport)
