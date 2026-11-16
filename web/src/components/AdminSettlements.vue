<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { createSettlement, deleteSettlement, viewAsSettlements } from '../lib/api'
import { formatDateTime, formatEUR } from '../lib/format'
import type { Settlements } from '../lib/types'

const props = defineProps<{ guestId: string; guestName: string }>()

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
    data.value = await viewAsSettlements(props.guestId)
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
  if (!window.confirm(`Abrechnung für ${props.guestName} zum ${formatDateTime(when.toISOString())} anlegen? Alles bis dahin gilt als bezahlt.`)) return
  run(() => createSettlement(props.guestId, when.toISOString()), 'Abrechnung fehlgeschlagen')
}

function remove(id: string, at: string) {
  if (!window.confirm(`Abrechnung vom ${formatDateTime(at)} löschen?`)) return
  run(() => deleteSettlement(props.guestId, id), 'Löschen fehlgeschlagen')
}

onMounted(load)
</script>

<template>
  <div class="space-y-2 rounded-[10px] bg-gray-100 p-3 dark:bg-gray-950">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm">
        <span class="text-gray-400">Offen<template v-if="data?.open.since"> seit {{ formatDateTime(data.open.since) }}</template>:</span>
        <span class="ml-1 font-semibold">{{ data ? formatEUR(data.open.amountEUR) : '...' }}</span>
      </p>
      <form class="flex flex-wrap items-center gap-2" @submit.prevent="create">
        <input
          v-model="at"
          type="datetime-local"
          aria-label="Abrechnungsdatum"
          required
          :max="maxAt"
          class="field !w-auto"
        />
        <button
          type="submit"
          :disabled="busy || !data"
          class="btn-primary"
        >
          Abrechnung anlegen
        </button>
      </form>
    </div>
    <ul v-if="rows.length" class="divide-y divide-gray-200 text-sm dark:divide-gray-800">
      <li v-for="s in rows" :key="s.id" class="flex flex-wrap items-center justify-between gap-2 py-1.5">
        <span>{{ formatDateTime(s.at) }}</span>
        <span class="font-medium">{{ formatEUR(s.amountEUR) }}</span>
        <button
          type="button"
          :disabled="busy"
          class="btn-danger"
          @click="remove(s.id, s.at)"
        >
          Löschen
        </button>
      </li>
    </ul>
    <p v-else-if="data" class="text-sm text-gray-400">Noch keine Abrechnungen.</p>
    <p v-if="error" role="alert" class="alert-error">{{ error }}</p>
  </div>
</template>
