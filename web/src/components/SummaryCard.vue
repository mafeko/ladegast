<script setup lang="ts">
import type { Summary } from '../lib/types'
import { formatEUR, formatKWh, solarSharePercent } from '../lib/format'

const props = defineProps<{ title: string; summary: Summary }>()
</script>

<template>
  <section>
    <h3>{{ props.title }}</h3>
    <p>{{ formatKWh(props.summary.chargedKWh) }} geladen</p>
    <p>davon {{ solarSharePercent(props.summary.solarKWh, props.summary.chargedKWh) }}% Sonne</p>
    <p>{{ formatEUR(props.summary.costEUR) }}</p>
  </section>
</template>
