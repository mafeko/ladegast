<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listFriends } from '../lib/api'
import type { CreatedFriend, Friend } from '../lib/types'
import AdminLogin from '../components/AdminLogin.vue'
import AdminFriendForm from '../components/AdminFriendForm.vue'
import AdminFriendRow from '../components/AdminFriendRow.vue'

const checking = ref(true)
const authenticated = ref(false)
const friends = ref<Friend[]>([])
const error = ref('')
const lastCreated = ref<CreatedFriend | null>(null)

async function loadFriends() {
  error.value = ''
  try {
    friends.value = await listFriends()
    authenticated.value = true
  } catch {
    authenticated.value = false
  } finally {
    checking.value = false
  }
}

function onLoginSuccess() {
  authenticated.value = true
  loadFriends()
}

function onCreated(friend: CreatedFriend) {
  lastCreated.value = friend
  loadFriends()
}

function onUpdated(updated: Friend) {
  const i = friends.value.findIndex((f) => f.id === updated.id)
  if (i !== -1) friends.value[i] = updated
}

onMounted(loadFriends)
</script>

<template>
  <main>
    <div v-if="checking" class="flex min-h-screen items-center justify-center text-sm text-gray-500 dark:text-gray-400">Lade...</div>
    <AdminLogin v-else-if="!authenticated" @success="onLoginSuccess" />
    <div v-else class="mx-auto max-w-5xl px-4 py-10">
      <h1 class="mb-6 text-2xl font-semibold tracking-tight">Freunde verwalten</h1>
      <div class="mb-6 rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900">
        <h3 class="mb-4 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Neuen Freund anlegen</h3>
        <AdminFriendForm @created="onCreated" />
      </div>
      <p v-if="lastCreated" class="mb-6 rounded-lg bg-indigo-600 px-4 py-3 text-sm break-all text-white">
        Neuer Link für {{ lastCreated.id }}: {{ lastCreated.link }} (PIN: {{ lastCreated.pin }})
      </p>
      <div class="overflow-x-auto rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
              <th class="px-3 py-2">Name</th>
              <th class="px-3 py-2">Fahrzeug</th>
              <th class="px-3 py-2">EUR/kWh Sonne</th>
              <th class="px-3 py-2">EUR/kWh Netz</th>
              <th class="px-3 py-2">Aktionen</th>
            </tr>
          </thead>
          <tbody>
            <AdminFriendRow v-for="f in friends" :key="f.id" :friend="f" @updated="onUpdated" />
          </tbody>
        </table>
      </div>
      <p v-if="error" role="alert" class="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </div>
  </main>
</template>
