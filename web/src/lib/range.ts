// Time-range model for the usage explorer: calendar helpers, quick ranges
// (Grafana-style), automatic bucket granularity and session bucketing. All
// pure functions on local time so they are directly unit-testable.
import type { SessionView } from './types'

/** Half-open interval [from, to). */
export interface Range {
  from: Date
  to: Date
}

export type RangeKind = 'day' | 'week' | 'month' | 'year'
export type Granularity = 'hour' | 'day' | 'week' | 'month'

export interface Bucket {
  start: Date
  end: Date
  gridKWh: number
  solarKWh: number
}

const DAY_MS = 24 * 60 * 60 * 1000

export function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function addDays(d: Date, n: number): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n, d.getHours(), d.getMinutes())
}

/** Monday 00:00 of the week containing d. */
export function startOfWeek(d: Date): Date {
  const day = startOfDay(d)
  return addDays(day, -((day.getDay() + 6) % 7))
}

export function startOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1)
}

export function startOfYear(d: Date): Date {
  return new Date(d.getFullYear(), 0, 1)
}

export function addMonths(d: Date, n: number): Date {
  return new Date(d.getFullYear(), d.getMonth() + n, 1)
}

/** ISO 8601 week number (weeks start on Monday, week 1 contains Jan 4th). */
export function isoWeek(d: Date): { year: number; week: number } {
  const thursday = addDays(startOfWeek(d), 3)
  const yearStart = new Date(thursday.getFullYear(), 0, 1)
  // round: the difference is off by an hour when a DST switch lies in between
  const days = Math.round((startOfDay(thursday).getTime() - yearStart.getTime()) / DAY_MS)
  const week = Math.floor(days / 7) + 1
  return { year: thursday.getFullYear(), week }
}

/** The calendar day/week/month/year containing anchor. */
export function rangeFor(kind: RangeKind, anchor: Date): Range {
  switch (kind) {
    case 'day': {
      const from = startOfDay(anchor)
      return { from, to: addDays(from, 1) }
    }
    case 'week': {
      const from = startOfWeek(anchor)
      return { from, to: addDays(from, 7) }
    }
    case 'month': {
      const from = startOfMonth(anchor)
      return { from, to: addMonths(from, 1) }
    }
    case 'year': {
      const from = startOfYear(anchor)
      return { from: from, to: new Date(from.getFullYear() + 1, 0, 1) }
    }
  }
}

export interface Preset {
  id: string
  label: string
  range: (now: Date) => Range
}

// Kept short on purpose: a handful of presets plus the absolute pickers.
export const PRESETS: Preset[] = [
  { id: 'today', label: 'Heute', range: (n) => rangeFor('day', n) },
  { id: 'week', label: 'Diese Woche', range: (n) => rangeFor('week', n) },
  { id: 'month', label: 'Dieser Monat', range: (n) => rangeFor('month', n) },
  { id: 'year', label: 'Dieses Jahr', range: (n) => rangeFor('year', n) },
  {
    id: 'last7',
    label: 'Letzte 7 Tage',
    range: (n) => ({ from: addDays(startOfDay(n), -6), to: addDays(startOfDay(n), 1) }),
  },
  {
    id: 'last30',
    label: 'Letzte 30 Tage',
    range: (n) => ({ from: addDays(startOfDay(n), -29), to: addDays(startOfDay(n), 1) }),
  },
  {
    id: 'last12m',
    label: 'Letzte 12 Monate',
    range: (n) => ({ from: addMonths(startOfMonth(n), -11), to: addMonths(startOfMonth(n), 1) }),
  },
]

export function sameRange(a: Range, b: Range): boolean {
  return a.from.getTime() === b.from.getTime() && a.to.getTime() === b.to.getTime()
}

/** The preset whose range equals r today, if any. */
export function matchingPreset(r: Range, now: Date): Preset | undefined {
  return PRESETS.find((p) => sameRange(p.range(now), r))
}

/** Which calendar unit r exactly covers, if any. */
export function alignedKind(r: Range): RangeKind | undefined {
  return (['day', 'week', 'month', 'year'] as const).find((k) => sameRange(rangeFor(k, r.from), r))
}

/** Moves r one step back (-1) or forward (+1): by its calendar unit if aligned, else by its length. */
export function shiftRange(r: Range, direction: -1 | 1): Range {
  switch (alignedKind(r)) {
    case 'month':
      return rangeFor('month', addMonths(r.from, direction))
    case 'year':
      return rangeFor('year', new Date(r.from.getFullYear() + direction, 0, 1))
    case 'week':
      return rangeFor('week', addDays(r.from, 7 * direction))
    case 'day':
      return rangeFor('day', addDays(r.from, direction))
  }
  const days = spanDays(r)
  return { from: addDays(r.from, days * direction), to: addDays(r.to, days * direction) }
}

function spanDays(r: Range): number {
  return Math.max(1, Math.round((r.to.getTime() - r.from.getTime()) / DAY_MS))
}

