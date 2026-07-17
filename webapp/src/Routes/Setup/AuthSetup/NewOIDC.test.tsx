import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { NewOIDC } from './NewOIDC'
import { renderWithProviders } from '../../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('NewOIDC', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates an OIDC provider and closes the form', async () => {
    const setShowNewOIDCProvider = vi.fn()
    renderWithProviders(<NewOIDC setShowNewOIDCProvider={setShowNewOIDCProvider} />)

    expect(screen.getByText('New OIDC Provider')).toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText('Client ID'), 'client-123')
    await userEvent.type(screen.getByPlaceholderText('Client Secret'), 'super-secret')
    await userEvent.type(screen.getByPlaceholderText('discoveryURI'), 'https://issuer.example.com/.well-known/openid-configuration')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/oidc')
    expect(postMock.mock.calls[0][1]).toMatchObject({
      name: 'myprovider',
      clientId: 'client-123',
      clientSecret: 'super-secret',
      discoveryURI: 'https://issuer.example.com/.well-known/openid-configuration',
    })
    await waitFor(() => expect(setShowNewOIDCProvider).toHaveBeenCalledWith(false))
  })
})
