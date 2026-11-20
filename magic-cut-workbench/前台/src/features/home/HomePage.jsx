import { Alert, Button, Empty, Spin } from 'antd'
import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { CreateProjectDialog } from '../project-create/CreateProjectDialog.jsx'
import { createHomeModel } from './homeModel.js'

export function HomePage() {
  const copy = createHomeModel()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [projects, setProjects] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    apiRequest('/api/magic-cut/projects')
      .then(result => setProjects(result.projects || []))
      .catch(requestError => setError(requestError.message || '工程列表读取失败'))
      .finally(() => setLoading(false))
  }, [])

  function projectCreated(project) {
    setProjects(current => [project, ...current.filter(item => item.id !== project.id)])
    setDialogOpen(false)
    window.history.pushState({}, '', `/projects/${project.id}/edit`)
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  return (
    <main className="magic-home">
      <header className="magic-home__header">
        <div>
          <p className="magic-home__eyebrow">MAGIC CUT WORKSPACE</p>
          <h1>{copy.sectionTitle}</h1>
        </div>
        <Button type="primary" size="large" icon={<Plus size={17} />} onClick={() => setDialogOpen(true)}>
          {copy.primaryAction}
        </Button>
      </header>
      {error ? <Alert className="project-list-error" type="warning" showIcon message={error} /> : null}
      {loading ? <section className="magic-home__empty"><Spin size="large" /></section> : projects.length ? (
        <section className="project-grid" aria-label={copy.sectionTitle}>
          {projects.map(project => (
            <button key={project.id} className="project-card" type="button" onClick={() => projectCreated(project)}>
              <span className="project-card__preview">{project.width} × {project.height}</span>
              <b>{project.name}</b>
              <small>{project.frameRate} fps</small>
            </button>
          ))}
        </section>
      ) : (
        <section className="magic-home__empty" aria-label={copy.sectionTitle}>
          <Empty description={copy.emptyTitle} />
          <p>{copy.emptyDescription}</p>
        </section>
      )}
      <CreateProjectDialog open={dialogOpen} onCancel={() => setDialogOpen(false)} onCreated={projectCreated} />
    </main>
  )
}
