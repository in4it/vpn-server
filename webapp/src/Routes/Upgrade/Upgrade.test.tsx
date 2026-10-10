import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { act, screen, waitFor } from '@testing-library/react'
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

  describe('while upgrading', () => {
    let upgradeResponse: { status: number; body: unknown }

    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true })
      upgradeResponse = { status: 200, body: { currentVersion: '1.0.0', newVersionAvailable: true, newVersion: '1.1.0' } }
      vi.stubGlobal('fetch', vi.fn(async () => ({
        ok: upgradeResponse.status < 400,
        status: upgradeResponse.status,
        json: async () => upgradeResponse.body,
      })))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    const nextCheck = () => act(async () => { await vi.advanceTimersByTimeAsync(20000) })

    it('only reports completion once the new version answers', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderWithProviders(<Upgrade />)

      await user.click(await screen.findByRole('button', { name: 'Upgrade' }))
      expect(await screen.findByText(/Waiting for new version to become available/i)).toBeInTheDocument()

      // the server is restarting: a 502 is not a completed upgrade
      upgradeResponse = { status: 502, body: { error: 'bad gateway' } }
      await nextCheck()
      expect(screen.queryByText('Upgrade completed')).not.toBeInTheDocument()
      expect(screen.getByText('Current Version: 1.0.0')).toBeInTheDocument()

      // the old server is still answering with the old version
      upgradeResponse = { status: 200, body: { currentVersion: '1.0.0', newVersionAvailable: true, newVersion: '1.1.0' } }
      await nextCheck()
      expect(screen.queryByText('Upgrade completed')).not.toBeInTheDocument()

      // the new version is up
      upgradeResponse = { status: 200, body: { currentVersion: '1.1.0', newVersionAvailable: false, newVersion: '' } }
      await nextCheck()
      expect(await screen.findByText('Upgrade completed')).toBeInTheDocument()
      expect(screen.getByText('Current Version: 1.1.0')).toBeInTheDocument()
      expect(screen.queryByText(/Waiting for new version to become available/i)).not.toBeInTheDocument()
    })
  })

  it('shows an upgrade specific error when the upgrade info cannot be loaded', async () => {
    installFetchMock([{ match: '/upgrade', response: { error: 'forbidden' }, status: 403 }])
    renderWithProviders(<Upgrade />)
    expect(await screen.findByText('cannot retrieve upgrade information: forbidden')).toBeInTheDocument()
  })
})
