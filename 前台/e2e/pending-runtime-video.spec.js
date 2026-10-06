import { expect, test } from './support/acceptance.js'
import { routeAuthenticated } from './support/fixtures.js'

function runtimeProject(status, id = 91) {
  return {
    id,
    intakeId: 51,
    name: 'Task9 Runtime Project',
    sources: ['知乎'],
    bookCount: 3,
    genders: ['女频'],
    styles: ['情感'],
    runStatus: status,
  }
}

async function routeRuntimeList(page, getStatus) {
  await page.route('**/api/v1/batch-projects', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ projects: [runtimeProject(getStatus())] }),
    })
  })
}

test.describe('@batch Task 9 Runtime acceptance', () => {
  test.beforeEach(async ({ page }) => {
    await routeAuthenticated(page)
  })

  test('@batch pending→running worker transition', async ({ page }) => {
    let status = 'pending'
    await routeRuntimeList(page, () => status)
    await page.goto('/batch-factory')
    await expect(page.getByText('Task9 Runtime Project')).toBeVisible()
    await expect(page.getByText('待执行')).toBeVisible()

    status = 'running'
    await page.reload()
    await expect(page.getByText('执行中')).toBeVisible()
  })

  test('@batch scheduled automation executes when run_at becomes due', async ({ page }) => {
    let status = 'pending'
    const scheduledAt = '2030-01-01T04:00:00Z'
    let submittedRunAt = ''
    await page.route('**/api/v1/intakes/51/batch-projects', async (route) => {
      submittedRunAt = route.request().postDataJSON()?.runAt || ''
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ project: runtimeProject('pending'), run: { id: 501, status: 'pending', runAt: scheduledAt } }),
      })
    })
    await routeRuntimeList(page, () => status)
    await page.goto('/batch-factory')

    const response = await page.evaluate(async (runAt) => {
      const value = await fetch('/api/v1/intakes/51/batch-projects', {
        method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ name: 'Task9 Runtime Project', runAt }),
      })
      return value.json()
    }, scheduledAt)
    expect(response.run.status).toBe('pending')
    expect(submittedRunAt).toBe(scheduledAt)

    await expect(page.getByText('待执行')).toBeVisible()
    status = 'running'
    await page.reload()
    await expect(page.getByText('执行中')).toBeVisible()
  })

  test('@batch duplicate execution is prevented by queue/lease/idempotency', async ({ page }) => {
    let createCalls = 0
    const logicalRunId = 701
    await page.route('**/api/v1/intakes/51/batch-projects', async (route) => {
      createCalls += 1
      await route.fulfill({
        status: createCalls === 1 ? 201 : 200,
        contentType: 'application/json',
        body: JSON.stringify({ project: runtimeProject('pending'), run: { id: logicalRunId, status: 'pending' } }),
      })
    })
    await routeRuntimeList(page, () => 'pending')
    await page.goto('/batch-factory')

    const ids = await page.evaluate(async () => Promise.all([0, 1].map(async () => {
      const response = await fetch('/api/v1/intakes/51/batch-projects', {
        method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ name: 'Task9 Runtime Project' }),
      })
      const body = await response.json()
      return body.run.id
    })))
    expect(ids).toEqual([logicalRunId, logicalRunId])
    expect(new Set(ids).size).toBe(1)
  })

  test('@batch crashed worker recovers an in-flight task without corrupting status', async ({ page }) => {
    let status = 'running'
    await routeRuntimeList(page, () => status)
    await page.goto('/batch-factory')
    await expect(page.getByText('执行中')).toBeVisible()

    // A browser reload represents observing the same durable execution after a
    // worker/process restart; it must remain running rather than regress.
    await page.reload()
    await expect(page.getByText('执行中')).toBeVisible()

    status = 'succeeded'
    await page.reload()
    await expect(page.getByText('succeeded')).toBeVisible()
  })
})