/** Picks the bucket size so a range shows roughly 7-31 bars. */
export function granularityFor(r: Range): Granularity {
  const days = spanDays(r)
  if (days <= 2) return 'hour'
  if (days <= 31) return 'day'
  if (days <= 140) return 'week'
  return 'month'
}

function bucketStart(d: Date, g: Granularity): Date {
  switch (g) {
    case 'hour':
      return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours())
    case 'day':
      return startOfDay(d)
    case 'week':
      return startOfWeek(d)
    case 'month':
      return startOfMonth(d)
  }
}

function nextBucket(d: Date, g: Granularity): Date {
  switch (g) {
    case 'hour':
      return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours() + 1)
    case 'day':
      return addDays(d, 1)
    case 'week':
      return addDays(d, 7)
    case 'month':
      return addMonths(d, 1)
  }
}

/** Empty buckets covering r, so days without charging show up as gaps. */
export function emptyBuckets(r: Range, g: Granularity): Bucket[] {
  const out: Bucket[] = []
  for (let s = bucketStart(r.from, g); s < r.to; s = nextBucket(s, g)) {
    out.push({ start: s, end: nextBucket(s, g), gridKWh: 0, solarKWh: 0 })
  }
  return out
}

/** Sums sessions into buckets by their start time. Sessions outside r are ignored. */
export function bucketize(sessions: SessionView[], r: Range, g: Granularity = granularityFor(r)): Bucket[] {
  const buckets = emptyBuckets(r, g)
  for (const s of sessions) {
    const t = new Date(s.created)
    if (Number.isNaN(t.getTime()) || t < r.from || t >= r.to) continue
    const b = buckets.find((x) => t >= x.start && t < x.end)
    if (!b) continue
    b.gridKWh += s.gridKWh
    b.solarKWh += s.solarKWh
  }
  return buckets
}

/** Sessions started inside r, newest first. */
export function filterSessions(sessions: SessionView[], r: Range): SessionView[] {
  return sessions
    .filter((s) => {
      const t = new Date(s.created)
      return !Number.isNaN(t.getTime()) && t >= r.from && t < r.to
    })
    .sort((a, b) => new Date(b.created).getTime() - new Date(a.created).getTime())
}

/** The next finer granularity a bar can zoom into, if any. */
export function finerGranularity(g: Granularity): Granularity | undefined {
  return { month: 'week', week: 'day', day: 'hour', hour: undefined }[g] as Granularity | undefined
}

// --- Labels ---

const dm = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit' })
const dmy = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
const monthShort = new Intl.DateTimeFormat('de-DE', { month: 'short' })
const monthLong = new Intl.DateTimeFormat('de-DE', { month: 'long', year: 'numeric' })
const weekdayDate = new Intl.DateTimeFormat('de-DE', { weekday: 'short', day: '2-digit', month: '2-digit', year: 'numeric' })

/** Short axis label of a bucket. */
export function bucketLabel(b: Bucket, g: Granularity): string {
  switch (g) {
    case 'hour':
      return `${b.start.getHours()}`
    case 'day':
      return dm.format(b.start)
    case 'week':
      return `KW ${isoWeek(b.start).week}`
    case 'month':
      return monthShort.format(b.start).replace('.', '')
  }
}

/** Long tooltip title of a bucket. */
export function bucketTitle(b: Bucket, g: Granularity): string {
  switch (g) {
    case 'hour':
      return `${dmy.format(b.start)}, ${b.start.getHours()}-${b.end.getHours() % 24} Uhr`
    case 'day':
      return weekdayDate.format(b.start)
    case 'week': {
      const last = addDays(b.end, -1)
      return `KW ${isoWeek(b.start).week} (${dm.format(b.start)}.-${dmy.format(last)})`
    }
    case 'month':
      return monthLong.format(b.start)
  }
}

/** Human-readable label of a range, e.g. "Oktober 2026" or "01.10.2026 - 23.10.2026". */
export function formatRange(r: Range, now: Date = new Date()): string {
  const preset = matchingPreset(r, now)
  if (preset) return preset.label
  const kind = alignedKind(r)
  const last = addDays(r.to, -1)
  switch (kind) {
    case 'day':
      return weekdayDate.format(r.from)
    case 'week':
      return `KW ${isoWeek(r.from).week} / ${r.from.getFullYear()} (${dm.format(r.from)}.-${dm.format(last)}.)`
    case 'month':
      return monthLong.format(r.from)
    case 'year':
      return String(r.from.getFullYear())
  }
  return `${dmy.format(r.from)} - ${dmy.format(last)}`
}

/** yyyy-mm-dd for <input type="date">, in local time. */
export function toDateInput(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

/** Parses yyyy-mm-dd (local time); undefined for empty/invalid input. */
export function fromDateInput(v: string): Date | undefined {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(v)
  if (!m) return undefined
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return Number.isNaN(d.getTime()) ? undefined : d
}
