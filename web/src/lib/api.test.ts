import { afterEach, describe, expect, it, vi } from 'vitest'
import * as api from './api'
import { ApiError } from './api'

function mockFetchOnce(status: number, body: unknown, init: Partial<Response> = {}) {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
    ...init,
  } as Response)
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('request error handling (via any endpoint)', () => {
  it('resolves with the parsed JSON body on success', async () => {
    mockFetchOnce(200, { authenticated: true })
    const result = await api.getFriendStatus('uuid-1')
    expect(result).toEqual({ authenticated: true })
  })

  it('throws ApiError with the server-provided message on failure', async () => {
    mockFetchOnce(401, { error: 'invalid PIN' })
    await expect(api.friendLogin('uuid-1', '000000')).rejects.toMatchObject({
      name: 'ApiError',
      status: 401,
      message: 'Falsche PIN.',
    })
  })

  it('falls back to a generic message when the error body is not JSON', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: async () => {
        throw new SyntaxError('unexpected end of input')
      },
    } as unknown as Response)
    vi.stubGlobal('fetch', fetchMock)

    await expect(api.getFriendStatus('uuid-1')).rejects.toThrow('Anfrage fehlgeschlagen (Status 500)')
  })

  it('returns undefined for a 204 No Content response', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
      json: async () => {
        throw new Error('should not be called')
      },
    } as unknown as Response)
    vi.stubGlobal('fetch', fetchMock)

    await expect(api.friendLogout('uuid-1')).resolves.toBeUndefined()
  })
})

describe('friend-facing endpoints', () => {
  it('getFriendStatus calls GET /public/l/{uuid}/status', async () => {
    const fetchMock = mockFetchOnce(200, { authenticated: false })
    await api.getFriendStatus('abc')
    expect(fetchMock).toHaveBeenCalledWith(
      '/public/l/abc/status',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    )
  })

  it('friendLogin posts the PIN', async () => {
    const fetchMock = mockFetchOnce(200, { name: 'Alex' })
    await api.friendLogin('abc', '123456')
    expect(fetchMock).toHaveBeenCalledWith(
      '/public/l/abc/login',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ pin: '123456' }) }),
    )
  })

  it('getDashboard calls the dashboard endpoint', async () => {
    const fetchMock = mockFetchOnce(200, {})
    await api.getDashboard('abc')
    expect(fetchMock).toHaveBeenCalledWith('/public/l/abc/dashboard', expect.anything())
  })

  it('getSessions calls the sessions endpoint', async () => {
    const fetchMock = mockFetchOnce(200, [])
    await api.getSessions('abc')
    expect(fetchMock).toHaveBeenCalledWith('/public/l/abc/sessions', expect.anything())
  })

  it('setMode posts the chosen mode', async () => {
    const fetchMock = mockFetchOnce(200, { mode: 'now' })
    const result = await api.setMode('abc', 'now')
    expect(result).toEqual({ mode: 'now' })
    expect(fetchMock).toHaveBeenCalledWith(
      '/public/l/abc/mode',
      expect.objectContaining({ body: JSON.stringify({ mode: 'now' }) }),
    )
  })
})

describe('admin endpoints', () => {
  it('adminLogin posts username and password', async () => {
    const fetchMock = mockFetchOnce(200, { username: 'admin' })
    await api.adminLogin('admin', 'secret')
    expect(fetchMock).toHaveBeenCalledWith(
      '/admin/login',
      expect.objectContaining({ body: JSON.stringify({ username: 'admin', password: 'secret' }) }),
    )
  })

  it('adminLogout posts to the logout endpoint', async () => {
    const fetchMock = mockFetchOnce(204, undefined)
    await api.adminLogout()
    expect(fetchMock).toHaveBeenCalledWith('/admin/logout', expect.objectContaining({ method: 'POST' }))
  })

  it('listFriends calls GET /admin/friends', async () => {
    const fetchMock = mockFetchOnce(200, [])
    await api.listFriends()
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends', expect.objectContaining({ method: 'GET' }))
  })

  it('createFriend posts the friend input', async () => {
    const input = { name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35 }
    const fetchMock = mockFetchOnce(201, { id: '1', pin: '123456', link: 'http://x/l/1' })
    const result = await api.createFriend(input)
    expect(result.pin).toBe('123456')
    expect(fetchMock).toHaveBeenCalledWith(
      '/admin/friends',
      expect.objectContaining({ method: 'POST', body: JSON.stringify(input) }),
    )
  })

  it('updateFriend PUTs to /admin/friends/{id}', async () => {
    const input = { name: 'Alexandra', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.2, rateGridPerKWh: 0.4 }
    const fetchMock = mockFetchOnce(200, { id: '1', ...input, createdAt: '', linkRegeneratedAt: '' })
    await api.updateFriend('1', input)
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends/1', expect.objectContaining({ method: 'PUT' }))
  })

  it('regenerateLink posts to the regenerate-link endpoint', async () => {
    const fetchMock = mockFetchOnce(200, { id: '2', link: 'http://x/l/2' })
    const result = await api.regenerateLink('1')
    expect(result.id).toBe('2')
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends/1/regenerate-link', expect.objectContaining({ method: 'POST' }))
  })

  it('resetPin posts to the reset-pin endpoint', async () => {
    const fetchMock = mockFetchOnce(200, { pin: '654321' })
    const result = await api.resetPin('1')
    expect(result.pin).toBe('654321')
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends/1/reset-pin', expect.objectContaining({ method: 'POST' }))
  })

  it('viewAsDashboard calls the view-as dashboard endpoint', async () => {
    const fetchMock = mockFetchOnce(200, {})
    await api.viewAsDashboard('1')
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends/1/view-as/dashboard', expect.objectContaining({ method: 'GET' }))
  })

  it('viewAsSessions calls the view-as sessions endpoint', async () => {
    const fetchMock = mockFetchOnce(200, [])
    await api.viewAsSessions('1')
    expect(fetchMock).toHaveBeenCalledWith('/admin/friends/1/view-as/sessions', expect.objectContaining({ method: 'GET' }))
  })
})

describe('ApiError', () => {
  it('carries status and message', () => {
    const err = new ApiError('boom', 500)
    expect(err.message).toBe('boom')
    expect(err.status).toBe(500)
    expect(err.name).toBe('ApiError')
  })
})
