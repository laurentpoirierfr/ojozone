<script setup>
import { ref, onMounted } from 'vue'
import { currentUser, refreshTokens, loadCurrentUser, state } from '../auth/store.js'
import { listMyContributions } from '../api/client.js'

const error = ref('')
const contributions = ref([])

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

onMounted(async () => {
  error.value = ''
  try {
    if (!currentUser.value) {
      await refreshTokens()
      await loadCurrentUser()
    }
    const data = await listMyContributions(state.accessToken)
    contributions.value = data.data
  } catch (err) {
    error.value = err.message
  }
})
</script>

<template>
  <section v-if="currentUser">
    <h1>Mon compte</h1>
    <dl class="meta">
      <div><dt>E-mail</dt><dd>{{ currentUser.email }}</dd></div>
      <div><dt>Pseudonyme</dt><dd>{{ currentUser.display_name || '—' }}</dd></div>
      <div><dt>Rôle</dt><dd>{{ currentUser.role }}</dd></div>
      <div><dt>Langue</dt><dd>{{ currentUser.locale }}</dd></div>
    </dl>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <h2>Mes contributions</h2>
    <p v-if="contributions.length === 0" class="muted">Aucune contribution pour le moment.</p>
    <table v-else>
      <thead>
        <tr>
          <th>Type</th>
          <th>Statut</th>
          <th>Objet</th>
          <th>Créée le</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in contributions" :key="item.id">
          <td>{{ item.type }}</td>
          <td>{{ item.status }}</td>
          <td>{{ item.subject }}</td>
          <td>{{ formatDate(item.created_at) }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  margin: 0 0 2rem;
}

.meta dt {
  font-size: 0.75rem;
  text-transform: uppercase;
  color: var(--muted);
}

.meta dd {
  margin: 0.125rem 0 0;
}

.error {
  color: var(--err);
}
</style>