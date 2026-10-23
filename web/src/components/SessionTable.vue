<script setup lang="ts">
import type { SessionView } from '../lib/types'
import { formatDateTime, formatEUR, formatKWh } from '../lib/format'

defineProps<{ sessions: SessionView[] }>()
</script>

<template>
  <table class="w-full text-left text-sm">
    <thead>
      <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
        <th class="px-3 py-2">Datum</th>
        <th class="px-3 py-2">Geladen</th>
        <th class="px-3 py-2">Sonne</th>
        <th class="px-3 py-2">Netz</th>
        <th class="px-3 py-2">Kosten</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="(s, i) in sessions"
        :key="i"
        class="border-b border-gray-100 last:border-0 dark:border-gray-800"
      >
        <td class="px-3 py-2">{{ formatDateTime(s.created) }}</td>
        <td class="px-3 py-2">{{ formatKWh(s.chargedKWh) }}</td>
        <td class="px-3 py-2">{{ formatKWh(s.solarKWh) }}</td>
        <td class="px-3 py-2">{{ formatKWh(s.gridKWh) }}</td>
        <td class="px-3 py-2">{{ formatEUR(s.costEUR) }}</td>
      </tr>
      <tr v-if="sessions.length === 0">
        <td colspan="5" class="px-3 py-6 text-center text-gray-500 dark:text-gray-400">Keine Ladevorgänge</td>
      </tr>
    </tbody>
  </table>
</template>
