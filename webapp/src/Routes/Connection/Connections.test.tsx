import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { Connections } from './Connections'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })), delete: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('Connections', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([
      { match: '/vpn/connectionlicense', response: { licenseUserCount: 10, connectionCount: 0 } },
      { match: '/vpn/connections', response: [{ id: 'c1', name: 'laptop' }] },
    ])
  })

  it('lists a regular user\'s connections and creates a new one', async () => {
    renderWithProviders(<Connections />, { authInfo: { login: 'john', role: 'user' } })

    expect(await screen.findByText('VPN Connections')).toBeInTheDocument()
    expect(await screen.findByText('laptop')).toBeInTheDocument()

    const newConnection = await screen.findByRole('button', { name: 'New VPN Connection' })
    await userEvent.click(newConnection)

    await waitFor(() => expect(postMock).toHaveBeenCalledWith(
      '/api/vpn/connections',
      {},
      { headers: { Authorization: 'Bearer test-token' } }
    ))
  })

  it('tells the admin user to log in with another account', () => {
    renderWithProviders(<Connections />, { authInfo: { login: 'admin', role: 'admin' } })
    expect(screen.getByText(/admin user cannot create new connections/i)).toBeInTheDocument()
  })
})
