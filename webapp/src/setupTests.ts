import '@testing-library/jest-dom/vitest'
import { afterEach, vi } from 'vitest'

// Remove any global stubs (e.g. fetch) installed by a test so they don't leak.
afterEach(() => {
  vi.unstubAllGlobals()
})

// Mantine's ScrollArea and other components use ResizeObserver, absent in jsdom.
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
window.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver

// Some screens call window.scrollTo(0, 0) after saving; jsdom has no impl.
window.scrollTo = vi.fn() as unknown as typeof window.scrollTo

// Mantine components rely on window.matchMedia, which jsdom does not implement.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }),
})
