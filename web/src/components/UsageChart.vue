<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { bucketLabel, bucketTitle, type Bucket, type Granularity } from '../lib/range'
import { formatKWh, solarSharePercent } from '../lib/format'

const props = defineProps<{ buckets: Bucket[]; granularity: Granularity; canZoom?: boolean }>()
const emit = defineEmits<{ zoom: [bucket: Bucket] }>()

const HEIGHT = 260
const MARGIN = { top: 22, right: 8, bottom: 28, left: 44 }

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
const innerH = HEIGHT - MARGIN.top - MARGIN.bottom

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
const y = (v: number) => MARGIN.top + innerH - (v / yMax.value) * innerH

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

function describe(b: Bucket): string {
  return `${bucketTitle(b, props.granularity)}: Netz ${formatKWh(b.gridKWh)}, Sonne ${formatKWh(b.solarKWh)}, gesamt ${formatKWh(total(b))}`
}

function onZoom() {
  if (activeBucket.value && props.canZoom) emit('zoom', activeBucket.value)
}
</script>

<template>
  <div>
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2 text-sm">
      <p class="text-gray-600 dark:text-gray-300">
        <span class="font-semibold text-gray-900 dark:text-gray-100">{{ formatKWh(sum.total) }}</span>
        gesamt
        <span v-if="hasData" class="text-gray-500 dark:text-gray-400">· {{ solarSharePercent(sum.solar, sum.total) }}% Sonne</span>
      </p>
      <ul class="flex gap-4 text-xs text-gray-600 dark:text-gray-300">
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
            :y="MARGIN.top"
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
      </svg>

      <p v-if="!hasData" class="pointer-events-none absolute inset-x-0 top-1/3 text-center text-sm text-gray-500 dark:text-gray-400">
        Keine Ladevorgänge im Zeitraum
      </p>

      <div
        v-if="activeBucket"
        role="status"
        class="absolute top-0 z-10 w-44 -translate-x-1/2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs shadow-lg dark:border-gray-700 dark:bg-gray-900"
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
          class="mt-2 w-full rounded-md bg-indigo-600 px-2 py-1 text-white hover:bg-indigo-500"
          @click="onZoom"
        >
          Hineinzoomen
        </button>
      </div>
    </div>
  </div>
</template>
