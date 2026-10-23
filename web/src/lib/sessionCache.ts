// In-memory cache for the (potentially long) session history so changing
// the time range filters locally instead of calling the backend - and thus
// EVCC - again. Concurrent callers share one in-flight request.
import type { SessionView } from './types'

const DEFAULT_TTL_MS = 5 * 60 * 1000

interface Entry {
  at: number
  data: SessionView[]
}

const entries = new Map<string, Entry>()
const inflight = new Map<string, Promise<SessionView[]>>()

export interface CacheOptions {
  /** Ignore a cached value and fetch again. */
  force?: boolean
  ttlMs?: number
  now?: () => number
}

export async function cachedSessions(
  key: string,
  loader: () => Promise<SessionView[]>,
  { force = false, ttlMs = DEFAULT_TTL_MS, now = Date.now }: CacheOptions = {},
): Promise<SessionView[]> {
  const hit = entries.get(key)
  if (!force && hit && now() - hit.at < ttlMs) return hit.data

  const pending = inflight.get(key)
  if (pending) return pending

  const p = loader()
    .then((data) => {
      entries.set(key, { at: now(), data })
      return data
    })
    .finally(() => inflight.delete(key))
  inflight.set(key, p)
  return p
}

/** Timestamp (ms) of the cached value for key, if any. */
export function cachedAt(key: string): number | undefined {
  return entries.get(key)?.at
}

export function clearSessionCache(key?: string): void {
  if (key === undefined) entries.clear()
  else entries.delete(key)
}
