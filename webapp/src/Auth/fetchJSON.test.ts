import { describe, it, expect, afterEach, vi } from 'vitest'
import { fetchJSON, HttpError } from './fetchJSON'
import { installFetchMock } from '../test-utils'

describe('fetchJSON', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns the json body and sends the token', async () => {
    const fetchMock = installFetchMock([{ match: '/api/users', response: [{ id: '1' }] }])
    await expect(fetchJSON('/api/users', 'abc')).resolves.toEqual([{ id: '1' }])
    expect(fetchMock).toHaveBeenCalledWith('/api/users', {
      headers: { "Content-Type": "application/json", "Authorization": "Bearer abc" },
    })
  })

  it('does not send an Authorization header without a token', async () => {
    const fetchMock = installFetchMock([{ match: '/api/authmethods', response: { localAuthDisabled: false } }])
    await fetchJSON('/api/authmethods')
    expect(fetchMock).toHaveBeenCalledWith('/api/authmethods', {
      headers: { "Content-Type": "application/json" },
    })
  })

  it('throws an HttpError with the backend error message', async () => {
    installFetchMock([{ match: '/api/users', response: { error: 'token expired' }, status: 401 }])
    const err = await fetchJSON('/api/users', 'abc').catch((e) => e)
    expect(err).toBeInstanceOf(HttpError)
    expect(err.status).toBe(401)
    expect(err.message).toBe('token expired')
  })

  it('throws an HttpError when the error body is not json', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: false,
      status: 502,
      json: async () => { throw new SyntaxError('Unexpected token <') },
    })))
    const err = await fetchJSON('/api/users', 'abc').catch((e) => e)
    expect(err).toBeInstanceOf(HttpError)
    expect(err.status).toBe(502)
    expect(err.message).toBe('request failed with status 502')
  })
})
