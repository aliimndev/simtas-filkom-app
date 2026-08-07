import { test, expect } from '@playwright/test'

const EMAIL = process.env.E2E_EMAIL || 'mahasiswa@test.local'
const PASSWORD = process.env.E2E_PASSWORD || 'Password123!'

// Smoke E2E: login → dashboard renders → navigate to a protected page.
test('login and land on dashboard', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill(EMAIL)
  await page.getByLabel('Password').fill(PASSWORD)
  await page.getByRole('button', { name: /masuk/i }).click()

  await expect(page).toHaveURL(/\/dashboard/, { timeout: 15_000 })
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
})

test('unauthenticated user is redirected to login', async ({ page }) => {
  await page.goto('/dashboard')
  await expect(page).toHaveURL(/\/login/, { timeout: 15_000 })
})
