import { describe, it, expect } from 'vitest'
import { parseDateParam } from './dateParam'

describe('parseDateParam', () => {
  it('parses yyyy-mm-dd as a local date', () => {
    const date = parseDateParam('2024-03-05')
    expect(date?.getFullYear()).toBe(2024)
    expect(date?.getMonth()).toBe(2)
    expect(date?.getDate()).toBe(5)
  })

  it.each([null, '', 'not-a-date', '2024-3-5', '2024-02-31', '2024-13-01', '2024-03-05T00:00:00Z'])('rejects %s', (param) => {
    expect(parseDateParam(param)).toBeNull()
  })
})
