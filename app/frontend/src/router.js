import { createRouter, createWebHistory } from 'vue-router'
import { initAuth, currentUser, refreshTokens, loadCurrentUser } from './auth/store.js'
import HomeView from './views/HomeView.vue'
import ProductView from './views/ProductView.vue'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import AccountView from './views/AccountView.vue'
import ContributeView from './views/ContributeView.vue'
import ModerationView from './views/ModerationView.vue'
import AdminView from './views/AdminView.vue'
import NotFoundView from './views/NotFoundView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/products/:id', name: 'product', component: ProductView, props: true },
    { path: '/login', name: 'login', component: LoginView, meta: { guest: true } },
    { path: '/register', name: 'register', component: RegisterView, meta: { guest: true } },
    { path: '/account', name: 'account', component: AccountView, meta: { requiresAuth: true } },
    { path: '/contribuer', name: 'contribute', component: ContributeView, meta: { requiresAuth: true } },
    { path: '/moderation', name: 'moderation', component: ModerationView, meta: { roles: ['moderator', 'admin'] } },
    { path: '/admin', name: 'admin', component: AdminView, meta: { roles: ['admin'] } },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFoundView },
  ],
})

async function ensureSession() {
  await initAuth()
  if (currentUser.value) return
  try {
    await refreshTokens()
    await loadCurrentUser()
  } catch {
    // session indisponible
  }
}

router.beforeEach(async (to) => {
  await ensureSession()
  if (to.meta.guest && currentUser.value) {
    return { name: 'home' }
  }
  if (to.meta.requiresAuth && !currentUser.value) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.roles && (!currentUser.value || !to.meta.roles.includes(currentUser.value.role))) {
    return { name: 'not-found' }
  }
  return true
})

export default router