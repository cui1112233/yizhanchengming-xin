import { useEffect, useRef, useState } from 'react';
import { AudioLines, Bot, Clapperboard, FilePenLine, Menu, X } from 'lucide-react';
import { BrandLogo } from '../../shared/components/BrandLogo';
import { Link } from '../../shared/components/Link';
import { HomeHandwriting } from '../components/HomeHandwriting';
import './cinematic-home.css';

const scenes = [0, 1, 2, 3].map(index => '/assets/cinematic/scene-' + index + '.mp4');
const quickActions = [
  { href: '/script', icon: FilePenLine, title: '私人定制', desc: '小说转剧本与漫画素材' },
  { href: '/batch-factory', icon: Clapperboard, title: '水货生产', desc: '批次管理与漫剧生产' },
  { href: '/agent', icon: Bot, title: 'CM 画布', desc: '串联创作与生成流程' },
  { href: '/tts', icon: AudioLines, title: '配音工坊', desc: '为你的故事赋予声音' }
];
const navItems = [
  { href: '/script', label: '私人定制' }, { href: '/novel-fetch', label: '小说获取' },
  { href: '/batch-factory', label: '水货生产' }, { href: '/agent', label: 'CM 画布' }
];

export function HomePage({ isLoggedIn, onOpenLogin }) {
  const [active, setActive] = useState(0);
  const [menuOpen, setMenuOpen] = useState(false);
  const videoRefs = useRef([]);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let disposed = false, timer, fading;
    const current = videoRefs.current[active];
    const next = videoRefs.current[(active + 1) % scenes.length];
    const start = () => {
      clearTimeout(timer);
      if (document.hidden || disposed) return;
      current?.play().catch(() => {});
      timer = setTimeout(() => {
        if (next?.readyState >= 2) setActive(index => (index + 1) % scenes.length);
        else { next?.load(); start(); }
      }, 14000);
    };
    // Preload just the next scene; only the active video plays.
    if (next) next.preload = 'auto';
    fading = setTimeout(() => videoRefs.current.forEach((video, index) => { if (index !== active) video?.pause(); }), 1000);
    const visibility = () => {
      if (document.hidden) { clearTimeout(timer); videoRefs.current.forEach(video => video?.pause()); }
      else start();
    };
    current?.addEventListener('loadeddata', start);
    document.addEventListener('visibilitychange', visibility);
    setFailed(false);
    start();
    return () => { disposed = true; clearTimeout(timer); clearTimeout(fading); current?.removeEventListener('loadeddata', start); document.removeEventListener('visibilitychange', visibility); };
  }, [active]);

  function requireLogin(event) {
    setMenuOpen(false);
    if (isLoggedIn) return;
    event.preventDefault();
    onOpenLogin?.();
  }

  return <div className="cinematic-home">
    <section className="scene" aria-label="一战晟铭创作首页">
      <div className="videos" aria-hidden="true">{scenes.map((src, index) => <video key={src} ref={element => { videoRefs.current[index] = element; }} className={`video${active === index ? ' active' : ''}`} src={src} autoPlay={index === 0} muted loop playsInline preload={index === 0 ? 'auto' : 'metadata'} onError={() => { if (index === active) setFailed(true); }} />)}</div>
      <div className="shade" />
      <img className="foreground" src="/assets/cinematic/train-window.png" alt="" aria-hidden="true" />
      <div className="content">
        <nav className="nav" aria-label="首页导航">
          <Link href="/" className="logo"><BrandLogo /><span className="shine">一战晟铭</span></Link>
          <div className="glass navlinks">{navItems.map(item => <Link key={item.href} href={item.href} onClick={requireLogin}>{item.label}</Link>)}<Link href="/script" className="white" onClick={requireLogin}>{isLoggedIn ? '进入工作台' : '登录 / 进入工作台'}</Link></div>
          <button className="glass hamburger" aria-label="打开导航菜单" aria-expanded={menuOpen} onClick={() => setMenuOpen(true)}><Menu size={21} /></button>
        </nav>
        <main className="hero">
          <div className="journey-label sans">一战晟铭 · 下一站，你的故事</div>
          <HomeHandwriting />
          <p className="sans subtext">让灵感随沿途的光影抵达，把心中的故事写成剧本、绘成画面。<br />从第一段文字到第一声对白，在这里，开启你的创作旅程。</p>
          <div className="home-entry"><Link className="white entry-main" href="/script" onClick={requireLogin}>开启创作旅程　↗</Link><Link className="glass entry-secondary" href="/script" onClick={requireLogin}>开启旅程</Link></div>
        </main>
        <footer className="home-tools sans">{quickActions.map(action => { const Icon = action.icon; return <Link key={action.href} href={action.href} onClick={requireLogin} className="glass tool"><Icon size={25} strokeWidth={1.4} className="tool-symbol" /><div><b>{action.title}</b><small>{action.desc}</small></div><span className="arrow">↗</span></Link>; })}</footer>
      </div>
      {failed && <div className="loading" role="status">背景暂时无法播放，创作入口仍可正常使用</div>}
    </section>
    {menuOpen && <div className="mobile open"><button className="close" aria-label="关闭导航菜单" onClick={() => setMenuOpen(false)}><X size={22} /></button>{navItems.map(item => <Link key={item.href} href={item.href} onClick={requireLogin}>{item.label}</Link>)}<Link href="/script" className="white" onClick={requireLogin}>{isLoggedIn ? '进入工作台' : '登录 / 进入工作台'}</Link></div>}
  </div>;
}
