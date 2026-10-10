import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import App from './App'
import { installFetchMock } from './test-utils'

vi.mock('react-chartjs-2', () => ({ Chart: () => <div data-testid="chart" /> }))

describe('App routing', () => {
  it('sends a regular user on the observability server to /logs instead of a blank page', async () => {
    installFetchMock([
      { match: '/context', response: { serverType: 'observability', setupCompleted: true } },
      { match: '/userinfo', response: { login: 'john', role: 'user', userType: 'saml' } },
      { match: '/version', response: { version: 'v1.0.0' } },
      {
        match: '/observability/logs',
        response: { enabled: true, logEntries: [{ data: 'hello log entry', timestamp: '2026-01-01T00:00:00Z', tags: [] }], environments: [], nextPos: -1, tags: [] },
      },
    ])
    // e.g. a bookmarked vpn url: unknown routes go back to the home page,
    // which redirects users to their landing page
    window.history.pushState({}, '', '/connection')
    render(<App />)

    expect(await screen.findByText('hello log entry')).toBeInTheDocument()
    await waitFor(() => expect(window.location.pathname).toBe('/logs'))
    // SAML users have no local password/MFA: no profile link
    expect(screen.queryByText('Profile')).not.toBeInTheDocument()
    expect(screen.getByText('Logout')).toBeInTheDocument()
  })
})
