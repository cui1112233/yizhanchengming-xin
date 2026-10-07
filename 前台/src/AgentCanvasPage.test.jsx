// @vitest-environment jsdom
import React from 'react'
import { afterEach, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import AgentCanvasPage from './AgentCanvasPage.jsx'
vi.mock('./api.js', () => ({ agentAttachmentContentURL: vi.fn(() => '/download'), continueAgentProject: vi.fn(), deleteAgentAttachment: vi.fn(), getAgentCanvas: vi.fn(), listAgentAttachments: vi.fn(), listAgentCanvasVersions: vi.fn(), listAgentExecutions: vi.fn(() => Promise.resolve({executions:[]})), listAgentMessages: vi.fn(() => Promise.resolve({messages:[]})), listAgentSkills: vi.fn(() => Promise.resolve({skills:[]})), restoreAgentCanvas: vi.fn(), saveAgentCanvas: vi.fn(), uploadAgentAttachment: vi.fn() }))
import { deleteAgentAttachment, getAgentCanvas, listAgentAttachments, listAgentCanvasVersions, restoreAgentCanvas, saveAgentCanvas, uploadAgentAttachment } from './api.js'
afterEach(() => vi.clearAllMocks())

it('loads, uploads and deletes project-scoped attachments', async () => {
  window.history.pushState({}, '', '/agent/canvas?projectId=7')
  listAgentAttachments.mockResolvedValue({ attachments: [{ id: 3, filename: 'note.txt' }] })
  getAgentCanvas.mockResolvedValue({ canvas: { projectId: 7, revision: 0, document: { nodes: [], edges: [] } } })
  listAgentCanvasVersions.mockResolvedValue({ versions: [] })
  uploadAgentAttachment.mockResolvedValue({})
  deleteAgentAttachment.mockResolvedValue({})
  render(<AgentCanvasPage />)
  expect(await screen.findByText('note.txt')).toBeTruthy()
  const input = document.querySelector('input[type=file]')
  fireEvent.change(input, { target: { files: [new File(['x'], 'draft.txt', { type: 'text/plain' })] } })
  await waitFor(() => expect(uploadAgentAttachment).toHaveBeenCalledWith(7, expect.any(File)))
  fireEvent.click(screen.getByRole('button', { name: /删\s*除/ }))
  await waitFor(() => expect(deleteAgentAttachment).toHaveBeenCalledWith(7, 3))
})

it('shows an explicit forbidden attachment state', async () => {
  window.history.pushState({}, '', '/agent/canvas?projectId=7')
  const forbidden = new Error('forbidden')
  forbidden.status = 403
  listAgentAttachments.mockRejectedValue(forbidden)
  getAgentCanvas.mockRejectedValue(forbidden)
  listAgentCanvasVersions.mockRejectedValue(forbidden)
  render(<AgentCanvasPage />)
  expect(await screen.findByText('无权限查看或操作该项目附件。')).toBeTruthy()
})

it('loads, saves and restores a durable canvas revision', async () => {
  window.history.pushState({}, '', '/agent/canvas?projectId=7')
  listAgentAttachments.mockResolvedValue({ attachments: [] })
  getAgentCanvas.mockResolvedValue({ canvas: { projectId: 7, revision: 1, document: { nodes: ['first'], edges: [] } } })
  listAgentCanvasVersions.mockResolvedValue({ versions: [{ id: 1, revision: 1, document: { nodes: ['first'], edges: [] } }] })
  saveAgentCanvas.mockResolvedValue({ canvas: { projectId: 7, revision: 2, document: { nodes: ['second'], edges: [] } } })
  restoreAgentCanvas.mockResolvedValue({ canvas: { projectId: 7, revision: 3, document: { nodes: ['first'], edges: [] } } })
  render(<AgentCanvasPage />)
  expect(await screen.findByDisplayValue(/"first"/)).toBeTruthy()
  fireEvent.change(screen.getByLabelText('画布 JSON'), { target: { value: '{"nodes":["second"],"edges":[]}' } })
  fireEvent.click(screen.getByRole('button', { name: '保存画布' }))
  await waitFor(() => expect(saveAgentCanvas).toHaveBeenCalledWith(7, { revision: 1, document: { nodes: ['second'], edges: [] } }))
  fireEvent.click(await screen.findByRole('button', { name: '恢复版本 1' }))
  await waitFor(() => expect(restoreAgentCanvas).toHaveBeenCalledWith(7, 1))
})
