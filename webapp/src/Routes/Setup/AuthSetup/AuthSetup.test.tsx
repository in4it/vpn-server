import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { AuthSetup } from './AuthSetup'
import { renderWithProviders, installFetchMock } from '../../../test-utils'

const oidcProvider = {
  id: 'o1',
  name: 'google',
  clientId: 'client-id-123',
  scope: 'openid',
  redirectURI: 'https://vpn.example.com/api/authmethods/oidc/o1/redirect-long',
  loginURL: 'https://accounts.google.com/o/oauth2/v2/auth?client_id=long',
}

const samlProvider = {
  id: 's1',
  name: 'okta',
  audience: 'https://vpn.example.com/saml/aud/s1-longvalue-padding-here',
  issuer: 'https://vpn.example.com/saml/metadata/s1-longvalue-padding',
  acs: 'https://vpn.example.com/saml/acs/s1-longvalue-padding-here-more',
  metadataURL: 'https://okta.example.com/app/metadata/s1-longvalue-padding',
  allowMissingAttributes: false,
}

describe('AuthSetup', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/oidc', response: [oidcProvider] },
      { match: '/saml-setup', response: [samlProvider] },
      { match: '/scim-setup', response: { enabled: false } },
    ])
  })

  it('renders OIDC, SAML and provisioning tabs with their data', async () => {
    renderWithProviders(<AuthSetup />)

    expect(screen.getByText('Authentication & Provisioning')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'New OIDC Connection' })).toBeInTheDocument()

    // OIDC provider row (active tab)
    expect(await screen.findByText('google')).toBeInTheDocument()
    // SAML and provisioning panels are kept mounted, so their content is present too
    expect(await screen.findByText('okta')).toBeInTheDocument()
    expect(await screen.findByText('Enable SCIM v2 endpoint')).toBeInTheDocument()
  })
})
