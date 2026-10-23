<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getFriendStatus } from '../lib/api'
import FriendLogin from '../components/FriendLogin.vue'
import FriendDashboard from '../components/FriendDashboard.vue'

const props = defineProps<{ uuid: string }>()

const checking = ref(true)
const authenticated = ref(false)
const notFound = ref(false)

async function checkStatus() {
  checking.value = true
  notFound.value = false
  try {
    const status = await getFriendStatus(props.uuid)
    authenticated.value = status.authenticated
  } catch {
    notFound.value = true
  } finally {
    checking.value = false
  }
}

function onLoginSuccess() {
  authenticated.value = true
}

function onLoggedOut() {
  authenticated.value = false
}

onMounted(checkStatus)
</script>

<template>
  <main>
    <div v-if="checking" class="flex min-h-screen items-center justify-center text-sm text-gray-500 dark:text-gray-400">Lade...</div>
    <div v-else-if="notFound" class="flex min-h-screen items-center justify-center text-sm text-gray-600 dark:text-gray-300">Dieser Link ist ungültig.</div>
    <FriendDashboard v-else-if="authenticated" :uuid="props.uuid" @logged-out="onLoggedOut" />
    <FriendLogin v-else :uuid="props.uuid" @success="onLoginSuccess" />
  </main>
</template>
