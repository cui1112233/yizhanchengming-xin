import { expect, test } from '@playwright/test'

test.describe('@auth login/session acceptance', () => {
  test('@auth @public-smoke first unauthenticated visit shows login without expired-session warning', async ({ page }, testInfo) => {
    const consoleErrors = []
    const failedRequests = []

    page.on('console', (message) => {
      if (message.type() === 'error') consoleErrors.push(message.text())
    })
    page.on('requestfailed', (request) => {
      failedRequests.push(`${request.method()} ${request.url()} :: ${request.failure()?.errorText || 'request failed'}`)
    })

    await page.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ code: 'UNAUTHENTICATED', message: 'authentication required' }),
      })
    })
    await page.route('**/api/auth/refresh', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ code: 'SESSION_MISSING', message: 'no refresh session' }),
      })
    })

    await page.goto('/batch-factory')

    await expect(page.getByText('一战晟铭登录')).toBeVisible()
    await expect(page.getByText('当前登录状态已过期，请重新登录。')).toHaveCount(0)

    await testInfo.attach('current-url', {
      body: Buffer.from(page.url()),
      contentType: 'text/plain',
    })
    await testInfo.attach('console-errors', {
      body: Buffer.from(consoleErrors.join('\n') || '(none)'),
      contentType: 'text/plain',
    })
    await testInfo.attach('failed-network-requests', {
      body: Buffer.from(failedRequests.join('\n') || '(none)'),
      contentType: 'text/plain',
    })
  })
})
