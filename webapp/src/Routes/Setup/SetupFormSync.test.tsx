import { describe, it, expect, beforeEach, vi } from 'vitest'
import { act, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ReactElement } from 'react'
import { GeneralSetup } from './GeneralSetup'
import { VPNSetup } from './VPNSetup'
import { TemplateSetup } from './TemplateSetup'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

type FormCase = {
  name: string
  url: string
  queryKey: string
  render: () => ReactElement
  field: string // placeholder or label of the input under test
  initial: Record<string, unknown>
  updated: Record<string, unknown>
  initialValue: string
  updatedValue: string
}

const cases: FormCase[] = [
  {
    name: 'GeneralSetup', url: '/setup/general', queryKey: 'general-setup', render: () => <GeneralSetup />, field: 'Hostname',
    initial: { hostname: 'old.example.com', enableTLS: false, redirectToHttps: false, disableLocalAuth: false, enableOIDCTokenRenewal: false },
    updated: { hostname: 'new.example.com', enableTLS: false, redirectToHttps: false, disableLocalAuth: false, enableOIDCTokenRenewal: false },
    initialValue: 'old.example.com', updatedValue: 'new.example.com',
  },
  {
    name: 'VPNSetup', url: '/vpn/setup/vpn', queryKey: 'vpn-setup', render: () => <VPNSetup />, field: 'hostname',
    initial: { vpnEndpoint: 'old.example.com', routes: '', addressRange: '', clientAddressPrefix: '', port: '', externalInterface: '', nameservers: '', disableNAT: false, enablePacketLogs: false, packetLogsTypes: [], packetLogsRetention: '' },
    updated: { vpnEndpoint: 'new.example.com', routes: '', addressRange: '', clientAddressPrefix: '', port: '', externalInterface: '', nameservers: '', disableNAT: false, enablePacketLogs: false, packetLogsTypes: [], packetLogsRetention: '' },
    initialValue: 'old.example.com', updatedValue: 'new.example.com',
  },
  {
    name: 'TemplateSetup', url: '/vpn/setup/templates', queryKey: 'templates-setup', render: () => <TemplateSetup />, field: 'VPN Client config template',
    initial: { clientTemplate: 'old client template', serverTemplate: 'server' },
    updated: { clientTemplate: 'new client template', serverTemplate: 'server' },
    initialValue: 'old client template', updatedValue: 'new client template',
  },
]

const getField = (field: string) => screen.queryByPlaceholderText(field) ?? screen.getByLabelText(field)

describe.each(cases)('$name form', (c) => {
  let route: { match: string; response: unknown }

  beforeEach(() => {
    vi.clearAllMocks()
    route = { match: c.url, response: c.initial }
    installFetchMock([route])
  })

  it('updates the form when the server data changes', async () => {
    const { queryClient } = renderWithProviders(c.render())
    await waitFor(() => expect(getField(c.field)).toHaveValue(c.initialValue))

    route.response = c.updated
    await act(() => queryClient.invalidateQueries({ queryKey: [c.queryKey] }))

    await waitFor(() => expect(getField(c.field)).toHaveValue(c.updatedValue))
  })

  it('keeps unsaved edits when the server data changes', async () => {
    const { queryClient } = renderWithProviders(c.render())
    await waitFor(() => expect(getField(c.field)).toHaveValue(c.initialValue))

    await userEvent.clear(getField(c.field))
    await userEvent.type(getField(c.field), 'typed by the user')

    route.response = c.updated
    await act(() => queryClient.invalidateQueries({ queryKey: [c.queryKey] }))

    expect(getField(c.field)).toHaveValue('typed by the user')
  })

  it('shows the backend error when loading fails', async () => {
    installFetchMock([{ match: c.url, response: { error: 'settings unavailable' }, status: 500 }])
    renderWithProviders(c.render())
    expect(await screen.findByText(/settings unavailable/)).toBeInTheDocument()
  })
})
