import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { Upgrade } from './Upgrade'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('Upgrade', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([
      { match: '/upgrade', response: { currentVersion: '1.0.0', newVersionAvailable: true, newVersion: '1.1.0' } },
    ])
  })

  it('shows the available version and starts an upgrade', async () => {
    renderWithProviders(<Upgrade />)

    expect(await screen.findByText('Upgrade VPN Server')).toBeInTheDocument()
    expect(screen.getByText('Current Version: 1.0.0')).toBeInTheDocument()
    expect(screen.getByText(/New Version available: 1.1.0/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Upgrade' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/upgrade')
    // match the alert body specifically (the alert title also says "Upgrade In Progress")
    expect(await screen.findByText(/Waiting for new version to become available/i)).toBeInTheDocument()
  })
})
