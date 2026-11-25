import { Avatar } from 'antd';
import { ChartNoAxesCombined, ChevronLeft, Crown, ExternalLink, KeyRound, LogOut, ShieldCheck, UserRound, UsersRound } from 'lucide-react';
import { Link } from '../components/Link';
import '../styles/profile-grey-glass.css';

function roleLabel(role) {
  if (role === 'dev') return 'DEV';
  if (role === 'manager') return 'MANAGER';
  return 'MEMBER';
}

export function AccountCenterLayout({ account, username, avatar, pathname, onReturn, onLogout, children }) {
  const canManage = ['dev', 'manager'].includes(account?.role);
  const canAccessAdmin = account?.role === 'dev';

  // Roll out the new shell only to Profile; other account pages keep their layout.
  if (pathname === '/profile') {
    const entries = [
      { href: '/profile', label: '个人资料' },
      { href: '/member', label: '会员中心' },
      { href: '/security', label: '账号安全' },
      ...(canManage ? [{ href: '/advanced-team-admin', label: '联合治理' }, { href: '/api-config', label: 'API 配置' }] : []),
      { href: 'http://newapi.yzcm.site/', label: 'YZCM API', external: true },
      { href: '/usage', label: '用量与制作' },
      ...(canManage ? [{ href: '/team', label: '组员管理' }] : []),
      ...(canAccessAdmin ? [{ href: '/accounts', label: '账号与角色', developer: true }, { href: '/admin/presets', label: '管理后台', reload: true, developer: true }] : [])
    ];
    return <div className="account-center-workspace profile-glass-workspace">
      <header className="pg-shell-header">
        <button type="button" className="pg-brand" onClick={onReturn} aria-label="一战晟铭，返回工作台"><b aria-hidden="true">Y</b><span>一战晟铭</span></button>
        <div className="pg-shell-account"><Avatar size={32} src={account?.avatarUrl} style={{ backgroundColor: avatar?.background }}>{avatar?.emoji}</Avatar><span>{account?.displayName || username}</span><em>{roleLabel(account?.role)}</em></div>
      </header>
      <nav className="pg-navigation" aria-label="个人中心导航">
        {entries.map(entry => entry.external
          ? <a key={entry.href} href={entry.href} target="_blank" rel="noreferrer" title="一战晟铭专属 API 控制台">{entry.label}<ExternalLink size={12} aria-hidden="true" /></a>
          : <Link key={entry.href} href={entry.href} reload={entry.reload} aria-current={pathname === entry.href ? 'page' : undefined} className={pathname === entry.href ? 'active' : ''}>{entry.label}{entry.developer ? <small>DEV</small> : null}</Link>)}
      </nav>
      <main className="account-center-main pg-main">{children}</main>
      <footer className="pg-shell-footer"><button type="button" onClick={onReturn} aria-label="返回工作台"><ChevronLeft size={16} />返回工作台</button><button type="button" onClick={onLogout}><LogOut size={16} />退出登录</button></footer>
    </div>;
  }

  return <div className="account-center-workspace">
    <aside className="account-center-sidebar" aria-label="个人中心导航">
      <div className="account-center-sidebar-head">
        <button type="button" className="account-center-return" aria-label="返回工作台" onClick={onReturn}><ChevronLeft size={19} /><span>返回工作台</span></button>
        <strong>个人中心</strong>
      </div>
      <div className="account-center-sidebar-profile">
        <Avatar size={54} src={account?.avatarUrl} style={{ backgroundColor: avatar?.background }}>{avatar?.emoji}</Avatar>
        <div><strong>{account?.displayName || username}</strong><small>@{username}</small><em>{roleLabel(account?.role)}</em></div>
      </div>
      <nav className="account-center-sidebar-nav">
        <div className="account-center-sidebar-group">
          <Link href="/profile" className={pathname === '/profile' ? 'active' : ''} aria-current={pathname === '/profile' ? 'page' : undefined}><UserRound size={18} />个人资料</Link>
          <div className="account-center-sidebar-submenu"><Link href="/member" className={pathname === '/member' ? 'active' : ''} aria-current={pathname === '/member' ? 'page' : undefined}><Crown size={17} />会员中心</Link><Link href="/security" className={pathname === '/security' ? 'active' : ''} aria-current={pathname === '/security' ? 'page' : undefined}><ShieldCheck size={17} />账号安全</Link>{canManage ? <Link href="/advanced-team-admin" className={pathname === '/advanced-team-admin' ? 'active' : ''} aria-current={pathname === '/advanced-team-admin' ? 'page' : undefined}><UsersRound size={17} />联合治理</Link> : null}</div>
        </div>
        <div className="account-center-sidebar-group">{canManage ? <Link href="/api-config" className={pathname === '/api-config' ? 'active' : ''} aria-current={pathname === '/api-config' ? 'page' : undefined}><KeyRound size={17} />API 配置</Link> : null}<a href="http://newapi.yzcm.site/" target="_blank" rel="noreferrer" title="一战晟铭专属 API 控制台"><ExternalLink size={17} />YZCM API</a><Link href="/usage" className={pathname === '/usage' ? 'active' : ''} aria-current={pathname === '/usage' ? 'page' : undefined}><ChartNoAxesCombined size={17} />用量与制作</Link>{canManage ? <Link href="/team" className={pathname === '/team' ? 'active' : ''} aria-current={pathname === '/team' ? 'page' : undefined}><UsersRound size={17} />组员管理</Link> : null}</div>
        {canAccessAdmin ? <div className="account-center-sidebar-group account-center-sidebar-developer"><span>开发者工具</span><Link href="/accounts" className={pathname === '/accounts' ? 'active' : ''} aria-current={pathname === '/accounts' ? 'page' : undefined}><UsersRound size={17} />账号与角色</Link><Link href="/admin/presets" reload><ShieldCheck size={17} />管理后台</Link></div> : null}
      </nav>
      <button type="button" className="account-center-sidebar-logout" onClick={onLogout}><LogOut size={17} />退出登录</button>
    </aside>
    <main className="account-center-main">{children}</main>
  </div>;
}
