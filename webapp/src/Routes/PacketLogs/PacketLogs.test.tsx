import { describe, it, expect } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { PacketLogs } from './PacketLogs'
import { renderWithProviders, installFetchMock } from '../../test-utils'

describe('PacketLogs', () => {
  it('renders the packet logs page when logging is enabled', async () => {
    installFetchMock([
      {
        match: '/vpn/stats/packetlogs',
        response: {
          enabled: true,
          logTypes: ['dns', 'tcp'],
          logData: { schema: { columns: [] }, rows: [], nextPos: -1 },
          users: { all: 'All Users', u1: 'john' },
        },
      },
    ])

    renderWithProviders(<PacketLogs />, { route: '/packetlogs' })

    expect(await screen.findByText('Packet Logs')).toBeInTheDocument()
    // with the default "all" user, the table prompts to pick a user
    expect(await screen.findByText('Select a user to see log data.')).toBeInTheDocument()
    expect(screen.getByText('Timestamp')).toBeInTheDocument()
  })

  it('shows an activation prompt when packet logging is disabled', async () => {
    installFetchMock([
      {
        match: '/vpn/stats/packetlogs',
        response: {
          enabled: false,
          logTypes: [],
          logData: { schema: { columns: [] }, rows: [], nextPos: -1 },
          users: {},
        },
      },
    ])

    renderWithProviders(<PacketLogs />, { route: '/packetlogs' })

    expect(await screen.findByText(/Packet Logs are not activated/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /VPN Settings/i })).toBeInTheDocument()
  })

  describe('query parameters', () => {
    const enabledResponse = {
      enabled: true,
      logTypes: ['dns', 'tcp'],
      logData: { schema: { columns: [] }, rows: [{ t: '2024-03-05T10:00:00Z', d: ['dns', '10.0.0.2'] }], nextPos: -1 },
      users: { all: 'All Users', 'u1-abc': 'john' },
    }
    const requestedURLs = (fetchMock: ReturnType<typeof installFetchMock>) => fetchMock.mock.calls.map((call) => String(call[0]))

    it('does not put a path traversal user param in the API path', async () => {
      const fetchMock = installFetchMock([{ match: '/vpn/stats/packetlogs', response: enabledResponse }])
      renderWithProviders(<PacketLogs />, { route: '/packetlogs?user=../../setup/vpn' })

      expect(await screen.findByText('Select a user to see log data.')).toBeInTheDocument()
      const urls = requestedURLs(fetchMock)
      expect(urls.length).toBeGreaterThan(0)
      for (const url of urls) {
        expect(url).toMatch(/^\/api\/vpn\/stats\/packetlogs\/all\//)
        expect(url).not.toContain('setup')
      }
    })

    it('loads the logs of a user from the users list', async () => {
      const fetchMock = installFetchMock([{ match: '/vpn/stats/packetlogs', response: enabledResponse }])
      renderWithProviders(<PacketLogs />, { route: '/packetlogs?user=u1-abc&date=2024-03-05' })

      expect(await screen.findByText('10.0.0.2')).toBeInTheDocument()
      expect(requestedURLs(fetchMock)[0]).toMatch(/^\/api\/vpn\/stats\/packetlogs\/u1-abc\/2024-03-05\?/)
    })

    it('rejects a user that is not in the users list', async () => {
      installFetchMock([{ match: '/vpn/stats/packetlogs', response: enabledResponse }])
      renderWithProviders(<PacketLogs />, { route: '/packetlogs?user=not-a-user' })

      expect(await screen.findByText(/Unknown user in the URL/)).toBeInTheDocument()
      expect(screen.getByText('Select a user to see log data.')).toBeInTheDocument()
      expect(screen.queryByText('10.0.0.2')).not.toBeInTheDocument()
    })

    it('falls back to today for an invalid date param', async () => {
      const fetchMock = installFetchMock([{ match: '/vpn/stats/packetlogs', response: enabledResponse }])
      renderWithProviders(<PacketLogs />, { route: '/packetlogs?date=not-a-date' })

      await waitFor(() => expect(fetchMock).toHaveBeenCalled())
      const now = new Date()
      const today = now.getFullYear() + '-' + String(now.getMonth() + 1).padStart(2, '0') + '-' + String(now.getDate()).padStart(2, '0')
      const url = requestedURLs(fetchMock)[0]
      expect(url).not.toContain('NaN')
      expect(url).toContain('/all/' + today + '?')
    })

    it('shows the backend error', async () => {
      installFetchMock([{ match: '/vpn/stats/packetlogs', response: { error: 'storage unavailable' }, status: 500 }])
      renderWithProviders(<PacketLogs />, { route: '/packetlogs' })
      expect(await screen.findByText(/storage unavailable/)).toBeInTheDocument()
    })
  })
})
