import { describe, expect, it } from 'vitest'
import {
  PRESETS,
  alignedKind,
  bucketize,
  filterSessions,
  formatRange,
  fromDateInput,
  granularityFor,
  isoWeek,
  matchingPreset,
  rangeFor,
  shiftRange,
  toDateInput,
} from './range'
import type { SessionView } from './types'

const d = (y: number, m: number, day: number, h = 0, min = 0) => new Date(y, m - 1, day, h, min)
const session = (created: Date, solar: number, grid: number): SessionView => ({
  created: created.toISOString(),
  finished: created.toISOString(),
  chargedKWh: solar + grid,
  solarKWh: solar,
  gridKWh: grid,
  costEUR: 0,
})

describe('rangeFor', () => {
  it('covers a calendar month half-open', () => {
    const r = rangeFor('month', d(2026, 2, 17))
    expect(r.from).toEqual(d(2026, 2, 1))
    expect(r.to).toEqual(d(2026, 3, 1))
  })
  it('starts weeks on Monday', () => {
    const r = rangeFor('week', d(2026, 10, 23)) // Friday
    expect(r.from).toEqual(d(2026, 10, 19))
    expect(r.to).toEqual(d(2026, 10, 26))
  })
  it('covers a year', () => {
    const r = rangeFor('year', d(2026, 6, 1))
    expect(r.from).toEqual(d(2026, 1, 1))
    expect(r.to).toEqual(d(2027, 1, 1))
  })
})

describe('isoWeek', () => {
  it('numbers weeks per ISO 8601', () => {
    expect(isoWeek(d(2026, 10, 23))).toEqual({ year: 2026, week: 43 })
    expect(isoWeek(d(2026, 1, 1))).toEqual({ year: 2026, week: 1 })
    expect(isoWeek(d(2027, 1, 1))).toEqual({ year: 2026, week: 53 })
    expect(isoWeek(d(2024, 12, 30))).toEqual({ year: 2025, week: 1 })
  })
})

describe('shiftRange', () => {
  it('moves calendar months by one month, across year boundaries', () => {
    expect(shiftRange(rangeFor('month', d(2026, 1, 10)), -1)).toEqual(rangeFor('month', d(2025, 12, 1)))
    expect(shiftRange(rangeFor('month', d(2026, 12, 10)), 1)).toEqual(rangeFor('month', d(2027, 1, 1)))
  })
  it('moves years, weeks and days by their unit', () => {
    expect(shiftRange(rangeFor('year', d(2026, 5, 5)), -1)).toEqual(rangeFor('year', d(2025, 1, 1)))
    expect(shiftRange(rangeFor('week', d(2026, 10, 23)), 1).from).toEqual(d(2026, 10, 26))
    expect(shiftRange(rangeFor('day', d(2026, 3, 1)), -1).from).toEqual(d(2026, 2, 28))
  })
  it('moves custom ranges by their length', () => {
    const r = { from: d(2026, 10, 1), to: d(2026, 10, 11) } // 10 days
    expect(shiftRange(r, 1)).toEqual({ from: d(2026, 10, 11), to: d(2026, 10, 21) })
    expect(shiftRange(r, -1)).toEqual({ from: d(2026, 9, 21), to: d(2026, 10, 1) })
  })
})

