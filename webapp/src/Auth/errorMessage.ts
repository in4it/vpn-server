/**
 * Returns a human readable message for a failed request. For axios errors the
 * backend's {"error": "..."} response body is preferred over the generic
 * "Request failed with status code 4xx" message.
 */
export function getErrorMessage(error: unknown): string {
  const data = (error as { response?: { data?: unknown } } | null)?.response?.data
  if (data !== null && typeof data === 'object' && 'error' in data) {
    const backendError = (data as { error: unknown }).error
    if (typeof backendError === 'string' && backendError !== '') {
      return backendError
    }
  }
  if (error instanceof Error) {
    return error.message
  }
  return String(error)
}
