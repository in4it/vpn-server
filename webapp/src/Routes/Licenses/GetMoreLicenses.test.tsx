import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GetMoreLicenses } from './GetMoreLicenses'
import { NewUserButton } from '../Users/NewUserButton'
import { renderWithProviders, installFetchMock } from '../../test-utils'

describe('GetMoreLicenses', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/license/get-more', response: { currentUserCount: 3, licenseUserCount: 10, key: 'stripe-key-abc' } },
    ])
  })

  it('shows license usage and refreshes on demand', async () => {
    renderWithProviders(<GetMoreLicenses />)

    expect(await screen.findByText('Get More Licenses')).toBeInTheDocument()
    expect(screen.getByText('Current Users: 3')).toBeInTheDocument()
    expect(screen.getByText('Current User licenses: 10')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Buy More Licenses' })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Refresh License' }))
    expect(await screen.findByText(/Refreshed licenses/)).toBeInTheDocument()
  })

  it('does not share its cache entry with GET /license', async () => {
    const fetchMock = installFetchMock([
      { match: '/license/get-more', response: { currentUserCount: 3, licenseUserCount: 10, key: 'stripe-key-abc' } },
      { match: /\/license$/, response: { currentUserCount: 10, licenseUserCount: 10 } },
    ])
    renderWithProviders(
      <>
        <NewUserButton setShowNewUser={() => {}} localAuthDisabled={false} />
        <GetMoreLicenses />
      </>
    )

    expect(await screen.findByText('Current Users: 3')).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: 'New Local User (license user count reached)' })).toBeDisabled()
    const urls = fetchMock.mock.calls.map((call) => String(call[0]))
    expect(urls).toContain('/api/license')
    expect(urls).toContain('/api/license/get-more')
  })

  it('shows the backend error', async () => {
    installFetchMock([{ match: '/license/get-more', response: { error: 'license server unreachable' }, status: 502 }])
    renderWithProviders(<GetMoreLicenses />)
    expect(await screen.findByText(/license server unreachable/)).toBeInTheDocument()
  })
})
