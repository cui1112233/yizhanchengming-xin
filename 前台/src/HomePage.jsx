import React, { useEffect, useState } from 'react'
import { Alert, Button, Empty, Spin, Tag, Typography } from 'antd'
import { listBatchProjects } from './api.js'

const quickActions = [
  { href: '/script', icon: '✎', title: '新建剧本项目', desc: '导入小说，生成可编辑剧本与制作素材' },
  { href: '/novel-panel', icon: '▦', title: '小说面板', desc: '分析人物、场景与分镜，保存完整项目' },
  { href: '/batch-factory', icon: '↯', title: '批量工厂', desc: '多篇小说批量导演、审核并生成视频方案' },
  { href: '/shuihuo-production', icon: '▶', title: '水货生产', desc: '分段、提示词、素材和视频任务生产' },
  { href: '/agent', icon: '◇', title: 'AI 智能 Agent', desc: 'Agent 将重新设计，当前不可用', status: '待重新设计' },
  { href: '/tts', icon: '♫', title: '声音配音工坊', desc: '多音色情感合成，让你的画面声临其境' },
]

function RouteLink({ href, onNavigate, className, children, ...rest }) {
  const handleClick = (event) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(href)
  }
  return <a href={href} className={className} onClick={handleClick} {...rest}>{children}</a>
}

export default function HomePage({ onNavigate }) {
  const [recent, setRecent] = useState({ loading: true, error: '', projects: [] })
  const loadRecent = async () => {
    setRecent((value) => ({ ...value, loading: true, error: '' }))
    try {
      const payload = await listBatchProjects()
      setRecent({ loading: false, error: '', projects: (payload.projects || []).slice(0, 6) })
    } catch (error) {
      setRecent((value) => ({ ...value, loading: false, error: error?.message || '读取最近项目失败' }))
    }
  }
  useEffect(() => { void loadRecent() }, [])
  return (
    <main className="home-page">
      <section className="home-video-hero" aria-labelledby="home-title">
        <video className="home-hero-video" aria-label="首页视觉背景" autoPlay muted loop playsInline poster="/assets/brand-logo-white.png" src="/assets/home-hero.mp4" />
        <div className="home-video-overlay" />
        <div className="home-hero-glow home-hero-glow-left" />
        <div className="home-hero-glow home-hero-glow-right" />
        <div className="home-hero-content">
          <p className="home-hero-kicker">NOVEL VISUAL SCRIPT STUDIO</p>
          <Typography.Title id="home-title" level={1} className="home-hero-title">
            让小说章节直接进入可视化剧本工作流
          </Typography.Title>
          <Typography.Paragraph className="home-hero-subtitle">
            从原文提取人物、场景、节奏和镜头结构，生成可继续编辑、导出和配音的短剧制作素材。
          </Typography.Paragraph>
          <div className="home-hero-actions">
            <Button type="primary" size="large" className="home-primary-action" onClick={() => onNavigate('/script')}>开始生成</Button>
            <Button size="large" className="home-secondary-action" onClick={() => onNavigate('/tts')}>进入配音</Button>
          </div>
        </div>
      </section>

      <section className="home-section" aria-labelledby="quick-actions-title">
        <div className="home-section-heading">
          <div>
            <span className="home-section-eyebrow">CREATION ENTRANCES</span>
            <Typography.Title level={2} id="quick-actions-title">创作工作台</Typography.Title>
          </div>
          <Typography.Text type="secondary">从一个入口开始，进入对应的创作流程</Typography.Text>
        </div>
        <div className="home-quick-actions">
          {quickActions.map((action) => (
            <RouteLink key={action.href} href={action.href} onNavigate={onNavigate} className="quick-action-card">
              <span className="quick-action-icon" aria-hidden="true">{action.icon}</span>
              <span className="quick-action-copy"><strong>{action.title}</strong>{action.status ? <Tag>{action.status}</Tag> : null}<span>{action.desc}</span></span>
              <span className="quick-action-arrow" aria-hidden="true">→</span>
            </RouteLink>
          ))}
        </div>
      </section>

      <section className="home-section home-recent-section" aria-labelledby="recent-title">
        <div className="home-section-heading">
          <div>
            <span className="home-section-eyebrow">RECENT PROJECTS</span>
            <Typography.Title level={2} id="recent-title">最近创作项目</Typography.Title>
          </div>
          <div className="home-recent-actions"><Button onClick={() => void loadRecent()} loading={recent.loading} aria-label="刷新最近项目">刷新</Button><Button onClick={() => onNavigate('/history')}>查看全部历史记录</Button></div>
        </div>
        {recent.loading ? <div className="recent-project-entry"><Spin size="small" /> <span>正在读取你的最近项目…</span></div> : null}
        {recent.error ? <Alert type="error" showIcon message={recent.error} action={<Button size="small" onClick={() => void loadRecent()}>重试</Button>} /> : null}
        {!recent.loading && !recent.error && recent.projects.length === 0 ? <Empty description="暂无可访问的最近项目" /> : null}
        {!recent.loading && !recent.error ? recent.projects.map((project) => <RouteLink key={project.id} href={`/shuihuo-production?projectId=${project.id}`} onNavigate={onNavigate} className="recent-project-entry"><span className="recent-project-mark" aria-hidden="true">◫</span><span><strong>{project.name || `项目 #${project.id}`}</strong><small>{project.bookCount || 0} 本小说 · {project.runStatus || '尚未执行'}</small></span><span aria-hidden="true">→</span></RouteLink>) : null}
      </section>
    </main>
  )
}

export { quickActions }
