<script setup>
import { ref, onMounted } from 'vue'
import { state } from '../auth/store.js'
import { listModerationQueue, approveContribution, rejectContribution } from '../api/client.js'

const items = ref([])
const status = ref('pending')
const error = ref('')
const acting = ref('')

async function load() {
  error.value = ''
  try {
    const data = await listModerationQueue(state.accessToken, status.value)
    items.value = data.data
  } catch (err) {
    error.value = err.message
  }
}

async function decide(id, note, approve) {
  acting.value = id
  error.value = ''
  try {
    if (approve) {
      await approveContribution(state.accessToken, id, note)
    } else {
      await rejectContribution(state.accessToken, id, note)
    }
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
  <section>
    <h1>Modération</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <div class="filters">
      <label for="status">Statut</label>
      <select id="status" v-model="status" @change="load">
        <option value="pending">En attente</option>
        <option value="approved">Approuvées</option>
        <option value="rejected">Rejetées</option>
      </select>
    </div>

    <p v-if="items.length === 0" class="muted">Aucune contribution dans cette file.</p>
    <ul v-else class="product-list">
      <li v-for="item in items" :key="item.id" class="card">
        <div class="moderation-row">
          <div>
            <strong>{{ item.subject }}</strong>
            <p class="muted">
              {{ item.type }} · {{ item.amount }} {{ item.currency }} · {{ item.location.name }}
            </p>
          </div>
          <div class="moderation-actions">
            <span class="badge" :class="item.status === 'approved' ? 'badge-ok' : item.status === 'rejected' ? 'badge-warn' : ''">
              {{ item.status }}
            </span>
            <template v-if="item.status === 'pending'">
              <button v-if="acting !== item.id" type="button" @click="decide(item.id, '', true)">Approuver</button>
              <button v-if="acting !== item.id" type="button" class="danger" @click="decide(item.id, 'rejeté', false)">Rejeter</button>
              <span v-else class="muted">Traitement…</span>
            </template>
          </div>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.moderation-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.moderation-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.danger {
  background: var(--err);
  border-color: var(--err);
}

.error {
  color: var(--err);
}
</style>