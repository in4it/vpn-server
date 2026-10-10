import { describe, it, expect, vi } from 'vitest'
import { createQueryClient } from './queryClient'
import { HttpError } from './Auth/fetchJSON'

describe('createQueryClient', () => {
  it('calls onUnauthorized when a query fails with 401', async () => {
    const onUnauthorized = vi.fn()
    const client = createQueryClient(onUnauthorized)
    await client.fetchQuery({ queryKey: ['a'], queryFn: () => Promise.reject(new HttpError(401, 'expired')) }).catch(() => {})
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('does not log out on other errors', async () => {
    const onUnauthorized = vi.fn()
    const client = createQueryClient(onUnauthorized)
    await client.fetchQuery({ queryKey: ['b'], queryFn: () => Promise.reject(new HttpError(403, 'forbidden')) }).catch(() => {})
    expect(onUnauthorized).not.toHaveBeenCalled()
  })

  it('does not retry client errors', async () => {
    const client = createQueryClient(vi.fn())
    const retry = client.getDefaultOptions().queries?.retry as (count: number, error: Error) => boolean
    expect(retry(0, new HttpError(400, 'bad request'))).toBe(false)
    expect(retry(0, new HttpError(401, 'expired'))).toBe(false)
    expect(retry(0, new HttpError(500, 'server error'))).toBe(true)
    expect(retry(0, new TypeError('Failed to fetch'))).toBe(true)
    expect(retry(3, new TypeError('Failed to fetch'))).toBe(false)
  })
})
