/**
 * Parses a yyyy-mm-dd query parameter as a local date (not UTC midnight, which
 * shows the previous day in negative UTC offsets). Returns null when the
 * parameter is missing or not a valid date.
 */
export function parseDateParam(param: string | null): Date | null {
  if (param === null) return null
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(param)
  if (match === null) return null
  const [year, month, day] = [Number(match[1]), Number(match[2]), Number(match[3])]
  const date = new Date(year, month - 1, day)
  // reject overflowing dates like 2024-02-31
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) return null
  return date
}
