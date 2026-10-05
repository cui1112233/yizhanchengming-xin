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

test.describe('@video Task 14 VIDEO/Merge acceptance blockers', () => {
  test('@video provider status exposes unconfigured/available/unavailable/auth_failed without secrets', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_3_6: provider/status API is not merged to main')
  })

  test('@video Start and Poll expose provider/model/status/attempts/error/output', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_1_2_8_9: VIDEO task/status/start/poll are not merged to main')
  })

  test('@video Retry retries only failed VIDEO and Cancel stops active VIDEO', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_10: VIDEO retry/cancel behavior is not merged to main')
  })

  test('@video unconfigured provider never turns Batch Factory into a page-level 500/503', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_6_15: provider status resilience is not merged to main')
  })

  test('@video successful VIDEO fragments merge queued→running→succeeded without regenerating VIDEO', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_12_14: merge worker/final result are not merged to main')
  })

  test('@video missing ffmpeg reports Merge error without preventing API startup', async () => {
    test.skip(true, 'BLOCKED_BY_TASK_14_12_13: ffmpeg merge chain is not merged to main')
  })
})
