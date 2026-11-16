import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminGuestDialog from './AdminGuestDialog.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

beforeEach(() => {
  vi.stubGlobal('confirm', vi.fn(() => true))
})

const guest = {
  id: '1',
  name: 'Alex',
  vehicleName: 'vehicle_1',
  rateSolarPerKWh: 0.1,
  rateGridPerKWh: 0.35,
  createdAt: '',
  linkRegeneratedAt: '', link: 'http://x/l/1',
}

function mountRow() {
  return mount(AdminGuestDialog, {
    props: { guest },
    global: { stubs: { RouterLink: true, AdminSettlements: true } },
  })
}

function button(wrapper: ReturnType<typeof mountRow>, text: string) {
  const b = wrapper.findAll('button').find((x) => x.text() === text)
  if (!b) throw new Error(`button "${text}" not found`)
  return b
}

describe('AdminGuestDialog', () => {
  it('disables Speichern until something changed', async () => {
    const wrapper = mountRow()
    expect(button(wrapper, 'Speichern').attributes('disabled')).toBeDefined()
    await wrapper.find('input[aria-label="Name"]').setValue('Alexa')
    expect(button(wrapper, 'Speichern').attributes('disabled')).toBeUndefined()
  })


  it('saves all edited fields and emits updated', async () => {
    const updated = { ...guest, name: 'Alexandra', rateSolarPerKWh: 0.2 }
    vi.mocked(api.updateGuest).mockResolvedValue(updated)
    const wrapper = mountRow()

    await wrapper.find('input[aria-label="Name"]').setValue('Alexandra')
    await wrapper.find('input[aria-label="Preis Sonne"]').setValue(0.2)
    await button(wrapper, 'Speichern').trigger('click')
    await flushMicrotasks()

    expect(api.updateGuest).toHaveBeenCalledWith('1', {
      name: 'Alexandra',
      vehicleName: 'vehicle_1',
      rateSolarPerKWh: 0.2,
      rateGridPerKWh: 0.35,
    })
    expect(wrapper.emitted('updated')).toEqual([[updated]])
    expect(wrapper.text()).toContain('Gespeichert.')
  })

  it('shows an error message when saving fails', async () => {
    vi.mocked(api.updateGuest).mockRejectedValue(new Error('internal error'))
    const wrapper = mountRow()

    await wrapper.find('input[aria-label="Name"]').setValue('Other')
    await button(wrapper, 'Speichern').trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('internal error')
    expect(wrapper.emitted('updated')).toBeUndefined()
  })

  it('copies the guest link at any time', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })
    const wrapper = mountRow()

    await button(wrapper, 'Link kopieren').trigger('click')
    await flushMicrotasks()

    expect(writeText).toHaveBeenCalledWith('http://x/l/1')
    expect(wrapper.text()).toContain('Kopiert')
  })

  it('regenerates the link and emits it to the parent', async () => {
    vi.mocked(api.regenerateLink).mockResolvedValue({ id: '2', link: 'http://x/l/2' })
    const wrapper = mountRow()

    await button(wrapper, 'Link neu').trigger('click')
    await flushMicrotasks()

    expect(wrapper.emitted('regenerated')).toEqual([['http://x/l/2']])
  })

  it('shows an error message when regenerating the link fails', async () => {
    vi.mocked(api.regenerateLink).mockRejectedValue(new Error('guest not found'))
    const wrapper = mountRow()

    await button(wrapper, 'Link neu').trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('guest not found')
  })

  it('sets a chosen PIN via the PIN form', async () => {
    vi.mocked(api.resetPin).mockResolvedValue({ pin: '4711' })
    const wrapper = mountRow()

    await button(wrapper, 'PIN ändern').trigger('click')
    await wrapper.find('input[aria-label="Neue PIN"]').setValue('4711')
    await wrapper.findAll('form')[1].trigger('submit')
    await flushMicrotasks()

    expect(api.resetPin).toHaveBeenCalledWith('1', '4711')
    expect(wrapper.text()).toContain('4711')
    expect(wrapper.findAll('form').length).toBe(1)
  })

  it('generates a random PIN when asked to', async () => {
    vi.mocked(api.resetPin).mockResolvedValue({ pin: '654321' })
    const wrapper = mountRow()

    await button(wrapper, 'PIN ändern').trigger('click')
    await button(wrapper, 'Zufällige PIN').trigger('click')
    await flushMicrotasks()

    expect(api.resetPin).toHaveBeenCalledWith('1', undefined)
    expect(wrapper.text()).toContain('654321')
  })

  it('shows an error message when setting the PIN fails', async () => {
    vi.mocked(api.resetPin).mockRejectedValue(new Error('internal error'))
    const wrapper = mountRow()

    await button(wrapper, 'PIN ändern').trigger('click')
    await button(wrapper, 'Zufällige PIN').trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('internal error')
  })

  it('deletes the guest after confirmation and emits deleted', async () => {
    vi.mocked(api.deleteGuest).mockResolvedValue(undefined)
    const wrapper = mountRow()

    await button(wrapper, 'Löschen').trigger('click')
    await flushMicrotasks()

    expect(api.deleteGuest).toHaveBeenCalledWith('1')
    expect(wrapper.emitted('deleted')).toEqual([[guest]])
  })

  it('shows an error message when deleting fails', async () => {
    vi.mocked(api.deleteGuest).mockRejectedValue(new Error('guest not found'))
    const wrapper = mountRow()

    await button(wrapper, 'Löschen').trigger('click')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('guest not found')
    expect(wrapper.emitted('deleted')).toBeUndefined()
  })

  it('does nothing when a confirmation is declined', async () => {
    vi.stubGlobal('confirm', vi.fn(() => false))
    const wrapper = mountRow()

    await button(wrapper, 'Link neu').trigger('click')
    await button(wrapper, 'Löschen').trigger('click')
    await button(wrapper, 'PIN ändern').trigger('click')
    await button(wrapper, 'Zufällige PIN').trigger('click')
    await flushMicrotasks()

    expect(api.regenerateLink).not.toHaveBeenCalled()
    expect(api.deleteGuest).not.toHaveBeenCalled()
    expect(api.resetPin).not.toHaveBeenCalled()
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
