<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  PRESETS,
  addDays,
  formatRange,
  fromDateInput,
  isoWeek,
  matchingPreset,
  rangeFor,
  shiftRange,
  startOfDay,
  toDateInput,
  type Range,
  type RangeKind,
} from '../lib/range'

const props = defineProps<{ modelValue: Range }>()
const emit = defineEmits<{ 'update:modelValue': [range: Range] }>()

type Tab = RangeKind | 'custom'
const TABS: { id: Tab; label: string }[] = [
  { id: 'day', label: 'Tag' },
  { id: 'week', label: 'Woche' },
  { id: 'month', label: 'Monat' },
  { id: 'year', label: 'Jahr' },
  { id: 'custom', label: 'Von - Bis' },
]
const MONTHS = ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez']

const open = ref(false)
const tab = ref<Tab>('month')
const root = ref<HTMLElement | null>(null)
const pickerYear = ref(props.modelValue.from.getFullYear())
const dayInput = ref(toDateInput(props.modelValue.from))
const customFrom = ref(toDateInput(props.modelValue.from))
const customTo = ref(toDateInput(addDays(props.modelValue.to, -1)))

const label = computed(() => formatRange(props.modelValue))
const activePresetId = computed(() => matchingPreset(props.modelValue, new Date())?.id)
const weekHint = computed(() => {
  const d = fromDateInput(dayInput.value)
  if (!d) return ''
  const r = rangeFor('week', d)
  const { week, year } = isoWeek(r.from)
  return `KW ${week} / ${year}: ${fmt(r.from)} - ${fmt(addDays(r.to, -1))}`
})
const customError = computed(() => {
  const from = fromDateInput(customFrom.value)
  const to = fromDateInput(customTo.value)
  if (!from || !to) return 'Bitte beide Daten angeben.'
  if (to < from) return 'Das Ende liegt vor dem Beginn.'
  return ''
})

function fmt(d: Date): string {
  return d.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

// Keep the inputs in sync when the range changes from outside (arrows, bar zoom).
watch(
  () => props.modelValue,
  (r) => {
    pickerYear.value = r.from.getFullYear()
    dayInput.value = toDateInput(r.from)
    customFrom.value = toDateInput(r.from)
    customTo.value = toDateInput(addDays(r.to, -1))
  },
)

function select(r: Range) {
  emit('update:modelValue', r)
  open.value = false
}

function pickDay() {
  const d = fromDateInput(dayInput.value)
  if (d) select(rangeFor(tab.value === 'week' ? 'week' : 'day', d))
}

function pickMonth(m: number) {
  select(rangeFor('month', new Date(pickerYear.value, m, 1)))
}

function pickYear() {
  select(rangeFor('year', new Date(pickerYear.value, 0, 1)))
}

function applyCustom() {
  const from = fromDateInput(customFrom.value)
  const to = fromDateInput(customTo.value)
  if (!from || !to || customError.value) return
  select({ from: startOfDay(from), to: addDays(startOfDay(to), 1) })
}

function onDocumentClick(e: MouseEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) open.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})

// Mobile first: 44px touch targets and 16px input text (iOS zooms into smaller inputs); sm: tightens both.
const arrowClass =
  'min-h-11 min-w-11 rounded-lg border border-gray-300 bg-white px-3 text-base font-medium text-gray-700 transition hover:bg-gray-50 sm:min-h-0 sm:min-w-0 sm:py-2 sm:text-sm dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800'
const inputClass =
  'min-h-11 rounded-lg border border-gray-300 bg-white px-3 py-2 text-base focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:min-h-0 sm:text-sm dark:border-gray-700 dark:bg-gray-950'
const primaryClass =
  'min-h-11 rounded-lg bg-indigo-600 px-4 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:opacity-50 sm:min-h-0 sm:py-2'
</script>

