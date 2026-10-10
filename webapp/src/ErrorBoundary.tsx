import { Component, ErrorInfo, ReactNode } from 'react'
import { Alert, Button, Container } from '@mantine/core'

type Props = {
  children?: ReactNode
}

type State = {
  error: Error | null
}

/**
 * Catches render errors so a single broken page shows an error message instead
 * of unmounting the whole app (white screen).
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("render error:", error, errorInfo.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <Container my={40}>
          <Alert color="red" title="Something went wrong">
            {this.state.error.message}
          </Alert>
          <Button mt="md" onClick={() => window.location.reload()}>Reload</Button>
        </Container>
      )
    }
    return this.props.children
  }
}
