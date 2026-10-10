import { Cookies } from 'react-cookie'

export const TOKEN_COOKIE = 'token'

/**
 * Options for the token cookie: scoped to the whole app, never sent on
 * cross-site requests, and https-only when the app is served over https.
 */
export function tokenCookieOptions() {
  return {
    path: '/',
    sameSite: 'strict' as const,
    secure: window.location.protocol === 'https:',
  }
}

/** Options to remove the token cookie (must match the path it was set with). */
export const tokenCookieRemoveOptions = { path: '/' }

/** Removes the token cookie outside of React (e.g. from the query client). */
export function removeTokenCookie() {
  new Cookies().remove(TOKEN_COOKIE, tokenCookieRemoveOptions)
}
