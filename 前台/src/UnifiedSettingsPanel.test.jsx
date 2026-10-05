// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeAll, describe, expect, it, vi } from 'vitest'
import UnifiedSettingsPanel from './UnifiedSettingsPanel.jsx'

beforeAll(() => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
})

describe('UnifiedSettingsPanel', () => {
  it('opens production, publishing and version profile in drawers without parse-input entry', async () => {
    const api = {
      getUnifiedSettings: vi.fn().mockResolvedValue({ projectId: 9, effective: {}, project: {}, profile: { name: '默认配置档', version: 'v1', settings: {} } }),
      saveProductionSettings: vi.fn().mockResolvedValue({ projectId: 9, effective: {} }),
      savePublishingSettings: vi.fn().mockResolvedValue({ projectId: 9, effective: {} }),
      saveVersionProfile: vi.fn().mockResolvedValue({ projectId: 9, profile: { name: '默认配置档', version: 'v1', settings: {} } }),
      sync121Config: vi.fn().mockResolvedValue({ projectId: 9, profile: { name: '默认配置档', version: 'v1', settings: { website121: { synced: true } } } }),
      syncStyleTypes: vi.fn().mockResolvedValue({ projectId: 9, profile: { name: '默认配置档', version: 'v1', settings: { styleTypes: { styles: ['剧情'] } } } }),
    }

    render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)
    expect(api.getUnifiedSettings).not.toHaveBeenCalled()
    expect(screen.queryByText('解析输入')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: '生产统一设置' }))
    await waitFor(() => expect(api.getUnifiedSettings).toHaveBeenCalledWith(9))
    expect(await screen.findByText('保存生产统一设置')).toBeTruthy()
    expect(screen.getByText('AI 文案唯一控制位置')).toBeTruthy()

    fireEvent.click(screen.getByLabelText('Close'))
    fireEvent.click(screen.getByRole('button', { name: '发布统一设置' }))
    expect(await screen.findByText('保存发布统一设置')).toBeTruthy()

    fireEvent.click(screen.getByLabelText('Close'))
    fireEvent.click(screen.getByRole('button', { name: '版本对应配置档' }))
    expect(await screen.findByText('处理规则提示词')).toBeTruthy()
    expect(screen.getByText('知识库提示词')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: '同步' }))
    expect(await screen.findByText('同步 121 网站配置')).toBeTruthy()
    expect(screen.getByText('同步批量风格类型')).toBeTruthy()
  })
})
