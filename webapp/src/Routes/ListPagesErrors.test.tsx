import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { ReactElement } from 'react'
import { renderWithProviders, installFetchMock } from '../test-utils'
import { ListConnections } from './Connection/ListConnections'
import { ListUsers } from './Users/ListUsers'
import { ListFactors } from './Profile/ListFactors'
import { ListOIDCProviders } from './Setup/AuthSetup/ListOIDCProviders'
import { ListSAMLProviders } from './Setup/AuthSetup/ListSAMLProviders'

// Before the fix, an error response (e.g. 401 after the token expired) was
// handed to data.map, which threw and unmounted the whole app.
const pages: [string, string, () => ReactElement][] = [
  ['ListConnections', '/vpn/connections', () => <ListConnections />],
  ['ListUsers', '/users', () => <ListUsers localAuthDisabled={false} />],
  ['ListFactors', '/profile/factors', () => <ListFactors />],
  ['ListOIDCProviders', '/oidc', () => <ListOIDCProviders />],
  ['ListSAMLProviders', '/saml-setup', () => <ListSAMLProviders />],
]

describe('list pages with error responses', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  for (const [name, url, render] of pages) {
    for (const status of [400, 401, 500]) {
      it(`${name} shows the backend error on ${status}`, async () => {
        installFetchMock([{ match: url, response: { error: `failure ${status}` }, status }])
        renderWithProviders(render())
        expect(await screen.findByText(new RegExp(`failure ${status}`))).toBeInTheDocument()
      })
    }
  }
})
