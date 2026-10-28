<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { createSettlement, deleteSettlement, viewAsSettlements } from '../lib/api'
import { formatDateTime, formatEUR } from '../lib/format'
import type { Settlements } from '../lib/types'

const props = defineProps<{ friendId: string; friendName: string }>()

const data = ref<Settlements | null>(null)
const error = ref('')
const busy = ref(false)

// datetime-local value ("YYYY-MM-DDTHH:mm") in local time
function localNow(): string {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().slice(0, 16)
}
const maxAt = localNow()
const at = ref(maxAt)

const rows = computed(() => [...(data.value?.settlements ?? [])].reverse())

async function load() {
  error.value = ''
  try {
    data.value = await viewAsSettlements(props.friendId)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Laden fehlgeschlagen'
  }
}

async function run(action: () => Promise<unknown>, failure: string) {
  busy.value = true
  error.value = ''
  try {
    await action()
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : failure
  } finally {
    busy.value = false
  }
}

function create() {
  if (!at.value) return
  const when = new Date(at.value)
  if (!window.confirm(`Abrechnung für ${props.friendName} zum ${formatDateTime(when.toISOString())} anlegen? Alles bis dahin gilt als bezahlt.`)) return
  run(() => createSettlement(props.friendId, when.toISOString()), 'Abrechnung fehlgeschlagen')
}

function remove(id: string, at: string) {
  if (!window.confirm(`Abrechnung vom ${formatDateTime(at)} löschen?`)) return
  run(() => deleteSettlement(props.friendId, id), 'Löschen fehlgeschlagen')
}

onMounted(load)
</script>

<template>
  <div class="space-y-2 rounded-lg border border-gray-200 p-3 dark:border-gray-800">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm">
        <span class="text-gray-500 dark:text-gray-400">Offen<template v-if="data?.open.since"> seit {{ formatDateTime(data.open.since) }}</template>:</span>
        <span class="ml-1 font-semibold">{{ data ? formatEUR(data.open.amountEUR) : '...' }}</span>
      </p>
      <form class="flex flex-wrap items-center gap-2" @submit.prevent="create">
        <input
          v-model="at"
          type="datetime-local"
          aria-label="Abrechnungsdatum"
          required
          :max="maxAt"
          class="rounded-lg border border-gray-300 bg-white min-h-11 px-2 py-1 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
        <button
          type="submit"
          :disabled="busy || !data"
          class="rounded-lg bg-indigo-600 min-h-11 px-3 py-1.5 text-xs font-medium sm:min-h-0 text-white transition hover:bg-indigo-500 disabled:opacity-50"
        >
          Abrechnung anlegen
        </button>
      </form>
    </div>
    <ul v-if="rows.length" class="divide-y divide-gray-100 text-sm dark:divide-gray-800">
      <li v-for="s in rows" :key="s.id" class="flex flex-wrap items-center justify-between gap-2 py-1.5">
        <span>{{ formatDateTime(s.at) }}</span>
        <span class="font-medium">{{ formatEUR(s.amountEUR) }}</span>
        <button
          type="button"
          :disabled="busy"
          class="rounded-lg border border-red-300 bg-white min-h-11 px-2 py-1 text-xs font-medium sm:min-h-0 text-red-600 transition hover:bg-red-50 disabled:opacity-50 dark:border-red-900 dark:bg-gray-900 dark:text-red-400 dark:hover:bg-red-950/40"
          @click="remove(s.id, s.at)"
        >
          Löschen
        </button>
      </li>
    </ul>
    <p v-else-if="data" class="text-sm text-gray-500 dark:text-gray-400">Noch keine Abrechnungen.</p>
    <p v-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">{{ error }}</p>
  </div>
</template>
