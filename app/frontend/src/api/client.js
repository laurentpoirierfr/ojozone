const BASE = '/api/v1'

async function jsonRequest(path, options = {}) {
  const headers = { Accept: 'application/json', ...(options.headers || {}) }
  const config = { ...options, headers }

  if (options.accessToken) {
    headers.Authorization = `Bearer ${options.accessToken}`
  }
  if (options.body !== undefined && options.body !== null) {
    headers['Content-Type'] = 'application/json'
    config.body = JSON.stringify(options.body)
  }
  delete config.accessToken

  const response = await fetch(`${BASE}${path}`, config)
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

function addQuery(path, params) {
  if (!params) return path
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      query.set(key, String(value))
    }
  }
  const encoded = query.toString()
  return encoded ? `${path}?${encoded}` : path
}

export class ApiError extends Error {
  constructor(status, problem = null) {
    super(problem?.detail || `Erreur HTTP ${status}`)
    this.status = status
    this.problem = problem
  }
}

export function listProducts(search = '', limit = 50) {
  return jsonRequest(addQuery('/products', { q: search, limit }))
}

export function listLocations(limit = 200) {
  return jsonRequest(addQuery('/locations', { limit }))
}

export function listFuelTypes(limit = 200) {
  return jsonRequest(addQuery('/fuel-types', { limit }))
}

export function listUnits(limit = 200) {
  return jsonRequest(addQuery('/units', { limit }))
}

export function getProduct(id) {
  return jsonRequest(`/products/${id}`)
}

export function listProductPrices(productID, limit = 50) {
  return jsonRequest(addQuery(`/products/${productID}/prices`, { limit }))
}

export function registerRequest(email, password, locale) {
  return jsonRequest('/auth/register', { method: 'POST', body: { email, password, locale } })
}

export function loginRequest(email, password) {
  return jsonRequest('/auth/login', { method: 'POST', body: { email, password } })
}

export function refreshRequest(refreshToken) {
  return jsonRequest('/auth/refresh', { method: 'POST', body: { refresh_token: refreshToken } })
}

export function logoutRequest(refreshToken) {
  return jsonRequest('/auth/logout', { method: 'POST', body: { refresh_token: refreshToken } })
}

export function getMe(accessToken) {
  return jsonRequest('/me', { accessToken })
}

export function listMyContributions(accessToken, limit = 50) {
  return jsonRequest(addQuery('/me/contributions', { limit }), { accessToken })
}

export function submitProductPriceContribution(accessToken, payload) {
  return jsonRequest('/contributions/product-prices', {
    method: 'POST',
    accessToken,
    body: payload,
  })
}

export function submitFuelPriceContribution(accessToken, payload) {
  return jsonRequest('/contributions/fuel-prices', {
    method: 'POST',
    accessToken,
    body: payload,
  })
}

export function patchContribution(accessToken, id, payload) {
  return jsonRequest(`/contributions/${id}`, {
    method: 'PATCH',
    accessToken,
    body: payload,
  })
}

export function deleteContribution(accessToken, id) {
  return jsonRequest(`/contributions/${id}`, {
    method: 'DELETE',
    accessToken,
  })
}

export function listModerationQueue(accessToken, status = '', limit = 50) {
  return jsonRequest(addQuery('/moderation/queue', { status, limit }), { accessToken })
}

export function approveContribution(accessToken, id, note = '') {
  return jsonRequest(`/moderation/contributions/${id}/approve`, {
    method: 'POST',
    accessToken,
    body: note ? { note } : {},
  })
}

export function rejectContribution(accessToken, id, note) {
  return jsonRequest(`/moderation/contributions/${id}/reject`, {
    method: 'POST',
    accessToken,
    body: { note },
  })
}

export function listImports(accessToken, limit = 50) {
  return jsonRequest(addQuery('/admin/imports', { limit }), { accessToken })
}

export function listAdminUsers(accessToken) {
  return jsonRequest('/admin/users', { accessToken })
}

export function createImport(accessToken, resourceType, rows) {
  return jsonRequest('/admin/imports', {
    method: 'POST',
    accessToken,
    body: { resource_type: resourceType, rows },
  })
}

export function validateImport(accessToken, id) {
  return jsonRequest(`/admin/imports/${id}/validate`, { method: 'POST', accessToken })
}

export function publishImport(accessToken, id) {
  return jsonRequest(`/admin/imports/${id}/publish`, { method: 'POST', accessToken })
}