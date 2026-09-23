<script setup>
import { ref, onMounted } from 'vue'
import { state } from '../auth/store.js'
import {
  listImports,
  listAdminUsers,
  createImport,
  validateImport,
  publishImport,
} from '../api/client.js'

const imports = ref([])
const users = ref([])
const error = ref('')
const busy = ref('')

const resourceTypes = [
  'geo-areas',
  'sources',
  'categories',
  'units',
  'merchants',
  'locations',
  'fuel-types',
  'fuel-prices',
  'housing-observations',
  'income-observations',
  'products',
  'product-prices',
]

const resourceType = ref('units')
const rowsText = ref('')

async function load() {
  error.value = ''
  try {
    const importsData = await listImports(state.accessToken)
    imports.value = importsData.data
    const usersData = await listAdminUsers(state.accessToken)
    users.value = usersData.data
  } catch (err) {
    error.value = err.message
  }
}

async function submitCreate() {
  error.value = ''
  let rows
  try {
    rows = rowsText.value
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line.length > 0)
      .map((line) => JSON.parse(line))
    if (rows.length === 0) throw new Error('Ajoutez au moins une ligne JSON.')
  } catch (err) {
    error.value = `Lignes JSON invalides : ${err.message}`
    return
  }
  try {
    await createImport(state.accessToken, resourceType.value, rows)
    rowsText.value = ''
    await load()
  } catch (err) {
    error.value = err.message
  }
}

async function run(id, action) {
  busy.value = `${action}:${id}`
  error.value = ''
  try {
    if (action === 'validate') await validateImport(state.accessToken, id)
    else await publishImport(state.accessToken, id)
    await load()
  } catch (err) {
    error.value = err.message
  } finally {
    busy.value = ''
  }
}

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function unit(key) {
  const labels = {
    draft: 'Brouillon',
    validated: 'Validé',
    published: 'Publié',
  }
  return labels[key] || key
}

onMounted(load)
</script>

<template>
  <section>
    <h1>Administration</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <h2>Nouvel import</h2>
    <form class="import-form" @submit.prevent="submitCreate">
      <label for="resource-type">Type de ressource</label>
      <select id="resource-type" v-model="resourceType">
        <option v-for="type in resourceTypes" :key="type" :value="type">{{ type }}</option>
      </select>
      <label for="rows">Lignes (un objet JSON par ligne)</label>
      <textarea id="rows" v-model="rowsText" rows="6" placeholder='{"code":"kg","dimension":"mass","to_base_factor":"1"}'></textarea>
      <button type="submit">Créer l'import</button>
    </form>

    <h2>Imports</h2>
    <p v-if="imports.length === 0" class="muted">Aucun import pour le moment.</p>
    <ul v-else class="product-list">
      <li v-for="item in imports" :key="item.id" class="card">
        <div class="import-row">
          <div>
            <strong>{{ item.resource_type }}</strong>
            <p class="muted">
              {{ item.line_count }} lignes · {{ item.valid_count }} valides · {{ item.invalid_count }} invalides ·
              {{ formatDate(item.created_at) }}
            </p>
            <p v-if="item.report && item.report.length > 0" class="report">
              <span v-for="entry in item.report" :key="entry.line">{{ entry.line }}: {{ entry.error }}</span>
            </p>
          </div>
          <div class="import-actions">
            <span class="badge" :class="{ 'badge-ok': item.status === 'published', 'badge-warn': item.status === 'validated' }">
              {{ unit(item.status) }}
            </span>
            <button
              v-if="item.status === 'draft'"
              type="button"
              :disabled="busy === `validate:${item.id}`"
              @click="run(item.id, 'validate')"
            >Valider</button>
            <button
              v-if="item.status === 'validated'"
              type="button"
              :disabled="busy === `publish:${item.id}`"
              @click="run(item.id, 'publish')"
            >Publier</button>
          </div>
        </div>
      </li>
    </ul>

    <h2>Utilisateurs</h2>
    <p v-if="users.length === 0" class="muted">Aucun utilisateur.</p>
    <table v-else>
      <thead>
        <tr>
          <th>E-mail</th>
          <th>Rôle</th>
          <th>Date d'inscription</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.id">
          <td>{{ user.email }}</td>
          <td>{{ user.role }}</td>
          <td>{{ formatDate(user.created_at) }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.import-form {
  display: grid;
  gap: 0.5rem;
  margin-bottom: 2rem;
}

.import-form label {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--muted);
}

.import-form textarea {
  font: inherit;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--fg);
}

.import-form button {
  width: fit-content;
}

.import-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.import-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.report {
  display: grid;
  gap: 0.125rem;
  color: var(--err);
  font-size: 0.85rem;
  margin: 0.5rem 0 0;
}

.error {
  color: var(--err);
}
</style>