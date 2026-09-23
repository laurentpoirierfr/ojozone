<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { state } from '../auth/store.js'
import {
  listProducts,
  listLocations,
  listFuelTypes,
  listUnits,
  getProduct,
  submitProductPriceContribution,
  submitFuelPriceContribution,
} from '../api/client.js'

const route = useRoute()

const kind = ref(route.query.type === 'fuel' ? 'fuel' : 'product')
const products = ref([])
const locations = ref([])
const fuelTypes = ref([])
const units = ref([])
const productSearch = ref('')
const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const result = ref(null)

const form = ref({
  product_id: route.query.product_id || '',
  fuel_type_id: route.query.fuel_type_id || '',
  location_id: route.query.location_id || '',
  amount: '',
  currency: 'EUR',
  quantity: '1.000',
  unit_code: '',
  is_promotion: false,
  amount_per_litre: '',
  observed_at: defaultObservedAt(),
})

function defaultObservedAt() {
  const now = new Date()
  now.setMinutes(now.getMinutes() - now.getTimezoneOffset())
  return now.toISOString().slice(0, 16)
}

const filteredProducts = computed(() => {
  const q = productSearch.value.trim().toLowerCase()
  if (!q) return products.value
  return products.value.filter((p) => p.name.toLowerCase().includes(q))
})

watch(
  () => form.value.product_id,
  (id) => {
    const product = products.value.find((p) => p.id === id)
    if (product && product.reference_unit && !form.value.unit_code) {
      form.value.unit_code = product.reference_unit
    }
  },
)

