<script setup lang="ts">
import { computed } from 'vue'
import type { Settlements } from '../lib/types'
import IconPlug from './IconPlug.vue'
import IconSun from './IconSun.vue'
import { formatDateTime, formatEUR, formatKWh, formatKWhValue } from '../lib/format'

const shortFormat = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit' })

/** Compact date for phones: "28.10.26 12:11". */
function formatShort(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : shortFormat.format(d).replace(',', '')
}

const th = 'px-1 py-2 sm:px-3'
const td = 'px-1 py-2 whitespace-nowrap sm:px-3'

const props = defineProps<{ data: Settlements }>()

// Newest first.
const rows = computed(() => [...props.data.settlements].reverse())
const settled = computed(() => ({
  solar: props.data.settlements.reduce((a, s) => a + s.solarKWh, 0),
  grid: props.data.settlements.reduce((a, s) => a + s.gridKWh, 0),
  kWh: props.data.settlements.reduce((a, s) => a + s.chargedKWh, 0),
  eur: props.data.settlements.reduce((a, s) => a + s.amountEUR, 0),
}))
const total = computed(() => ({
  solar: settled.value.solar + props.data.open.solarKWh,
  grid: settled.value.grid + props.data.open.gridKWh,
  kWh: settled.value.kWh + props.data.open.chargedKWh,
  eur: settled.value.eur + props.data.open.amountEUR,
}))
</script>

<template>
  <table class="w-full text-left text-[11px] sm:text-sm">
    <thead>
      <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
        <th :class="th"><span class="sm:hidden">Datum</span><span class="hidden sm:inline">Abgerechnet am</span></th>
        <th :class="th">Geladen</th>
        <th :class="[th, 'text-right']">Summe</th>
      </tr>
    </thead>
    <tbody>
      <tr class="border-b border-gray-100 dark:border-gray-800">
        <td :class="[td, 'text-gray-500 dark:text-gray-400']">
          <template v-if="data.open.since">
            <span class="sm:hidden">Offen seit<br />{{ formatShort(data.open.since) }}</span>
            <span class="hidden sm:inline">Offen seit {{ formatDateTime(data.open.since) }}</span>
          </template>
          <template v-else>Offen</template>
        </td>
        <td :class="td">
          {{ formatKWh(data.open.chargedKWh) }}
          <span class="block text-gray-500 sm:ml-1 sm:inline dark:text-gray-400" data-testid="split">(<IconSun class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(data.open.solarKWh) }} <IconPlug class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(data.open.gridKWh) }})</span>
        </td>
        <td :class="[td, 'text-right']">{{ formatEUR(data.open.amountEUR) }}</td>
      </tr>
      <tr v-for="s in rows" :key="s.id" class="border-b border-gray-100 last:border-0 dark:border-gray-800">
        <td :class="td">
          <span class="sm:hidden">{{ formatShort(s.at) }}</span>
          <span class="hidden sm:inline">{{ formatDateTime(s.at) }}</span>
        </td>
        <td :class="td">
          {{ formatKWh(s.chargedKWh) }}
          <span class="block text-gray-500 sm:ml-1 sm:inline dark:text-gray-400" data-testid="split">(<IconSun class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(s.solarKWh) }} <IconPlug class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(s.gridKWh) }})</span>
        </td>
        <td :class="[td, 'text-right']">{{ formatEUR(s.amountEUR) }}</td>
      </tr>
    </tbody>
    <tfoot>
      <tr data-testid="settled-total" class="border-t border-gray-300 dark:border-gray-700">
        <td :class="td"><span class="sm:hidden">Abgerechnet</span><span class="hidden sm:inline">Bereits abgerechnet</span></td>
        <td :class="td">
          {{ formatKWh(settled.kWh) }}
          <span class="block text-gray-500 sm:ml-1 sm:inline dark:text-gray-400" data-testid="split">(<IconSun class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(settled.solar) }} <IconPlug class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(settled.grid) }})</span>
        </td>
        <td :class="[td, 'text-right']">{{ formatEUR(settled.eur) }}</td>
      </tr>
      <tr data-testid="open-total">
        <td :class="td"><span class="sm:hidden">Offen</span><span class="hidden sm:inline">Noch offen</span></td>
        <td :class="td">
          {{ formatKWh(data.open.chargedKWh) }}
          <span class="block text-gray-500 sm:ml-1 sm:inline dark:text-gray-400" data-testid="split">(<IconSun class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(data.open.solarKWh) }} <IconPlug class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(data.open.gridKWh) }})</span>
        </td>
        <td :class="[td, 'text-right']">{{ formatEUR(data.open.amountEUR) }}</td>
      </tr>
      <tr data-testid="grand-total" class="border-t border-gray-300 font-semibold dark:border-gray-700">
        <td :class="td">Gesamt</td>
        <td :class="td">
          {{ formatKWh(total.kWh) }}
          <span class="block font-normal text-gray-500 sm:ml-1 sm:inline dark:text-gray-400" data-testid="split">(<IconSun class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(total.solar) }} <IconPlug class="inline size-3.5 align-text-bottom" /> {{ formatKWhValue(total.grid) }})</span>
        </td>
        <td :class="[td, 'text-right']">{{ formatEUR(total.eur) }}</td>
      </tr>
    </tfoot>
  </table>
</template>
