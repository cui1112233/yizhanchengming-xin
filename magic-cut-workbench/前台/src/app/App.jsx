import { ConfigProvider, theme } from 'antd'

import { HomePage } from '../features/home/HomePage.jsx'
import { WorkModeMenu } from '../shared/work-mode/WorkModeMenu.jsx'
import './app.css'

export function App() {
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
      <div className="magic-shell">
        <aside className="magic-sidebar">
          <WorkModeMenu />
          <nav className="magic-sidebar__nav" aria-label="魔剪功能">
            <a className="is-current" href="/">首页</a>
            <a href="/tasks">任务中心</a>
          </nav>
          <p className="magic-sidebar__version">4040 独立开发版</p>
        </aside>
        <HomePage />
      </div>
    </ConfigProvider>
  )
}
