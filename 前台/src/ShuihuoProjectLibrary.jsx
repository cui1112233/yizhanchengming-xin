import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AppstoreOutlined, ClockCircleOutlined, FolderOpenOutlined, InboxOutlined, PlusOutlined, SearchOutlined, ThunderboltOutlined, UserOutlined } from '@ant-design/icons'
import { Alert, Button, Empty, Input, Popconfirm, Spin } from 'antd'
import { archiveBatchProject, listBatchProjects } from './api.js'
import { batchError, batchUpdatedAt } from './batchFactoryPresentation.js'
import './shuihuo-project-library.css'

// Ported from the V88 Shuihuo ProjectsView layout.  The V88 Node calls are
// deliberately replaced by the existing Go-owned BatchProject projection.
function projectDetail(project) {
  const count = Number(project?.bookCount || 0)
  if (count > 0) return `${count} 本小说`
  if (project?.runStatus === 'completed') return '已完成生产'
  if (project?.runStatus === 'running') return '正在生产'
  return '待开始生产'
}

export default function ShuihuoProjectLibrary({ onOpen, onCreate, onBatch }) {
  const [projects, setProjects] = useState([])
  const [query, setQuery] = useState('')
  const [sortOrder, setSortOrder] = useState('desc')
  const [loading, setLoading] = useState(true)
  const [refreshKey, setRefreshKey] = useState(0)
  const [error, setError] = useState(null)
  const [archivingID, setArchivingID] = useState(null)
  const requestRef = useRef(0)

  const refresh = useCallback(() => setRefreshKey((value) => value + 1), [])

  useEffect(() => {
    const request = ++requestRef.current
    setLoading(true)
    setError(null)
    listBatchProjects({ archived: 'active', limit: 100, sort: 'updated_desc' })
      .then((payload) => {
        if (request !== requestRef.current) return
        setProjects(Array.isArray(payload?.projects) ? payload.projects : [])
      })
      .catch((reason) => {
        if (request !== requestRef.current) return
        setProjects([])
        setError(batchError(reason, '读取作品列表失败，请稍后重试。'))
      })
      .finally(() => { if (request === requestRef.current) setLoading(false) })
    return () => { requestRef.current += 1 }
  }, [refreshKey])

  const visibleProjects = useMemo(() => {
    const keyword = query.trim().toLocaleLowerCase()
    return [...projects]
      .filter((project) => !keyword || String(project.name || '').toLocaleLowerCase().includes(keyword))
      .sort((left, right) => {
        const leftTime = new Date(left.updatedAt || left.createdAt || 0).getTime()
        const rightTime = new Date(right.updatedAt || right.createdAt || 0).getTime()
        return sortOrder === 'desc' ? rightTime - leftTime : leftTime - rightTime
      })
  }, [projects, query, sortOrder])

  const archive = async (project) => {
    setArchivingID(project.id)
    setError(null)
    try {
      await archiveBatchProject(project.id)
      refresh()
    } catch (reason) {
      setError(batchError(reason, '归档作品失败，请稍后重试。'))
    } finally {
      setArchivingID(null)
    }
  }

  return <main className="shuihuo-library-page">
    <section className="shuihuo-project-library" aria-label="漫剧解说项目库">
      <div className="shuihuo-project-library-heading">
        <div className="shuihuo-project-library-title"><h1>漫剧解说</h1><p>管理和创建您的漫剧解说作品</p></div>
        <div className="shuihuo-create-actions">
          <Button className="shuihuo-create-project" type="primary" icon={<PlusOutlined />} onClick={onCreate}>创作漫剧</Button>
          <Button className="shuihuo-create-batch" icon={<ThunderboltOutlined />} onClick={onBatch}>批量工厂</Button>
        </div>
      </div>
      <div className="shuihuo-project-library-section-title"><UserOutlined /> <strong>个人作品</strong></div>
      <div className="shuihuo-project-library-toolbar">
        <Input className="shuihuo-project-search" prefix={<SearchOutlined />} value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索作品..." aria-label="搜索作品" allowClear />
        <button className="shuihuo-project-filter" type="button" onClick={() => setSortOrder((value) => value === 'desc' ? 'asc' : 'desc')}><ClockCircleOutlined /><span>{sortOrder === 'desc' ? '按时间降序' : '按时间升序'}</span></button>
        <Button className="shuihuo-project-view-toggle" type="text" icon={<AppstoreOutlined />} aria-label="网格视图" title="网格视图" />
      </div>
      {error ? <Alert className="shuihuo-project-library-alert" type="error" showIcon message={error.message} description={error.requestId ? `请求编号：${error.requestId}` : undefined} action={<Button size="small" onClick={refresh}>重试</Button>} /> : null}
      <Spin spinning={loading}>
        <div className="shuihuo-project-grid">
          {visibleProjects.map((project) => <article className="shuihuo-project-card" key={project.id}>
            <button type="button" className="shuihuo-project-card-open" onClick={() => onOpen(project)} aria-label={`打开${project.name || '未命名作品'}`}>
              <div className="shuihuo-project-card-cover"><span className="shuihuo-project-card-mode">水货生产</span><span>{projectDetail(project)}</span></div>
              <div className="shuihuo-project-card-meta"><strong>{project.name || '未命名作品'}</strong><span>{batchUpdatedAt(project.updatedAt || project.createdAt)}</span></div>
            </button>
            <div className="shuihuo-project-card-actions"><Button type="text" icon={<FolderOpenOutlined />} onClick={() => onOpen(project)} aria-label={`打开工作台 ${project.name || '未命名作品'}`} /><Popconfirm title="归档作品？" description="小说与生产记录会保留；恢复前项目将只读。" okText="归档" cancelText="取消" onConfirm={() => void archive(project)}><Button type="text" loading={archivingID === project.id} danger icon={<InboxOutlined />} aria-label={`归档${project.name || '未命名作品'}`} /></Popconfirm></div>
          </article>)}
          {!loading && visibleProjects.length === 0 ? <div className="shuihuo-empty">{projects.length ? <><SearchOutlined /><p>没有匹配的作品</p></> : <><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有项目" /><Button onClick={onCreate}>从小说获取开始</Button><Button className="shuihuo-create-batch" onClick={onBatch}>批量工厂</Button></>}</div> : null}
        </div>
      </Spin>
    </section>
  </main>
}
