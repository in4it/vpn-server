import { QueryCache, QueryClient } from '@tanstack/react-query'
import { HttpError } from './Auth/fetchJSON'

/**
 * Called when the backend rejects the token (expired, password changed,
 * user suspended): clear the token and go back to the login page.
 */
export function handleUnauthorized() {
  document.cookie = "token=; path=/; max-age=0"
  window.location.assign("/")
}

export function createQueryClient(onUnauthorized: () => void = handleUnauthorized) {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error) => {
        if (error instanceof HttpError && error.status === 401) {
          onUnauthorized()
        }
      },
    }),
    defaultOptions: {
      queries: {
        // client errors (4xx) won't succeed on retry
        retry: (failureCount, error) => {
          if (error instanceof HttpError && error.status < 500) {
            return false
          }
          return failureCount < 3
        },
      },
    },
  })
}
