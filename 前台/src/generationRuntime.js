import { useCallback, useEffect, useRef, useState } from 'react'
import { getGenerationRun } from './api.js'

const TERMINAL = new Set(['completed', 'failed', 'partial_failed', 'cancelled'])
const ACTIVE = new Set(['queued', 'pending', 'running', 'retryable_failed'])

function positiveID(value) {
  const parsed = Number(value)
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : null
}

export function isTerminalGenerationRun(run) {
  return run?.terminal === true || TERMINAL.has(String(run?.status || '').toLowerCase())
}

export function activeGenerationRunId(summary, bookId = null) {
  const books = Array.isArray(summary?.books) ? summary.books : []
  const candidates = bookId == null ? books : books.filter(book => Number(book?.bookId) === Number(bookId))
  for (const book of candidates) {
    const status = String(book?.run?.status || '').toLowerCase()
    const runId = positiveID(book?.run?.runId)
    if (runId && ACTIVE.has(status)) return runId
  }
  return null
}

function newIdempotencyKey() {
  if (typeof globalThis.crypto?.randomUUID === 'function') return globalThis.crypto.randomUUID()
  const bytes = new Uint8Array(16)
  globalThis.crypto?.getRandomValues?.(bytes)
  return Array.from(bytes, value => value.toString(16).padStart(2, '0')).join('')
}

export function useGenerationRunPolling({ projectId, intervalMs = 1200, onTerminal, onError } = {}) {
  const [target, setTarget] = useState(null)
  const [run, setRun] = useState(null)
  const [submitting, setSubmitting] = useState(false)
  const [restart, setRestart] = useState(0)
  const targetRef = useRef(null)
  const onTerminalRef = useRef(onTerminal)
  const onErrorRef = useRef(onError)
  const submitInFlight = useRef(false)
  const intentKeys = useRef(new Map())
  onTerminalRef.current = onTerminal
  onErrorRef.current = onError

  const begin = useCallback((accepted) => {
    const runId = positiveID(accepted?.runId)
    const scope = positiveID(projectId)
    if (!scope || !runId) throw new Error('生成服务未返回有效运行标识')
    const next = { ...accepted, runId, batchProjectId: scope }
    setRun(next)
    if (isTerminalGenerationRun(next)) {
      targetRef.current = null
      setTarget(null)
      onTerminalRef.current?.(next)
    } else {
      targetRef.current = { projectId: scope, runId }
      setTarget(targetRef.current)
    }
    return next
  }, [projectId])

  const resume = useCallback((runId, sourceProjectId = projectId) => {
    const scope = positiveID(projectId)
    const sourceScope = positiveID(sourceProjectId)
    const exactRun = positiveID(runId)
    if (!scope || !sourceScope || sourceScope !== scope || !exactRun) return false
    if (targetRef.current?.projectId === scope) return targetRef.current.runId === exactRun
    targetRef.current = { projectId: scope, runId: exactRun }
    setTarget(targetRef.current)
    return true
  }, [projectId])

  const admit = useCallback(async (intent, submitter) => {
    if (submitInFlight.current || targetRef.current || typeof submitter !== 'function') return null
    submitInFlight.current = true
    setSubmitting(true)
    const intentName = String(intent || 'generation')
    const key = intentKeys.current.get(intentName) || newIdempotencyKey()
    intentKeys.current.set(intentName, key)
    try {
      const accepted = await submitter(key)
      intentKeys.current.delete(intentName)
      return begin(accepted)
    } catch (error) {
      // A real HTTP response is a certain rejection. A fetch/network exception
      // is ambiguous, so the next explicit retry must reuse the same key.
      if (Number(error?.status) > 0) intentKeys.current.delete(intentName)
      throw error
    } finally {
      submitInFlight.current = false
      setSubmitting(false)
    }
  }, [begin])

  useEffect(() => {
    const scope = positiveID(projectId)
    if (targetRef.current && targetRef.current.projectId !== scope) targetRef.current = null
    setTarget(current => current && current.projectId !== scope ? null : current)
    setRun(current => current && current.batchProjectId !== scope ? null : current)
  }, [projectId])

  useEffect(() => {
    if (!target) return undefined
    let stopped = false
    let timer = null
    let controller = null

    const poll = async () => {
      if (stopped) return
      controller = new AbortController()
      let scheduleNext = false
      try {
        const value = await getGenerationRun(target.projectId, target.runId, { signal: controller.signal })
        if (stopped) return
        if (positiveID(value?.runId) !== target.runId || positiveID(value?.batchProjectId) !== target.projectId) {
          throw Object.assign(new Error('生成状态与当前项目不匹配'), { status: 409, code: 'GENERATION_RUN_SCOPE_MISMATCH' })
        }
        setRun(value)
        if (isTerminalGenerationRun(value)) {
          targetRef.current = null
          setTarget(null)
          onTerminalRef.current?.(value)
          return
        }
        scheduleNext = true
      } catch (error) {
        if (stopped || error?.name === 'AbortError') return
        onErrorRef.current?.(error)
        scheduleNext = true
      }
      if (scheduleNext && !stopped) timer = window.setTimeout(poll, intervalMs)
    }

    void poll()
    return () => {
      stopped = true
      if (timer != null) window.clearTimeout(timer)
      controller?.abort()
    }
  }, [target?.projectId, target?.runId, intervalMs, restart])

  return {
    active: Boolean(target),
    submitting,
    run,
    begin,
    resume,
    admit,
    retryPolling: () => setRestart(value => value + 1),
  }
}
