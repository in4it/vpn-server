import { describe, it, expect, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { Logs } from './Logs'
import { renderWithProviders, installFetchMock } from '../../test-utils'

describe('Logs', () => {
  beforeEach(() => {
    installFetchMock([
      {
        match: '/observability/logs',
        response: {
          enabled: true,
          logEntries: [{ data: 'hello log entry', timestamp: '2026-01-01T00:00:00Z', tags: [] }],
          environments: [],
          nextPos: -1,
          tags: [],
        },
      },
    ])
  })

  it('renders log entries', async () => {
    renderWithProviders(<Logs />, { route: '/logs' })

    expect(await screen.findByText('Logs')).toBeInTheDocument()
    expect(await screen.findByText('hello log entry')).toBeInTheDocument()
    expect(screen.getByText('Timestamp')).toBeInTheDocument()
  })
})
