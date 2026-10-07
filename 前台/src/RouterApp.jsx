import React, { useEffect, useState } from 'react'
import { Button, Result } from 'antd'
import IntakeWorkbench from './App.jsx'
import HomePage from './HomePage.jsx'
import UserShell from './UserShell.jsx'
import NovelFetchWorkshop from './NovelFetchWorkshop.jsx'
import './home.css'

const routeFoundations = {
  '/script': { title: '剧本生成', description: '路由基础已恢复；剧本生成实际业务由对应迁移任务接入。' },
  '/novel-panel': { title: '小说面板', description: '路由基础已恢复；小说面板实际业务由对应迁移任务接入。' },
  '/agent': { title: 'Agent 工作区', description: '路由基础已恢复；Agent 实际业务由对应迁移任务接入。' },
  '/tts': { title: '配音', description: '路由基础已恢复；配音实际业务由对应迁移任务接入。' },
  '/history': { title: '历史', description: '历史入口已恢复；真实历史项目读取由后续项目聚合接口接入。' },
  '/issues': { title: '问题日志', description: '问题日志入口已恢复；实际日志页面不属于本任务范围。' },
  '/settings': { title: '设置', description: '设置入口已恢复；实际设置页不属于本任务范围。' },
  '/member': { title: '用户入口', description: '用户入口已恢复；账号中心页面不属于本任务范围。' },
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

export default function RouterApp({ theme, onToggleTheme }) {
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

  let page
  if (pathname === '/') {
    page = <HomePage onNavigate={navigate} />
  } else if (pathname === '/novel-fetch' || pathname === '/batch-factory' || pathname === '/shuihuo-production') {
    page = <IntakeWorkbench />
  } else if (pathname === '/novel-fetch-workshop') {
    page = <NovelFetchWorkshop />
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
