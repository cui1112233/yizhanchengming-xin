import { test } from './support/acceptance.js'

test.describe('@batch Task 9 Runtime acceptance blockers', () => {
  test('@batch pending→running worker transition', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_9_4: Run worker/status transition is not merged to main')
  })

  test('@batch scheduled automation executes when run_at becomes due', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_9_4_7: Scheduler/Redis Worker is not merged to main')
  })

  test('@batch duplicate execution is prevented by queue/lock/idempotency', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_9_4_6_9: idempotency and Redis Lock are not merged to main')
  })

  test('@batch crashed worker recovers an in-flight task without corrupting status', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_9_4_10: worker recovery strategy is not merged to main')
  })
})
