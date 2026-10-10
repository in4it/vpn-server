export class HttpError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'HttpError'
    this.status = status
  }
}

/**
 * GET a JSON endpoint with the bearer token. Throws an HttpError on non-2xx
 * responses, so react-query takes its error path instead of handing an
 * {"error": ...} body to components that expect data.
 */
export async function fetchJSON<T = any>(url: string, token: string): Promise<T> { // eslint-disable-line @typescript-eslint/no-explicit-any
  const res = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + token,
    },
  })
  if (!res.ok) {
    let message = `request failed with status ${res.status}`
    try {
      const body = await res.json()
      if (body?.error) {
        message = body.error
      }
    } catch {
      // body is not json, keep the default message
    }
    throw new HttpError(res.status, message)
  }
  return res.json()
}
