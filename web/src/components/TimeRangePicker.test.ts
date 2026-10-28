import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TimeRangePicker from './TimeRangePicker.vue'
import { rangeFor } from '../lib/range'

// A fixed past month: the current month would be labelled "Dieser Monat".
const oct = rangeFor('month', new Date(2026, 2, 10))
const mountPicker = () => mount(TimeRangePicker, { props: { modelValue: oct }, attachTo: document.body })
const byText = (w: ReturnType<typeof mountPicker>, t: string) => w.findAll('button').find((b) => b.text() === t)!
const lastEmitted = (w: ReturnType<typeof mountPicker>) => w.emitted('update:modelValue')!.at(-1)![0] as ReturnType<typeof rangeFor>

describe('TimeRangePicker', () => {
  it('shows the current range label and steps with the arrows', async () => {
    const w = mountPicker()
    expect(w.text()).toContain('März 2026')

    await w.find('button[aria-label="Voriger Zeitraum"]').trigger('click')
    expect(lastEmitted(w)).toEqual(rangeFor('month', new Date(2026, 1, 1)))
    await w.find('button[aria-label="Nächster Zeitraum"]').trigger('click')
    expect(lastEmitted(w)).toEqual(rangeFor('month', new Date(2026, 3, 1)))
  })

  it('selects a preset and closes the popover', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    expect(w.find('[role="dialog"]').exists()).toBe(true)

    await byText(w, 'Heute').trigger('click')
    expect(lastEmitted(w)).toEqual(rangeFor('day', new Date()))
    expect(w.find('[role="dialog"]').exists()).toBe(false)
  })

  it('picks a month from the grid and steps the year', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    await w.find('button[aria-label="Vorjahr"]').trigger('click')
    await byText(w, 'Mai').trigger('click')
    expect(lastEmitted(w)).toEqual(rangeFor('month', new Date(2025, 4, 1)))
  })

  it('picks a whole year', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    await byText(w, 'Jahr').trigger('click')
    await byText(w, '2026 anzeigen').trigger('click')
    expect(lastEmitted(w)).toEqual(rangeFor('year', new Date(2026, 0, 1)))
  })

  it('picks the week containing a chosen date', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    await byText(w, 'Woche').trigger('click')
    await w.find('input[type="date"]').setValue('2026-10-23')
    expect(lastEmitted(w)).toEqual(rangeFor('week', new Date(2026, 9, 23)))
  })

  it('shows the calendar week of the current range in the week tab', async () => {
    const w = mount(TimeRangePicker, { props: { modelValue: rangeFor('week', new Date(2026, 9, 23)) } })
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    await byText(w, 'Woche').trigger('click')
    expect(w.text()).toContain('KW 43 / 2026: 19.10.2026 - 25.10.2026')
  })

  it('applies a custom range inclusive of its end date and validates order', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    await byText(w, 'Von - Bis').trigger('click')
    const [from, to] = w.findAll('input[type="date"]')

    await to.setValue('2026-10-02')
    await from.setValue('2026-10-05')
    expect(w.text()).toContain('Das Ende liegt vor dem Beginn.')
    expect(byText(w, 'Übernehmen').attributes('disabled')).toBeDefined()

    await from.setValue('2026-10-01')
    await to.setValue('2026-10-10')
    await w.find('form').trigger('submit')
    expect(lastEmitted(w)).toEqual({ from: new Date(2026, 9, 1), to: new Date(2026, 9, 11) })
  })

  it('closes on Escape and outside click', async () => {
    const w = mountPicker()
    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await w.vm.$nextTick()
    expect(w.find('[role="dialog"]').exists()).toBe(false)

    await w.find('button[aria-haspopup="dialog"]').trigger('click')
    document.body.click()
    await w.vm.$nextTick()
    expect(w.find('[role="dialog"]').exists()).toBe(false)
  })

  describe('mobile first', () => {
    const classesOf = (el: { classes: () => string[] }) => el.classes()
    const openPicker = async () => {
      const w = mountPicker()
      await w.find('button[aria-haspopup="dialog"]').trigger('click')
      return w
    }

    // Regression: an unprefixed `top-full` pushed the fixed sheet below the
    // viewport on phones, so the dropdown looked like it did nothing.
    it('anchors the dialog to the bottom of the viewport on phones, not below it', async () => {
      const dialog = classesOf((await openPicker()).find('[role="dialog"]'))
      expect(dialog).toContain('fixed')
      expect(dialog).toContain('inset-x-0')
      expect(dialog).toContain('bottom-0')
      expect(dialog).not.toContain('top-full')
      expect(dialog).not.toContain('absolute')
      // popover positioning only from the sm breakpoint up
      expect(dialog).toContain('sm:absolute')
      expect(dialog).toContain('sm:top-full')
    })

    it('limits the sheet height so it scrolls inside small screens', async () => {
      const dialog = classesOf((await openPicker()).find('[role="dialog"]'))
      expect(dialog).toContain('max-h-[85dvh]')
      expect(dialog).toContain('overflow-y-auto')
    })

    it('closes via backdrop tap', async () => {
      const w = await openPicker()
      const backdrop = w.find('[data-testid="picker-backdrop"]')
      expect(classesOf(backdrop)).toContain('sm:hidden')
      await backdrop.trigger('click')
      expect(w.find('[role="dialog"]').exists()).toBe(false)
      expect(w.find('[data-testid="picker-backdrop"]').exists()).toBe(false)
    })

    it('closes via the close button', async () => {
      const w = await openPicker()
      await w.find('button[aria-label="Schließen"]').trigger('click')
      expect(w.find('[role="dialog"]').exists()).toBe(false)
    })

    it('uses 44px touch targets and 16px inputs by default', async () => {
      const w = await openPicker()
      for (const label of ['Voriger Zeitraum', 'Nächster Zeitraum']) {
        expect(classesOf(w.find(`button[aria-label="${label}"]`))).toContain('min-h-11')
      }
      expect(classesOf(byText(w, 'Heute'))).toContain('min-h-11')
      expect(classesOf(byText(w, 'Mai'))).toContain('min-h-11')
      await byText(w, 'Tag').trigger('click')
      const input = classesOf(w.find('input[type="date"]'))
      expect(input).toContain('text-base')
      expect(input).toContain('sm:text-sm')
    })

    it('spans the full width on phones', () => {
      expect(classesOf(mountPicker().find('div'))).toContain('w-full')
    })
  })
})
