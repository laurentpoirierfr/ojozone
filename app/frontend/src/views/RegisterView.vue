<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { register } from '../auth/store.js'

const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await register(email.value.trim(), password.value, 'fr')
    router.push({ name: 'account' })
  } catch (err) {
    error.value = err.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-card card">
    <h1>Inscription</h1>
    <form class="auth-form" @submit.prevent="submit">
      <label for="email">Adresse e-mail</label>
      <input id="email" v-model="email" type="email" autocomplete="email" required />
      <label for="password">Mot de passe</label>
      <input id="password" v-model="password" type="password" autocomplete="new-password" required />
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button type="submit" :disabled="submitting">{{ submitting ? 'Inscription…' : 'Créer mon compte' }}</button>
    </form>
    <p class="muted">Déjà inscrit ? <RouterLink to="/login">Connectez-vous</RouterLink>.</p>
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