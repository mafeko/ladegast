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
    <p v-if="checking">Lade...</p>
    <p v-else-if="notFound">Dieser Link ist ungültig.</p>
    <FriendDashboard v-else-if="authenticated" :uuid="props.uuid" @logged-out="onLoggedOut" />
    <FriendLogin v-else :uuid="props.uuid" @success="onLoginSuccess" />
  </main>
</template>
