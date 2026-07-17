import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { Restart } from './Restart'
import { renderWithProviders } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('Restart', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('reloads the WireGuard VPN', async () => {
    renderWithProviders(<Restart />)

    await userEvent.click(screen.getByRole('button', { name: /Reload WireGuard/i }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/vpn/setup/restart-vpn')
    expect(await screen.findByText('VPN Restarted!')).toBeInTheDocument()
  })
})
