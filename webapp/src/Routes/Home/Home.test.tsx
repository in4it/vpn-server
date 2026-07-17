import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { Home } from './Home'
import { renderWithProviders, installFetchMock } from '../../test-utils'

// Chart.js needs a real <canvas>; stub the chart so jsdom doesn't choke on it.
vi.mock('react-chartjs-2', () => ({ Chart: () => <div data-testid="chart" /> }))

describe('Home', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/vpn/stats/user', response: { receivedBytes: { labels: [], datasets: null }, transmitBytes: { labels: [], datasets: null }, handshakes: { labels: [], datasets: null } } },
      { match: '/upgrade', response: { newVersionAvailable: false } },
      { match: '/license', response: { currentUserCount: 2, licenseUserCount: 10, cloudType: 'other' } },
    ])
  })

  it('renders the VPN status with active/licensed user counts', async () => {
    renderWithProviders(<Home />, { authInfo: { role: 'admin' } })

    expect(await screen.findByText('VPN Status')).toBeInTheDocument()
    await waitFor(() => expect(screen.getByText('2 / 10')).toBeInTheDocument())
    expect(screen.getByText('Get more licenses')).toBeInTheDocument()
  })

  it('redirects regular users away from the admin home', () => {
    renderWithProviders(<Home />, { authInfo: { role: 'user' }, route: '/' })
    // "user" role is redirected to /connection, so the status title never renders
    expect(screen.queryByText('VPN Status')).not.toBeInTheDocument()
  })
})
