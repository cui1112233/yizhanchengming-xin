import { ConfigProvider } from 'antd';
import { AudioLines, Bot, ChevronRight, Clapperboard, FilePenLine } from 'lucide-react';
import { useState } from 'react';
import { BrandLogo } from '../../shared/components/BrandLogo';
import { GradientButton } from '../../shared/components/GradientButton';
import { Link } from '../../shared/components/Link';
import { V88Button } from '../../shared/components/V88Button';
import WorkModeMenu from '../../shared/layouts/WorkModeMenu';
import { createV88AntTheme } from '../../shared/styles/theme';
import HomeSplashCursor from '../components/HomeSplashCursor';
import './home-v88.css';

const quickActions = [
  { href: '/script', className: 'card-write', icon: FilePenLine, title: '私人定制', desc: '粘贴内容，生成剧本或漫画制作素材' },
  { href: '/batch-factory', className: 'card-shuihuo', icon: Clapperboard, title: '水货生产', desc: '统一管理批次、单书配置、分镜生产与视频提交' },
  { href: '/agent', className: 'card-agent', icon: Bot, title: 'AI 智能 Agent', desc: '专属助手对话，辅助精细化剧本包装' },
  { href: '/tts', className: 'card-tts', icon: AudioLines, title: '声音配音工坊', desc: '多音色情感合成，让你的画面声临其境' }
];

const navItems = [
  { href: '/', label: '首页' },
  { href: '/script', label: '私人定制' },
  { href: '/novel-fetch', label: '小说获取' },
  { href: '/batch-factory', label: '水货生产' },
  { href: '/agent', label: 'CM 画布' },
  { href: '/tts', label: '配音' },
  { href: '/settings', label: '设置' }
];

export function HomePage({ isLoggedIn, onOpenLogin, theme = 'dark' }) {
  const pathname = window.location.pathname;

  function requireLogin(event) {
    if (isLoggedIn) return;
    event.preventDefault();
    onOpenLogin?.();
  }

  function openProtectedPage(href) {
    if (!isLoggedIn) {
      onOpenLogin?.();
      return;
    }
    window.location.href = href;
  }

  return (
    <ConfigProvider theme={createV88AntTheme(theme)}>
    <div className="home-page">
      <section className="home-video-hero">
        <video className="home-hero-video" autoPlay muted loop playsInline>
          <source src="/assets/home-hero.mp4" type="video/mp4" />
        </video>
        <div className="home-video-overlay" />
        <HomeSplashCursor />

        <nav className="home-hero-nav">
          <WorkModeMenu pathname={pathname} className="home-hero-brand"><BrandLogo className="home-brand-logo" /><span className="shine">一战晟铭</span></WorkModeMenu>
          <div className="home-hero-links">
            {navItems.map(item => <Link key={item.href} href={item.href} onClick={item.href === '/' ? undefined : requireLogin}>{item.label}</Link>)}
          </div>
          <div className="home-hero-actions-right">
            <GradientButton as="a" href="#contact">联系</GradientButton>
          </div>
        </nav>

        <div className="home-hero-content">
          <p className="home-hero-kicker">Novel Visual Script Studio</p>
          <h1 className="home-hero-title">让小说章节直接进入可视化剧本工作流</h1>
          <p className="home-hero-subtitle">从原文提取人物、场景、节奏和镜头结构，生成可继续编辑、导出和配音的短剧制作素材。</p>
          <div className="home-hero-actions">
            <GradientButton onClick={() => openProtectedPage('/script')}>开始生成</GradientButton>
            <GradientButton onClick={() => openProtectedPage('/tts')}>进入配音</GradientButton>
          </div>
        </div>
      </section>

      <section className="home-quick-actions">
        {quickActions.map(action => {
          const Icon = action.icon;
          return <Link key={action.href} href={action.href} onClick={requireLogin} className={`quick-action-card ${action.className}`}>
            <span className="card-icon"><Icon size={30} strokeWidth={1.7} aria-hidden="true" /></span>
            <span className="card-info"><span className="card-title">{action.title}</span><span className="card-desc">{action.desc}</span></span>
            <ChevronRight className="card-arrow" size={20} aria-hidden="true" />
          </Link>;
        })}
      </section>

      <section id="contact" className="home-contact-section"><div><h2>需要配置账号或部署环境？</h2><p>联系管理员获取登录账号、API 配置和本地部署支持。</p></div><V88Button variant="secondary" onClick={() => openProtectedPage('/settings')}>打开设置</V88Button></section>
    </div>
    </ConfigProvider>
  );
}
