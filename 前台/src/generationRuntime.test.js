// @vitest-environment jsdom
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('./api.js', () => ({ getGenerationRun: vi.fn() }))
import { getGenerationRun } from './api.js'
import { activeGenerationRunId, useGenerationRunPolling } from './generationRuntime.js'

describe('generation runtime controller', () => {
  afterEach(() => { vi.resetAllMocks(); vi.useRealTimers() })

  it('recovers only an active exact run id from the MySQL summary', () => {
    expect(activeGenerationRunId({ books: [
      { bookId: 1, run: { runId: 40, status: 'completed' } },
      { bookId: 2, run: { runId: 41, status: 'running' } },
    ] })).toBe(41)
    expect(activeGenerationRunId({ books: [{ run: { runId: 42, status: 'completed' } }] })).toBeNull()
  })

  it('polls without overlap and stops at the exact terminal run', async () => {
    let resolveFirst
    getGenerationRun
      .mockImplementationOnce(() => new Promise(resolve => { resolveFirst = resolve }))
      .mockResolvedValueOnce({ runId: 44, batchProjectId: 3, status: 'completed', terminal: true })
    const onTerminal = vi.fn()
    const { result } = renderHook(() => useGenerationRunPolling({ projectId: 3, intervalMs: 5, onTerminal }))

    act(() => result.current.resume(44))
    await waitFor(() => expect(getGenerationRun).toHaveBeenCalledTimes(1))
    await act(async () => { await new Promise(resolve => window.setTimeout(resolve, 20)) })
    expect(getGenerationRun).toHaveBeenCalledTimes(1)

    await act(async () => { resolveFirst({ runId: 44, batchProjectId: 3, status: 'running', terminal: false }); await Promise.resolve() })
    await waitFor(() => expect(getGenerationRun).toHaveBeenCalledTimes(2))
    await waitFor(() => expect(onTerminal).toHaveBeenCalledWith(expect.objectContaining({ runId: 44, terminal: true })))
    expect(result.current.active).toBe(false)
  })

  it('aborts an old project poll and ignores its late response', async () => {
    let resolveOld
    getGenerationRun.mockImplementationOnce((_projectId, _runId, { signal }) => new Promise(resolve => {
      resolveOld = { resolve, signal }
    })).mockResolvedValueOnce({ runId: 55, status: 'completed', terminal: true })
    const onTerminal = vi.fn()
    const { result, rerender, unmount } = renderHook(({ projectId }) => useGenerationRunPolling({ projectId, onTerminal }), { initialProps: { projectId: 3 } })
    act(() => result.current.resume(44))
    await waitFor(() => expect(resolveOld).toBeTruthy())
    rerender({ projectId: 4 })
    expect(resolveOld.signal.aborted).toBe(true)
    await act(async () => { resolveOld.resolve({ runId: 44, status: 'completed', terminal: true }); await Promise.resolve() })
    expect(onTerminal).not.toHaveBeenCalled()
    unmount()
  })

  it('reuses an idempotency key after an uncertain network failure and blocks double submit', async () => {
    let rejectFirst
    const submitter = vi.fn()
      .mockImplementationOnce(() => new Promise((_resolve, reject) => { rejectFirst = reject }))
      .mockResolvedValueOnce({ runId: 61, status: 'queued', terminal: false })
    const { result } = renderHook(() => useGenerationRunPolling({ projectId: 3 }))
    let first
    act(() => { first = result.current.admit('book:7', submitter) })
    let duplicate
    act(() => { duplicate = result.current.admit('book:7', submitter) })
    expect(await duplicate).toBeNull()
    const networkError = new TypeError('network failed')
    await act(async () => { rejectFirst(networkError); await expect(first).rejects.toBe(networkError) })
    await act(async () => { await result.current.admit('book:7', submitter) })
    expect(submitter).toHaveBeenCalledTimes(2)
    expect(submitter.mock.calls[1][0]).toBe(submitter.mock.calls[0][0])
  })

  it('does not let a stale recovered run replace a newly accepted exact run', async () => {
    getGenerationRun.mockImplementation(() => new Promise(() => {}))
    const submitter = vi.fn().mockResolvedValue({ runId: 61, status: 'queued', terminal: false })
    const { result } = renderHook(() => useGenerationRunPolling({ projectId: 3 }))
    await act(async () => { await result.current.admit('book:7', submitter) })
    act(() => result.current.resume(50))
    await waitFor(() => expect(getGenerationRun).toHaveBeenCalled())
    expect(getGenerationRun.mock.calls.every(([, runId]) => runId === 61)).toBe(true)
  })

  it('rejects a recovered run that belongs to a different project', () => {
    const { result } = renderHook(() => useGenerationRunPolling({ projectId: 3 }))
    let resumed
    act(() => { resumed = result.current.resume(50, 2) })
    expect(resumed).toBe(false)
    expect(result.current.active).toBe(false)
    expect(getGenerationRun).not.toHaveBeenCalled()
  })
})
