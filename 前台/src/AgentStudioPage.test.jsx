// @vitest-environment jsdom
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import AgentStudioPage from './AgentStudioPage.jsx'
vi.mock('./api.js', () => ({ createAgentProject: vi.fn(), continueAgentProject: vi.fn(), deleteAgentProject: vi.fn(), listAgentProjects: vi.fn() }))
import { createAgentProject, continueAgentProject, listAgentProjects } from './api.js'
it('creates a durable project then continues through the Go API', async () => { listAgentProjects.mockResolvedValue({projects:[]}); createAgentProject.mockResolvedValue({ project: { id: 8 } }); continueAgentProject.mockResolvedValue({}); const nav=vi.fn(); render(<AgentStudioPage onNavigate={nav}/>); fireEvent.change(screen.getByPlaceholderText('描述你的创作需求…'),{target:{value:'写开头'}}); fireEvent.click(screen.getByRole('button',{name:'开始创作'})); await waitFor(()=>expect(continueAgentProject).toHaveBeenCalledWith(8,{content:'写开头'})); expect(nav).toHaveBeenCalledWith('/agent/canvas?projectId=8') })
