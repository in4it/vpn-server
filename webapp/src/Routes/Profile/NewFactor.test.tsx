import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { NewFactor } from './NewFactor'
import { renderWithProviders } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('NewFactor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('adds a TOTP factor and returns to the profile', async () => {
    const setShowNewFactor = vi.fn()
    renderWithProviders(<NewFactor setShowNewFactor={setShowNewFactor} secret="JBSWY3DPEHPK3PXP" />)

    expect(screen.getByText('New Security Factor (MFA)')).toBeInTheDocument()

    // the form has two text inputs: name first, then the 6-digit code
    const [nameInput, codeInput] = screen.getAllByRole('textbox')
    await userEvent.type(nameInput, 'google-auth')
    await userEvent.type(codeInput, '123456')
    await userEvent.click(screen.getByRole('button', { name: 'Add' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/profile/factors')
    expect(postMock.mock.calls[0][1]).toMatchObject({
      name: 'google-auth',
      code: '123456',
      secret: 'JBSWY3DPEHPK3PXP',
      type: 'totp',
    })
    await waitFor(() => expect(setShowNewFactor).toHaveBeenCalledWith(false))
  })

  it('rejects a non-numeric code', async () => {
    const setShowNewFactor = vi.fn()
    renderWithProviders(<NewFactor setShowNewFactor={setShowNewFactor} secret="JBSWY3DPEHPK3PXP" />)

    const [nameInput, codeInput] = screen.getAllByRole('textbox')
    await userEvent.type(nameInput, 'google-auth')
    await userEvent.type(codeInput, 'abc')
    await userEvent.click(screen.getByRole('button', { name: 'Add' }))

    expect(await screen.findByText(/Invalid code/i)).toBeInTheDocument()
    expect(postMock).not.toHaveBeenCalled()
  })
})
