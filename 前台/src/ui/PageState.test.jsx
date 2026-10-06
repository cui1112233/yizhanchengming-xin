// @vitest-environment jsdom
import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import PageState from './PageState.jsx'

describe('PageState', () => {
  it('renders a retry action for failed pages', () => {
    const retry = vi.fn()
    render(<PageState state="failed" title="读取失败" onRetry={retry} />)
    expect(screen.getByText('读取失败')).toBeTruthy()
    expect(screen.getByRole('button', { name: /重\s*试/ })).toBeTruthy()
  })
})
