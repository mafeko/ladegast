import { describe, expect, it } from 'vitest'
import { formatDateTime, formatEUR, formatKWh, solarSharePercent } from './format'

describe('formatKWh', () => {
  it('formats with two decimals and a unit', () => {
    expect(formatKWh(5.3119)).toBe('5,31 kWh')
  })
  it('formats zero', () => {
    expect(formatKWh(0)).toBe('0,00 kWh')
  })
})

describe('formatEUR', () => {
  it('formats a positive amount as German-locale currency', () => {
    expect(formatEUR(10.5)).toContain('10,50')
    expect(formatEUR(10.5)).toContain('€')
  })
  it('formats zero', () => {
    expect(formatEUR(0)).toContain('0,00')
  })
})

describe('formatDateTime', () => {
  it('formats a valid ISO timestamp', () => {
    const out = formatDateTime('2026-10-17T10:00:00Z')
    expect(out).not.toBe('2026-10-17T10:00:00Z')
    expect(out.length).toBeGreaterThan(0)
  })
  it('returns the input unchanged for an invalid date', () => {
    expect(formatDateTime('not-a-date')).toBe('not-a-date')
  })
})

describe('solarSharePercent', () => {
  it('computes the percentage', () => {
    expect(solarSharePercent(5, 10)).toBe(50)
  })
  it('rounds to the nearest integer', () => {
    expect(solarSharePercent(1, 3)).toBe(33)
  })
  it('returns 0 when chargedKWh is zero', () => {
    expect(solarSharePercent(0, 0)).toBe(0)
  })
  it('returns 0 when chargedKWh is negative (defensive)', () => {
    expect(solarSharePercent(5, -1)).toBe(0)
  })
})
