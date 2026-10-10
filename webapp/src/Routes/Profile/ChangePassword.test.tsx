import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import axios from 'axios'
import { ChangePassword } from './ChangePassword'

// Hoisted mocks so they can be referenced from the vi.mock factories below.
const { setAuthInfoMock, setCookieMock, removeCookieMock } = vi.hoisted(() => ({
  setAuthInfoMock: vi.fn(),
  setCookieMock: vi.fn(),
  removeCookieMock: vi.fn(),
}))

vi.mock('axios', () => ({
  default: { post: vi.fn() },
}))

// Provide a controllable auth context (token used for the Authorization header).
vi.mock('../../Auth/Auth', () => ({
  useAuthContext: () => ({
    authInfo: { login: 'john', role: 'user', token: 'test-token', userType: 'local' },
    setAuthInfo: setAuthInfoMock,
  }),
}))

// Capture cookie writes so we can assert the token is removed on logout.
vi.mock('react-cookie', () => ({
  useCookies: () => [{}, setCookieMock, removeCookieMock],
}))

const postMock = vi.mocked(axios.post)

function renderChangePassword() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MantineProvider>
        <ChangePassword />
      </MantineProvider>
    </QueryClientProvider>
  )
}

describe('ChangePassword', () => {
  const reloadMock = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    // jsdom does not implement navigation; stub reload so logout can call it.
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, reload: reloadMock },
    })
  })

  it('warns the user they will be logged out', () => {
    renderChangePassword()
    expect(
      screen.getByText(/log you out of this and any other active sessions/i)
    ).toBeInTheDocument()
  })

  it('does not submit when the passwords do not match', async () => {
    const user = userEvent.setup()
    renderChangePassword()

    await user.type(screen.getByPlaceholderText('Current Password'), 'oldpass1!')
    await user.type(screen.getByPlaceholderText('New Password'), 'newpass1!')
    await user.type(screen.getByPlaceholderText('Repeat Password'), 'different1!')
    await user.click(screen.getByRole('button', { name: 'Change Password' }))

    expect(screen.getByText("passwords don't match")).toBeInTheDocument()
    expect(postMock).not.toHaveBeenCalled()
  })

  it('does not submit a password without a special character', async () => {
    const user = userEvent.setup()
    renderChangePassword()

    await user.type(screen.getByPlaceholderText('Current Password'), 'oldpass1!')
    await user.type(screen.getByPlaceholderText('New Password'), 'newpass1')
    await user.type(screen.getByPlaceholderText('Repeat Password'), 'newpass1')
    await user.click(screen.getByRole('button', { name: 'Change Password' }))

    expect(screen.getByText(/at least 1 special character/i)).toBeInTheDocument()
    expect(postMock).not.toHaveBeenCalled()
  })

  it('changes the password, then logs out and redirects to login', async () => {
    postMock.mockResolvedValue({ data: { result: 'OK' } })
    const user = userEvent.setup()
    renderChangePassword()

    await user.type(screen.getByPlaceholderText('Current Password'), 'oldpass1!')
    await user.type(screen.getByPlaceholderText('New Password'), 'newpass1!')
    await user.type(screen.getByPlaceholderText('Repeat Password'), 'newpass1!')
    await user.click(screen.getByRole('button', { name: 'Change Password' }))

    // the request is sent to the profile password endpoint with the bearer token
    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock).toHaveBeenCalledWith(
      '/api/profile/password',
      { password: 'newpass1!', currentPassword: 'oldpass1!' },
      { headers: { Authorization: 'Bearer test-token' } }
    )

    // confirmation is shown to the user
    await waitFor(() =>
      expect(
        screen.getByText(/Password updated\. You will be logged out/i)
      ).toBeInTheDocument()
    )

    // after the short delay the session is torn down and the app reloads to login
    await waitFor(
      () => {
        // the cookie is removed (not set to an empty value)
        expect(removeCookieMock).toHaveBeenCalledWith('token', { path: '/' })
        expect(setCookieMock).not.toHaveBeenCalled()
        expect(setAuthInfoMock).toHaveBeenCalledWith({
          login: '',
          role: '',
          token: '',
          userType: '',
        })
        expect(reloadMock).toHaveBeenCalled()
      },
      { timeout: 3000 }
    )
  })

  it('requires the current password', async () => {
    const user = userEvent.setup()
    renderChangePassword()

    await user.type(screen.getByPlaceholderText('New Password'), 'newpass1!')
    await user.type(screen.getByPlaceholderText('Repeat Password'), 'newpass1!')
    await user.click(screen.getByRole('button', { name: 'Change Password' }))

    expect(screen.getByText('current password is empty')).toBeInTheDocument()
    expect(postMock).not.toHaveBeenCalled()
  })

  it('shows the backend error (e.g. wrong current password) and stays logged in', async () => {
    postMock.mockRejectedValue({
      message: 'Request failed with status code 403',
      response: { status: 403, data: { error: 'current password is incorrect' } },
    })
    const user = userEvent.setup()
    renderChangePassword()

    await user.type(screen.getByPlaceholderText('Current Password'), 'wrong1!')
    await user.type(screen.getByPlaceholderText('New Password'), 'newpass1!')
    await user.type(screen.getByPlaceholderText('Repeat Password'), 'newpass1!')
    await user.click(screen.getByRole('button', { name: 'Change Password' }))

    expect(await screen.findByText('current password is incorrect')).toBeInTheDocument()
    expect(screen.queryByText(/Password updated\. You will be logged out/i)).not.toBeInTheDocument()
    expect(removeCookieMock).not.toHaveBeenCalled()
  })
})
