// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import PageState from './PageState.jsx'
import { APIError } from '../api'

describe('PageState', () => {
  it('presents APIError message and correlation without rendering payload', () => {
    cleanup()
    const error = new APIError('服务暂时不可用', { requestId: 'req_page_123', payload: { password: 'payload-secret' } })
    const retry = vi.fn()
    const { container } = render(<PageState kind="error" message={error.message} requestId={error.requestId} payload={error.payload} onRetry={retry} />)
    expect(screen.getByText('服务暂时不可用')).toBeTruthy()
    expect(screen.getByText('请求编号：req_page_123')).toBeTruthy()
    expect(container.textContent).not.toContain('payload-secret')
    fireEvent.click(screen.getByRole('button', { name: /重\s*试/ }))
    expect(retry).toHaveBeenCalledTimes(1)
    cleanup()
  })
  it('renders a retry action for failed pages', () => {
    const retry = vi.fn()
    render(<PageState state="failed" title="读取失败" onRetry={retry} />)
    expect(screen.getByText('读取失败')).toBeTruthy()
    expect(screen.getByRole('button', { name: /重\s*试/ })).toBeTruthy()
  })
})
