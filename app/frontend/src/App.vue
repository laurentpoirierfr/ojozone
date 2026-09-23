<script setup>
import { RouterView } from 'vue-router'
import logo from './assets/logo.png'
</script>

<template>
  <header class="site-header">
    <RouterLink to="/" class="brand">
      <img class="brand-logo" :src="logo" alt="" width="28" height="28" />
      <span>OjoZone</span>
    </RouterLink>
    <nav aria-label="Principal">
      <RouterLink to="/">Rechercher</RouterLink>
    </nav>
  </header>
  <main class="site-main">
    <RouterView />
  </main>
  <footer class="site-footer">Données de démonstration pilotes.</footer>
</template>