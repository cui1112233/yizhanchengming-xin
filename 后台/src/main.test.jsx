import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { Modal } from 'antd'
import { AdminShell } from './main.jsx'

function response(status, body, headers = {}) {
  return { ok: status >= 200 && status < 300, status, headers: new Headers(headers), json: async () => body }
}

afterEach(() => {
  Modal.destroyAll()
  vi.restoreAllMocks()
})

test('renders an explicit no-permission state when capability bootstrap is forbidden', async () => {
  vi.spyOn(global, 'fetch').mockResolvedValue(response(403, { code: 'ADMIN_CAPABILITY_REQUIRED' }))

  render(<AdminShell />)

  expect(await screen.findByText('没有后台访问权限')).toBeInTheDocument()
  expect(screen.getByText('后台入口仅对拥有 admin.* capability 的账号开放。')).toBeInTheDocument()
})

test('loads safe prompt metadata and fetches body only after a view-authorized detail action', async () => {
  const fetchMock = vi.spyOn(global, 'fetch').mockImplementation(async (url) => {
    if (url === '/api/v1/admin/capabilities') {
      return response(200, { capabilities: ['admin.prompt.view', 'admin.prompt.edit', 'admin.prompt.publish'] })
    }
    if (url === '/api/v1/admin/prompts') {
      return response(200, { prompts: [{ id: 11, key: 'script.default', version: 3, lifecycle: 'published', enabled: true, seedSource: 'go_default', contentSha256: 'sha' }] })
    }
    if (url === '/api/v1/admin/prompts/script.default/versions/3') {
      return response(200, { prompt: { id: 11, key: 'script.default', version: 3, lifecycle: 'published', enabled: true, content: 'server-detail-body', contentSha256: 'sha' } })
    }
    throw new Error(`unexpected fetch ${url}`)
  })

  render(<AdminShell />)

  expect(await screen.findByText('script.default')).toBeInTheDocument()
  expect(screen.queryByText('server-detail-body')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: /查看正文/ }))
  expect(await screen.findByText('server-detail-body')).toBeInTheDocument()
  expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/prompts/script.default/versions/3', expect.objectContaining({ credentials: 'include' }))
})

test('shows loading failure and supports a refresh retry without browser storage', async () => {
  let listAttempts = 0
  vi.spyOn(global, 'fetch').mockImplementation(async (url) => {
    if (url === '/api/v1/admin/capabilities') return response(200, { capabilities: ['admin.prompt.view'] })
    if (url === '/api/v1/admin/prompts') {
      listAttempts += 1
      if (listAttempts === 1) return response(500, { code: 'ADMIN_OPERATION_FAILED', secret: 'private-config' }, { 'X-Request-ID': 'list-id' })
      return response(200, { prompts: [] })
    }
    throw new Error(`unexpected fetch ${url}`)
  })

  render(<AdminShell />)
  expect(await screen.findByText('提示词加载失败')).toBeInTheDocument()
  expect(screen.getByText(/请求编号：list-id/)).toBeInTheDocument()
  expect(screen.queryByText(/private-config/)).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '重新加载' }))
  expect(await screen.findByText('暂无提示词版本')).toBeInTheDocument()
  await waitFor(() => expect(window.localStorage.length).toBe(0))
})

test('keeps the dedicated expired-login bootstrap state', async () => {
  const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(response(401, { request_id: 'login-id' }))
  render(<AdminShell />)
  expect(await screen.findByText('登录已过期')).toBeInTheDocument()
  expect(fetchMock).toHaveBeenCalledTimes(1)
})

test('shows a request ID for detail errors without rendering sensitive payload', async () => {
  vi.spyOn(global, 'fetch').mockImplementation(async (url) => {
    if (url === '/api/v1/admin/capabilities') return response(200, { capabilities: ['admin.prompt.view'] })
    if (url === '/api/v1/admin/prompts') return response(200, { prompts: [{ key: 'script.default', version: 1, lifecycle: 'draft' }] })
    return response(403, { message: '正文读取失败', request_id: 'detail-id', secret: 'private-config' })
  })
  render(<AdminShell />)
  fireEvent.click(await screen.findByRole('button', { name: /查看正文/ }))
  expect(await screen.findByText(/请求编号：detail-id/)).toBeInTheDocument()
  expect(screen.queryByText(/private-config/)).not.toBeInTheDocument()
})

test.each(['save', 'publish', 'restore'])('shows a safe request ID for %s write errors', async (operation) => {
  const fetchMock = vi.spyOn(global, 'fetch').mockImplementation(async (url, options) => {
    if (url === '/api/v1/admin/capabilities') return response(200, { capabilities: ['admin.prompt.view', 'admin.prompt.edit', 'admin.prompt.publish'] })
    if (url === '/api/v1/admin/prompts') return response(200, { prompts: [{ key: 'script.default', version: 1, lifecycle: 'draft' }] })
    expect(options.credentials).toBe('include')
    expect(options.headers.has('Authorization')).toBe(false)
    return response(403, { message: '操作被拒绝', secret: 'private-config' }, { 'X-Request-ID': `${operation}-id` })
  })
  render(<AdminShell />)
  await screen.findByText('script.default')
  if (operation === 'save') {
    fireEvent.click(screen.getByRole('button', { name: '新建草稿' }))
    fireEvent.change(screen.getByLabelText('Prompt Key'), { target: { value: 'script.default' } })
    fireEvent.change(screen.getByLabelText('提示词正文'), { target: { value: 'draft' } })
    fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  } else {
    const label = operation === 'publish' ? /发\s*布/ : /恢\s*复/
    fireEvent.click(screen.getByRole('button', { name: label }))
    await waitFor(() => expect(screen.getAllByRole('button', { name: label })).toHaveLength(2))
    const buttons = screen.getAllByRole('button', { name: label })
    fireEvent.click(buttons[buttons.length - 1])
  }
  expect(await screen.findByText(`操作被拒绝（请求编号：${operation}-id）`)).toBeInTheDocument()
  expect(screen.queryByText(/private-config/)).not.toBeInTheDocument()
  const path = operation === 'save' ? '/api/v1/admin/prompts/script.default/drafts' : `/api/v1/admin/prompts/script.default/versions/1/${operation}`
  expect(fetchMock).toHaveBeenCalledWith(path, expect.objectContaining({
    method: 'POST', credentials: 'include', ...(operation === 'save' ? { body: '{"content":"draft"}' } : {}),
  }))
})
