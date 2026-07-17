import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { GeneralSetup } from './GeneralSetup'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('GeneralSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([
      { match: '/setup/general', response: { hostname: 'vpn.example.com', enableTLS: true, redirectToHttps: false, disableLocalAuth: false, enableOIDCTokenRenewal: false } },
    ])
  })

  it('loads settings and saves them', async () => {
    renderWithProviders(<GeneralSetup />)

    // hostname loaded from the backend into the form
    await waitFor(() => expect(screen.getByDisplayValue('vpn.example.com')).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/setup/general')
    expect(await screen.findByText('Settings Saved!')).toBeInTheDocument()
  })
})
