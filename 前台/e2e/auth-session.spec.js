import { expect, test } from './support/acceptance.js'
import { defaultUser, routeAuthenticated, routeAnonymous } from './support/fixtures.js'

const isPublic = Boolean(process.env.PUBLIC_BASE_URL?.trim())
const publicUsername = process.env.E2E_USERNAME?.trim()
const publicPassword = process.env.E2E_PASSWORD?.trim()

async function loginThroughUI(page, username = 'fixture-user', password = 'fixture-password') {
  await page.getByLabel('用户名').fill(username)
  await page.getByLabel('密码').fill(password)
  await page.getByRole('button', { name: '登录' }).click()
}

test.describe('@auth login/session acceptance', () => {
  test('@auth @public-smoke Case 1 first anonymous visit is not reported as expired session', async ({ page }) => {
    test.fail(true, 'TASK16-AUTH-001 First anonymous visit incorrectly reported as expired session')
    if (!isPublic) await routeAnonymous(page)

    await page.goto('/batch-factory')

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toHaveCount(0)
  })

  test('@auth @public-smoke Case 2 login succeeds and enters the requested page', async ({ page }) => {
    if (isPublic) {
      test.skip(!publicUsername || !publicPassword, 'Public login requires E2E_USERNAME and E2E_PASSWORD secrets')
    } else {
      await routeAnonymous(page)
      await page.route('**/api/auth/login', async (route) => {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
      })
    }

    await page.goto('/batch-factory')
    await loginThroughUI(page, publicUsername || undefined, publicPassword || undefined)

    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await expect(page).toHaveURL(/\/batch-factory(?:\?.*)?$/)
  })

  test('@auth Case 3 login preserves /batch-factory?project=123', async ({ page }) => {
    await routeAnonymous(page)
    await page.route('**/api/auth/login', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory?project=123')
    await loginThroughUI(page)

    await expect(page).toHaveURL(/\/batch-factory\?project=123$/)
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
  })

  test('@auth Case 4 a 401 performs exactly one refresh and retries once', async ({ page }) => {
    let currentUserCalls = 0
    let refreshCalls = 0

    await page.route('**/api/auth/current-user', async (route) => {
      currentUserCalls += 1
      if (currentUserCalls === 1) {
        await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
        return
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ refreshed: true }) })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')

    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    expect(refreshCalls).toBe(1)
    expect(currentUserCalls).toBe(2)
  })

  test('@auth Case 5 a 403 does not refresh and does not jump to Login', async ({ page }) => {
    test.fail(true, 'TASK16-AUTH-002 Forbidden auth bootstrap currently falls back to Login instead of preserving 403 semantics')
    let refreshCalls = 0

    await page.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'forbidden' }) })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' })
    })

    await page.goto('/batch-factory')
    await page.waitForTimeout(100)

    expect(refreshCalls).toBe(0)
    await expect(page.getByText('一战晟铭登录')).toHaveCount(0)
  })

  test('@auth Case 6 browser refresh restores an existing session without Login flash', async ({ page }) => {
    await routeAuthenticated(page)
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await page.reload()

    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await expect(page.getByText('一战晟铭登录')).toHaveCount(0)
  })

  test('@auth logout prevents the old session from being restored', async ({ page }) => {
    let sessionActive = true
    await page.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({
        status: sessionActive ? 200 : 401,
        contentType: 'application/json',
        body: JSON.stringify(sessionActive ? { user: defaultUser } : { code: 'UNAUTHENTICATED' }),
      })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_MISSING' }) })
    })
    await page.route('**/api/auth/logout', async (route) => {
      sessionActive = false
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ ok: true }) })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await page.evaluate(() => fetch('/api/auth/logout', { method: 'POST', credentials: 'include' }))
    await page.reload()

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
  })

  test('@auth refresh loop protection never refreshes more than once per request', async ({ page }) => {
    let refreshCalls = 0
    let currentUserCalls = 0
    await page.route('**/api/auth/current-user', async (route) => {
      currentUserCalls += 1
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ refreshed: true }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByText('一战晟铭登录')).toBeVisible()

    expect(refreshCalls).toBe(1)
    expect(currentUserCalls).toBe(2)
  })
})
