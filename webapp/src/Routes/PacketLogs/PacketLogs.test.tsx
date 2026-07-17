import { describe, it, expect } from 'vitest'
import { screen } from '@testing-library/react'
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
})
