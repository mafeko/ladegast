<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getDashboard, getSessions, friendLogout } from '../lib/api'
import type { Dashboard, SessionView } from '../lib/types'
import SummaryCard from './SummaryCard.vue'
import SessionTable from './SessionTable.vue'
import ModeSwitch from './ModeSwitch.vue'

const props = defineProps<{ uuid: string }>()
const emit = defineEmits<{ loggedOut: [] }>()

const dashboard = ref<Dashboard | null>(null)
const sessions = ref<SessionView[]>([])
const error = ref('')
const loading = ref(true)

async function load() {
  loading.value = true
  error.value = ''
  try {
    dashboard.value = await getDashboard(props.uuid)
    sessions.value = await getSessions(props.uuid)
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
  <div>
    <button type="button" @click="logout">Abmelden</button>
    <p v-if="loading">Lade...</p>
    <p v-else-if="error" role="alert">{{ error }}</p>
    <template v-else-if="dashboard">
      <h2>Hallo {{ dashboard.friend.name }}</h2>
      <ModeSwitch :uuid="props.uuid" @changed="load" />
      <SummaryCard title="Heute" :summary="dashboard.today" />
      <SummaryCard title="Diese Woche" :summary="dashboard.thisWeek" />
      <SummaryCard title="Dieser Monat" :summary="dashboard.thisMonth" />
      <SummaryCard title="Dieses Jahr" :summary="dashboard.thisYear" />
      <SessionTable :sessions="sessions" />
    </template>
  </div>
</template>
