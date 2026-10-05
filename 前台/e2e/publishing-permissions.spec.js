import { expect, test } from './support/acceptance.js'
import { routeAuthenticated } from './support/fixtures.js'

const isPublic = Boolean(process.env.PUBLIC_BASE_URL?.trim())

async function appRequest(page, path, options = {}) {
  return page.evaluate(async ({ requestPath, requestOptions }) => {
    const api = await import('/src/api.js')
    try {
      const payload = await api.requestJSON(requestPath, requestOptions)
      return { ok: true, payload }
    } catch (error) {
      return { ok: false, status: error?.status || 0, code: error?.code || '', message: error?.message || '' }
    }
  }, { requestPath: path, requestOptions: options })
}

async function loginPublic(page) {
  const username = process.env.E2E_USERNAME?.trim()
  const password = process.env.E2E_PASSWORD?.trim()
  test.skip(!username || !password, 'Public publishing permission smoke requires E2E_USERNAME and E2E_PASSWORD')
  await page.goto('/batch-factory')
  if (await page.getByText('一战晟铭登录').count()) {
    await page.getByLabel('用户名').fill(username)
    await page.getByLabel('密码').fill(password)
    await page.locator('button[type="submit"]').click()
  }
}

test.describe('@publish publishing permission acceptance', () => {
  test('@publish 403 remains 403, does not refresh, and does not force Login', async ({ page }) => {
    let refreshCalls = 0
    if (!isPublic) {
      await routeAuthenticated(page)
      await page.route('**/api/auth/refresh', async (route) => {
        refreshCalls += 1
        await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' })
      })
      await page.route('**/api/v1/publishing/accounts', async (route) => {
        await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'publish.account.configure required' }) })
      })
    } else {
      test.skip(process.env.E2E_EXPECT_RESTRICTED_PUBLISH !== '1', 'Public restricted-publishing smoke is explicit opt-in')
      await loginPublic(page)
    }

    await page.goto('/batch-factory')
    const result = await appRequest(page, '/api/v1/publishing/accounts')

    expect(result.ok).toBe(false)
    expect(result.status).toBe(403)
    expect(refreshCalls).toBe(0)
    await expect(page.getByText('一战晟铭登录')).toHaveCount(0)
  })

  test('@publish publish.execute does not imply publish.account.configure', async ({ page }) => {
    await routeAuthenticated(page)
    let intentCalls = 0
    let accountCalls = 0
    await page.route('**/api/v1/publishing/intents', async (route) => {
      intentCalls += 1
      await route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify({ intent: { id: 41, status: 'pending' } }) })
    })
    await page.route('**/api/v1/publishing/accounts', async (route) => {
      accountCalls += 1
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'publish.account.configure required' }) })
    })
    await page.goto('/batch-factory')

    const execute = await appRequest(page, '/api/v1/publishing/intents', { method: 'POST', body: JSON.stringify({ batchProjectId: 123, publishingAccountId: 9 }) })
    const configure = await appRequest(page, '/api/v1/publishing/accounts', { method: 'POST', body: JSON.stringify({ name: 'should-not-configure' }) })

    expect(execute.ok).toBe(true)
    expect(configure.ok).toBe(false)
    expect(configure.status).toBe(403)
    expect(intentCalls).toBe(1)
    expect(accountCalls).toBe(1)
  })

  test('@publish ordinary user cannot view publishing audit without publish.audit.view', async ({ page }) => {
    await routeAuthenticated(page)
    await page.route('**/api/v1/publishing/audits', async (route) => {
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'publish.audit.view required' }) })
    })
    await page.goto('/batch-factory')

    const result = await appRequest(page, '/api/v1/publishing/audits')
    expect(result.ok).toBe(false)
    expect(result.status).toBe(403)
  })

  test('@publish another users publishing account remains forbidden', async ({ page }) => {
    await routeAuthenticated(page)
    await page.route('**/api/v1/publishing/intents', async (route) => {
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'FORBIDDEN', message: 'publishing account is not owned by current principal' }) })
    })
    await page.goto('/batch-factory')

    const result = await appRequest(page, '/api/v1/publishing/intents', { method: 'POST', body: JSON.stringify({ batchProjectId: 123, publishingAccountId: 99999 }) })
    expect(result.ok).toBe(false)
    expect(result.status).toBe(403)
    expect(result.message).toMatch(/not owned|forbidden/i)
  })
})
