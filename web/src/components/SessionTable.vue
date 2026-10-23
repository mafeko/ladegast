<script setup lang="ts">
import type { SessionView } from '../lib/types'
import { formatDateTime, formatEUR, formatKWh } from '../lib/format'

defineProps<{ sessions: SessionView[] }>()
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
  </table>
</template>
