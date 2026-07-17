import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { TemplateSetup } from './TemplateSetup'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('TemplateSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([
      { match: '/vpn/setup/templates', response: { clientTemplate: 'client-template', serverTemplate: 'server-template' } },
    ])
  })

  it('loads the templates and saves them', async () => {
    renderWithProviders(<TemplateSetup />)

    await waitFor(() => expect(screen.getByDisplayValue('client-template')).toBeInTheDocument())
    expect(screen.getByDisplayValue('server-template')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/vpn/setup/templates')
    expect(await screen.findByText('Settings Saved!')).toBeInTheDocument()
  })
})
