import { expect, test } from './support/acceptance.js'
import { defaultUser, routeAuthenticated, routeAnonymous } from './support/fixtures.js'

const isPublic = Boolean(process.env.PUBLIC_BASE_URL?.trim())
const publicUsername = process.env.E2E_USERNAME?.trim()
const publicPassword = process.env.E2E_PASSWORD?.trim()

async function loginThroughUI(page, username = 'fixture-user', password = 'fixture-password') {
  await page.getByLabel('用户名').fill(username)
  await page.getByLabel('密码').fill(password)
  await page.locator('button[type="submit"]').click()
}

test.describe('@auth login/session acceptance', () => {
  test('@auth @public-smoke Case 1 first anonymous visit is not reported as expired session', async ({ page }) => {
    let refreshCalls = 0
    if (!isPublic) {
      await page.route('**/api/auth/current-user', async (route) => {
        await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'UNAUTHENTICATED', message: 'authentication required' }) })
      })
      await page.route('**/api/auth/refresh', async (route) => {
        refreshCalls += 1
        await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_MISSING', message: 'no refresh session' }) })
      })
    }

    await page.goto('/batch-factory')

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toHaveCount(0)
    if (!isPublic) expect(refreshCalls).toBe(0)
  })

  test('@auth @public-smoke Case 2 login succeeds and enters the requested page', async ({ page }) => {
    if (isPublic) {
      test.skip(!publicUsername || !publicPassword, 'Public login requires E2E_USERNAME and E2E_PASSWORD secrets')
    } else {
      await routeAnonymous(page)
      await page.route('**/api/auth/login', async (route) => {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
      })
      await page.route('**/api/v1/batch-projects', async (route) => {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
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

  test('@auth Case 4 an existing session 401 performs exactly one refresh and recovers', async ({ page }) => {
    let recoveryStarted = false
    let recoveryCurrentUserCalls = 0
    let refreshCalls = 0

    await page.route('**/api/auth/current-user', async (route) => {
      if (!recoveryStarted) {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
        return
      }
      recoveryCurrentUserCalls += 1
      if (recoveryCurrentUserCalls === 1) {
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

    recoveryStarted = true
    await page.reload()

    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    expect(refreshCalls).toBe(1)
    expect(recoveryCurrentUserCalls).toBeGreaterThanOrEqual(2)
  })

  test('@auth an existing session with a real refresh failure reports expired login', async ({ page }) => {
    let expired = false
    let refreshCalls = 0

    await page.route('**/api/auth/current-user', async (route) => {
      if (!expired) {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
        return
      }
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()

    expired = true
    await page.reload()

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toBeVisible()
    expect(refreshCalls).toBe(1)
  })

  test('@auth Case 5 an authenticated user receiving 403 sees Forbidden without refresh or Login', async ({ page }) => {
    let forbidden = false
    let refreshCalls = 0

    await page.route('**/api/auth/current-user', async (route) => {
      if (!forbidden) {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
        return
      }
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'forbidden' }) })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()

    forbidden = true
    await page.reload()

    expect(refreshCalls).toBe(0)
    await expect(page.getByText('一战晟铭登录')).toHaveCount(0)
    await expect(page.getByText('Forbidden')).toBeVisible()
    await expect(page.getByText('无权限访问此页面。')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toHaveCount(0)
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

  test('@auth logout prevents the old session from being restored or reported as expired', async ({ page }) => {
    let sessionActive = true
    let refreshCalls = 0
    await page.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({
        status: sessionActive ? 200 : 401,
        contentType: 'application/json',
        body: JSON.stringify(sessionActive ? { user: defaultUser } : { code: 'UNAUTHENTICATED' }),
      })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
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
    await page.evaluate(async () => {
      const api = await import('/src/api.js')
      await api.logout()
    })
    await page.reload()

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toHaveCount(0)
    expect(refreshCalls).toBe(0)
  })

  test('@auth refresh loop protection never refreshes more than once during bootstrap recovery', async ({ page }) => {
    let recoveryStarted = false
    let refreshCalls = 0
    let recoveryCurrentUserCalls = 0
    await page.route('**/api/auth/current-user', async (route) => {
      if (!recoveryStarted) {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
        return
      }
      recoveryCurrentUserCalls += 1
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
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

    recoveryStarted = true
    await page.reload()
    await expect(page.getByText('一战晟铭登录')).toBeVisible()

    expect(refreshCalls).toBe(1)
    expect(recoveryCurrentUserCalls).toBeGreaterThanOrEqual(2)
  })
})
