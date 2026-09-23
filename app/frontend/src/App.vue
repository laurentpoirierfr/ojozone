<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { currentUser, logout, hasRole } from './auth/store.js'
import logo from './assets/logo.png'

function roleLabel(user) {
  const labels = { member: 'Membre', moderator: 'Modérateur', admin: 'Administrateur' }
  return labels[user?.role] || user?.role || 'Invité'
}

async function handleLogout() {
  await logout()
}
</script>

<template>
  <header class="site-header">
    <RouterLink to="/" class="brand">
      <img class="brand-logo" :src="logo" alt="" width="28" height="28" />
      <span>OjoZone</span>
    </RouterLink>
    <nav aria-label="Principal">
      <RouterLink to="/">Rechercher</RouterLink>
      <RouterLink v-if="currentUser" to="/account">Mon compte</RouterLink>
      <RouterLink v-if="hasRole('moderator', 'admin')" to="/moderation">Modération</RouterLink>
      <RouterLink v-if="hasRole('admin')" to="/admin">Administration</RouterLink>
    </nav>
    <div class="header-actions">
      <template v-if="currentUser">
        <span class="user-name">{{ currentUser.display_name || currentUser.email }}</span>
        <span class="badge badge-role">{{ roleLabel(currentUser) }}</span>
        <button type="button" class="link-button" @click="handleLogout">Déconnexion</button>
      </template>
      <template v-else>
        <RouterLink to="/login">Connexion</RouterLink>
        <RouterLink to="/register">Inscription</RouterLink>
      </template>
    </div>
  </header>
  <main class="site-main">
    <RouterView />
  </main>
  <footer class="site-footer">Données de démonstration pilotes.</footer>
</template>