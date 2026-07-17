import { describe, it, expect, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { Setup } from './Setup'
import { renderWithProviders, installFetchMock } from '../../test-utils'

describe('Setup', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/vpn/setup/vpn', response: { routes: '', vpnEndpoint: '', addressRange: '', clientAddressPrefix: '', port: '', externalInterface: '', nameservers: '', disableNAT: false, enablePacketLogs: false, packetLogsTypes: [], packetLogsRetention: '' } },
      { match: '/vpn/setup/templates', response: { clientTemplate: '', serverTemplate: '' } },
      { match: '/setup/general', response: { hostname: '', enableTLS: false, redirectToHttps: false, disableLocalAuth: false, enableOIDCTokenRenewal: false } },
    ])
  })

  it('renders the setup tabs and the general settings form', async () => {
    renderWithProviders(<Setup />)

    expect(screen.getByText('VPN Setup')).toBeInTheDocument()
    // tab labels
    expect(screen.getByRole('tab', { name: /General/ })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /VPN/ })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Templates/ })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Restart/ })).toBeInTheDocument()
    // general panel content is rendered (re-query inside waitFor: the form
    // remounts its inputs after loading data, which would detach a held node)
    await waitFor(() => expect(screen.getByText('VPN Server Hostname')).toBeInTheDocument())
  })
})
