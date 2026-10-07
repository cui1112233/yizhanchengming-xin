// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import BatchFactoryHome from './BatchFactoryHome.jsx'

vi.mock('./api.js', () => ({ listBatchProjects: vi.fn(), createIntake: vi.fn(), executeIntake: vi.fn(), createBatchProject: vi.fn() }))
import * as api from './api.js'

describe('BatchFactoryHome', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 6, name: '来源测试', sources: ['知乎付费'], bookCount: 2, genders: ['女频'], runStatus: 'completed' }, { id: 7, name: '另一个', sources: ['番茄免费'], bookCount: 1, runStatus: 'failed' }] })
  })
  afterEach(() => vi.restoreAllMocks())
  it('uses server project facts, filters them and opens the chosen project', async () => {
    const open = vi.fn(); render(<BatchFactoryHome onOpenProject={open} />)
    await screen.findByText('来源测试')
    fireEvent.change(screen.getByRole('textbox', { name: '搜索批量项目' }), { target: { value: '来源' } })
    expect(screen.queryByText('另一个')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: '进入项目' }))
    expect(open).toHaveBeenCalledWith(6)
  })
  it('creates intake, executes it, then creates project only after completed', async () => {
    api.createIntake.mockResolvedValue({ intake: { id: 22 } })
    api.executeIntake.mockResolvedValue({ status: 'completed' })
    api.createBatchProject.mockResolvedValue({ project: { id: 9 } })
    render(<BatchFactoryHome onOpenProject={() => {}} />)
    await screen.findByText('来源测试')
    fireEvent.click(screen.getByRole('button', { name: '新建批量' }))
    fireEvent.change(screen.getByLabelText('项目名称'), { target: { value: '新项目' } })
    fireEvent.mouseDown(screen.getByLabelText('书城来源'))
    fireEvent.click((await screen.findAllByText('知乎付费')).at(-1))
    fireEvent.change(screen.getByLabelText('Book ID'), { target: { value: '1001' } })
    fireEvent.click(screen.getByRole('button', { name: '获取并创建项目' }))
    await waitFor(() => expect(api.createBatchProject).toHaveBeenCalledWith(22, { name: '新项目' }))
  })
})
