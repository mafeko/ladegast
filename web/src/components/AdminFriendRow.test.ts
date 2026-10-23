import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminFriendRow from './AdminFriendRow.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

const friend = {
  id: '1',
  name: 'Alex',
  vehicleName: 'vehicle_1',
  rateSolarPerKWh: 0.1,
  rateGridPerKWh: 0.35,
  createdAt: '',
  linkRegeneratedAt: '',
}

function mountRow() {
  return mount(AdminFriendRow, {
    props: { friend },
    global: { stubs: { RouterLink: true } },
  })
}

describe('AdminFriendRow', () => {
  it('saves updated rates and emits updated', async () => {
    const updated = { ...friend, rateSolarPerKWh: 0.2 }
    vi.mocked(api.updateFriend).mockResolvedValue(updated)

    const wrapper = mountRow()
    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click') // "Speichern"
    await flushMicrotasks()

    expect(api.updateFriend).toHaveBeenCalledWith('1', {
      name: 'Alex',
      vehicleName: 'vehicle_1',
      rateSolarPerKWh: 0.1,
      rateGridPerKWh: 0.35,
    })
    expect(wrapper.emitted('updated')).toEqual([[updated]])
    expect(wrapper.text()).toContain('Gespeichert.')
  })

  it('shows an error message when saving rates fails', async () => {
    vi.mocked(api.updateFriend).mockRejectedValue(new Error('internal error'))
    const wrapper = mountRow()

    await wrapper.findAll('button')[0].trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('internal error')
    expect(wrapper.emitted('updated')).toBeUndefined()
  })

  it('shows an error message when regenerating the link fails', async () => {
    vi.mocked(api.regenerateLink).mockRejectedValue(new Error('friend not found'))
    const wrapper = mountRow()

    await wrapper.findAll('button')[1].trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('friend not found')
  })

  it('regenerates the link and shows the new one', async () => {
    vi.mocked(api.regenerateLink).mockResolvedValue({ id: '2', link: 'http://x/l/2' })
    const wrapper = mountRow()

    await wrapper.findAll('button')[1].trigger('click') // "Link neu"
    await flushMicrotasks()

    expect(wrapper.text()).toContain('http://x/l/2')
  })

  it('resets the PIN and shows the new one', async () => {
    vi.mocked(api.resetPin).mockResolvedValue({ pin: '654321' })
    const wrapper = mountRow()

    await wrapper.findAll('button')[2].trigger('click') // "PIN zurücksetzen"
    await flushMicrotasks()

    expect(wrapper.text()).toContain('654321')
  })

  it('shows an error message when an action fails', async () => {
    vi.mocked(api.resetPin).mockRejectedValue(new Error('internal error'))
    const wrapper = mountRow()

    await wrapper.findAll('button')[2].trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('internal error')
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
