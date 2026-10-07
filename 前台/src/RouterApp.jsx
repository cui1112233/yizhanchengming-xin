import React, { useEffect, useState } from 'react'
import { Button, Result } from 'antd'
import IntakeWorkbench from './App.jsx'
import HomePage from './HomePage.jsx'
import UserShell from './UserShell.jsx'
import NovelFetchWorkshop from './NovelFetchWorkshop.jsx'
import ScriptWorkspace from './ScriptWorkspace.jsx'
import NovelPanelWorkbench from './novel-panel/NovelPanelWorkbench.jsx'
import TtsPage from './TtsPage.jsx'
import HistoryPage from './HistoryPage.jsx'
import AccountCenterPage from './AccountCenterPage.jsx'
import IssuesPage from './IssuesPage.jsx'
import { getNovelPanel, listNovelPanelHistory, restoreNovelPanelHistory, saveNovelPanel } from './api.js'
import './home.css'

const routeFoundations = {
  '/agent': { title: 'Agent 工作区', description: '路由基础已恢复；Agent 实际业务由对应迁移任务接入。' },
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
  const [pathname, setPathname] = useState(() => (
    typeof window === 'undefined' ? '/' : window.location.pathname
  ))

  useEffect(() => {
    const handlePopState = () => setPathname(window.location.pathname)
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const navigate = (path) => {
    if (window.location.pathname !== path) {
      window.history.pushState({}, '', path)
    }
    setPathname(path)
  }
  const [novelPanel, setNovelPanel] = useState({ loading: false, loadedProject: 0, workspace: null, error: null })
  const panelProjectId = Number(new URLSearchParams(typeof window === 'undefined' ? '' : window.location.search).get('projectId'))
  useEffect(() => {
    if (pathname !== '/novel-panel' || !Number.isSafeInteger(panelProjectId) || panelProjectId <= 0 || novelPanel.loadedProject === panelProjectId || novelPanel.loading) return
    setNovelPanel((value) => ({ ...value, loading: true, error: null }))
    getNovelPanel(panelProjectId).then((result) => setNovelPanel({ loading: false, loadedProject: panelProjectId, workspace: result.workspace, error: null })).catch((error) => setNovelPanel({ loading: false, loadedProject: panelProjectId, workspace: null, error }))
  }, [pathname, panelProjectId, novelPanel.loadedProject, novelPanel.loading])

  let page
  if (pathname === '/') {
    page = <HomePage onNavigate={navigate} />
  } else if (pathname === '/novel-fetch' || pathname === '/batch-factory' || pathname === '/shuihuo-production') {
    page = <IntakeWorkbench />
  } else if (pathname === '/novel-fetch-workshop') {
    page = <NovelFetchWorkshop />
  } else if (pathname === '/script') {
    page = <ScriptWorkspace />
  } else if (pathname === '/tts') {
    page = <TtsPage />
  } else if (pathname === '/history') {
    page = <HistoryPage />
  } else if (pathname === '/issues') {
    page = <IssuesPage />
  } else if (pathname === '/profile') {
    page = <AccountCenterPage user={currentUser} onLogout={onLogout} />
  } else if (pathname === '/member') {
    page = <AccountCenterPage user={currentUser} onLogout={onLogout} mode="member" />
  } else if (pathname === '/novel-panel') {
    page = !Number.isSafeInteger(panelProjectId) || panelProjectId <= 0
      ? <RouteFoundation title="小说面板" description="请从批量项目进入小说面板（需要 projectId）。" onNavigate={navigate} />
      : novelPanel.loading ? <RouteFoundation title="小说面板" description="正在读取工作区…" onNavigate={navigate} />
        : novelPanel.error ? <RouteFoundation title="小说面板" description={novelPanel.error.message || '读取失败，请重试。'} onNavigate={navigate} />
          : <NovelPanelWorkbench projectId={panelProjectId} initialWorkspace={novelPanel.workspace} api={{ saveWorkspace: saveNovelPanel, listHistory: listNovelPanelHistory, restoreHistory: restoreNovelPanelHistory }} />
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
    <UserShell pathname={pathname} theme={theme} onToggleTheme={onToggleTheme} onNavigate={navigate}>
      {page}
    </UserShell>
  )
}

export { routeFoundations }
