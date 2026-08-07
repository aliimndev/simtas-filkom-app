// SIMTAS FILKOM — API smoke test collection (Fase 3 #1)
//
// Covers: auth flow, RBAC (401), pagination limits, error shape, key list
// endpoints. Runs against a seeded environment.
//
//   k6 run tests/k6/api-smoke.js \
//     -e BASE_URL=http://localhost:8080 \
//     -e EMAIL=mahasiswa@filkom.unida.ac.id \
//     -e PASSWORD=... \
//     -e ADMIN_EMAIL=admin@filkom.unida.ac.id \
//     -e ADMIN_PASSWORD=...
import http from 'k6/http'
import { check } from 'k6'
import { Counter } from 'k6/metrics'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'
const EMAIL = __ENV.EMAIL || 'mahasiswa@test.local'
const PASSWORD = __ENV.PASSWORD || 'Password123!'
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@test.local'
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'Password123!'

const failures = new Counter('failed_checks')

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: { failed_checks: ['count==0'] },
}

function login(email, password) {
  const res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({ email, password }), {
    headers: { 'Content-Type': 'application/json' },
  })
  check(res, { [`login ${email} is 200`]: (r) => r.status === 200 })
  return res.json('data.access_token')
}

function authed(path, token, params = {}) {
  return http.get(`${BASE_URL}/api/v1${path}`, {
    headers: { Authorization: `Bearer ${token}` },
    ...params,
  })
}

function record(name, ok) {
  check(null, { [name]: () => ok })
  if (!ok) failures.add(1)
}

export default function () {
  // 1. Auth: login → me → notifications
  const token = login(EMAIL, PASSWORD)
  record('login returns token', !!token)

  const me = authed('/users/me', token)
  record('GET /users/me is 200', me.status === 200 && me.json('success') === true)

  const notifs = authed('/notifications', token)
  record('GET /notifications is 200', notifs.status === 200)

  // 2. Pagination: huge per_page must be clamped, not OOM
  const paged = authed('/theses?page=1&per_page=10000', token)
  record('per_page=10000 clamped (200)', paged.status === 200)
  const perPage = paged.json('meta.per_page')
  record('meta.per_page <= 100', typeof perPage === 'number' && perPage <= 100)

  // 3. Key list endpoints
  record('GET /theses is 200', authed('/theses?page=1&per_page=10', token).status === 200)
  record('GET /archives is 200', authed('/archives?q=skripsi', token).status === 200)

  // 4. RBAC: no token → 401; error shape {success:false, message}
  const anon = http.get(`${BASE_URL}/api/v1/theses`)
  record('anonymous GET /theses is 401', anon.status === 401)
  record('error shape has success=false', anon.json('success') === false)
  record('error shape has message', typeof anon.json('message') === 'string')

  // 5. Admin/Kaprodi dashboard summary
  const adminToken = login(ADMIN_EMAIL, ADMIN_PASSWORD)
  const summary = authed('/dashboard/summary', adminToken)
  record('GET /dashboard/summary is 200', summary.status === 200)
}
