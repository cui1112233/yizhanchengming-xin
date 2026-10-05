// @vitest-environment jsdom
import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
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

afterEach(() => cleanup())

function createPersistentApi() {
  let project = {
    production: { productionMode: 'original', aiCopyEnabled: true, aiCopyCount: 2 },
    publishing: { uploadVideoType: 'merged', materialReuse: false, versionProfile: '初始发布档' },
  }
  let profile = {
    name: '默认配置档',
    version: 'v1',
    settings: {
      processingRulePromptRef: 'processing-v1',
      knowledgePromptRef: 'knowledge-v1',
      website121: {},
      styleTypes: {},
    },
  }
  const payload = () => ({ projectId: 9, effective: { ...project }, project, profile, priority: ['project', 'version_profile', 'system_default'] })
  return {
    getUnifiedSettings: vi.fn(async () => payload()),
    saveProductionSettings: vi.fn(async (_id, value) => { project = { ...project, production: value }; return payload() }),
    savePublishingSettings: vi.fn(async (_id, value) => { project = { ...project, publishing: value }; return payload() }),
    saveVersionProfile: vi.fn(async (_id, value) => { profile = { ...profile, ...value, settings: { ...profile.settings, ...(value.settings || {}) } }; return payload() }),
    sync121Config: vi.fn(async () => { profile = { ...profile, settings: { ...profile.settings, website121: { sources: [{ source: '知乎', platformId: '4' }] } } }; return payload() }),
    syncStyleTypes: vi.fn(async () => { profile = { ...profile, settings: { ...profile.settings, styleTypes: { styles: ['剧情'], genres: ['都市'], genders: ['女频'] } } }; return payload() }),
  }
}

describe('UnifiedSettingsPanel', () => {
  it('loads only on click and exposes the three drawers without parse-input entry', async () => {
    const api = createPersistentApi()
    render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)

    expect(api.getUnifiedSettings).not.toHaveBeenCalled()
    expect(screen.queryByText('解析输入')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: '生产统一设置' }))
    await waitFor(() => expect(api.getUnifiedSettings).toHaveBeenCalledTimes(1))
    expect(await screen.findByRole('button', { name: '保存生产统一设置' })).toBeTruthy()
    expect(screen.getByText('AI 文案唯一控制位置')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: '发布统一设置' }))
    await waitFor(() => expect(api.getUnifiedSettings).toHaveBeenCalledTimes(2))
    expect(await screen.findByRole('button', { name: '保存发布统一设置' })).toBeTruthy()
    expect(await screen.findByRole('textbox', { name: '发布网站配置档' })).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: '版本对应配置档' }))
    await waitFor(() => expect(api.getUnifiedSettings).toHaveBeenCalledTimes(3))
    expect(await screen.findByRole('button', { name: '保存配置档' })).toBeTruthy()
    expect(screen.getByText('处理规则提示词')).toBeTruthy()
    expect(screen.getByText('知识库提示词')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: '同步' }))
    expect(await screen.findByText('同步 121 网站配置')).toBeTruthy()
    expect(screen.getByText('同步批量风格类型')).toBeTruthy()
  }, 15000)

  it('re-reads saved production and publishing values after close and remount without losing workbench context', async () => {
    const api = createPersistentApi()
    window.history.replaceState({}, '', '/batch-factory?project=9')
    const view = render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)

    fireEvent.click(screen.getByRole('button', { name: '生产统一设置' }))
    const countInput = await screen.findByRole('spinbutton', { name: 'AI 文案数量' })
    expect(countInput.value).toBe('2')
    fireEvent.change(countInput, { target: { value: '5' } })
    fireEvent.click(screen.getByRole('button', { name: '保存生产统一设置' }))
    await waitFor(() => expect(api.saveProductionSettings).toHaveBeenCalledWith(9, expect.objectContaining({ aiCopyCount: 5 })))
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')

    fireEvent.click(screen.getByRole('button', { name: '生产统一设置' }))
    expect((await screen.findByRole('spinbutton', { name: 'AI 文案数量' })).value).toBe('5')

    fireEvent.click(screen.getByRole('button', { name: '发布统一设置' }))
    const profileInput = await screen.findByRole('textbox', { name: '发布网站配置档' })
    expect(profileInput.value).toBe('初始发布档')
    fireEvent.change(profileInput, { target: { value: '保存后的发布档' } })
    fireEvent.click(screen.getByRole('button', { name: '保存发布统一设置' }))
    await waitFor(() => expect(api.savePublishingSettings).toHaveBeenCalledWith(9, expect.objectContaining({ versionProfile: '保存后的发布档' })))
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')

    view.unmount()
    render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)
    fireEvent.click(screen.getByRole('button', { name: '发布统一设置' }))
    expect((await screen.findByRole('textbox', { name: '发布网站配置档' })).value).toBe('保存后的发布档')
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')
  }, 15000)

  it('persists version profile and both sync actions through backend API results', async () => {
    const api = createPersistentApi()
    window.history.replaceState({}, '', '/batch-factory?project=9')
    const view = render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)

    fireEvent.click(screen.getByRole('button', { name: '版本对应配置档' }))
    const nameInput = await screen.findByRole('textbox', { name: '配置档名称' })
    fireEvent.change(nameInput, { target: { value: '女频短剧版' } })
    fireEvent.change(screen.getByRole('textbox', { name: '处理规则提示词' }), { target: { value: 'processing-v3' } })
    fireEvent.change(screen.getByRole('textbox', { name: '知识库提示词' }), { target: { value: 'knowledge-v7' } })
    fireEvent.click(screen.getByRole('button', { name: '保存配置档' }))
    await waitFor(() => expect(api.saveVersionProfile).toHaveBeenCalledWith(9, expect.objectContaining({
      name: '女频短剧版',
      settings: expect.objectContaining({ processingRulePromptRef: 'processing-v3', knowledgePromptRef: 'knowledge-v7' }),
    })))
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')

    fireEvent.click(screen.getByRole('button', { name: '版本对应配置档' }))
    fireEvent.click(await screen.findByRole('tab', { name: '同步' }))
    fireEvent.click(screen.getByRole('button', { name: '同步 121 网站配置' }))
    await waitFor(() => expect(api.sync121Config).toHaveBeenCalledWith(9))
    fireEvent.click(screen.getByRole('button', { name: '同步批量风格类型' }))
    await waitFor(() => expect(api.syncStyleTypes).toHaveBeenCalledWith(9))
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')

    view.unmount()
    render(<UnifiedSettingsPanel project={{ id: 9, name: '项目九' }} api={api} />)
    fireEvent.click(screen.getByRole('button', { name: '版本对应配置档' }))
    expect((await screen.findByRole('textbox', { name: '配置档名称' })).value).toBe('女频短剧版')
    expect(screen.getByRole('textbox', { name: '处理规则提示词' }).value).toBe('processing-v3')
    expect(screen.getByRole('textbox', { name: '知识库提示词' }).value).toBe('knowledge-v7')
    fireEvent.click(screen.getByRole('tab', { name: '同步' }))
    expect(await screen.findByText('剧情')).toBeTruthy()
    expect(screen.getByText('都市')).toBeTruthy()
    expect(screen.getByText('女频')).toBeTruthy()
    expect(`${window.location.pathname}${window.location.search}`).toBe('/batch-factory?project=9')
  }, 30000)
})
