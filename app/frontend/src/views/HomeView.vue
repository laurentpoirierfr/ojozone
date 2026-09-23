<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { listProducts } from '../api/client.js'

const loading = ref(false)
const error = ref('')
const products = ref([])
const meta = ref({ limit: 0, offset: 0 })

async function searchProducts(search = '') {
  loading.value = true
  error.value = ''
  try {
    const data = await listProducts(search)
    products.value = data.data
    meta.value = data.meta
  } catch (err) {
    error.value = err.message
    products.value = []
    meta.value = { limit: 0, offset: 0 }
  } finally {
    loading.value = false
  }
}

const q = ref('')

function submitSearch() {
  searchProducts(q.value.trim())
}

onMounted(() => searchProducts(''))
</script>

<template>
  <section class="hero">
    <h1>Comparez le coût de la vie en zone euro</h1>
    <form class="search-form" role="search" @submit.prevent="submitSearch">
      <label for="search" class="visually-hidden">Rechercher un produit</label>
      <input id="search" v-model="q" type="search" placeholder="Produit, marque…" />
      <button type="submit">Rechercher</button>
    </form>
  </section>

  <p v-if="loading" class="muted" aria-live="polite">Chargement…</p>
  <p v-else-if="error" class="error" role="alert">{{ error }}</p>

  <ul v-else class="product-list">
    <li v-for="product in products" :key="product.id" class="card">
      <RouterLink :to="`/products/${product.id}`">
        <strong>{{ product.name }}</strong>
        <span v-if="product.category_slug" class="muted"> · {{ product.category_slug }}</span>
      </RouterLink>
    </li>
  </ul>
  <p v-if="!loading && products.length === 0 && !error" class="muted">Aucun produit trouvé.</p>
</template>

<style scoped>
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
}

.hero h1 {
  font-size: 1.75rem;
  margin: 0 0 1rem;
}

.search-form {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.search-form input {
  flex: 1;
}

.error {
  color: var(--err);
}
</style>