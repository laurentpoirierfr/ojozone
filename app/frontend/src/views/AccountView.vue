<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { currentUser, refreshTokens, loadCurrentUser, state } from '../auth/store.js'
import { listMyContributions, patchContribution, deleteContribution } from '../api/client.js'

const error = ref('')
const contributions = ref([])
const editing = ref('')
const edit = ref({ amount: '', quantity: '', unit_code: '' })
const acting = ref('')

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function label(item) {
  if (item.type === 'product_price') return `Prix produit · ${item.subject}`
  if (item.type === 'fuel_price') return `Prix carburant · ${item.subject}`
  return item.type
}

async function load() {
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
}

function startEdit(item) {
  editing.value = item.id
  edit.value = { amount: item.amount, quantity: item.quantity || '', unit_code: item.unit_code || '' }
}

function cancelEdit() {
  editing.value = ''
  edit.value = { amount: '', quantity: '', unit_code: '' }
}

async function saveEdit(id) {
  acting.value = id
  error.value = ''
  try {
    const payload = { amount: edit.value.amount }
    if (edit.value.quantity) payload.quantity = edit.value.quantity
    if (edit.value.unit_code) payload.unit_code = edit.value.unit_code
    await patchContribution(state.accessToken, id, payload)
    cancelEdit()
    await load()
  } catch (err) {
    error.value = err.message
  } finally {
    acting.value = ''
  }
}

async function withdraw(id) {
  if (!window.confirm('Retirer cette contribution ?')) return
  acting.value = id
  error.value = ''
  try {
    await deleteContribution(state.accessToken, id)
    await load()
  } catch (err) {
    error.value = err.message
  } finally {
    acting.value = ''
  }
}

onMounted(load)
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
    <p class="muted">
      <RouterLink to="/contribuer">Soumettre un prix</RouterLink>
    </p>
    <p v-if="contributions.length === 0" class="muted">Aucune contribution pour le moment.</p>
    <table v-else>
      <thead>
        <tr>
          <th>Type</th>
          <th>Statut</th>
          <th>Objet</th>
          <th>Créée le</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in contributions" :key="item.id">
          <td>{{ item.type }}</td>
          <td>{{ item.status }}</td>
          <td>
            <template v-if="editing === item.id">
              <input
                v-model="edit.amount"
                class="edit-input"
                type="text"
                inputmode="decimal"
                :aria-label="`Nouveau prix pour ${item.subject}`"
              />
              {{ item.currency }}
            </template>
            <template v-else>{{ label(item) }}</template>
          </td>
          <td>{{ formatDate(item.created_at) }}</td>
          <td>
            <div v-if="item.status === 'pending'" class="row-actions">
              <template v-if="editing === item.id">
                <button type="button" :disabled="acting === item.id" @click="saveEdit(item.id)">Enregistrer</button>
                <button type="button" class="link-button" @click="cancelEdit">Annuler</button>
              </template>
              <template v-else>
                <button type="button" :disabled="acting === item.id" @click="startEdit(item)">Corriger</button>
                <button
                  type="button"
                  class="danger"
                  :disabled="acting === item.id"
                  @click="withdraw(item.id)"
                >
                  Retirer
                </button>
              </template>
            </div>
            <span v-else class="muted">—</span>
          </td>
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

.row-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.edit-input {
  width: 6rem;
  margin-right: 0.5rem;
}

.danger {
  background: var(--err);
  border-color: var(--err);
}
</style>