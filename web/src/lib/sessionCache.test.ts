import { beforeEach, describe, expect, it, vi } from 'vitest'
import { cachedAt, cachedSessions, clearSessionCache } from './sessionCache'

beforeEach(() => clearSessionCache())

describe('sessionCache', () => {
  it('loads once and serves later calls from the cache', async () => {
    const loader = vi.fn().mockResolvedValue([])
    await cachedSessions('k', loader)
    await cachedSessions('k', loader)
    expect(loader).toHaveBeenCalledTimes(1)
    expect(cachedAt('k')).toBeTypeOf('number')
  })

  it('keeps separate entries per key', async () => {
    const loader = vi.fn().mockResolvedValue([])
    await cachedSessions('a', loader)
    await cachedSessions('b', loader)
    expect(loader).toHaveBeenCalledTimes(2)
  })

  it('shares one in-flight request between concurrent callers', async () => {
    const loader = vi.fn().mockResolvedValue([])
    await Promise.all([cachedSessions('k', loader), cachedSessions('k', loader)])
    expect(loader).toHaveBeenCalledTimes(1)
  })

  it('reloads after the TTL or when forced', async () => {
    const loader = vi.fn().mockResolvedValue([])
    let t = 0
    const opts = { ttlMs: 1000, now: () => t }
    await cachedSessions('k', loader, opts)
    t = 999
    await cachedSessions('k', loader, opts)
    expect(loader).toHaveBeenCalledTimes(1)
    t = 1000
    await cachedSessions('k', loader, opts)
    expect(loader).toHaveBeenCalledTimes(2)
    await cachedSessions('k', loader, { ...opts, force: true })
    expect(loader).toHaveBeenCalledTimes(3)
  })

  it('does not cache failures', async () => {
    const loader = vi.fn().mockRejectedValueOnce(new Error('boom')).mockResolvedValue([])
    await expect(cachedSessions('k', loader)).rejects.toThrow('boom')
    await cachedSessions('k', loader)
    expect(loader).toHaveBeenCalledTimes(2)
  })
})
