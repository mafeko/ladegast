<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { bucketLabel, bucketTitle, type Bucket, type Granularity } from '../lib/range'
import { formatEUR, formatKWh, solarSharePercent } from '../lib/format'
import type { Settlement } from '../lib/types'

const props = defineProps<{ buckets: Bucket[]; granularity: Granularity; canZoom?: boolean; settlements?: Settlement[] }>()
const emit = defineEmits<{ zoom: [bucket: Bucket] }>()

const PLOT_H = 196
const MARGIN = { right: 8, bottom: 28, left: 44 }
const LANE_H = 12
const CHAR_W = 6.4 // approx. width of an 11px label glyph

const container = ref<HTMLElement | null>(null)
const width = ref(600)
const active = ref<number | null>(null)
let observer: ResizeObserver | undefined

onMounted(() => {
  if (container.value && typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver(([entry]) => {
      if (entry.contentRect.width > 0) width.value = entry.contentRect.width
    })
    observer.observe(container.value)
  }
})
onBeforeUnmount(() => observer?.disconnect())

const innerW = computed(() => Math.max(0, width.value - MARGIN.left - MARGIN.right))
const innerH = PLOT_H

const total = (b: Bucket) => b.gridKWh + b.solarKWh

/** Round the axis maximum up to a "nice" number so ticks are readable. */
function niceMax(max: number): number {
  if (max <= 0) return 10
  const pow = Math.pow(10, Math.floor(Math.log10(max)))
  const n = max / pow
  const step = n <= 1 ? 1 : n <= 2 ? 2 : n <= 5 ? 5 : 10
  return step * pow
}

const yMax = computed(() => niceMax(Math.max(0, ...props.buckets.map(total))))
const ticks = computed(() => [0, 1, 2, 3, 4].map((i) => (yMax.value / 4) * i))
const y = (v: number) => top.value + innerH - (v / yMax.value) * innerH

const band = computed(() => (props.buckets.length ? innerW.value / props.buckets.length : 0))
const barW = computed(() => Math.max(2, Math.min(48, band.value * 0.7)))
const x = (i: number) => MARGIN.left + band.value * i + (band.value - barW.value) / 2

// Thin out axis labels so they never overlap.
const labelEvery = computed(() => {
  const minPx = props.granularity === 'week' ? 44 : 38
  return Math.max(1, Math.ceil(minPx / Math.max(1, band.value)))
})

const sum = computed(() => {
  const solar = props.buckets.reduce((a, b) => a + b.solarKWh, 0)
  const grid = props.buckets.reduce((a, b) => a + b.gridKWh, 0)
  return { solar, grid, total: solar + grid }
})
const hasData = computed(() => sum.value.total > 0)

const activeBucket = computed(() => (active.value === null ? null : props.buckets[active.value] ?? null))
const tooltipLeft = computed(() => {
  if (active.value === null) return 0
  const center = MARGIN.left + band.value * (active.value + 0.5)
  return Math.min(Math.max(center, 90), Math.max(90, width.value - 90))
})

// Settlements inside the visible range become vertical lines, positioned
// proportionally within the bucket they fall into. Labels that would touch
// are spread over extra lanes above the plot, so none overlap.
const settlementLines = computed(() => {
  const lines: { id: string; x: number; x1: number; x2: number; label: string; title: string; lane: number; anchor: 'start' | 'middle' | 'end' }[] = []
  const laneEnds: number[] = []
  const list = [...(props.settlements ?? [])].sort((a, b) => a.at.localeCompare(b.at))
  for (const s of list) {
    const t = new Date(s.at).getTime()
    const i = props.buckets.findIndex((b) => t >= b.start.getTime() && t < b.end.getTime())
    if (i === -1) continue
    const b = props.buckets[i]
    const lx = MARGIN.left + band.value * (i + (t - b.start.getTime()) / (b.end.getTime() - b.start.getTime()))
    const label = formatEUR(s.amountEUR)
    const w = label.length * CHAR_W
    const anchor = lx + w / 2 > width.value - 2 ? 'end' : lx - w / 2 < MARGIN.left ? 'start' : 'middle'
    const x1 = anchor === 'end' ? lx - w : anchor === 'start' ? lx : lx - w / 2
    let lane = laneEnds.findIndex((end) => x1 >= end + 6)
    if (lane === -1) lane = laneEnds.length
    laneEnds[lane] = x1 + w
    lines.push({ id: s.id, x: lx, x1, x2: x1 + w, label, title: `Abrechnung · ${label}`, lane, anchor })
  }
  return lines
})
const lanes = computed(() => Math.max(0, ...settlementLines.value.map((l) => l.lane + 1)))
const top = computed(() => (lanes.value ? 8 + lanes.value * LANE_H + 6 : 20))
const HEIGHT = computed(() => top.value + innerH + MARGIN.bottom)

function describe(b: Bucket): string {
  return `${bucketTitle(b, props.granularity)}: Netz ${formatKWh(b.gridKWh)}, Sonne ${formatKWh(b.solarKWh)}, gesamt ${formatKWh(total(b))}`
}

function onZoom() {
  if (activeBucket.value && props.canZoom) emit('zoom', activeBucket.value)
}
</script>

