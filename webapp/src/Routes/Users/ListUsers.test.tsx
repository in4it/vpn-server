import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { ListUsers } from './ListUsers'
import { renderWithProviders, installFetchMock } from '../../test-utils'

vi.mock('axios', () => ({
  default: { patch: vi.fn(), delete: vi.fn() },
}))

const patchMock = vi.mocked(axios.patch)
const deleteMock = vi.mocked(axios.delete)

const baseUser = {
  role: 'user',
  oidcID: '',
  samlID: '',
  provisioned: false,
  suspended: false,
  lastLogin: '',
  lastTokenRenewal: '',
  connectionsDisabledOnAuthFailure: false,
}
const users = [
  { ...baseUser, id: 'a', login: 'alice' },
  { ...baseUser, id: 'b', login: 'bob' },
]

const backendError = (message: string) => ({
  message: 'Request failed with status code 400',
  response: { status: 400, data: { error: message } },
})

async function openMenuItem(login: string, item: string) {
  const row = (await screen.findByText(login)).closest('tr') as HTMLElement
  await userEvent.click(within(row).getByRole('button'))
  await userEvent.click(await screen.findByRole('menuitem', { name: item }))
}

describe('ListUsers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    installFetchMock([{ match: '/users', response: users }])
  })

  describe('change password modal', () => {
    it('does not reuse the password typed for another user', async () => {
      patchMock.mockResolvedValue({ data: {} })
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('alice', 'Change Password')
      await userEvent.type(await screen.findByPlaceholderText('New Password'), 'secret-for-alice1!')
      // close without submitting
      await userEvent.keyboard('{Escape}')
      await waitFor(() => expect(screen.queryByPlaceholderText('New Password')).not.toBeInTheDocument())

      await openMenuItem('bob', 'Change Password')
      const input = await screen.findByPlaceholderText('New Password')
      expect(input).toHaveValue('')

      await userEvent.click(screen.getByRole('button', { name: 'Change Password' }))
      expect(await screen.findByText('Password cannot be empty')).toBeInTheDocument()
      expect(patchMock).not.toHaveBeenCalled()
    })

    it('submits the password for the selected user', async () => {
      patchMock.mockResolvedValue({ data: {} })
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('bob', 'Change Password')
      await userEvent.type(await screen.findByPlaceholderText('New Password'), 'newpass1!')
      await userEvent.click(screen.getByRole('button', { name: 'Change Password' }))

      await waitFor(() => expect(patchMock).toHaveBeenCalledWith('/api/user/b', { id: 'b', password: 'newpass1!' }, expect.anything()))
      expect(await screen.findByText('Password Updated!')).toBeInTheDocument()
    })

    it('shows the backend error when the password change fails', async () => {
      patchMock.mockRejectedValue(backendError('password too weak'))
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('bob', 'Change Password')
      await userEvent.type(await screen.findByPlaceholderText('New Password'), 'x')
      await userEvent.click(screen.getByRole('button', { name: 'Change Password' }))

      expect(await screen.findByText('password too weak')).toBeInTheDocument()
      expect(screen.queryByText('Password Updated!')).not.toBeInTheDocument()
    })
  })

  describe('delete user', () => {
    it('asks for confirmation before deleting', async () => {
      deleteMock.mockResolvedValue({ data: {} })
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('alice', 'Delete User')
      expect(deleteMock).not.toHaveBeenCalled()
      expect(await screen.findByText(/Are you sure you want to delete the user "alice"/)).toBeInTheDocument()

      await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
      await waitFor(() => expect(deleteMock).toHaveBeenCalledWith('/api/user/a', expect.anything()))
    })

    it('shows the backend error when the delete fails', async () => {
      deleteMock.mockRejectedValue(backendError('cannot delete the last admin'))
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('alice', 'Delete User')
      await userEvent.click(await screen.findByRole('button', { name: 'Delete' }))
      expect(await screen.findByText('Could not delete user: cannot delete the last admin')).toBeInTheDocument()
    })
  })

  describe('update user', () => {
    it('reverts the role select and shows the error when the role change fails', async () => {
      patchMock.mockRejectedValue(backendError('cannot change your own role'))
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      const roleSelect = await screen.findByRole('textbox', { name: 'Role of alice' })
      expect(roleSelect).toHaveValue('user')
      await userEvent.click(roleSelect)
      await userEvent.click(await screen.findByRole('option', { name: 'admin' }))

      expect(await screen.findByText('Could not update user: cannot change your own role')).toBeInTheDocument()
      await waitFor(() => expect(screen.getByRole('textbox', { name: 'Role of alice' })).toHaveValue('user'))
    })

    it('shows the error when suspending fails', async () => {
      patchMock.mockRejectedValue(backendError('suspend failed'))
      renderWithProviders(<ListUsers localAuthDisabled={false} />)

      await openMenuItem('alice', 'Suspend User')
      expect(await screen.findByText('Could not update user: suspend failed')).toBeInTheDocument()
    })
  })
})
