import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { Download } from './Download'
import { ListConnections } from './ListConnections'
import { NewConnection } from './NewConnection'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(), delete: vi.fn() },
}))

const postMock = vi.mocked(axios.post)
const deleteMock = vi.mocked(axios.delete)

describe('Download', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.URL.createObjectURL = vi.fn(() => 'blob:config')
    window.URL.revokeObjectURL = vi.fn()
  })

  it('shows the backend error and does not save it as a .conf file', async () => {
    installFetchMock([{ match: '/vpn/connection/c1', response: { error: 'token expired' }, status: 401 }])
    renderWithProviders(<Download id="c1" name="laptop" />)

    await userEvent.click(screen.getByText('Download New Config'))

    expect(await screen.findByText(/Could not download the config: token expired/)).toBeInTheDocument()
    expect(window.URL.createObjectURL).not.toHaveBeenCalled()
  })

  it('saves the config when the request succeeds', async () => {
    installFetchMock([{ match: '/vpn/connection/c1', response: '[Interface]' }])
    // jsdom can't navigate to the blob url
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    renderWithProviders(<Download id="c1" name="laptop" />)

    await userEvent.click(screen.getByText('Download New Config'))

    await waitFor(() => expect(window.URL.createObjectURL).toHaveBeenCalledTimes(1))
    expect(clickSpy).toHaveBeenCalled()
    expect(screen.queryByText(/Could not download the config/)).not.toBeInTheDocument()
    clickSpy.mockRestore()
  })
})

describe('ListConnections delete', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/vpn/connections', response: [{ id: 'c1', name: 'laptop' }] }])
  })

  it('asks for confirmation before deleting a connection', async () => {
    deleteMock.mockResolvedValue({ data: {} })
    renderWithProviders(<ListConnections />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete connection laptop' }))
    expect(deleteMock).not.toHaveBeenCalled()
    expect(await screen.findByText(/Are you sure you want to delete the VPN connection "laptop"/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith('/api/vpn/connection/c1', expect.anything()))
  })

  it('does not delete when the confirmation is cancelled', async () => {
    renderWithProviders(<ListConnections />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete connection laptop' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Cancel' }))
    expect(deleteMock).not.toHaveBeenCalled()
  })

  it('shows the backend error when the delete fails', async () => {
    deleteMock.mockRejectedValue({ message: 'Request failed with status code 500', response: { data: { error: 'storage unavailable' } } })
    renderWithProviders(<ListConnections />)

    await userEvent.click(await screen.findByRole('button', { name: 'Delete connection laptop' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(await screen.findByText(/An error occurred while deleting: storage unavailable/)).toBeInTheDocument()
  })
})

describe('NewConnection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/vpn/connectionlicense', response: { licenseUserCount: 10, connectionCount: 0 } }])
  })

  it('shows the connection limit error above the button and clears it on retry', async () => {
    postMock
      .mockRejectedValueOnce({ message: 'Request failed with status code 403', response: { status: 403, data: { error: 'connection limit reached (max 2)' } } })
      .mockResolvedValueOnce({ data: {} })
    renderWithProviders(<NewConnection />, { authInfo: { role: 'user' } })

    await userEvent.click(await screen.findByRole('button', { name: 'New VPN Connection' }))
    expect(await screen.findByText('Error: connection limit reached (max 2)')).toBeInTheDocument()
    // the button is still there so the user can retry
    const button = screen.getByRole('button', { name: 'New VPN Connection' })

    await userEvent.click(button)
    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(2))
    await waitFor(() => expect(screen.queryByText(/connection limit reached/)).not.toBeInTheDocument())
  })

  it('shows a backend error when the license info cannot be loaded', async () => {
    installFetchMock([{ match: '/vpn/connectionlicense', response: { error: 'license unavailable' }, status: 500 }])
    renderWithProviders(<NewConnection />, { authInfo: { role: 'user' } })
    expect(await screen.findByText('license unavailable')).toBeInTheDocument()
  })
})
