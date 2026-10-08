import React, { useEffect, useRef, useState } from 'react'
import { Button, Result } from 'antd'
import NovelFetchPage from './NovelFetchPage.jsx'
import BatchFactoryHome from './BatchFactoryHome.jsx'
import BatchProjectListPage from './BatchProjectListPage.jsx'
import ShuihuoProductionPage from './ShuihuoProductionPage.jsx'
import HomePage from './HomePage.jsx'
import UserShell from './UserShell.jsx'
import NovelFetchWorkshop from './NovelFetchWorkshop.jsx'
import ScriptWorkspace from './ScriptWorkspace.jsx'
import NovelPanelWorkbench from './novel-panel/NovelPanelWorkbench.jsx'
import TtsPage from './TtsPage.jsx'
import HistoryPage from './HistoryPage.jsx'
import AccountCenterPage from './AccountCenterPage.jsx'
import SettingsPage from './SettingsPage.jsx'
import AgentReservedPage from './AgentReservedPage.jsx'
import { getNovelPanel, listNovelPanelHistory, restoreNovelPanelHistory, saveNovelPanel } from './api.js'
import './home.css'

const routeFoundations = {
  '/settings': { title: '设置', description: '设置入口已恢复；实际设置页不属于本任务范围。' },
}

function RouteFoundation({ title, description, onNavigate }) {
  return (
    <main className="route-foundation-page">
      <Result
        status="info"
        title={title}
        subTitle={description}
        extra={<Button type="primary" onClick={() => onNavigate('/')}>返回首页</Button>}
      />
    </main>
  )
}

export default function RouterApp({ theme, onToggleTheme, currentUser, onLogout }) {
  const [location, setLocation] = useState(() => ({
    pathname: typeof window === 'undefined' ? '/' : window.location.pathname,
    search: typeof window === 'undefined' ? '' : window.location.search,
  }))
  const { pathname, search } = location

  useEffect(() => {
    const handlePopState = () => setLocation({ pathname: window.location.pathname, search: window.location.search })
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const navigate = (path) => {
    let target
    try { target = new URL(path, window.location.href) } catch { return }
    if (target.origin !== window.location.origin) return
    const next = `${target.pathname}${target.search}${target.hash}`
    if (`${window.location.pathname}${window.location.search}${window.location.hash}` !== next) window.history.pushState({}, '', next)
    setLocation({ pathname: target.pathname, search: target.search })
  }
  const [novelPanel, setNovelPanel] = useState({ loading: false, loadedProject: 0, workspace: null, error: null })
  const novelPanelRequest = useRef(0)
  const panelProjectId = Number(new URLSearchParams(search).get('projectId'))
  useEffect(() => {
    const requestID = ++novelPanelRequest.current
    if (pathname !== '/novel-panel' || !Number.isSafeInteger(panelProjectId) || panelProjectId <= 0) {
      setNovelPanel({ loading: false, loadedProject: 0, workspace: null, error: null })
      return undefined
    }
    let active = true
    setNovelPanel({ loading: true, loadedProject: 0, workspace: null, error: null })
    getNovelPanel(panelProjectId)
      .then((result) => {
        if (active && requestID === novelPanelRequest.current) setNovelPanel({ loading: false, loadedProject: panelProjectId, workspace: result?.workspace || null, error: null })
      })
      .catch((error) => {
        if (active && requestID === novelPanelRequest.current) setNovelPanel({ loading: false, loadedProject: panelProjectId, workspace: null, error })
      })
    return () => { active = false }
  }, [pathname, panelProjectId])

  let page
  if (pathname === '/') {
    page = <HomePage onNavigate={navigate} />
  } else if (pathname === '/novel-fetch') {
    page = <NovelFetchPage />
  } else if (pathname === '/batch-factory') {
    const parsed = parseBatchProjectSearch(search)
    page = parsed.projectId
      ? <BatchProjectListPage key={parsed.projectId} initialProjectId={parsed.projectId} onClearProject={() => navigate('/batch-factory')} />
      : <BatchFactoryHome initialError={parsed.error} initialArchived={parsed.archived || 'active'} onArchivedChange={(archived) => navigate(archived === 'archived' ? '/batch-factory?archived=archived' : '/batch-factory')} onOpenProject={(id) => navigate(`/batch-factory?projectId=${id}`)} />
  } else if (pathname === '/shuihuo-production') {
    page = <ShuihuoProductionPage />
  } else if (pathname === '/novel-fetch-workshop') {
    page = <NovelFetchWorkshop />
  } else if (pathname === '/script') {
    page = <ScriptWorkspace />
  } else if (pathname === '/tts') {
    page = <TtsPage />
  } else if (pathname === '/history') {
    page = <HistoryPage />
  } else if (pathname === '/settings') {
    page = <SettingsPage theme={theme} onTheme={onToggleTheme} />
  } else if (pathname === '/profile') {
    page = <AccountCenterPage user={currentUser} onLogout={onLogout} />
  } else if (pathname === '/member') {
    page = <AccountCenterPage user={currentUser} onLogout={onLogout} mode="member" />
  } else if (pathname === '/agent' || pathname === '/agent/canvas') {
    page = <AgentReservedPage onNavigate={navigate} />
  } else if (pathname === '/novel-panel') {
    page = !Number.isSafeInteger(panelProjectId) || panelProjectId <= 0
      ? <RouteFoundation title="小说面板" description="请从批量项目进入小说面板（需要 projectId）。" onNavigate={navigate} />
      : novelPanel.loading || novelPanel.loadedProject !== panelProjectId ? <RouteFoundation title="小说面板" description="正在读取工作区…" onNavigate={navigate} />
        : novelPanel.error ? <RouteFoundation title="小说面板" description={novelPanel.error.message || '读取失败，请重试。'} onNavigate={navigate} />
          : <NovelPanelWorkbench key={`novel-panel-${panelProjectId}`} projectId={panelProjectId} initialWorkspace={novelPanel.workspace} api={{ saveWorkspace: saveNovelPanel, listHistory: listNovelPanelHistory, restoreHistory: restoreNovelPanelHistory }} />
  } else if (routeFoundations[pathname]) {
    page = <RouteFoundation {...routeFoundations[pathname]} onNavigate={navigate} />
  } else {
    page = (
      <main className="route-foundation-page">
        <Result
          status="404"
          title="页面不存在"
          subTitle="该地址尚未接入新主线。"
          extra={<Button type="primary" onClick={() => navigate('/')}>返回首页</Button>}
        />
      </main>
    )
  }

  return (
    <UserShell pathname={pathname} theme={theme} onToggleTheme={onToggleTheme} onNavigate={navigate} currentUser={currentUser} onLogout={onLogout}>
      {page}
    </UserShell>
  )
}

function parseBatchProjectSearch(search) {
  if (!search) return { projectId: null, error: '' }
  if (search === '?archived=archived') return { projectId: null, archived: 'archived', error: '' }
  const matched = search.match(/^\?projectId=([1-9]\d*)$/)
  if (!matched) return { projectId: null, error: '项目入口无效' }
  const projectId = Number(matched[1])
  return Number.isSafeInteger(projectId) ? { projectId, error: '' } : { projectId: null, error: '项目入口无效' }
}

export { parseBatchProjectSearch, routeFoundations }