describe('presets', () => {
  const now = d(2026, 10, 23, 14, 30)
  it('computes today as a single day', () => {
    expect(PRESETS.find((p) => p.id === 'today')!.range(now)).toEqual({ from: d(2026, 10, 23), to: d(2026, 10, 24) })
  })
  it('computes the last 7 days including today', () => {
    const r = PRESETS.find((p) => p.id === 'last7')!.range(now)
    expect(r.from).toEqual(d(2026, 10, 17))
    expect(r.to).toEqual(d(2026, 10, 24))
  })
  it('computes the last 12 months including the current one', () => {
    const r = PRESETS.find((p) => p.id === 'last12m')!.range(now)
    expect(r.from).toEqual(d(2025, 11, 1))
    expect(r.to).toEqual(d(2026, 11, 1))
  })
  it('recognises a preset range and labels it', () => {
    const r = rangeFor('month', now)
    expect(matchingPreset(r, now)?.id).toBe('month')
    expect(formatRange(r, now)).toBe('Dieser Monat')
  })
  it('labels other aligned and custom ranges', () => {
    expect(formatRange(rangeFor('month', d(2026, 3, 1)), now)).toBe('März 2026')
    expect(formatRange(rangeFor('year', d(2025, 3, 1)), now)).toBe('2025')
    expect(formatRange({ from: d(2026, 10, 1), to: d(2026, 10, 11) }, now)).toBe('01.10.2026 - 10.10.2026')
    expect(alignedKind({ from: d(2026, 10, 1), to: d(2026, 10, 11) })).toBeUndefined()
  })
})

describe('granularityFor', () => {
  it('picks a bucket size by span', () => {
    expect(granularityFor(rangeFor('day', d(2026, 10, 1)))).toBe('hour')
    expect(granularityFor(rangeFor('week', d(2026, 10, 1)))).toBe('day')
    expect(granularityFor(rangeFor('month', d(2026, 10, 1)))).toBe('day')
    expect(granularityFor({ from: d(2026, 1, 1), to: d(2026, 4, 1) })).toBe('week')
    expect(granularityFor(rangeFor('year', d(2026, 1, 1)))).toBe('month')
  })
})

describe('bucketize', () => {
  const oct = rangeFor('month', d(2026, 10, 1))
  it('creates one bucket per day, including days without charging', () => {
    const b = bucketize([], oct)
    expect(b).toHaveLength(31)
    expect(b.every((x) => x.gridKWh === 0 && x.solarKWh === 0)).toBe(true)
  })
  it('stacks grid and solar energy per bucket and ignores sessions outside the range', () => {
    const b = bucketize(
      [
        session(d(2026, 10, 5, 9), 4, 1),
        session(d(2026, 10, 5, 18), 1, 2),
        session(d(2026, 10, 6, 10), 3, 0),
        session(d(2026, 9, 30, 23), 99, 99), // before
        session(d(2026, 11, 1, 0), 99, 99), // end is exclusive
      ],
      oct,
    )
    expect(b[4]).toMatchObject({ solarKWh: 5, gridKWh: 3 })
    expect(b[5]).toMatchObject({ solarKWh: 3, gridKWh: 0 })
    expect(b.reduce((a, x) => a + x.solarKWh + x.gridKWh, 0)).toBe(11)
  })
  it('buckets a single day by hour', () => {
    const b = bucketize([session(d(2026, 10, 5, 13, 40), 2, 0)], rangeFor('day', d(2026, 10, 5)))
    expect(b).toHaveLength(24)
    expect(b[13].solarKWh).toBe(2)
  })
  it('buckets a year by month', () => {
    const b = bucketize([session(d(2026, 3, 15), 1, 1)], rangeFor('year', d(2026, 1, 1)))
    expect(b).toHaveLength(12)
    expect(b[2]).toMatchObject({ solarKWh: 1, gridKWh: 1 })
  })
})

describe('filterSessions', () => {
  it('keeps sessions inside the range, newest first', () => {
    const a = session(d(2026, 10, 2), 1, 0)
    const b = session(d(2026, 10, 20), 1, 0)
    const out = session(d(2026, 8, 1), 1, 0)
    expect(filterSessions([a, out, b], rangeFor('month', d(2026, 10, 1)))).toEqual([b, a])
  })
})

describe('date input helpers', () => {
  it('round-trips yyyy-mm-dd in local time', () => {
    expect(toDateInput(d(2026, 3, 7))).toBe('2026-03-07')
    expect(fromDateInput('2026-03-07')).toEqual(d(2026, 3, 7))
  })
  it('rejects empty and malformed input', () => {
    expect(fromDateInput('')).toBeUndefined()
    expect(fromDateInput('07.03.2026')).toBeUndefined()
  })
})
