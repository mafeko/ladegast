<script setup lang="ts">
import { computed } from 'vue'
import type { SessionView, Settlement } from '../lib/types'
import IconPlug from './IconPlug.vue'
import IconSun from './IconSun.vue'
import { formatDateTime, formatEUR, formatKWhValue } from '../lib/format'

const props = defineProps<{ sessions: SessionView[]; settlements?: Settlement[] }>()

type Row = { kind: 'session'; at: string; s: SessionView } | { kind: 'settlement'; at: string; st: Settlement }

// Sessions and settlements interleaved, newest first. A settlement sorts
// above sessions at the same instant: it closes everything before it.
const rows = computed<Row[]>(() => {
  const all: Row[] = [
    ...props.sessions.map((s): Row => ({ kind: 'session', at: s.created, s })),
    ...(props.settlements ?? []).map((st): Row => ({ kind: 'settlement', at: st.at, st })),
  ]
  const ms = (r: Row) => new Date(r.at).getTime()
  return all.sort((a, b) => ms(b) - ms(a) || (a.kind === 'settlement' ? -1 : 1) - (b.kind === 'settlement' ? -1 : 1))
})

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
          <IconSun class="ml-auto" />
        </th>
        <th :class="th" title="Netz (kWh)" aria-label="Netz (kWh)">
          <IconPlug class="ml-auto" />
        </th>
        <th :class="th" title="Kosten" aria-label="Kosten"><span class="text-base leading-4" data-icon="euro" aria-hidden="true">€</span></th>
      </tr>
    </thead>
    <tbody>
      <template v-for="(r, i) in rows" :key="i">
        <tr
          v-if="r.kind === 'settlement'"
          data-testid="settlement-row"
          class="border-b border-gray-100 bg-emerald-50 font-semibold text-emerald-700 last:border-0 dark:border-gray-800 dark:bg-emerald-950/30 dark:text-emerald-400"
        >
          <td class="px-0.5 py-2 whitespace-nowrap sm:px-3">
            <span class="sm:hidden">{{ formatShort(r.st.at) }}</span>
            <span class="hidden sm:inline">{{ formatDateTime(r.st.at) }}</span>
          </td>
          <td colspan="3" class="px-0.5 py-2 text-right sm:px-3">
            Abrechnung · {{ num(r.st.chargedKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px]">kWh</span>
          </td>
          <td :class="td">{{ eur(r.st.amountEUR) }}<span class="hidden sm:inline"> €</span></td>
        </tr>
        <tr v-else class="border-b border-gray-100 last:border-0 dark:border-gray-800">
          <td class="px-0.5 py-2 whitespace-nowrap sm:px-3">
            <span class="sm:hidden">{{ formatShort(r.s.created) }}</span>
            <span class="hidden sm:inline">{{ formatDateTime(r.s.created) }}</span>
          </td>
          <td :class="td">{{ num(r.s.chargedKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
          <td :class="td">{{ num(r.s.solarKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
          <td :class="td">{{ num(r.s.gridKWh) }}<span class="ml-0.5 text-[9px] font-normal sm:text-[10px] text-gray-500 dark:text-gray-400">kWh</span></td>
          <td :class="td">{{ eur(r.s.costEUR) }}<span class="hidden sm:inline"> €</span></td>
        </tr>
      </template>
      <tr v-if="rows.length === 0">
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
