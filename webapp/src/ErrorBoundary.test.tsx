import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { ErrorBoundary } from './ErrorBoundary'
import { renderWithProviders } from './test-utils'

function Broken(): never {
  throw new Error('data.map is not a function')
}

describe('ErrorBoundary', () => {
  it('shows the error instead of unmounting the app', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    renderWithProviders(<ErrorBoundary><Broken /></ErrorBoundary>)
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('data.map is not a function')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Reload' })).toBeInTheDocument()
  })

  it('renders children when there is no error', () => {
    renderWithProviders(<ErrorBoundary><div>page content</div></ErrorBoundary>)
    expect(screen.getByText('page content')).toBeInTheDocument()
  })
})
