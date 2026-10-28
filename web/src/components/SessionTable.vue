<script setup lang="ts">
import { computed } from 'vue'
import type { SessionView } from '../lib/types'
import { formatDateTime, formatEUR, formatKWh } from '../lib/format'

const props = defineProps<{ sessions: SessionView[] }>()

const total = computed(() => ({
  chargedKWh: props.sessions.reduce((a, s) => a + s.chargedKWh, 0),
  solarKWh: props.sessions.reduce((a, s) => a + s.solarKWh, 0),
  gridKWh: props.sessions.reduce((a, s) => a + s.gridKWh, 0),
  costEUR: props.sessions.reduce((a, s) => a + s.costEUR, 0),
}))
</script>

<template>
  <table class="block w-full text-left text-sm md:table">
    <thead class="hidden md:table-header-group">
      <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
        <th class="px-3 py-2">Datum</th>
        <th class="px-3 py-2">Geladen</th>
        <th class="px-3 py-2">Sonne</th>
        <th class="px-3 py-2">Netz</th>
        <th class="px-3 py-2">Kosten</th>
      </tr>
    </thead>
    <tbody class="block md:table-row-group">
      <tr
        v-for="(s, i) in sessions"
        :key="i"
        class="block border-b border-gray-100 py-2 last:border-0 md:table-row md:py-0 dark:border-gray-800"
      >
        <td data-label="Datum" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatDateTime(s.created) }}</td>
        <td data-label="Geladen" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(s.chargedKWh) }}</td>
        <td data-label="Sonne" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(s.solarKWh) }}</td>
        <td data-label="Netz" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(s.gridKWh) }}</td>
        <td data-label="Kosten" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatEUR(s.costEUR) }}</td>
      </tr>
      <tr v-if="sessions.length === 0" class="block md:table-row">
        <td colspan="5" class="block px-3 md:table-cell py-6 text-center text-gray-500 dark:text-gray-400">Keine Ladevorgänge</td>
      </tr>
    </tbody>
    <tfoot v-if="sessions.length > 0" class="block md:table-footer-group">
      <tr data-testid="session-total" class="block border-t border-gray-300 py-2 font-semibold md:table-row md:py-0 dark:border-gray-700">
        <td class="px-3 py-1 md:py-2">Gesamt</td>
        <td data-label="Geladen" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(total.chargedKWh) }}</td>
        <td data-label="Sonne" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(total.solarKWh) }}</td>
        <td data-label="Netz" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatKWh(total.gridKWh) }}</td>
        <td data-label="Kosten" class="flex justify-between gap-4 px-3 py-1 before:text-xs before:font-semibold before:tracking-wide before:text-gray-500 before:uppercase before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400">{{ formatEUR(total.costEUR) }}</td>
      </tr>
    </tfoot>
  </table>
</template>