<template>
  <div ref="root" class="relative flex w-full items-stretch gap-2 sm:inline-flex sm:w-auto sm:items-center">
    <button type="button" aria-label="Voriger Zeitraum" :class="arrowClass" @click="emit('update:modelValue', shiftRange(modelValue, -1))">‹</button>
    <button
      type="button"
      aria-haspopup="dialog"
      :aria-expanded="open"
      class="min-h-11 min-w-0 flex-1 truncate rounded-lg border border-gray-300 bg-white px-4 text-left text-sm font-medium text-gray-900 transition hover:bg-gray-50 sm:min-h-0 sm:min-w-48 sm:flex-none sm:py-2 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-100 dark:hover:bg-gray-800"
      @click="open = !open"
    >
      {{ label }} <span aria-hidden="true" class="float-right ml-3 text-gray-400">▾</span>
    </button>
    <button type="button" aria-label="Nächster Zeitraum" :class="arrowClass" @click="emit('update:modelValue', shiftRange(modelValue, 1))">›</button>

    <!-- Phones: bottom sheet over a backdrop. sm and up: popover under the button. -->
    <div v-if="open" data-testid="picker-backdrop" class="fixed inset-0 z-20 bg-black/40 sm:hidden" @click="open = false"></div>
    <div
      v-if="open"
      role="dialog"
      aria-label="Zeitraum wählen"
      class="fixed inset-x-0 bottom-0 z-30 flex max-h-[85dvh] flex-col overflow-y-auto rounded-t-2xl border border-gray-200 bg-white pb-[env(safe-area-inset-bottom)] shadow-xl sm:absolute sm:inset-x-auto sm:top-full sm:bottom-auto sm:left-0 sm:z-20 sm:mt-2 sm:max-h-none sm:w-[34rem] sm:flex-row sm:overflow-hidden sm:rounded-2xl sm:pb-0 dark:border-gray-800 dark:bg-gray-900"
    >
      <div class="flex items-center justify-between border-b border-gray-100 px-4 py-2 sm:hidden dark:border-gray-800">
        <span class="text-sm font-semibold">Zeitraum wählen</span>
        <button type="button" aria-label="Schließen" class="min-h-11 min-w-11 text-xl text-gray-500 dark:text-gray-400" @click="open = false">✕</button>
      </div>
      <div class="border-b border-gray-100 p-3 sm:w-44 sm:shrink-0 sm:border-r sm:border-b-0 dark:border-gray-800">
        <p class="mb-2 px-2 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Schnellauswahl</p>
        <ul class="grid grid-cols-2 gap-1 sm:grid-cols-1">
          <li v-for="p in PRESETS" :key="p.id">
            <button
              type="button"
              class="min-h-11 w-full rounded-lg px-3 text-left text-sm transition hover:bg-gray-100 sm:min-h-0 sm:px-2 sm:py-1.5 dark:hover:bg-gray-800"
              :class="activePresetId === p.id ? 'bg-indigo-50 font-medium text-indigo-700 dark:bg-indigo-950/50 dark:text-indigo-300' : ''"
              @click="select(p.range(new Date()))"
            >
              {{ p.label }}
            </button>
          </li>
        </ul>
      </div>

      <div class="flex-1 p-3">
        <p class="mb-2 px-1 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Auswählen</p>
        <div role="tablist" class="mb-3 flex flex-wrap gap-1">
          <button
            v-for="t in TABS"
            :key="t.id"
            type="button"
            role="tab"
            :aria-selected="tab === t.id"
            class="min-h-11 rounded-lg px-3 text-sm transition sm:min-h-0 sm:px-2.5 sm:py-1"
            :class="tab === t.id ? 'bg-indigo-600 font-medium text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'"
            @click="tab = t.id"
          >
            {{ t.label }}
          </button>
        </div>

        <div v-if="tab === 'day' || tab === 'week'" class="space-y-2">
          <label class="block text-sm text-gray-600 dark:text-gray-300">
            {{ tab === 'day' ? 'Datum' : 'Ein Tag der gewünschten Woche' }}
            <input v-model="dayInput" type="date" class="mt-1 block w-full" :class="inputClass" @change="pickDay" />
          </label>
          <p v-if="tab === 'week'" class="text-xs text-gray-500 dark:text-gray-400">{{ weekHint }}</p>
        </div>

        <div v-else-if="tab === 'month' || tab === 'year'" class="space-y-3">
          <div class="flex items-center justify-between">
            <button type="button" aria-label="Vorjahr" :class="arrowClass" @click="pickerYear--">‹</button>
            <span class="text-sm font-semibold">{{ pickerYear }}</span>
            <button type="button" aria-label="Folgejahr" :class="arrowClass" @click="pickerYear++">›</button>
          </div>
          <div v-if="tab === 'month'" class="grid grid-cols-3 gap-1 sm:grid-cols-4">
            <button
              v-for="(m, i) in MONTHS"
              :key="m"
              type="button"
              class="min-h-11 rounded-lg px-2 text-sm transition hover:bg-gray-100 sm:min-h-0 sm:py-2 dark:hover:bg-gray-800"
              :class="
                modelValue.from.getFullYear() === pickerYear && modelValue.from.getMonth() === i && modelValue.to.getTime() === new Date(pickerYear, i + 1, 1).getTime()
                  ? 'bg-indigo-50 font-medium text-indigo-700 dark:bg-indigo-950/50 dark:text-indigo-300'
                  : ''
              "
              @click="pickMonth(i)"
            >
              {{ m }}
            </button>
          </div>
          <button v-else type="button" :class="primaryClass" class="w-full" @click="pickYear">{{ pickerYear }} anzeigen</button>
        </div>

        <form v-else class="space-y-2" @submit.prevent="applyCustom">
          <div class="grid gap-2 min-[420px]:grid-cols-2">
            <label class="block text-sm text-gray-600 dark:text-gray-300">
              Von
              <input v-model="customFrom" type="date" required class="mt-1 block w-full" :class="inputClass" />
            </label>
            <label class="block text-sm text-gray-600 dark:text-gray-300">
              Bis
              <input v-model="customTo" type="date" required class="mt-1 block w-full" :class="inputClass" />
            </label>
          </div>
          <p v-if="customError" role="alert" class="text-xs text-red-600 dark:text-red-400">{{ customError }}</p>
          <button type="submit" :disabled="!!customError" :class="primaryClass" class="w-full sm:w-auto">Übernehmen</button>
        </form>
      </div>
    </div>
  </div>
</template>
