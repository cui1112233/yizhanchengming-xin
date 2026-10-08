// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import AgentReservedPage from './AgentReservedPage.jsx'

afterEach(() => { cleanup(); vi.restoreAllMocks() })

it('explains Agent unavailability without network calls and returns home', () => {
  const fetch = vi.spyOn(globalThis, 'fetch')
  const onNavigate = vi.fn()
  render(<AgentReservedPage onNavigate={onNavigate} />)
  expect(screen.getByText('Agent 工作区待重新设计')).toBeTruthy()
  expect(screen.getByText('Agent 将重新设计，当前不可用')).toBeTruthy()
  expect(screen.getByText('当前版本不执行 Agent 任务，也不会创建项目或调用 Provider。')).toBeTruthy()
  fireEvent.click(screen.getByRole('button', { name: '返回首页' }))
  expect(onNavigate).toHaveBeenCalledWith('/')
  expect(fetch).not.toHaveBeenCalled()
})
