// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ShuihuoProjectLibrary from './ShuihuoProjectLibrary.jsx'

vi.mock('./api.js', () => ({ archiveBatchProject: vi.fn(), listBatchProjects: vi.fn() }))
import * as api from './api.js'

describe('ShuihuoProjectLibrary', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    api.listBatchProjects.mockResolvedValue({ projects: [{ id: 41, name: '真实项目', bookCount: 3, runStatus: 'running', updatedAt: '2026-10-09T04:05:06Z' }] })
  })
  afterEach(cleanup)

  it('projects the Go-owned catalog into the V88 card layout and opens the selected project', async () => {
    const onOpen = vi.fn()
    render(<ShuihuoProjectLibrary onOpen={onOpen} onCreate={() => {}} onBatch={() => {}} />)

    expect(await screen.findByRole('heading', { name: '漫剧解说' })).toBeTruthy()
    expect(api.listBatchProjects).toHaveBeenCalledWith({ archived: 'active', limit: 100, sort: 'updated_desc' })
    expect(screen.getByText('真实项目')).toBeTruthy()
    expect(screen.getByText('3 本小说')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: '打开真实项目' }))
    expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: 41, name: '真实项目' }))
  })

})
