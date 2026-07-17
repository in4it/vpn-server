import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { VPNSetup } from './VPNSetup'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

const vpnSettings = {
  routes: '0.0.0.0/0',
  vpnEndpoint: 'vpn.example.com',
  addressRange: '10.0.0.0/21',
  clientAddressPrefix: '/32',
  port: '51820',
  externalInterface: 'eth0',
  nameservers: '1.1.1.1',
  disableNAT: false,
  enablePacketLogs: false,
  packetLogsTypes: [],
  packetLogsRetention: '',
}

describe('VPNSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/vpn/setup/vpn', response: vpnSettings }])
  })

  it('loads VPN settings and saves them', async () => {
    renderWithProviders(<VPNSetup />)

    expect(await screen.findByText('VPN Endpoint to use')).toBeInTheDocument()
    await waitFor(() => expect(screen.getByDisplayValue('10.0.0.0/21')).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/vpn/setup/vpn')
    expect(await screen.findByText('Settings Saved!')).toBeInTheDocument()
  })
})
