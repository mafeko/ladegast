import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminPage from './AdminPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

beforeEach(() => {
  vi.stubGlobal('confirm', vi.fn(() => true))
})

// The edit dialog is teleported to <body>, so query it there.
async function openDialog(wrapper: ReturnType<typeof mount>) {
  await wrapper.find('button[aria-label$="bearbeiten"]').trigger('click')
  await flushMicrotasks()
}
function dialogButton(text: string) {
  const b = [...document.body.querySelectorAll('[role="dialog"] button')].find((x) => x.textContent?.trim() === text)
  if (!b) throw new Error(`dialog button ${text} not found`)
  return b as HTMLButtonElement
}

afterEach(() => {
  document.body.innerHTML = ''
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('AdminPage', () => {
  it('shows the login form when listGuests is unauthorized', async () => {
    vi.mocked(api.listGuests).mockRejectedValue(new Error('not authenticated'))
    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.find('#username').exists()).toBe(true)
  })

  it('shows the guest management UI once authenticated', async () => {
    vi.mocked(api.listGuests).mockResolvedValue([
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
    ])
    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Gäste verwalten')
    expect(wrapper.text()).toContain('Alex')
    await openDialog(wrapper)
    expect((document.body.querySelector('[role="dialog"] input[aria-label="Name"]') as HTMLInputElement).value).toBe('Alex')
  })

  it('reloads the guest list after creating a guest', async () => {
    vi.mocked(api.listGuests).mockResolvedValueOnce([]).mockResolvedValueOnce([
      { id: '1', name: 'Bea', vehicleName: 'vehicle_2', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
    ])
    vi.mocked(api.createGuest).mockResolvedValue({ id: '1', pin: '123456', link: 'http://x/l/1' })

    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    await wrapper.findAll('button').find((b) => b.text() === '+ Neuer Gast')!.trigger('click')
    await wrapper.find('#name').setValue('Bea')
    await wrapper.find('#vehicleName').setValue('vehicle_2')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Bea')
    expect(wrapper.text()).toContain('http://x/l/1')
  })

  it('shows an empty state when there are no guests', async () => {
    vi.mocked(api.listGuests).mockResolvedValue([])
    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Noch keine Gäste angelegt.')
  })

  it('reloads the list and shows the new link after regenerating', async () => {
    const guest = { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' }
    vi.mocked(api.listGuests).mockResolvedValueOnce([guest]).mockResolvedValueOnce([{ ...guest, id: '2' }])
    vi.mocked(api.regenerateLink).mockResolvedValue({ id: '2', link: 'http://x/l/2' })

    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()
    await openDialog(wrapper)
    dialogButton('Link neu').click()
    await flushMicrotasks()

    expect(api.listGuests).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('http://x/l/2')
  })

  it('removes a guest from the list after deleting', async () => {
    const guest = { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' }
    vi.mocked(api.listGuests).mockResolvedValue([guest])
    vi.mocked(api.deleteGuest).mockResolvedValue(undefined)

    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()
    await openDialog(wrapper)
    dialogButton('Löschen').click()
    await flushMicrotasks()

    expect(api.deleteGuest).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('Noch keine Gäste angelegt.')
    expect(wrapper.text()).toContain('Alex wurde gelöscht.')
  })

  it('logs out and shows the login form again', async () => {
    vi.mocked(api.listGuests).mockResolvedValue([])
    vi.mocked(api.adminLogout).mockResolvedValue(undefined)
    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    await wrapper.findAll('button').find((b) => b.text() === 'Abmelden')!.trigger('click')
    await flushMicrotasks()

    expect(api.adminLogout).toHaveBeenCalled()
    expect(wrapper.find('#username').exists()).toBe(true)
  })

  it('reflects a row update without re-fetching the whole list', async () => {
    vi.mocked(api.listGuests).mockResolvedValue([
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
    ])
    vi.mocked(api.updateGuest).mockResolvedValue({
      id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.5, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1',
    })

    const wrapper = mount(AdminPage, { attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    await openDialog(wrapper)
    const nameInput = document.body.querySelector('[role="dialog"] input[aria-label="Name"]') as HTMLInputElement
    nameInput.value = 'Alexa'
    nameInput.dispatchEvent(new Event('input'))
    await flushMicrotasks()
    dialogButton('Speichern').click()
    await flushMicrotasks()

    expect(api.updateGuest).toHaveBeenCalled()
    expect(document.body.textContent).toContain('Gespeichert.')
  })
})
