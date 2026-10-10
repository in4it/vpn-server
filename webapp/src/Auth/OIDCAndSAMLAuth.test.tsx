import { describe, it, expect, vi, beforeEach } from 'vitest'
import { StrictMode } from 'react'
import { render, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import axios from 'axios'
import { AuthContext } from './Auth'
import { OIDCAndSAMLAuth } from './OIDCAndSAMLAuth'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: { authenticated: true, token: 'new-token' } })) },
}))

const postMock = vi.mocked(axios.post)

describe('OIDCAndSAMLAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // the backend only accepts a code/state once, so a second request (React
  // StrictMode runs effects twice) would fail the login
  it('exchanges the code only once in StrictMode', async () => {
    const setAuthInfo = vi.fn()
    render(
      <StrictMode>
        <MemoryRouter initialEntries={['/callback/oidc/provider-1?code=abc&state=xyz']}>
          <QueryClientProvider client={new QueryClient()}>
            <AuthContext.Provider value={{ authInfo: { login: '', role: '', token: '', userType: '' }, setAuthInfo }}>
              <Routes>
                <Route path="/callback/:type/:id" element={<OIDCAndSAMLAuth />} />
              </Routes>
            </AuthContext.Provider>
          </QueryClientProvider>
        </MemoryRouter>
      </StrictMode>
    )
    await waitFor(() => expect(setAuthInfo).toHaveBeenCalledWith(expect.objectContaining({ token: 'new-token' })))
    expect(postMock).toHaveBeenCalledTimes(1)
    expect(postMock).toHaveBeenCalledWith('/api/authmethods/oidc/provider-1', {
      code: 'abc',
      state: 'xyz',
      redirectURI: '/callback/oidc/provider-1',
    })
  })
})
