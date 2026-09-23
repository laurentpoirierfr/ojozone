const BASE = '/api/v1'

async function jsonRequest(path, options = {}) {
  const response = await fetch(`${BASE}${path}`, {
    headers: { Accept: 'application/json' },
    ...options,
  })
  if (!response.ok) {
    let problem = null
    try {
      problem = await response.json()
    } catch {
      // corps non JSON
    }
    throw new ApiError(response.status, problem)
  }
  const contentType = response.headers.get('content-type') || ''
  if (!contentType.includes('application/json')) {
    return null
  }
  return response.json()
}

export class ApiError extends Error {
  constructor(status, problem = null) {
    super(problem?.detail || `Erreur HTTP ${status}`)
    this.status = status
    this.problem = problem
  }
}

export function listProducts(search = '', limit = 50) {
  const params = new URLSearchParams()
  if (search) params.set('q', search)
  params.set('limit', String(limit))
  return jsonRequest(`/products?${params.toString()}`)
}

export function getProduct(id) {
  return jsonRequest(`/products/${id}`)
}

export function listProductPrices(productID, limit = 50) {
  const params = new URLSearchParams({ limit: String(limit) })
  return jsonRequest(`/products/${productID}/prices?${params.toString()}`)
}