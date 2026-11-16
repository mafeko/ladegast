import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminGuestRow from './AdminGuestRow.vue'

vi.mock('../lib/api')

const guest = {
  id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35,
  createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1',
}

function mountRow() {
  return mount(AdminGuestRow, {
    props: { guest },
    attachTo: document.body,
    global: { stubs: { RouterLink: true, AdminSettlements: true } },
  })
}

describe('AdminGuestRow', () => {
  it('shows a read-only summary without edit inputs', () => {
    const w = mountRow()
    expect(w.text()).toContain('Alex')
    expect(w.text()).toContain('vehicle_1')
    expect(w.text()).toContain('0,35')
    expect(w.find('input').exists()).toBe(false)
    w.unmount()
  })

  it('opens the edit dialog on click and closes it again', async () => {
    const w = mountRow()
    await w.find('button[aria-label="Alex bearbeiten"]').trigger('click')
    expect(document.body.querySelector('[role="dialog"]')).not.toBeNull()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await w.vm.$nextTick()
    expect(document.body.querySelector('[role="dialog"]')).toBeNull()
    w.unmount()
  })
})
