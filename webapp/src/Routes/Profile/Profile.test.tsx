import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { Profile } from './Profile'
import { renderWithProviders, installFetchMock } from '../../test-utils'

describe('Profile', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/profile/factors', response: [{ name: 'phone', type: 'totp' }] },
    ])
  })

  it('shows password change and MFA factors for local users', async () => {
    renderWithProviders(<Profile />, { authInfo: { login: 'john', userType: 'local' } })

    // "Change Password" appears as both a heading and a button; target the heading
    expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
    // the factors list (and its heading) render once the factors request resolves
    expect(await screen.findByText('phone')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Multifactor authentication' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'New Factor (MFA)' })).toBeInTheDocument()
  })

  it('shows no profile info for OIDC users', () => {
    renderWithProviders(<Profile />, { authInfo: { login: 'jane', userType: 'oidc' } })
    expect(screen.getByText('No profile information available for OIDC users.')).toBeInTheDocument()
  })
})
