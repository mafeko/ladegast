import { createRouter, createWebHistory } from 'vue-router'
import GuestPage from './views/GuestPage.vue'
import AdminPage from './views/AdminPage.vue'
import AdminViewAsPage from './views/AdminViewAsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/admin' },
    { path: '/public/l/:uuid', name: 'guest', component: GuestPage, props: true },
    { path: '/admin', name: 'admin', component: AdminPage },
    { path: '/admin/view-as/:id', name: 'admin-view-as', component: AdminViewAsPage, props: true },
  ],
})
