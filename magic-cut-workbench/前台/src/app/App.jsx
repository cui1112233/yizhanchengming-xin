import { Button, ConfigProvider, Spin, theme } from 'antd'
import { useEffect, useState } from 'react'

import { HomePage } from '../features/home/HomePage.jsx'
import { EditorPage } from '../features/editor/EditorPage.jsx'
import { projectIDFromPath } from '../features/editor/editorModel.js'
import { TaskCenter } from '../features/task-center/TaskCenter.jsx'
import { loadSession } from '../shared/auth/session.js'
import { WorkModeMenu } from '../shared/work-mode/WorkModeMenu.jsx'
import { createSessionGateModel } from './sessionGateModel.js'
import './app.css'

export function App() {
  const [session, setSession] = useState({ loading: true, authenticated: false, account: null, message: '' })
  const [pathname, setPathname] = useState(window.location.pathname)

  useEffect(() => {
    let active = true
    loadSession().then(result => {
      if (active) setSession({ loading: false, ...result })
    })
    return () => { active = false }
  }, [])

  useEffect(() => {
    const updatePath = () => setPathname(window.location.pathname)
    window.addEventListener('popstate', updatePath)
    return () => window.removeEventListener('popstate', updatePath)
  }, [])

  const projectID = projectIDFromPath(pathname)
  const sessionGate = createSessionGateModel(session, window.location)

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#6f8ff7',
          colorBgBase: '#111827',
          borderRadius: 10,
          fontFamily: 'Inter, "Microsoft YaHei", sans-serif',
        },
      }}
    >
      {session.loading ? (
          <main className="account-gate"><Spin size="large" /><p>正在读取一战晟铭账号……</p></main>
        ) : !session.authenticated ? (
          <main className="account-gate">
            <div className="account-gate__card">
              <span className="account-gate__mark">M</span>
              <h1>{sessionGate.title}</h1>
              <p>{sessionGate.message}</p>
              <Button type="primary" size="large" href={sessionGate.returnHref}>{sessionGate.returnLabel}</Button>
            </div>
          </main>
        ) : pathname === '/' ? (
          <HomePage account={session.account} />
        ) : (
          <div className="magic-shell">
            <aside className="magic-sidebar">
              <WorkModeMenu />
              <nav className="magic-sidebar__nav" aria-label="魔剪功能">
                <a href="/">首页</a>
                <a className={pathname === '/tasks' ? 'is-current' : ''} href="/tasks">任务中心</a>
              </nav>
              <p className="magic-sidebar__version">4040 独立开发版</p>
            </aside>
            <div className="magic-content">
              <div className="account-chip">{session.account?.displayName || session.account?.username}</div>
              {pathname === '/tasks' ? <TaskCenter /> : <EditorPage projectID={projectID} />}
            </div>
          </div>
        )}
    </ConfigProvider>
  )
}
