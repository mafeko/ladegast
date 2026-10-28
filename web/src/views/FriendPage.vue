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
    <div v-if="checking" class="flex min-h-dvh items-center justify-center text-sm text-gray-500 dark:text-gray-400">Lade...</div>
    <div v-else-if="notFound" class="flex min-h-dvh items-center justify-center px-4">
      <div class="w-full max-w-sm rounded-2xl border border-gray-200 bg-white p-6 shadow-sm sm:p-8 dark:border-gray-800 dark:bg-gray-900">
        <h1 class="text-lg font-semibold tracking-tight">Link ungültig</h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Dieser Link existiert nicht mehr oder wurde erneuert. Frag die Person, die ihn dir gegeben hat, nach dem aktuellen Link.
        </p>
      </div>
    </div>
    <FriendDashboard v-else-if="authenticated" :uuid="props.uuid" @logged-out="onLoggedOut" />
    <FriendLogin v-else :uuid="props.uuid" @success="onLoginSuccess" />
  </main>
</template>
