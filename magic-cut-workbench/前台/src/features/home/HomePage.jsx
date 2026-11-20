import { Alert, Button, Spin } from 'antd'
import {
  BarChart3, BookOpen, Bot, CheckCircle2, ChevronDown, ChevronRight,
  Clapperboard, Cloud, Film, Grid2X2, Image, LayoutDashboard,
  Lightbulb, ListChecks, MonitorPlay, Plus, Settings, Sparkles, Video,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { WorkModeMenu } from '../../shared/work-mode/WorkModeMenu.jsx'
import { CreateProjectDialog } from '../project-create/CreateProjectDialog.jsx'
import { resolveHomeToolAction } from './homeInteractionModel.js'
import { createHomeModel } from './homeModel.js'

const primaryIcons = [MonitorPlay, Clapperboard, BarChart3]
const secondaryIcons = [Film, Sparkles, Grid2X2, Clapperboard, Bot, Image, Cloud]
const sidebarIcons = {
  layout: LayoutDashboard,
  cloud: Cloud,
  video: Video,
  bulb: Lightbulb,
  chart: BarChart3,
  tasks: ListChecks,
}

export function HomePage({ account }) {
  const copy = createHomeModel()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [projects, setProjects] = useState([])
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  useEffect(() => {
    apiRequest('/api/magic-cut/projects')
      .then(result => setProjects(result.projects || []))
      .catch(requestError => setError(requestError.message || '工程列表读取失败'))
      .finally(() => setLoading(false))
  }, [])

  function openProject(project) {
    window.history.pushState({}, '', `/projects/${project.id}/edit`)
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  function projectCreated(project) {
    setProjects(current => [project, ...current.filter(item => item.id !== project.id)])
    setDialogOpen(false)
    openProject(project)
  }

  function activateTool(tool) {
    const action = resolveHomeToolAction(tool)
    if (action === 'create-project') {
      setDialogOpen(true)
      return
    }
    if (action === 'navigate' && tool.href) {
      window.history.pushState({}, '', tool.href)
      window.dispatchEvent(new PopStateEvent('popstate'))
      return
    }
    setNotice(`${tool.title}正在按原版功能开发，完成前不会伪装成可用。`)
  }

  const displayName = account?.displayName || account?.username || '一战晟铭'
  const avatarText = displayName.slice(0, 1)

  return (
    <div className="original-home">
      <header className="original-home__topbar">
        <WorkModeMenu />
        <div className="original-home__top-actions">
          <span className="top-action">行业 <b>短剧漫剧</b> <ChevronDown size={13} /></span>
          <span className="top-action"><BookOpen size={14} /> 使用指引</span>
          <span className="top-action"><Settings size={14} /> 设置</span>
          <a className="top-action" href="/tasks"><ListChecks size={14} /> 任务中心</a>
          <span className="top-action original-home__user-dot">{avatarText}</span>
          <b className="original-home__user-name">{displayName}</b>
        </div>
      </header>

      <div className="original-home__body">
        <aside className="original-home__sidebar">
          {copy.sidebarGroups.map(group => (
            <section className="home-nav-group" key={group.label || 'create'}>
              {group.label ? <p>{group.label}</p> : null}
              {group.items.map(item => {
                const Icon = sidebarIcons[item.icon]
                return item.href ? (
                  <a className={`home-nav-item ${item.indent ? 'is-indent' : ''}`} href={item.href} key={item.title}>
                    {Icon ? <Icon size={16} /> : null}<span>{item.title}</span>
                  </a>
                ) : (
                  <button
                    className={`home-nav-item ${item.active ? 'is-active' : ''} ${item.indent ? 'is-indent' : ''}`}
                    key={item.title}
                    onClick={() => activateTool(item)}
                    type="button"
                  >
                    {Icon ? <Icon size={16} /> : null}<span>{item.title}</span>
                  </button>
                )
              })}
            </section>
          ))}
        </aside>

        <main className="original-home__content">
          {notice ? <Alert className="home-notice" closable message={notice} onClose={() => setNotice('')} showIcon type="info" /> : null}
          {error ? <Alert className="home-notice" message={error} showIcon type="warning" /> : null}

          <section className="primary-tools" aria-label="主要创作工具">
            {copy.primaryTools.map((tool, index) => {
              const Icon = primaryIcons[index]
              return (
                <button className={`primary-tool is-${tool.tone}`} key={tool.id} onClick={() => activateTool(tool)} type="button">
                  {tool.badge ? <span className="tool-badge">{tool.badge}</span> : null}
                  <span className="primary-tool__icon"><Icon size={25} /></span>
                  <b>{tool.title}</b>
                  <small>{tool.description}</small>
                  <span className="primary-tool__arrow"><ChevronRight size={18} /></span>
                </button>
              )
            })}
          </section>

          <p className="home-section-label">更多工具</p>
          <section className="secondary-tools" aria-label="更多工具">
            {copy.secondaryTools.map((tool, index) => {
              const Icon = secondaryIcons[index]
              return (
                <button className="secondary-tool" key={tool.id} onClick={() => activateTool(tool)} type="button">
                  {tool.badge ? <span className="secondary-tool__badge">{tool.badge}</span> : null}
                  <span className="secondary-tool__icon"><Icon size={20} /></span>
                  <span><b>{tool.title}</b><small>{tool.description}</small></span>
                </button>
              )
            })}
          </section>

          <section className="home-overview-row">
            <div className="home-overview">
              <header><b>概览</b><button type="button">昨天 <ChevronDown size={13} /></button></header>
              <div className="home-stats">
                {copy.stats.map(item => <div key={item.label}><strong>{item.value}</strong><span>{item.label}</span></div>)}
              </div>
              <small className="home-overview__hint">仅统计权限范围内且为魔剪产出的素材</small>
            </div>
            <div className="home-promo">
              <div><strong>起量素材 一键出 <em>15国字幕</em></strong><span>每日 1000 条限免，50 条素材批量转译</span></div>
              <Button danger ghost shape="round" size="small">免费试用 →</Button>
              <div className="home-promo__art"><Cloud size={38} /><CheckCircle2 size={22} /></div>
            </div>
          </section>

          <section className="recent-projects">
            <header><b>{copy.sectionTitle}</b><button type="button">查看更多</button></header>
            <div className="recent-projects__row">
              <button className="create-project-card" onClick={() => setDialogOpen(true)} type="button">
                <span><Plus size={28} /></span><b>{copy.primaryAction}</b>
              </button>
              {loading ? <div className="recent-projects__loading"><Spin /></div> : projects.map(project => (
                <button className="recent-project-card" key={project.id} onClick={() => openProject(project)} type="button">
                  <span>{project.width} × {project.height}</span>
                  <b>{project.name}</b>
                  <small>{project.frameRate} fps</small>
                </button>
              ))}
            </div>
          </section>
        </main>
      </div>
      <CreateProjectDialog open={dialogOpen} onCancel={() => setDialogOpen(false)} onCreated={projectCreated} />
    </div>
  )
}
