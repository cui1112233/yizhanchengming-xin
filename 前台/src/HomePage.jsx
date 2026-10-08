import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Empty, Spin, Tag, Typography } from 'antd'
import { listWorkspaceRecent } from './api.js'
import WorkspaceIcon from './WorkspaceIcon.jsx'

const quickActions = [
  { href: '/script', icon: 'script', title: '新建剧本项目', desc: '导入小说，生成可编辑剧本与制作素材' },
  { href: '/novel-panel', icon: 'novel', title: '小说面板', desc: '分析人物、场景与分镜，保存完整项目' },
  { href: '/batch-factory', icon: 'batch', title: '批量工厂', desc: '多篇小说批量导演、审核并生成视频方案' },
  { href: '/shuihuo-production', icon: 'shuihuo', title: '水货生产', desc: '分段、提示词、素材和视频任务生产' },
  { href: '/agent', icon: 'agent', title: 'AI 智能 Agent', desc: 'Agent 将重新设计，当前不可用', status: '待重新设计' },
  { href: '/tts', icon: 'tts', title: '声音配音工坊', desc: '多音色情感合成，让你的画面声临其境' },
]

const kindLabels = {
  batch: '批量项目',
  intake: '小说获取',
  script: '剧本生成',
  novel_panel: '小说面板',
  tts: '配音',
  shuihuo_image: '水货图片',
  shuihuo_video: '水货视频',
  video: '视频生成',
}

const statusLabels = {
  pending: '待执行', queued: '排队中', running: '执行中', completed: '已完成', failed: '失败',
  partial_failed: '部分失败', retryable_failed: '可重试失败', saved: '已保存', measured: '已测量',
  ready: '已就绪', pending_executor: '等待执行器', executor_unavailable: '执行器不可用', cancelled: '已取消',
}

function RouteLink({ href, onNavigate, className, children, ...rest }) {
  const handleClick = (event) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(href)
  }
  return <a href={href} className={className} onClick={handleClick} {...rest}>{children}</a>
}

function safeRecentHref(value) {
  if (typeof value !== 'string') return ''
  const matched = value.match(/^\/batch-factory\?projectId=([1-9]\d*)$/)
  if (!matched) return ''
  const projectId = Number(matched[1])
  if (!Number.isSafeInteger(projectId)) return ''
  try {
    const base = new URL('http://workspace.local/')
    const target = new URL(value, base)
    if (target.origin !== base.origin || target.pathname !== '/batch-factory' || target.hash || [...target.searchParams].length !== 1) return ''
  } catch {
    return ''
  }
  return `/batch-factory?projectId=${projectId}`
}

function displayUpdatedAt(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '更新时间未知'
  return date.toLocaleString('zh-CN', { hour12: false })
}

function publicRecentError(error) {
  return {
    message: error?.status && error?.message ? error.message : '最近创作暂时无法读取，请稍后重试。',
    requestId: typeof error?.requestId === 'string' ? error.requestId : '',
  }
}

function RecentEntry({ item, onNavigate }) {
  const href = safeRecentHref(item?.href)
  const icon = item?.kind === 'tts' ? 'tts' : item?.kind === 'script' ? 'script' : item?.kind === 'novel_panel' ? 'novel' : item?.kind?.includes('shuihuo') ? 'shuihuo' : 'batch'
  const content = <><span className="recent-project-mark"><WorkspaceIcon name={icon} /></span><span className="recent-project-copy"><strong>{item?.title || '未命名项目'}</strong><small>{kindLabels[item?.kind] || '创作项目'} · {statusLabels[item?.status] || item?.status || '状态未知'}</small><time dateTime={item?.updatedAt || undefined}>{displayUpdatedAt(item?.updatedAt)}</time></span>{href ? <WorkspaceIcon name="arrow" className="recent-project-arrow" /> : <Tag>入口不可用</Tag>}</>
  if (!href) return <article className="recent-project-entry is-disabled">{content}</article>
  return <RouteLink href={href} onNavigate={onNavigate} className="recent-project-entry">{content}</RouteLink>
}

