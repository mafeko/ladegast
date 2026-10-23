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
  <main class="mx-auto max-w-5xl px-4 py-10">
    <RouterLink to="/admin" class="text-sm font-medium text-indigo-600 hover:underline dark:text-indigo-400">
      ← Zurück
    </RouterLink>
    <p v-if="loading" class="mt-6 text-sm text-gray-500 dark:text-gray-400">Lade...</p>
    <p v-else-if="error" role="alert" class="mt-6 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
      {{ error }}
    </p>
    <template v-else-if="dashboard">
      <h2 class="mt-6 mb-6 text-2xl font-semibold tracking-tight">Ansicht als {{ dashboard.friend.name }}</h2>
      <div class="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <SummaryCard title="Heute" :summary="dashboard.today" />
        <SummaryCard title="Diese Woche" :summary="dashboard.thisWeek" />
        <SummaryCard title="Dieser Monat" :summary="dashboard.thisMonth" />
        <SummaryCard title="Dieses Jahr" :summary="dashboard.thisYear" />
      </div>
      <div class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900">
        <SessionTable :sessions="sessions" />
      </div>
    </template>
  </main>
</template>
