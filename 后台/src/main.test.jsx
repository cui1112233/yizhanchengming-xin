import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { AdminShell } from './main.jsx'

function response(status, body) {
  return { ok: status >= 200 && status < 300, status, json: async () => body }
}

afterEach(() => vi.restoreAllMocks())

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
      if (listAttempts === 1) return response(500, { code: 'ADMIN_OPERATION_FAILED' })
      return response(200, { prompts: [] })
    }
    throw new Error(`unexpected fetch ${url}`)
  })

  render(<AdminShell />)
  expect(await screen.findByText('提示词加载失败')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '重新加载' }))
  expect(await screen.findByText('暂无提示词版本')).toBeInTheDocument()
  await waitFor(() => expect(window.localStorage.length).toBe(0))
})
