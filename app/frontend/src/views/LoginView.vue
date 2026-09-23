<script setup>
import { ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { login } from '../auth/store.js'

const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await login(email.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(redirect)
  } catch (err) {
    error.value = err.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-card card">
    <h1>Connexion</h1>
    <form class="auth-form" @submit.prevent="submit">
      <label for="email">Adresse e-mail</label>
      <input id="email" v-model="email" type="email" autocomplete="email" required />
      <label for="password">Mot de passe</label>
      <input id="password" v-model="password" type="password" autocomplete="current-password" required />
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button type="submit" :disabled="submitting">{{ submitting ? 'Connexion…' : 'Se connecter' }}</button>
    </form>
    <p class="muted">Pas encore de compte ? <RouterLink to="/register">Inscrivez-vous</RouterLink>.</p>
  </section>
</template>

<style scoped>
.auth-card {
  max-width: 24rem;
  margin: 0 auto;
}

.auth-form {
  display: grid;
  gap: 0.5rem;
}

.auth-form label {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--muted);
}

.error {
  color: var(--err);
  margin: 0.5rem 0 0;
}
</style>