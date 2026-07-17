import { ReactElement, ReactNode } from 'react'
import { render } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import { AuthContext } from './Auth/Auth'

type AuthInfo = {
  login: string
  role: string
  token: string
  userType: string
}

const defaultAuthInfo: AuthInfo = {
  login: 'admin',
  role: 'admin',
  token: 'test-token',
  userType: 'local',
}

type RenderOptions = {
  authInfo?: Partial<AuthInfo>
  route?: string
}

/**
 * Renders a component wrapped in all providers the app relies on: a fresh
 * react-query client (retries disabled for deterministic tests), MantineProvider,
 * an in-memory router, and an auth context with a configurable authInfo.
 */
export function renderWithProviders(ui: ReactElement, options: RenderOptions = {}) {
  const authInfo = { ...defaultAuthInfo, ...options.authInfo }
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  const wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[options.route ?? '/']}>
      <QueryClientProvider client={queryClient}>
        <MantineProvider>
          <AuthContext.Provider value={{ authInfo, setAuthInfo: vi.fn() }}>
            {children}
          </AuthContext.Provider>
        </MantineProvider>
      </QueryClientProvider>
    </MemoryRouter>
  )

  return render(ui, { wrapper })
}

type FetchRoute = {
  match: string | RegExp
  response: unknown
  status?: number
}

/**
 * Installs a global fetch mock that routes requests by URL. Routes are matched
 * in order (first match wins), so list the most specific paths first. Returns
 * the mock so tests can assert on the calls that were made.
 */
export function installFetchMock(routes: FetchRoute[]) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input)
    const route = routes.find((r) =>
      typeof r.match === 'string' ? url.includes(r.match) : r.match.test(url)
    )
    if (!route) {
      return {
        ok: false,
        status: 404,
        json: async () => ({ error: `no fetch mock for ${url}` }),
        blob: async () => new Blob(),
      } as unknown as Response
    }
    return {
      ok: (route.status ?? 200) < 400,
      status: route.status ?? 200,
      json: async () => route.response,
      blob: async () => new Blob([JSON.stringify(route.response)]),
    } as unknown as Response
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}
