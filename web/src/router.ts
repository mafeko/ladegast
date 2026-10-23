import { createRouter, createWebHistory } from 'vue-router'
import FriendPage from './views/FriendPage.vue'
import AdminPage from './views/AdminPage.vue'
import AdminViewAsPage from './views/AdminViewAsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/admin' },
    { path: '/public/l/:uuid', name: 'friend', component: FriendPage, props: true },
    { path: '/admin', name: 'admin', component: AdminPage },
    { path: '/admin/view-as/:id', name: 'admin-view-as', component: AdminViewAsPage, props: true },
  ],
})
