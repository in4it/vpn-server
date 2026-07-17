import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { NewUser } from './NewUser'
import { renderWithProviders } from '../../test-utils'

vi.mock('axios', () => ({
  default: { post: vi.fn(() => Promise.resolve({ data: {} })) },
}))

const postMock = vi.mocked(axios.post)

describe('NewUser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates a local user and returns to the list', async () => {
    const setShowNewUser = vi.fn()
    renderWithProviders(<NewUser setShowNewUser={setShowNewUser} />)

    await userEvent.type(screen.getByPlaceholderText('Login'), 'john')
    await userEvent.type(screen.getByPlaceholderText('Password'), 'secret1!')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1))
    expect(postMock.mock.calls[0][0]).toBe('/api/users')
    expect(postMock.mock.calls[0][1]).toMatchObject({ login: 'john', password: 'secret1!', role: 'user' })
    // onSuccess hides the form again
    await waitFor(() => expect(setShowNewUser).toHaveBeenCalledWith(false))
  })

  it('blocks submit for an invalid (non-alphanumeric) login', async () => {
    const setShowNewUser = vi.fn()
    renderWithProviders(<NewUser setShowNewUser={setShowNewUser} />)

    await userEvent.type(screen.getByPlaceholderText('Login'), 'bad login!')
    await userEvent.type(screen.getByPlaceholderText('Password'), 'secret1!')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))

    expect(await screen.findByText(/Invalid login/i)).toBeInTheDocument()
    expect(postMock).not.toHaveBeenCalled()
  })
})
