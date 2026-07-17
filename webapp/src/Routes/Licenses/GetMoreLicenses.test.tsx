import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GetMoreLicenses } from './GetMoreLicenses'
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
})
