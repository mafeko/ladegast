<script setup lang="ts">
import { computed } from 'vue'
import type { SessionView } from '../lib/types'
import { formatDateTime, formatEUR, formatKWhValue } from '../lib/format'

const props = defineProps<{ sessions: SessionView[] }>()

const total = computed(() => ({
  chargedKWh: props.sessions.reduce((a, s) => a + s.chargedKWh, 0),
  solarKWh: props.sessions.reduce((a, s) => a + s.solarKWh, 0),
  gridKWh: props.sessions.reduce((a, s) => a + s.gridKWh, 0),
  costEUR: props.sessions.reduce((a, s) => a + s.costEUR, 0),
}))

const shortFormat = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })

/** Compact date for phones: "23.10. 13:08". */
function formatShort(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : shortFormat.format(d).replace(',', '')
}

const num = formatKWhValue
const eur = (v: number) => formatEUR(v).replace(/\s*€/, '')

const th = 'px-0.5 py-2 text-right font-semibold sm:px-3'
const td = 'px-0.5 py-2 text-right whitespace-nowrap tabular-nums sm:px-3'
</script>

<template>
  <table class="w-full table-fixed text-[11px] sm:table-auto sm:text-sm">
    <thead>
      <tr class="border-b border-gray-200 text-xs text-gray-500 dark:border-gray-800 dark:text-gray-400">
        <th class="w-[25%] px-0.5 py-2 text-left font-semibold tracking-wide uppercase sm:w-auto sm:px-3">Datum</th>
        <th :class="th" title="Geladen (kWh)" aria-label="Geladen (kWh)">
          <svg viewBox="0 0 24 24" class="ml-auto size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" data-icon="car">
            <path d="M5 17h14M3 13l2-6a2 2 0 0 1 2-1.5h10a2 2 0 0 1 2 1.5l2 6v4H3z" />
            <circle cx="7.5" cy="17" r="1.5" />
            <circle cx="16.5" cy="17" r="1.5" />
          </svg>
        </th>
        <th :class="th" title="Sonne (kWh)" aria-label="Sonne (kWh)">
          <svg viewBox="0 0 24 24" class="ml-auto size-4 text-amber-500" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true" data-icon="sun">
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
          </svg>
        </th>
        <th :class="th" title="Netz (kWh)" aria-label="Netz (kWh)">
          <svg viewBox="0 0 24 24" class="ml-auto size-4 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" data-icon="plug">
            <path d="M9 2v6M15 2v6M6 8h12v3a6 6 0 0 1-12 0zM12 17v5" />
          </svg>
        </th>
        <th :class="th" title="Kosten" aria-label="Kosten"><span class="text-base leading-4" data-icon="euro" aria-hidden="true">€</span></th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(s, i) in sessions" :key="i" class="border-b border-gray-100 last:border-0 dark:border-gray-800">
        <td class="px-0.5 py-2 whitespace-nowrap sm:px-3">
          <span class="sm:hidden">{{ formatShort(s.created) }}</span>
          <span class="hidden sm:inline">{{ formatDateTime(s.created) }}</span>
        </td>
        <td :class="td">{{ num(s.chargedKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ num(s.solarKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ num(s.gridKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ eur(s.costEUR) }}<span class="hidden sm:inline"> €</span></td>
      </tr>
      <tr v-if="sessions.length === 0">
        <td colspan="5" class="px-3 py-6 text-center text-gray-500 dark:text-gray-400">Keine Ladevorgänge</td>
      </tr>
    </tbody>
    <tfoot v-if="sessions.length > 0">
      <tr data-testid="session-total" class="border-t border-gray-300 font-semibold dark:border-gray-700">
        <td class="px-0.5 py-2 sm:px-3">Gesamt</td>
        <td :class="td">{{ num(total.chargedKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ num(total.solarKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ num(total.gridKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
        <td :class="td">{{ eur(total.costEUR) }}<span class="hidden sm:inline"> €</span></td>
      </tr>
    </tfoot>
  </table>
</template>
