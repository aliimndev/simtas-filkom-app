// SIMTAS FILKOM — Load test baseline (Fase 3 #2)
//
// Endpoints: theses list, dashboard summary, archive search, document upload.
// Targets (from PRODUCTION-READINESS-REVIEW §7E): p95 < 500ms at 50 VU.
//
//   k6 run tests/k6/load.ts \
//     -e BASE_URL=http://localhost:8080 \
//     -e EMAIL=admin@filkom.unida.ac.id \
//     -e PASSWORD=... \
//     -e THESIS_ID=<active thesis id>   # optional, enables upload branch
import http from 'k6/http'
import { check, sleep } from 'k6'
import { SharedArray } from 'k6/data'
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'
const EMAIL = __ENV.EMAIL || 'admin@test.local'
const PASSWORD = __ENV.PASSWORD || 'Password123!'
const THESIS_ID = __ENV.THESIS_ID || ''

// One admin token shared across all VUs (the load is on the API, not auth).
const tokens = new SharedArray('tokens', () => {
  const res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({ email: EMAIL, password: PASSWORD }), {
    headers: { 'Content-Type': 'application/json' },
  })
  const token = res.json('data.access_token')
  if (!token) throw new Error(`login failed: ${res.status} ${res.body}`)
  return [token]
})

export const options = {
  scenarios: {
    normal: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },
        { duration: '8m', target: 50 },
        { duration: '1m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.01'],
  },
}

const AUTH = { Authorization: `Bearer ${tokens[0]}` }

export default function () {
  const r = randomIntBetween(0, 99)

  // Weighted mix: 50% theses list, 25% archive search, 20% dashboard
  // summary, 5% document upload (only when THESIS_ID is provided).
  let res
  if (r < 50) {
    res = http.get(`${BASE_URL}/api/v1/theses?page=1&per_page=10`, { headers: AUTH })
  } else if (r < 75) {
    res = http.get(`${BASE_URL}/api/v1/archives?q=${encodeURIComponent('skripsi')}&page=1&per_page=12`, { headers: AUTH })
  } else if (r < 95) {
    res = http.get(`${BASE_URL}/api/v1/dashboard/summary`, { headers: AUTH })
  } else if (THESIS_ID) {
    const file = open('sample.pdf', 'b')
    res = http.post(
      `${BASE_URL}/api/v1/theses/${THESIS_ID}/documents`,
      { file: http.file(file, 'sample.pdf', 'application/pdf'), document_type: 'proposal' },
      { headers: AUTH },
    )
  } else {
    res = http.get(`${BASE_URL}/api/v1/notifications`, { headers: AUTH })
  }

  check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 })
  sleep(randomIntBetween(1, 3))
}
