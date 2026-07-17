import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { NewSAML } from './NewSAML'
import { renderWithProviders } from '../../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('NewSAML', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates a SAML provider and closes the form', async () => {
    const setShowNewSAMLProvider = vi.fn()
    renderWithProviders(<NewSAML setShowNewSAMLProvider={setShowNewSAMLProvider} />)

    expect(screen.getByText('New SAML Provider')).toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText('Metadata URL'), 'https://idp.example.com/metadata')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/saml-setup')
    expect(postMock.mock.calls[0][1]).toMatchObject({
      name: 'samlProvider',
      metadataURL: 'https://idp.example.com/metadata',
      allowMissingAttributes: false,
    })
    await waitFor(() => expect(setShowNewSAMLProvider).toHaveBeenCalledWith(false))
  })
})
