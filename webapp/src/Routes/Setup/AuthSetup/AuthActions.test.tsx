import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { ListOIDCProviders } from './ListOIDCProviders'
import { ListSAMLProviders } from './ListSAMLProviders'
import { Provisioning } from './Provisioning'
import { SAML } from './SAML'
import { ListFactors } from '../../Profile/ListFactors'
import { renderWithProviders, installFetchMock } from '../../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

const postMock = vi.mocked(axios.post)
const putMock = vi.mocked(axios.put)
const deleteMock = vi.mocked(axios.delete)

const backendError = (message: string) => ({
  message: 'Request failed with status code 400',
  response: { status: 400, data: { error: message } },
})

// GET /api/oidc no longer returns clientSecret
const oidcProvider = {
  id: 'o1', name: 'okta', clientId: 'client-1', scope: 'openid', discoveryURI: 'https://okta/.well-known',
  redirectURI: 'https://vpn.example.com/callback/oidc/o1', loginURL: 'https://vpn.example.com/login/oidc/o1',
}
const samlProvider = {
  id: 's1', name: 'adfs', metadataURL: 'https://adfs/metadata', allowMissingAttributes: false,
  acs: 'https://vpn.example.com/saml/acs/s1', audience: 'https://vpn.example.com/saml/aud', issuer: 'https://vpn.example.com/saml/issuer',
}

describe('ListOIDCProviders', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/oidc', response: [oidcProvider] }])
  })

  it('renders providers without a clientSecret in the response', async () => {
    renderWithProviders(<ListOIDCProviders />)
    expect(await screen.findByText('okta')).toBeInTheDocument()
    expect(screen.getByText('client-1')).toBeInTheDocument()
  })

  it('asks for confirmation before deleting a provider', async () => {
    deleteMock.mockResolvedValue({ data: {} })
    renderWithProviders(<ListOIDCProviders />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete OIDC provider okta' }))
    expect(deleteMock).not.toHaveBeenCalled()
    expect(await screen.findByText(/delete the OIDC connection "okta"/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith('/api/oidc/o1', expect.anything()))
  })

  it('shows the backend error when the delete fails', async () => {
    deleteMock.mockRejectedValue(backendError('provider in use'))
    renderWithProviders(<ListOIDCProviders />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete OIDC provider okta' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(await screen.findByText('An error occurred while deleting: provider in use')).toBeInTheDocument()
  })
})

describe('ListSAMLProviders', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/saml-setup', response: [samlProvider] }])
  })

  it('asks for confirmation before deleting a provider', async () => {
    deleteMock.mockResolvedValue({ data: {} })
    renderWithProviders(<ListSAMLProviders />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete SAML provider adfs' }))
    expect(deleteMock).not.toHaveBeenCalled()
    await userEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith('/api/saml-setup/s1', expect.anything()))
  })

  it('shows the backend error when updating a provider fails', async () => {
    putMock.mockRejectedValue(backendError('invalid metadata'))
    renderWithProviders(<ListSAMLProviders />)

    await userEvent.click(await screen.findByRole('textbox', { name: 'Allow missing attributes for adfs' }))
    await userEvent.click(await screen.findByRole('option', { name: 'True' }))
    expect(await screen.findByText('Could not update SAML connection: invalid metadata')).toBeInTheDocument()
  })
})

describe('Provisioning (SCIM)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/scim-setup', response: { enabled: true, token: 'scim-token', baseURL: 'https://vpn.example.com/scim' } }])
  })

  it('asks for confirmation before regenerating the token', async () => {
    postMock.mockResolvedValue({ data: {} })
    renderWithProviders(<Provisioning />)

    await userEvent.click(await screen.findByText('Regenerate Token'))
    expect(postMock).not.toHaveBeenCalled()
    expect(await screen.findByText(/The current token stops working immediately/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Regenerate' }))
    await waitFor(() => expect(postMock).toHaveBeenCalledWith('/api/scim-setup', { enabled: true, regenerateToken: true }, expect.anything()))
  })

  it('shows the backend error when toggling SCIM fails', async () => {
    postMock.mockRejectedValue(backendError('scim toggle failed'))
    renderWithProviders(<Provisioning />)

    await userEvent.click(await screen.findByText('Enable SCIM v2 endpoint'))
    expect(await screen.findByText('Could not update SCIM settings: scim toggle failed')).toBeInTheDocument()
  })

  it('shows the backend error when loading fails', async () => {
    installFetchMock([{ match: '/scim-setup', response: { error: 'scim unavailable' }, status: 500 }])
    renderWithProviders(<Provisioning />)
    expect(await screen.findByText(/scim unavailable/)).toBeInTheDocument()
  })
})

describe('SAML setup toggle', () => {
  it('shows the backend error when toggling SAML fails', async () => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/saml-setup', response: { enabled: false, metadataURL: '' } }])
    postMock.mockRejectedValue(backendError('saml toggle failed'))
    renderWithProviders(<SAML />)

    await userEvent.click(await screen.findByText('Enable SAML'))
    expect(await screen.findByText('Could not update SAML settings: saml toggle failed')).toBeInTheDocument()
  })
})

describe('ListFactors', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/profile/factors', response: [{ name: 'phone', type: 'totp' }] }])
  })

  async function clickDeleteFactor() {
    const row = (await screen.findByText('phone')).closest('tr') as HTMLElement
    await userEvent.click(within(row).getByRole('button'))
    await userEvent.click(await screen.findByRole('menuitem', { name: 'Delete Factor' }))
  }

  it('asks for confirmation before deleting a factor', async () => {
    deleteMock.mockResolvedValue({ data: {} })
    renderWithProviders(<ListFactors />)

    await clickDeleteFactor()
    expect(deleteMock).not.toHaveBeenCalled()
    expect(await screen.findByText(/delete the factor "phone"/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith('/api/profile/factors/phone', expect.anything()))
  })

  it('shows the backend error when the delete fails', async () => {
    deleteMock.mockRejectedValue(backendError('factor not found'))
    renderWithProviders(<ListFactors />)

    await clickDeleteFactor()
    await userEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(await screen.findByText('Could not delete factor: factor not found')).toBeInTheDocument()
  })
})
