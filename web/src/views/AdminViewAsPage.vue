<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { viewAsDashboard, viewAsSessions } from '../lib/api'
import type { Dashboard, SessionView } from '../lib/types'
import SummaryCard from '../components/SummaryCard.vue'
import SessionTable from '../components/SessionTable.vue'

const props = defineProps<{ id: string }>()

const dashboard = ref<Dashboard | null>(null)
const sessions = ref<SessionView[]>([])
const error = ref('')
const loading = ref(true)

async function load() {
  loading.value = true
  error.value = ''
  try {
    dashboard.value = await viewAsDashboard(props.id)
    sessions.value = await viewAsSessions(props.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Laden fehlgeschlagen'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <main>
    <RouterLink to="/admin">Zurück</RouterLink>
    <p v-if="loading">Lade...</p>
    <p v-else-if="error" role="alert">{{ error }}</p>
    <template v-else-if="dashboard">
      <h2>Ansicht als {{ dashboard.friend.name }}</h2>
      <SummaryCard title="Heute" :summary="dashboard.today" />
      <SummaryCard title="Diese Woche" :summary="dashboard.thisWeek" />
      <SummaryCard title="Dieser Monat" :summary="dashboard.thisMonth" />
      <SummaryCard title="Dieses Jahr" :summary="dashboard.thisYear" />
      <SessionTable :sessions="sessions" />
    </template>
  </main>
</template>