<template>
  <div>
    <div class="mb-3 flex flex-col gap-2 text-sm sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
      <p class="text-gray-600 dark:text-gray-300">
        <span class="font-semibold text-gray-900 dark:text-gray-100">{{ formatKWh(sum.total) }}</span>
        gesamt
        <span v-if="hasData" class="text-gray-500 dark:text-gray-400">· {{ solarSharePercent(sum.solar, sum.total) }}% Sonne</span>
      </p>
      <ul class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-600 dark:text-gray-300">
        <li v-if="settlementLines.length" class="flex items-center gap-1.5"><span class="inline-block h-3 border-l-2 border-dashed border-emerald-600 dark:border-emerald-400"></span>Abrechnung</li>
        <li class="flex items-center gap-1.5"><span class="inline-block size-3 rounded-sm bg-amber-400"></span>Sonne</li>
        <li class="flex items-center gap-1.5"><span class="inline-block size-3 rounded-sm bg-slate-500 dark:bg-slate-400"></span>Netz</li>
      </ul>
    </div>

    <div ref="container" class="relative" @mouseleave="active = null">
      <svg :width="width" :height="HEIGHT" role="img" :aria-label="`Gestapeltes Balkendiagramm der geladenen Energie, gesamt ${formatKWh(sum.total)}`" class="block select-none">
        <g class="text-gray-200 dark:text-gray-800">
          <line v-for="t in ticks" :key="t" :x1="MARGIN.left" :x2="width - MARGIN.right" :y1="y(t)" :y2="y(t)" stroke="currentColor" />
        </g>
        <g class="fill-gray-500 text-[11px] dark:fill-gray-400">
          <text v-for="t in ticks" :key="t" :x="MARGIN.left - 6" :y="y(t) + 4" text-anchor="end">{{ Number.isInteger(t) ? t : t.toFixed(1).replace('.', ',') }}</text>
          <text :x="4" :y="10" class="text-[10px]">kWh</text>
        </g>

        <g v-for="(b, i) in buckets" :key="b.start.getTime()">
          <!-- stack: grid at the bottom, solar on top -->
          <rect
            v-if="b.gridKWh > 0"
            :x="x(i)"
            :y="y(b.gridKWh)"
            :width="barW"
            :height="y(0) - y(b.gridKWh)"
            class="fill-slate-500 dark:fill-slate-400"
            :opacity="active === null || active === i ? 1 : 0.55"
          />
          <rect
            v-if="b.solarKWh > 0"
            :x="x(i)"
            :y="y(total(b))"
            :width="barW"
            :height="y(b.gridKWh) - y(total(b))"
            class="fill-amber-400"
            :opacity="active === null || active === i ? 1 : 0.55"
          />
          <text
            v-if="i % labelEvery === 0"
            :x="MARGIN.left + band * (i + 0.5)"
            :y="HEIGHT - 8"
            text-anchor="middle"
            class="fill-gray-500 text-[11px] dark:fill-gray-400"
          >
            {{ bucketLabel(b, granularity) }}
          </text>
          <!-- full-height hit area: hover, tap, keyboard focus -->
          <rect
            :x="MARGIN.left + band * i"
            :y="top"
            :width="band"
            :height="innerH"
            fill="transparent"
            tabindex="0"
            role="button"
            :aria-label="describe(b)"
            class="cursor-pointer focus:outline-none focus-visible:stroke-indigo-500"
            stroke-width="2"
            @mouseenter="active = i"
            @focus="active = i"
            @click="active = i"
          />
        </g>

        <!-- lines first, labels on top with a halo so crossing lines never cut through text -->
        <g class="pointer-events-none text-emerald-600 dark:text-emerald-400">
          <g v-for="l in settlementLines" :key="l.id" data-testid="settlement-line">
            <title>{{ l.title }}</title>
            <line :x1="l.x" :x2="l.x" :y1="8 + l.lane * LANE_H + 3" :y2="top + innerH" stroke="currentColor" stroke-width="2" stroke-dasharray="4 3" />
          </g>
          <text
            v-for="l in settlementLines"
            :key="'t' + l.id"
            :x="l.x"
            :y="8 + l.lane * LANE_H"
            :text-anchor="l.anchor"
            fill="currentColor"
            stroke-width="4"
            stroke-linejoin="round"
            class="stroke-white text-[11px] font-medium [paint-order:stroke] dark:stroke-gray-900"
          >{{ l.label }}</text>
        </g>
      </svg>

      <p v-if="!hasData" class="pointer-events-none absolute inset-x-0 top-1/3 text-center text-sm text-gray-500 dark:text-gray-400">
        Keine Ladevorgänge im Zeitraum
      </p>

      <div
        v-if="activeBucket"
        role="status"
        class="absolute top-0 z-10 w-48 -translate-x-1/2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs shadow-lg dark:border-gray-700 dark:bg-gray-900"
        :style="{ left: tooltipLeft + 'px' }"
      >
        <p class="mb-1 font-semibold">{{ bucketTitle(activeBucket, granularity) }}</p>
        <p class="flex justify-between"><span class="text-amber-600 dark:text-amber-400">Sonne</span>{{ formatKWh(activeBucket.solarKWh) }}</p>
        <p class="flex justify-between"><span class="text-slate-600 dark:text-slate-300">Netz</span>{{ formatKWh(activeBucket.gridKWh) }}</p>
        <p class="mt-1 flex justify-between border-t border-gray-100 pt-1 font-medium dark:border-gray-800">
          <span>Gesamt</span>{{ formatKWh(total(activeBucket)) }}
        </p>
        <button
          v-if="canZoom && total(activeBucket) > 0"
          type="button"
          class="mt-2 min-h-10 w-full rounded-md bg-indigo-600 px-2 py-1 text-white sm:min-h-0 hover:bg-indigo-500"
          @click="onZoom"
        >
          Hineinzoomen
        </button>
      </div>
    </div>
  </div>
</template>
