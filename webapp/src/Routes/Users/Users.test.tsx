import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { Users } from './Users'
import { renderWithProviders, installFetchMock } from '../../test-utils'

const user = {
  id: '1',
  login: 'john',
  role: 'user',
  oidcID: '',
  samlID: '',
  provisioned: false,
  suspended: false,
  lastLogin: '',
  lastTokenRenewal: '',
  connectionsDisabledOnAuthFailure: false,
}

describe('Users', () => {
  beforeEach(() => {
    installFetchMock([
      { match: '/setup/general', response: { disableLocalAuth: false } },
      { match: '/users', response: [user] },
      { match: '/license', response: { currentUserCount: 1, licenseUserCount: 10 } },
    ])
  })

  it('lists users and offers to add a new local user', async () => {
    renderWithProviders(<Users />)

    expect(await screen.findByText('john')).toBeInTheDocument()
    // column headers of the users table
    expect(screen.getByText('Login')).toBeInTheDocument()
    expect(screen.getByText('Last Web Login')).toBeInTheDocument()
    // the add button is enabled because local auth is on and licenses remain
    const newUserButton = await screen.findByRole('button', { name: 'New Local User' })
    expect(newUserButton).toBeEnabled()
  })
})
