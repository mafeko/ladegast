import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import UsageChart from './UsageChart.vue'
import { emptyBuckets, rangeFor } from '../lib/range'

function buckets() {
  const b = emptyBuckets(rangeFor('week', new Date(2026, 9, 23)), 'day')
  b[0].solarKWh = 6
  b[0].gridKWh = 2
  b[2].solarKWh = 3
  return b
}

describe('UsageChart', () => {
  it('draws grid below solar in one stacked bar per bucket with data', () => {
    const wrapper = mount(UsageChart, { props: { buckets: buckets(), granularity: 'day' } })
    const bars = wrapper.findAll('rect[class*="fill-"]')
    const grid = bars.filter((r) => r.classes().includes('fill-slate-500'))
    const solar = bars.filter((r) => r.classes().includes('fill-amber-400'))
    expect(grid).toHaveLength(1)
    expect(solar).toHaveLength(2)
    // solar segment sits on top of the grid segment
    expect(Number(solar[0].attributes('y')) + Number(solar[0].attributes('height'))).toBeCloseTo(Number(grid[0].attributes('y')))
  })

  it('summarises the total and solar share', () => {
    const wrapper = mount(UsageChart, { props: { buckets: buckets(), granularity: 'day' } })
    expect(wrapper.text()).toContain('11,00 kWh')
    expect(wrapper.text()).toContain('82% Sonne')
  })

  it('shows a tooltip with the bucket values on hover and focus', async () => {
    const wrapper = mount(UsageChart, { props: { buckets: buckets(), granularity: 'day' } })
    expect(wrapper.find('[role="status"]').exists()).toBe(false)

    await wrapper.findAll('rect[role="button"]')[0].trigger('mouseenter')
    const tip = wrapper.find('[role="status"]')
    expect(tip.text()).toContain('Sonne6,00 kWh')
    expect(tip.text()).toContain('Netz2,00 kWh')
    expect(tip.text()).toContain('Gesamt8,00 kWh')

    await wrapper.findAll('rect[role="button"]')[2].trigger('focus')
    expect(wrapper.find('[role="status"]').text()).toContain('3,00 kWh')
  })

  it('offers zooming only when allowed and emits the bucket', async () => {
    const noZoom = mount(UsageChart, { props: { buckets: buckets(), granularity: 'day' } })
    await noZoom.findAll('rect[role="button"]')[0].trigger('click')
    expect(noZoom.text()).not.toContain('Hineinzoomen')

    const wrapper = mount(UsageChart, { props: { buckets: buckets(), granularity: 'day', canZoom: true } })
    await wrapper.findAll('rect[role="button"]')[0].trigger('click')
    await wrapper.findAll('button').find((b) => b.text() === 'Hineinzoomen')!.trigger('click')
    expect(wrapper.emitted('zoom')?.[0][0]).toMatchObject({ solarKWh: 6, gridKWh: 2 })
  })

  it('shows an empty-state hint without data and exposes accessible bar labels', () => {
    const empty = emptyBuckets(rangeFor('week', new Date(2026, 9, 23)), 'day')
    const wrapper = mount(UsageChart, { props: { buckets: empty, granularity: 'day' } })
    expect(wrapper.text()).toContain('Keine Ladevorgänge im Zeitraum')
    expect(wrapper.find('rect[role="button"]').attributes('aria-label')).toContain('Netz 0,00 kWh')
  })
})
