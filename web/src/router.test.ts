import { describe, expect, it } from 'vitest'
import { router } from './router'

describe('router', () => {
  it('redirects the root path to /admin so "/" is never blank', async () => {
    await router.push('/')
    await router.isReady()
    expect(router.currentRoute.value.fullPath).toBe('/admin')
  })

  it('resolves the friend link route', () => {
    const resolved = router.resolve('/public/l/abc-123')
    expect(resolved.name).toBe('friend')
    expect(resolved.params.uuid).toBe('abc-123')
  })

  it('resolves the admin view-as route', () => {
    const resolved = router.resolve('/admin/view-as/42')
    expect(resolved.name).toBe('admin-view-as')
    expect(resolved.params.id).toBe('42')
  })
})
