<script setup lang="ts">
import type { Summary } from '../lib/types'
import { formatEUR, formatKWh, solarSharePercent } from '../lib/format'

const props = defineProps<{ title: string; summary: Summary }>()
</script>

<template>
  <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-800 dark:bg-gray-900">
    <h3 class="text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">{{ props.title }}</h3>
    <p class="mt-2 text-lg font-semibold tracking-tight">{{ formatKWh(props.summary.chargedKWh) }}</p>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
      davon {{ solarSharePercent(props.summary.solarKWh, props.summary.chargedKWh) }}% Sonne
    </p>
    <p class="mt-1 text-sm font-medium text-gray-700 dark:text-gray-300">{{ formatEUR(props.summary.costEUR) }}</p>
  </section>
</template>
