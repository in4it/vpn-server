import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { Users } from './Users/Users'
import { NewUserButton } from './Users/NewUserButton'
import { UserCount } from './Users/UserCount'
import { Logs } from './Logs/Logs'
import { Home } from './Home/Home'
import { MaxUsers } from './Home/MaxUsers'
import { UserStats } from './Home/UserStats'
import { renderWithProviders, installFetchMock } from '../test-utils'

vi.mock('react-chartjs-2', () => ({ Chart: () => <div data-testid="chart" /> }))

// Before these queries used fetchJSON, a non-2xx {"error": ...} body was handed
// to the component as if it were data (wrong numbers, enabled buttons, crashes).
describe('fetch based queries take the error path on non-2xx responses', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('Users shows the backend error', async () => {
    installFetchMock([{ match: '/setup/general', response: { error: 'forbidden' }, status: 403 }])
    renderWithProviders(<Users />)
    expect(await screen.findByText("can't load page, error: forbidden")).toBeInTheDocument()
  })

  it('NewUserButton does not offer a new user when the license cannot be read', async () => {
    installFetchMock([{ match: '/license', response: { error: 'license unavailable' }, status: 500 }])
    renderWithProviders(<NewUserButton setShowNewUser={vi.fn()} localAuthDisabled={false} />)
    expect(await screen.findByText('cannot retrieve licensed users')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'New Local User' })).not.toBeInTheDocument()
  })

  it('UserCount shows an error instead of a count', async () => {
    const setUserCount = vi.fn()
    installFetchMock([{ match: '/users', response: { error: 'forbidden' }, status: 403 }])
    renderWithProviders(<UserCount setUserCount={setUserCount} />)
    expect(await screen.findByText('cannot retrieve user count')).toBeInTheDocument()
    expect(setUserCount).not.toHaveBeenCalled()
  })

  it('MaxUsers shows an error instead of a count', async () => {
    const setLicenseUserCount = vi.fn()
    installFetchMock([{ match: '/license', response: { error: 'license unavailable' }, status: 500 }])
    renderWithProviders(<MaxUsers setLicenseUserCount={setLicenseUserCount} />)
    expect(await screen.findByText('cannot retrieve licensed users')).toBeInTheDocument()
    expect(setLicenseUserCount).not.toHaveBeenCalled()
  })

  it('Home shows the license error', async () => {
    installFetchMock([
      { match: '/vpn/stats/user', response: { receivedBytes: { labels: [], datasets: null }, transmitBytes: { labels: [], datasets: null }, handshakes: { labels: [], datasets: null } } },
      { match: '/upgrade', response: { newVersionAvailable: false } },
      { match: '/license', response: { error: 'license unavailable' }, status: 500 },
    ])
    renderWithProviders(<Home />, { authInfo: { role: 'admin' } })
    expect(await screen.findByText('cannot retrieve licensed users: license unavailable')).toBeInTheDocument()
  })

  it('UserStats shows an error instead of crashing', async () => {
    installFetchMock([{ match: '/vpn/stats/user', response: { error: 'stats unavailable' }, status: 500 }])
    renderWithProviders(<UserStats />, { authInfo: { role: 'admin' } })
    expect(await screen.findByText('cannot retrieve user stats: stats unavailable')).toBeInTheDocument()
  })

  it('Logs shows the backend error', async () => {
    installFetchMock([{ match: '/observability/logs', response: { error: 'logs unavailable' }, status: 500 }])
    renderWithProviders(<Logs />, { route: '/logs' })
    expect(await screen.findByText('A backend error has occurred: logs unavailable')).toBeInTheDocument()
  })
})
