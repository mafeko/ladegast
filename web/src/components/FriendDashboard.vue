<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getDashboard, getSessions, getSettlements, friendLogout } from '../lib/api'
import type { Dashboard } from '../lib/types'
import SummaryCard from './SummaryCard.vue'
import UsageExplorer from './UsageExplorer.vue'
// Charge mode switching (Schnell laden / Smart laden) is disabled for now.
// import ModeSwitch from './ModeSwitch.vue'

const props = defineProps<{ uuid: string }>()
const emit = defineEmits<{ loggedOut: [] }>()

const dashboard = ref<Dashboard | null>(null)
const error = ref('')
const loading = ref(true)

async function load(silent = false) {
  loading.value = !silent
  error.value = ''
  try {
    dashboard.value = await getDashboard(props.uuid)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Laden fehlgeschlagen'
  } finally {
    loading.value = false
  }
}

async function logout() {
  await friendLogout(props.uuid)
  emit('loggedOut')
}

onMounted(load)

defineExpose({ load })
</script>

<template>
  <div class="mx-auto max-w-5xl px-4 py-6 sm:py-10">
    <div class="mb-8 flex items-center justify-between gap-4">
      <h1 class="text-xl font-semibold sm:text-2xl tracking-tight">
        {{ dashboard ? `Hallo ${dashboard.friend.name}` : 'Ladegast' }}
      </h1>
      <button
        type="button"
        class="rounded-lg border border-gray-300 bg-white min-h-11 px-4 py-2 text-sm font-medium sm:min-h-0 text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
        @click="logout"
      >
        Abmelden
      </button>
    </div>
    <p v-if="loading" class="text-sm text-gray-500 dark:text-gray-400">Lade...</p>
    <p v-else-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
      {{ error }}
    </p>
    <template v-else-if="dashboard">
      <!-- Charge mode switching disabled for now (needs more development time)
      <div class="mb-6 rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
        <ModeSwitch :uuid="props.uuid" @changed="load(true)" />
      </div>
      -->
      <div class="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <SummaryCard title="Heute" :summary="dashboard.today" />
        <SummaryCard title="Diese Woche" :summary="dashboard.thisWeek" />
        <SummaryCard title="Dieser Monat" :summary="dashboard.thisMonth" />
        <SummaryCard title="Dieses Jahr" :summary="dashboard.thisYear" />
      </div>
      <UsageExplorer
        :cache-key="`friend:${props.uuid}`"
        :load="() => getSessions(props.uuid)"
        :load-settlements="() => getSettlements(props.uuid)"
      />
    </template>
  </div>
</template>