async function loadReferences() {
  loading.value = true
  error.value = ''
  try {
    const [productsData, locationsData, fuelTypesData, unitsData] = await Promise.all([
      listProducts('', 200),
      listLocations(),
      listFuelTypes(),
      listUnits(),
    ])
    products.value = productsData.data
    locations.value = locationsData.data
    fuelTypes.value = fuelTypesData.data
    units.value = unitsData.data

    const prefillProduct = route.query.product_id
    if (prefillProduct && !products.value.some((p) => p.id === prefillProduct)) {
      try {
        const extra = await getProduct(prefillProduct)
        products.value = [extra.data, ...products.value]
      } catch {
        // produit hors catalogue : le select restera vide
      }
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

async function submit() {
  error.value = ''
  result.value = null
  submitting.value = true
  try {
    const observedAt = form.value.observed_at ? new Date(form.value.observed_at).toISOString() : ''
    let data
    if (kind.value === 'product') {
      data = await submitProductPriceContribution(state.accessToken, {
        product_id: form.value.product_id,
        location_id: form.value.location_id,
        amount: form.value.amount,
        currency: form.value.currency,
        quantity: form.value.quantity,
        unit_code: form.value.unit_code,
        is_promotion: form.value.is_promotion,
        observed_at: observedAt,
      })
    } else {
      data = await submitFuelPriceContribution(state.accessToken, {
        fuel_type_id: form.value.fuel_type_id,
        location_id: form.value.location_id,
        amount_per_litre: form.value.amount_per_litre,
        currency: form.value.currency,
        observed_at: observedAt,
      })
    }
    result.value = data.data
  } catch (err) {
    error.value = err.message
  } finally {
    submitting.value = false
  }
}

onMounted(loadReferences)
</script>

<template>
  <section>
    <h1>Soumettre un prix</h1>
    <p class="muted">
      Votre contribution est créée au statut <strong>pending</strong> puis vérifiée par la
      modération.
    </p>

    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <div v-if="result" class="success" role="status">
      <p>
        <strong>Contribution reçue</strong> ({{ result.status }}) :
        <RouterLink to="/account">suivre dans mon compte</RouterLink>
      </p>
      <ul v-if="result.checks?.length" class="checks">
        <li v-for="check in result.checks" :key="check.code">{{ check.code }}</li>
      </ul>
    </div>

    <form class="contribute-form" @submit.prevent="submit">
      <fieldset :disabled="loading || submitting">
        <legend>Type de contribution</legend>
        <div class="kind-choice">
          <label>
            <input v-model="kind" type="radio" value="product" />
            Prix produit
          </label>
          <label>
            <input v-model="kind" type="radio" value="fuel" />
            Prix carburant
          </label>
        </div>

        <template v-if="kind === 'product'">
          <label for="product-search">Produit</label>
          <input
            id="product-search"
            v-model="productSearch"
            type="search"
            placeholder="Rechercher un produit…"
            autocomplete="off"
          />
          <label for="product-select" class="visually-hidden">Produit sélectionné</label>
          <select id="product-select" v-model="form.product_id" required>
            <option value="" disabled>Choisir un produit</option>
            <option v-for="product in filteredProducts" :key="product.id" :value="product.id">
              {{ product.name }}
            </option>
          </select>

          <label for="amount">Prix observé ({{ form.currency }})</label>
          <input
            id="amount"
            v-model="form.amount"
            type="text"
            inputmode="decimal"
            placeholder="2.35"
            required
          />

          <label for="quantity">Quantité</label>
          <input id="quantity" v-model="form.quantity" type="text" inputmode="decimal" required />

          <label for="unit">Unité</label>
          <select id="unit" v-model="form.unit_code" required>
            <option value="" disabled>Choisir une unité</option>
            <option v-for="unit in units" :key="unit.code" :value="unit.code">
              {{ unit.code }}
            </option>
          </select>

          <label class="checkbox">
            <input v-model="form.is_promotion" type="checkbox" />
            Promotion en cours
          </label>
        </template>

        <template v-else>
          <label for="fuel-type">Type de carburant</label>
          <select id="fuel-type" v-model="form.fuel_type_id" required>
            <option value="" disabled>Choisir un carburant</option>
            <option v-for="fuelType in fuelTypes" :key="fuelType.id" :value="fuelType.id">
              {{ fuelType.code }}
            </option>
          </select>

          <label for="amount-per-litre">Prix au litre ({{ form.currency }})</label>
          <input
            id="amount-per-litre"
            v-model="form.amount_per_litre"
            type="text"
            inputmode="decimal"
            placeholder="1.849"
            required
          />
        </template>

        <label for="location">Lieu</label>
        <select id="location" v-model="form.location_id" required>
          <option value="" disabled>Choisir un lieu</option>
          <option v-for="location in locations" :key="location.id" :value="location.id">
            {{ location.name }}
          </option>
        </select>

        <label for="currency">Devise</label>
        <input
          id="currency"
          v-model="form.currency"
          type="text"
          maxlength="3"
          minlength="3"
          required
        />

        <label for="observed-at">Observé le</label>
        <input id="observed-at" v-model="form.observed_at" type="datetime-local" required />

        <button type="submit" :disabled="loading || submitting">
          {{ submitting ? 'Envoi…' : 'Soumettre' }}
        </button>
        <span v-if="loading" class="muted" aria-live="polite">Chargement des référentiels…</span>
      </fieldset>
    </form>
  </section>
</template>

<style scoped>
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
}

.error {
  color: var(--err);
}

.success {
  border: 1px solid var(--ok, #2e7d32);
  border-radius: 6px;
  padding: 0.75rem 1rem;
  margin-bottom: 1.5rem;
}

.checks {
  margin: 0.5rem 0 0;
  padding-left: 1.25rem;
  font-size: 0.875rem;
  color: var(--muted);
}

.contribute-form fieldset {
  border: 1px solid var(--border, #d0d0d0);
  border-radius: 6px;
  padding: 1rem;
  max-width: 28rem;
}

.contribute-form legend {
  font-weight: 600;
  padding: 0 0.5rem;
}

.contribute-form label {
  display: block;
  margin-top: 0.75rem;
  font-size: 0.875rem;
}

.contribute-form label.checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.contribute-form input:not([type='radio']):not([type='checkbox']),
.contribute-form select {
  width: 100%;
  margin-top: 0.25rem;
}

.kind-choice {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 0.5rem;
}

.kind-choice label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0;
  cursor: pointer;
}

.contribute-form button {
  margin-top: 1.25rem;
}
</style>
