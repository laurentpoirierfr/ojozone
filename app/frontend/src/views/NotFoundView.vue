<script setup>
import { RouterLink } from 'vue-router'
</script>

<template>
  <section class="not-found">
    <h1>Page introuvable</h1>
    <p class="muted">Cette page n'existe pas ou ne vous est pas accessible.</p>
    <RouterLink to="/">Retour à la recherche</RouterLink>
  </section>
</template>