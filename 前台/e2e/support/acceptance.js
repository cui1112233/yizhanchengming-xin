import { expect, test as base } from '@playwright/test'

const SECRET_ASSIGNMENT = /(authorization|cookie|password|passwd|access[_-]?token|refresh[_-]?token|api[_-]?secret|client[_-]?secret|credential(?:_ciphertext)?|aes[_-]?nonce|dsn)\s*[:=]\s*["']?([^\s,;"'}]+)/ig
const BEARER = /Bearer\s+[A-Za-z0-9._~+\/-]{10,}/ig
const SECRET_JSON = /["']?(authorization|cookie|password|passwd|access[_-]?token|refresh[_-]?token|api[_-]?secret|client[_-]?secret|credential(?:_ciphertext)?|aes[_-]?nonce|dsn)["']?\s*:\s*["']([^"']{8,})["']/ig

function redactText(value) {
  return String(value || '')
    .replace(BEARER, 'Bearer [REDACTED]')
    .replace(SECRET_ASSIGNMENT, (_match, key) => `${key}=[REDACTED]`)
}

function safeURL(raw) {
  try {
    const url = new URL(raw)
    return `${url.origin}${url.pathname}`
  } catch {
    return String(raw || '').split('?')[0]
  }
}

function findSecretLikeValues(value, source) {
  const text = String(value || '')
  const findings = []
  for (const match of text.matchAll(SECRET_JSON)) {
    const key = String(match[1] || '').toLowerCase()
    const candidate = String(match[2] || '')
    if (key === 'passwordrequired') continue
    if (/^(true|false|null|undefined|redacted|masked|none)$/i.test(candidate)) continue
    findings.push(`${source}: secret-like value for ${key}`)
  }
  if (/Bearer\s+[A-Za-z0-9._~+\/-]{10,}/i.test(text)) {
    findings.push(`${source}: bearer credential-like value`)
  }
  return findings
}

export const test = base.extend({
  evidence: [async ({ page }, use, testInfo) => {
    const consoleErrors = []
    const pageErrors = []
    const failedRequests = []
    const httpFailures = []
    const secretFindings = []

    page.on('console', (message) => {
      const sanitized = redactText(message.text())
      if (message.type() === 'error') consoleErrors.push(sanitized)
      secretFindings.push(...findSecretLikeValues(message.text(), 'console'))
    })
    page.on('pageerror', (error) => {
      pageErrors.push(redactText(error?.message || error))
    })
    page.on('requestfailed', (request) => {
      failedRequests.push({
        method: request.method(),
        url: safeURL(request.url()),
        failure: redactText(request.failure()?.errorText || 'request failed'),
      })
    })
    page.on('response', async (response) => {
      const request = response.request()
      const url = new URL(response.url())
      if (response.status() >= 400) {
        httpFailures.push({ method: request.method(), url: `${url.origin}${url.pathname}`, status: response.status() })
      }
      if (!url.pathname.startsWith('/api/')) return
      const contentType = response.headers()['content-type'] || ''
      if (!/(json|text)/i.test(contentType)) return
      try {
        const body = await response.text()
        secretFindings.push(...findSecretLikeValues(body, `response ${request.method()} ${url.pathname}`))
      } catch {
        // Evidence collection must never break the application request.
      }
    })

    await use({ consoleErrors, pageErrors, failedRequests, httpFailures, secretFindings })

    try {
      const bodyText = await page.locator('body').innerText({ timeout: 1000 })
      secretFindings.push(...findSecretLikeValues(bodyText, 'ui'))
    } catch {
      // Navigation/crash evidence can still be attached without body text.
    }

    const snapshot = {
      test: testInfo.titlePath.join(' > '),
      status: testInfo.status,
      expectedStatus: testInfo.expectedStatus,
      timestamp: new Date().toISOString(),
      url: safeURL(page.url()),
      consoleErrors,
      pageErrors,
      failedRequests,
      httpFailures,
      secretFindings: [...new Set(secretFindings)],
    }

    await testInfo.attach('task16-evidence.json', {
      body: Buffer.from(JSON.stringify(snapshot, null, 2)),
      contentType: 'application/json',
    })

    expect(snapshot.secretFindings, 'browser/UI/network evidence must not contain secret-like values').toEqual([])
  }, { auto: true }],
})

export { expect }
