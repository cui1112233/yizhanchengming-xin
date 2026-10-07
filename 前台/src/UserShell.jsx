import React, { useState } from 'react'
import { Button, Drawer, Dropdown } from 'antd'

const primaryNavItems = [
  { href: '/', label: '首页' },
  { href: '/script', label: '剧本生成' },
  { href: '/novel-fetch', label: '小说获取' },
  { href: '/novel-panel', label: '小说面板' },
  { href: '/shuihuo-production', label: '水货生产' },
  { href: '/agent', label: 'Agent 工作区' },
  { href: '/history', label: '历史' },
  { href: '/issues', label: '问题日志' },
  { href: '/tts', label: '配音' },
]

function ShellLink({ href, active, onNavigate, className = '', children }) {
  const handleClick = (event) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(href)
  }
  return (
    <a href={href} className={`${className} ${active ? 'is-active' : ''}`.trim()} aria-current={active ? 'page' : undefined} onClick={handleClick}>
      {children}
    </a>
  )
}

export default function UserShell({ pathname, theme, onToggleTheme, onNavigate, currentUser, onLogout, children }) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const isHome = pathname === '/'
  const logo = theme === 'dark' || isHome ? '/assets/brand-logo-white.png' : '/assets/brand-logo-black.png'
  const navigate = (href) => { setMobileOpen(false); onNavigate(href) }
  const navigation = <nav className="user-shell-nav" aria-label="用户导航">{primaryNavItems.map((item) => <ShellLink key={item.href} href={item.href} active={pathname === item.href} onNavigate={navigate}>{item.label}</ShellLink>)}</nav>
  const menuItems = [{ key: 'profile', label: '个人资料' }, { key: 'member', label: '会员与权限' }, { type: 'divider' }, { key: 'logout', danger: true, label: '退出登录' }]
  return (
    <div className={`user-shell ${isHome ? 'user-shell-home' : ''}`}>
      <header className="user-shell-header">
        <ShellLink href="/" active={isHome} onNavigate={onNavigate} className="user-shell-brand">
          <img src={logo} alt="一战晟铭" /><strong>一战晟铭</strong>
        </ShellLink>
        {navigation}
        <div className="user-shell-actions">
          <Button type="text" className="shell-mobile-menu" aria-label="打开导航" onClick={() => setMobileOpen(true)}>导航</Button>
          <Button type="text" className="shell-action-button" onClick={onToggleTheme} aria-label="主题">主题</Button>
          <ShellLink href="/settings" active={pathname === '/settings'} onNavigate={onNavigate} className="shell-action-link">设置</ShellLink>
          <Dropdown menu={{ items: menuItems, onClick: ({ key }) => { if (key === 'logout') void onLogout?.(); else navigate(`/${key}`) } }} trigger={['click']}>
            <Button className="shell-user-entry">{currentUser?.displayName || currentUser?.username || '用户入口'}</Button>
          </Dropdown>
        </div>
      </header>
      <Drawer title="工作区导航" placement="left" open={mobileOpen} onClose={() => setMobileOpen(false)} className="shell-mobile-drawer">{navigation}</Drawer>
      <div className="user-shell-content">{children}</div>
    </div>
  )
}

export { primaryNavItems }
