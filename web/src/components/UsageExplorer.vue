<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { cachedAt, cachedSessions, clearSessionCache } from '../lib/sessionCache'
import {
  PRESETS,
  bucketize,
  filterSessions,
  finerGranularity,
  granularityFor,
  type Bucket,
  type Range,
} from '../lib/range'
import type { SessionView, Settlements } from '../lib/types'
import SessionTable from './SessionTable.vue'
import SettlementTable from './SettlementTable.vue'
import TimeRangePicker from './TimeRangePicker.vue'
import UsageChart from './UsageChart.vue'

// The whole history is loaded once per cacheKey; picking a range only
// filters it in memory.
const props = defineProps<{
  cacheKey: string
  load: () => Promise<SessionView[]>
  /** Optional: settlement history, drawn as lines in the chart and listed below. */
  loadSettlements?: () => Promise<Settlements>
}>()

const sessions = ref<SessionView[]>([])
const settlements = ref<Settlements | null>(null)
const loading = ref(true)
const error = ref('')
const updatedAt = ref<number | undefined>()
const range = ref<Range>(PRESETS.find((p) => p.id === 'month')!.range(new Date()))

const granularity = computed(() => granularityFor(range.value))
const buckets = computed(() => bucketize(sessions.value, range.value, granularity.value))
const visible = computed(() => filterSessions(sessions.value, range.value))
const canZoom = computed(() => finerGranularity(granularity.value) !== undefined)
const updatedLabel = computed(() =>
  updatedAt.value ? new Date(updatedAt.value).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' }) : '',
)

async function fetchSessions(force = false) {
  loading.value = true
  error.value = ''
  try {
    sessions.value = await cachedSessions(props.cacheKey, props.load, { force })
    updatedAt.value = cachedAt(props.cacheKey)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Laden fehlgeschlagen'
  } finally {
    loading.value = false
  }
  // Settlements are decoration: failing or empty must never block the chart.
  settlements.value = null
  if (props.loadSettlements) {
    try {
      const data = await props.loadSettlements()
      settlements.value = { ...data, settlements: data.settlements ?? [] }
    } catch {
      /* chart and sessions stay usable without settlements */
    }
  }
}

function refresh() {
  clearSessionCache(props.cacheKey)
  fetchSessions(true)
}

// Zoom into a bar: its bucket becomes the new range (month -> weeks -> days -> hours).
function zoom(b: Bucket) {
  range.value = { from: b.start < range.value.from ? range.value.from : b.start, to: b.end > range.value.to ? range.value.to : b.end }
}

onMounted(() => fetchSessions())
</script>

<template>
  <section class="space-y-6">
    <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
      <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <TimeRangePicker v-model="range" />
        <button
          type="button"
          :disabled="loading"
          class="rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-xs font-medium sm:min-h-0 text-gray-700 transition hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
          :title="updatedLabel ? `Stand ${updatedLabel} Uhr` : undefined"
          @click="refresh"
        >
          Aktualisieren
        </button>
      </div>
      <p v-if="loading && !sessions.length" class="py-10 text-center text-sm text-gray-500 dark:text-gray-400">Lade...</p>
      <p v-else-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
      <UsageChart v-else :buckets="buckets" :granularity="granularity" 
        :can-zoom="canZoom"
        :settlements="settlements?.settlements"
        @zoom="zoom" />
    </div>

    <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
      <h3 class="mb-3 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">
        Ladevorgänge im Zeitraum ({{ visible.length }})
      </h3>
      <SessionTable :sessions="visible" />
    </div>

    <div v-if="settlements" class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
      <h3 class="mb-3 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Abrechnungen</h3>
      <SettlementTable :data="settlements" />
    </div>
  </section>
</template>
