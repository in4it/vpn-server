import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { Auth, useAuthContext } from './Auth'
import { Logout } from './Logout'
import { tokenCookieOptions } from './cookies'
import { handleUnauthorized } from '../queryClient'
import { installFetchMock, renderWithProviders } from '../test-utils'

const { setCookieMock, removeCookieMock, cookieValues } = vi.hoisted(() => ({
  setCookieMock: vi.fn(),
  removeCookieMock: vi.fn(),
  cookieValues: { token: 'abc' } as Record<string, string>,
}))

vi.mock('react-cookie', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-cookie')>()),
  useCookies: () => [cookieValues, setCookieMock, removeCookieMock],
}))

function setProtocol(protocol: string) {
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: { ...window.location, protocol, reload: vi.fn(), assign: vi.fn() },
  })
}

function renderAuth(children: React.ReactNode) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <MantineProvider>
          <Auth>{children}</Auth>
        </MantineProvider>
      </QueryClientProvider>
    </MemoryRouter>
  )
}

describe('Auth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setProtocol('http:')
  })

  it('only renders the app once the role is known', async () => {
    installFetchMock([{ match: '/userinfo', response: { login: 'john', role: 'admin', userType: 'local' } }])
    const rolesSeen: string[] = []
    function App() {
      const { authInfo } = useAuthContext()
      rolesSeen.push(authInfo.role)
      return <div>app for {authInfo.login} ({authInfo.role}, {authInfo.userType}, {authInfo.token})</div>
    }
    renderAuth(<App />)

    expect(await screen.findByText('app for john (admin, local, abc)')).toBeInTheDocument()
    // the app was never rendered with an empty role (no 403 flash / wrong nav)
    expect(rolesSeen).not.toContain('')
  })

  it('sets the token cookie with sameSite=strict', async () => {
    installFetchMock([{ match: '/userinfo', response: { login: 'john', role: 'user', userType: 'local' } }])
    renderAuth(<div>app</div>)
    await screen.findByText('app')
    expect(setCookieMock).toHaveBeenCalledWith('token', 'abc', { path: '/', sameSite: 'strict', secure: false })
  })

  it('shows the login page when the token is rejected', async () => {
    installFetchMock([
      { match: '/userinfo', response: { error: 'token expired' }, status: 401 },
      { match: '/authmethods', response: { localAuthDisabled: false, oidcProviders: [], samlProviders: [] } },
    ])
    renderAuth(<div>app</div>)
    await waitFor(() => expect(screen.queryByText('app')).not.toBeInTheDocument())
    expect(setCookieMock).not.toHaveBeenCalled()
  })
})

describe('token cookie options', () => {
  it('marks the cookie secure on https', () => {
    setProtocol('https:')
    expect(tokenCookieOptions()).toEqual({ path: '/', sameSite: 'strict', secure: true })
  })

  it('does not mark the cookie secure on http', () => {
    setProtocol('http:')
    expect(tokenCookieOptions()).toEqual({ path: '/', sameSite: 'strict', secure: false })
  })
})

describe('Logout', () => {
  it('removes the token cookie instead of setting it to an empty value', async () => {
    vi.clearAllMocks()
    setProtocol('http:')
    renderWithProviders(<Logout />)
    await waitFor(() => expect(removeCookieMock).toHaveBeenCalledWith('token', { path: '/' }))
    expect(setCookieMock).not.toHaveBeenCalled()
    expect(window.location.reload).toHaveBeenCalled()
  })
})

describe('handleUnauthorized', () => {
  it('removes the token cookie and goes back to the login page', () => {
    setProtocol('http:')
    document.cookie = 'token=abc; path=/'
    expect(document.cookie).toContain('token=abc')
    handleUnauthorized()
    expect(document.cookie).not.toContain('token=')
    expect(window.location.assign).toHaveBeenCalledWith('/')
  })
})
