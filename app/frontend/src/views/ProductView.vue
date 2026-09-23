<script setup>
import { ref, onMounted } from 'vue'
import { getProduct, listProductPrices } from '../api/client.js'

const props = defineProps({
  id: { type: String, required: true },
})

const loading = ref(true)
const error = ref('')
const product = ref(null)
const prices = ref([])

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const productData = await getProduct(props.id)
    const pricesData = await listProductPrices(props.id)
    product.value = productData.data
    prices.value = pricesData.data
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <p v-if="loading" class="muted" aria-live="polite">Chargement…</p>
  <p v-else-if="error" class="error" role="alert">{{ error }}</p>

  <template v-else-if="product">
    <RouterLink to="/" aria-label="Retour à la recherche">← Retour</RouterLink>
    <h1>{{ product.name }}</h1>
    <dl class="meta">
      <div><dt>Unité de référence</dt><dd>{{ product.reference_unit }}</dd></div>
      <div><dt>Catégorie</dt><dd>{{ product.category_slug }}</dd></div>
      <div v-if="product.brand"><dt>Marque</dt><dd>{{ product.brand }}</dd></div>
      <div v-if="product.barcode"><dt>Code-barres</dt><dd>{{ product.barcode }}</dd></div>
    </dl>

    <h2>Prix récents</h2>
    <p v-if="prices.length === 0" class="muted">Aucun prix publié pour le moment.</p>
    <table v-else>
      <thead>
        <tr>
          <th>Prix</th>
          <th>Quantité</th>
          <th>Lieu</th>
          <th>Source</th>
          <th>Observé le</th>
          <th>Confiance</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="price in prices" :key="price.id">
          <td>
            {{ price.amount }} {{ price.currency }}
            <span v-if="price.is_promotion" class="badge badge-warn">promotion</span>
          </td>
          <td>{{ price.quantity }} {{ price.unit_code }}</td>
          <td>{{ price.location.name }}<span v-if="price.geo_area.name" class="muted"> · {{ price.geo_area.name }}</span></td>
          <td>{{ price.source.name }}</td>
          <td>{{ formatDate(price.observed_at) }}</td>
          <td>
            <span v-if="price.status === 'approved'" class="badge badge-ok">approuvé</span>
            <span v-else class="muted">{{ price.status }}</span>
          </td>
        </tr>
      </tbody>
    </table>
  </template>
</template>

<style scoped>
.error {
  color: var(--err);
}

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
</style>