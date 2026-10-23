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
    <p v-if="checking">Lade...</p>
    <AdminLogin v-else-if="!authenticated" @success="onLoginSuccess" />
    <template v-else>
      <h1>Freunde verwalten</h1>
      <AdminFriendForm @created="onCreated" />
      <p v-if="lastCreated">
        Neuer Link für {{ lastCreated.id }}: {{ lastCreated.link }} (PIN: {{ lastCreated.pin }})
      </p>
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Fahrzeug</th>
            <th>EUR/kWh Sonne</th>
            <th>EUR/kWh Netz</th>
            <th>Aktionen</th>
          </tr>
        </thead>
        <tbody>
          <AdminFriendRow v-for="f in friends" :key="f.id" :friend="f" @updated="onUpdated" />
        </tbody>
      </table>
      <p v-if="error" role="alert">{{ error }}</p>
    </template>
  </main>
</template>
