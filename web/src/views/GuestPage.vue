<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getGuestStatus } from '../lib/api'
import BrandHeader from '../components/BrandHeader.vue'
import GuestLogin from '../components/GuestLogin.vue'
import GuestDashboard from '../components/GuestDashboard.vue'

const props = defineProps<{ uuid: string }>()

const checking = ref(true)
const authenticated = ref(false)
const notFound = ref(false)

async function checkStatus() {
  checking.value = true
  notFound.value = false
  try {
    const status = await getGuestStatus(props.uuid)
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
      <div class="box w-full max-w-sm sm:p-8">
        <BrandHeader />
        <h1 class="heading mt-5">Link ungültig</h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Dieser Link existiert nicht mehr oder wurde erneuert. Frag die Person, die ihn dir gegeben hat, nach dem aktuellen Link.
        </p>
      </div>
    </div>
    <GuestDashboard v-else-if="authenticated" :uuid="props.uuid" @logged-out="onLoggedOut" />
    <GuestLogin v-else :uuid="props.uuid" @success="onLoginSuccess" />
  </main>
</template>