export default function HomePage({ onNavigate }) {
  const requestSequence = useRef(0)
  const mounted = useRef(true)
  const [recent, setRecent] = useState({ initialLoading: true, refreshing: false, error: null, items: [] })

  const loadRecent = useCallback(async () => {
    const requestID = ++requestSequence.current
    setRecent((value) => ({ ...value, initialLoading: value.items.length === 0, refreshing: value.items.length > 0, error: null }))
    try {
      const payload = await listWorkspaceRecent(6)
      if (!mounted.current || requestID !== requestSequence.current) return
      setRecent({ initialLoading: false, refreshing: false, error: null, items: Array.isArray(payload?.items) ? payload.items : [] })
    } catch (error) {
      if (!mounted.current || requestID !== requestSequence.current) return
      setRecent((value) => ({ ...value, initialLoading: false, refreshing: false, error: publicRecentError(error) }))
    }
  }, [])

  useEffect(() => {
    mounted.current = true
    void loadRecent()
    return () => { mounted.current = false; requestSequence.current += 1 }
  }, [loadRecent])

  return (
    <main className="home-page">
      <section className="home-video-hero" aria-labelledby="home-title">
        <video className="home-hero-video" aria-label="首页视觉背景" autoPlay muted loop playsInline poster="/assets/brand-logo-white.png" src="/assets/home-hero.mp4" />
        <div className="home-video-overlay" />
        <div className="home-hero-glow home-hero-glow-left" />
        <div className="home-hero-glow home-hero-glow-right" />
        <div className="home-hero-content">
          <p className="home-hero-kicker">NOVEL VISUAL SCRIPT STUDIO</p>
          <Typography.Title id="home-title" level={1} className="home-hero-title">让小说章节直接进入可视化剧本工作流</Typography.Title>
          <Typography.Paragraph className="home-hero-subtitle">从原文提取人物、场景、节奏和镜头结构，生成可继续编辑、导出和配音的短剧制作素材。</Typography.Paragraph>
          <div className="home-hero-actions"><Button type="primary" size="large" className="home-primary-action" onClick={() => onNavigate('/script')}>开始生成</Button><Button size="large" className="home-secondary-action" onClick={() => onNavigate('/tts')}>进入配音</Button></div>
        </div>
      </section>

      <section className="home-section" aria-labelledby="quick-actions-title">
        <div className="home-section-heading"><div><span className="home-section-eyebrow">CREATION ENTRANCES</span><Typography.Title level={2} id="quick-actions-title">创作工作台</Typography.Title></div><Typography.Text type="secondary">从一个入口开始，进入对应的创作流程</Typography.Text></div>
        <div className="home-quick-actions">
          {quickActions.map((action) => <RouteLink key={action.href} href={action.href} onNavigate={onNavigate} className="quick-action-card"><span className="quick-action-icon"><WorkspaceIcon name={action.icon} /></span><span className="quick-action-copy"><strong>{action.title}</strong>{action.status ? <Tag>{action.status}</Tag> : null}<span>{action.desc}</span></span><WorkspaceIcon name="arrow" className="quick-action-arrow" /></RouteLink>)}
        </div>
      </section>

      <section className="home-section home-recent-section" aria-labelledby="recent-title">
        <div className="home-section-heading"><div><span className="home-section-eyebrow">RECENT PROJECTS</span><Typography.Title level={2} id="recent-title">最近创作项目</Typography.Title></div><div className="home-recent-actions"><Button onClick={() => void loadRecent()} loading={recent.refreshing} aria-label="刷新最近项目">刷新</Button><Button onClick={() => onNavigate('/history')}>查看全部历史记录</Button></div></div>
        {recent.initialLoading ? <div className="recent-project-loading"><Spin size="small" /> <span>正在读取你的最近项目…</span></div> : null}
        {recent.error ? <Alert type="error" showIcon message={recent.error.message} description={recent.error.requestId ? `请求编号：${recent.error.requestId}` : undefined} action={<Button size="small" onClick={() => void loadRecent()}>重试</Button>} /> : null}
        {!recent.initialLoading && !recent.error && recent.items.length === 0 ? <Empty description="暂无可访问的最近项目" /> : null}
        {recent.items.length > 0 ? <div className="recent-project-grid">{recent.items.map((item) => <RecentEntry key={item.id} item={item} onNavigate={onNavigate} />)}</div> : null}
      </section>

      <section id="contact" className="home-contact" aria-labelledby="home-contact-title"><div><span className="home-section-eyebrow">WORKSPACE READY</span><Typography.Title level={2} id="home-contact-title">需要配置账号或部署环境？</Typography.Title><Typography.Paragraph>前往设置查看当前已选择配置与真实生效状态。</Typography.Paragraph></div><Button type="primary" className="gradient-btn" onClick={() => onNavigate('/settings')}>打开设置</Button></section>
    </main>
  )
}

export { quickActions, safeRecentHref }
