export class HttpError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'HttpError'
    this.status = status
  }
}

/**
 * Builds an HttpError from a non-2xx response, using the {"error": ...} body
 * the backend returns when present.
 */
export async function httpErrorFromResponse(res: Response): Promise<HttpError> {
  let message = `request failed with status ${res.status}`
  try {
    const body = await res.json()
    if (body?.error) {
      message = body.error
    }
  } catch {
    // body is not json, keep the default message
  }
  return new HttpError(res.status, message)
}

/**
 * GET a JSON endpoint with the bearer token. Throws an HttpError on non-2xx
 * responses, so react-query takes its error path instead of handing an
 * {"error": ...} body to components that expect data. Without a token (for
 * unauthenticated endpoints) no Authorization header is sent.
 */
export async function fetchJSON<T = any>(url: string, token?: string): Promise<T> { // eslint-disable-line @typescript-eslint/no-explicit-any
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (token !== undefined) {
    headers["Authorization"] = "Bearer " + token
  }
  const res = await fetch(url, { headers })
  if (!res.ok) {
    throw await httpErrorFromResponse(res)
  }
  return res.json()
}
