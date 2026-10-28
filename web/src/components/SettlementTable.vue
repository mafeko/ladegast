<script setup lang="ts">
import { computed } from 'vue'
import type { Settlements } from '../lib/types'
import { formatDateTime, formatEUR, formatKWh } from '../lib/format'

const props = defineProps<{ data: Settlements }>()

// Newest first.
const rows = computed(() => [...props.data.settlements].reverse())
const total = computed(() => props.data.settlements.reduce((a, s) => a + s.amountEUR, 0))
</script>

<template>
  <table class="w-full text-left text-sm">
    <thead>
      <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
        <th class="px-3 py-2">Abgerechnet am</th>
        <th class="px-3 py-2">Geladen</th>
        <th class="px-3 py-2 text-right">Summe</th>
      </tr>
    </thead>
    <tbody>
      <tr class="border-b border-gray-100 dark:border-gray-800">
        <td class="px-3 py-2 text-gray-500 dark:text-gray-400">
          Offen<template v-if="data.open.since"> seit {{ formatDateTime(data.open.since) }}</template>
        </td>
        <td class="px-3 py-2">{{ formatKWh(data.open.chargedKWh) }}</td>
        <td class="px-3 py-2 text-right">{{ formatEUR(data.open.amountEUR) }}</td>
      </tr>
      <tr v-for="s in rows" :key="s.id" class="border-b border-gray-100 last:border-0 dark:border-gray-800">
        <td class="px-3 py-2">{{ formatDateTime(s.at) }}</td>
        <td class="px-3 py-2">{{ formatKWh(s.chargedKWh) }}</td>
        <td class="px-3 py-2 text-right">{{ formatEUR(s.amountEUR) }}</td>
      </tr>
    </tbody>
    <tfoot>
      <tr class="border-t border-gray-300 font-semibold dark:border-gray-700">
        <td class="px-3 py-2" colspan="2">Gesamt abgerechnet</td>
        <td class="px-3 py-2 text-right">{{ formatEUR(total) }}</td>
      </tr>
    </tfoot>
  </table>
</template>
