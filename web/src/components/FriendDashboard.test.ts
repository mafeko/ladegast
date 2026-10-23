import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import FriendDashboard from './FriendDashboard.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

const dashboard = {
  friend: { id: 'abc', name: 'Alex' },
  today: { chargedKWh: 1, solarKWh: 1, gridKWh: 0, costEUR: 0.1 },
  thisWeek: { chargedKWh: 2, solarKWh: 1, gridKWh: 1, costEUR: 0.5 },
  thisMonth: { chargedKWh: 3, solarKWh: 2, gridKWh: 1, costEUR: 0.8 },
  thisYear: { chargedKWh: 4, solarKWh: 3, gridKWh: 1, costEUR: 1.2 },
}

describe('FriendDashboard', () => {
  it('loads and renders the dashboard and sessions on mount', async () => {
    vi.mocked(api.getDashboard).mockResolvedValue(dashboard)
    vi.mocked(api.getSessions).mockResolvedValue([])

    const wrapper = mount(FriendDashboard, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(api.getDashboard).toHaveBeenCalledWith('abc')
    expect(api.getSessions).toHaveBeenCalledWith('abc')
    expect(wrapper.text()).toContain('Hallo Alex')
  })

  it('shows an error message when loading fails', async () => {
    vi.mocked(api.getDashboard).mockRejectedValue(new Error('EVCC unreachable'))
    vi.mocked(api.getSessions).mockResolvedValue([])

    const wrapper = mount(FriendDashboard, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('EVCC unreachable')
  })

  it('logs out and emits loggedOut', async () => {
    vi.mocked(api.getDashboard).mockResolvedValue(dashboard)
    vi.mocked(api.getSessions).mockResolvedValue([])
    vi.mocked(api.friendLogout).mockResolvedValue(undefined)

    const wrapper = mount(FriendDashboard, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    await wrapper.find('button').trigger('click')
    await flushMicrotasks()

    expect(api.friendLogout).toHaveBeenCalledWith('abc')
    expect(wrapper.emitted('loggedOut')).toHaveLength(1)
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
