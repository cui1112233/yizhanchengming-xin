import { Avatar } from 'antd';
import { ChartNoAxesCombined, ChevronLeft, Crown, KeyRound, LogOut, ShieldCheck, UserRound, UsersRound } from 'lucide-react';
import { Link } from '../components/Link';

function roleLabel(role) {
  if (role === 'dev') return 'DEV';
  if (role === 'manager') return 'MANAGER';
  return 'MEMBER';
}

export function AccountCenterLayout({ account, username, avatar, pathname, onReturn, onLogout, children }) {
  const profileOpen = ['/profile', '/member', '/security', '/advanced-team-admin'].includes(pathname);
  const canManage = ['dev', 'manager'].includes(account?.role);
  const canAccessAdmin = account?.role === 'dev';

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
          {profileOpen ? <div className="account-center-sidebar-submenu"><Link href="/member" className={pathname === '/member' ? 'active' : ''} aria-current={pathname === '/member' ? 'page' : undefined}><Crown size={17} />会员中心</Link><Link href="/security" className={pathname === '/security' ? 'active' : ''} aria-current={pathname === '/security' ? 'page' : undefined}><ShieldCheck size={17} />账号安全</Link>{canManage ? <Link href="/advanced-team-admin" className={pathname === '/advanced-team-admin' ? 'active' : ''} aria-current={pathname === '/advanced-team-admin' ? 'page' : undefined}><UsersRound size={17} />联合治理</Link> : null}</div> : null}
        </div>
        <div className="account-center-sidebar-group"><Link href="/api-config" className={pathname === '/api-config' ? 'active' : ''} aria-current={pathname === '/api-config' ? 'page' : undefined}><KeyRound size={17} />API 配置</Link><Link href="/usage" className={pathname === '/usage' ? 'active' : ''} aria-current={pathname === '/usage' ? 'page' : undefined}><ChartNoAxesCombined size={17} />用量与制作</Link>{canManage ? <Link href="/team" className={pathname === '/team' ? 'active' : ''} aria-current={pathname === '/team' ? 'page' : undefined}><UsersRound size={17} />组员管理</Link> : null}</div>
        {canAccessAdmin ? <div className="account-center-sidebar-group account-center-sidebar-developer"><span>开发者工具</span><Link href="/accounts" className={pathname === '/accounts' ? 'active' : ''} aria-current={pathname === '/accounts' ? 'page' : undefined}><UsersRound size={17} />账号与角色</Link><Link href="/admin/presets" reload><ShieldCheck size={17} />管理后台</Link></div> : null}
      </nav>
      <button type="button" className="account-center-sidebar-logout" onClick={onLogout}><LogOut size={17} />退出登录</button>
    </aside>
    <main className="account-center-main">{children}</main>
  </div>;
}
