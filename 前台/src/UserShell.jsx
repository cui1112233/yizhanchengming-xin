import React, { useState } from 'react'
import { Button, Drawer, Dropdown } from 'antd'
import WorkspaceIcon from './WorkspaceIcon.jsx'

const primaryNavItems = [
  { href: '/', label: '首页', icon: 'batch' },
  { href: '/script', label: '剧本生成', icon: 'script' },
  { href: '/novel-fetch', label: '小说获取', icon: 'novel' },
  { href: '/novel-panel', label: '小说面板', icon: 'novel' },
  { href: '/shuihuo-production', label: '水货生产', icon: 'shuihuo' },
  { href: '/agent', label: 'Agent 工作区', icon: 'agent' },
  { href: '/history', label: '历史', icon: 'batch' },
  { href: '/tts', label: '配音', icon: 'tts' },
]

function ShellLink({ href, active, onNavigate, className = '', children }) {
  const handleClick = (event) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(href)
  }
  return <a href={href} className={`${className} ${active ? 'is-active' : ''}`.trim()} aria-current={active ? 'page' : undefined} onClick={handleClick}>{children}</a>
}

export default function UserShell({ pathname, theme, onToggleTheme, onNavigate, currentUser, onLogout, children }) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [decorationVisible, setDecorationVisible] = useState(true)
  const isHome = pathname === '/'
  const logo = theme === 'dark' || isHome ? '/assets/brand-logo-white.png' : '/assets/brand-logo-black.png'
  const navigate = (href) => { setMobileOpen(false); onNavigate(href) }
  const renderNavigation = (className, label, withIcons = false) => <nav className={className} aria-label={label}>{primaryNavItems.map((item) => <ShellLink key={item.href} href={item.href} active={pathname === item.href || (item.href === '/agent' && pathname === '/agent/canvas')} onNavigate={navigate}>{withIcons ? <><WorkspaceIcon name={item.icon} className="workspace-nav-icon" /><span className="workspace-nav-label">{item.label}</span></> : item.label}</ShellLink>)}</nav>
  const menuItems = [{ key: 'profile', label: '个人资料' }, { key: 'member', label: '会员与权限' }, { type: 'divider' }, { key: 'logout', danger: true, label: '退出登录' }]
  const accountName = currentUser?.name || currentUser?.username || '用户入口'

  return (
    <div className={`user-shell ${isHome ? 'user-shell-home' : 'user-shell-workspace'}`}>
      {!isHome && <aside className={`workspace-sidebar ${sidebarCollapsed ? 'is-collapsed' : ''}`}>
        <div className="workspace-sidebar-brand">
          <ShellLink href="/" active={false} onNavigate={navigate} className="workspace-sidebar-brand-link"><img src={logo} alt="一战晟铭" /><strong>一战晟铭</strong></ShellLink>
          <Button type="text" aria-label={sidebarCollapsed ? '展开侧边导航' : '收起侧边导航'} onClick={() => setSidebarCollapsed(value => !value)}>{sidebarCollapsed ? '›' : '‹'}</Button>
        </div>
        {renderNavigation('workspace-sidebar-nav', '工作区侧边导航', true)}
        <div className="workspace-sidebar-tools">
          <Button type="text" onClick={onToggleTheme} aria-label="主题">主题</Button>
          <ShellLink href="/settings" active={pathname === '/settings'} onNavigate={navigate}>设置</ShellLink>
        </div>
        <button type="button" className="workspace-sidebar-account" onClick={() => navigate('/profile')} aria-label="进入个人资料"><span>{accountName.slice(0, 1)}</span><strong>{accountName}</strong></button>
      </aside>}
      <div className="user-shell-stage">
        {isHome ? <header className="user-shell-header">
          <ShellLink href="/" active={isHome} onNavigate={navigate} className="user-shell-brand"><img src={logo} alt="一战晟铭" /><strong>一战晟铭</strong></ShellLink>
          {renderNavigation('user-shell-nav', '用户导航')}
          <div className="user-shell-actions">
            <Button type="text" className="shell-mobile-menu" aria-label="打开导航" onClick={() => setMobileOpen(true)}>导航</Button>
            <Button type="text" className="shell-action-button" onClick={onToggleTheme} aria-label="主题">主题</Button>
            <ShellLink href="/settings" active={pathname === '/settings'} onNavigate={navigate} className="shell-action-link">设置</ShellLink>
            <Dropdown menu={{ items: menuItems, onClick: ({ key }) => { if (key === 'logout') void onLogout?.(); else navigate(`/${key}`) } }} trigger={['click']}>
              <Button className="shell-user-entry">{accountName}</Button>
            </Dropdown>
          </div>
        </header> : <header className="user-shell-header workspace-mobile-header">
          <ShellLink href="/" active={false} onNavigate={navigate} className="user-shell-brand"><img src={logo} alt="一战晟铭" /><strong>一战晟铭</strong></ShellLink>
          <div className="user-shell-actions">
            <Button type="text" className="shell-mobile-menu" aria-label="打开导航" onClick={() => setMobileOpen(true)}>导航</Button>
            <Dropdown menu={{ items: menuItems, onClick: ({ key }) => { if (key === 'logout') void onLogout?.(); else navigate(`/${key}`) } }} trigger={['click']}><Button className="shell-user-entry">{accountName}</Button></Dropdown>
          </div>
        </header>}
        <Drawer title="工作区导航" placement="left" open={mobileOpen} onClose={() => setMobileOpen(false)} className="shell-mobile-drawer">{renderNavigation('user-shell-nav', '移动端工作区导航')}</Drawer>
        <div className="user-shell-content">{children}</div>
      </div>
      {decorationVisible ? <aside className="shell-decoration" aria-label="桌面装饰"><span aria-hidden="true" className="shell-decoration-orbit" /><span aria-hidden="true" className="shell-decoration-star">✦</span><button type="button" className="shell-decoration-close" onClick={() => setDecorationVisible(false)} aria-label="关闭桌面装饰">×</button></aside> : null}
    </div>
  )
}

export { primaryNavItems }
