<script setup lang="ts">
import type { SessionView } from '../lib/types'
import { formatDateTime, formatEUR, formatKWh } from '../lib/format'

defineProps<{ sessions: SessionView[] }>()
</script>

<template>
  <table>
    <thead>
      <tr>
        <th>Datum</th>
        <th>Geladen</th>
        <th>Sonne</th>
        <th>Netz</th>
        <th>Kosten</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(s, i) in sessions" :key="i">
        <td>{{ formatDateTime(s.created) }}</td>
        <td>{{ formatKWh(s.chargedKWh) }}</td>
        <td>{{ formatKWh(s.solarKWh) }}</td>
        <td>{{ formatKWh(s.gridKWh) }}</td>
        <td>{{ formatEUR(s.costEUR) }}</td>
      </tr>
      <tr v-if="sessions.length === 0">
        <td colspan="5">Keine Ladevorgänge</td>
      </tr>
    </tbody>
  </table>
</template>
