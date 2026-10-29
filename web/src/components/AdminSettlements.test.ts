import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminSettlements from './AdminSettlements.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

const data = {
  settlements: [{ id: 's1', at: '2026-10-01T10:00:00Z', chargedKWh: 10, solarKWh: 0, gridKWh: 0, amountEUR: 4.2 }],
  open: { since: '2026-10-01T10:00:00Z', chargedKWh: 3, solarKWh: 0, gridKWh: 0, amountEUR: 1.5 },
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(api.viewAsSettlements).mockResolvedValue(data)
  vi.spyOn(window, 'confirm').mockReturnValue(true)
})

describe('AdminSettlements', () => {
  it('lists settlements and the open amount', async () => {
    const wrapper = mount(AdminSettlements, { props: { friendId: 'f1', friendName: 'Alex' } })
    await flushPromises()
    expect(wrapper.text()).toContain('4,20')
    expect(wrapper.text()).toContain('1,50')
  })

  it('creates a settlement and reloads', async () => {
    vi.mocked(api.createSettlement).mockResolvedValue({ id: 's2', at: '2026-10-28T10:00:00Z' })
    const wrapper = mount(AdminSettlements, { props: { friendId: 'f1', friendName: 'Alex' } })
    await flushPromises()
    await wrapper.find('input[type="datetime-local"]').setValue('2026-10-20T08:30')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(api.createSettlement).toHaveBeenCalledWith('f1', new Date('2026-10-20T08:30').toISOString())
    expect(api.viewAsSettlements).toHaveBeenCalledTimes(2)
  })

  it('deletes a settlement after confirmation', async () => {
    vi.mocked(api.deleteSettlement).mockResolvedValue(undefined)
    const wrapper = mount(AdminSettlements, { props: { friendId: 'f1', friendName: 'Alex' } })
    await flushPromises()
    await wrapper.findAll('button').find((b) => b.text() === 'Löschen')!.trigger('click')
    await flushPromises()
    expect(api.deleteSettlement).toHaveBeenCalledWith('f1', 's1')
  })
})
