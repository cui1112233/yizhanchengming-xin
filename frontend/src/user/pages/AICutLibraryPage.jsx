import { formatWorkDateTime } from '../../shared/workDateTime.js';
import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Empty, Input, Modal, Select, Skeleton, message } from 'antd';
import { ArrowLeft, ArrowUpRight, Clapperboard, Film, FolderOpen, Plus, Search, LayoutGrid, List } from 'lucide-react';
import { createAICutProject, listAICutProjects, updateAICutProject } from '../../shared/api/aicut.js';
import { libraryProjects, projectEditorUrl } from './aicut/projectLibrary.js';
import './aicut/project-library.css';

export default function AICutLibraryPage() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [query, setQuery] = useState('');
  const [sort, setSort] = useState('recent');
  const [view, setView] = useState('grid');
  const [dialog, setDialog] = useState(null);
  const [title, setTitle] = useState('');
  const [ratio, setRatio] = useState('9:16');
  const [busy, setBusy] = useState(false);
  const visible = useMemo(() => libraryProjects(projects, query, sort), [projects, query, sort]);
  async function refresh() {
    setLoading(true); setError('');
    try { setProjects(await listAICutProjects()); }
    catch (failure) { setError(failure.message || '作品库读取失败'); }
    finally { setLoading(false); }
  }
  useEffect(() => { refresh(); }, []);
  function openCreate() { setTitle(''); setRatio('9:16'); setDialog({ type: 'create' }); }
  async function confirm() {
    if (!title.trim()) return message.warning('请输入工程名称');
    if (busy) return;
    setBusy(true);
    try {
      if (dialog.type === 'create') {
        const project = await createAICutProject({ title: title.trim(), aspectRatio: ratio });
        if (!project?.id) throw new Error('工程创建失败，请重试');
        window.location.assign(projectEditorUrl(project.id));
      } else {
        const project = await updateAICutProject(dialog.id, { title: title.trim() });
        setProjects(items => items.map(item => item.id === project.id ? project : item));
        setDialog(null); message.success('工程名称已更新');
      }
    } catch (failure) { message.error(failure.message || '操作失败'); }
    finally { setBusy(false); }
  }
  return <main className="aicut-library">
    <header className="aicut-library-top">
      <a href="/" className="aicut-library-back"><ArrowLeft size={16} />一战晟铭</a>
      <span className="aicut-library-wordmark"><Clapperboard size={17} /> AI-CUT <small>/ 作品库</small></span>
    </header>
    <section className="aicut-library-hero">
      <div><h1>我的作品 <small>{projects.length} 个工程</small></h1></div>
      <Button type="primary" icon={<Plus size={15} />} onClick={openCreate}>新建工程</Button>
    </section>
    <section aria-label="剪辑工程" className="aicut-library-content">
      <div className="aicut-library-toolbar"><h2><FolderOpen size={15} />全部工程</h2><div>
        <Input allowClear prefix={<Search size={16} />} placeholder="搜索工程名称" aria-label="搜索工程名称" value={query} onChange={event => setQuery(event.target.value)} />
        <Select aria-label="工程排序" value={sort} onChange={setSort} options={[{ value: 'recent', label: '最近修改' }, { value: 'name', label: '名称排序' }]} />
        <div className="aicut-library-views" aria-label="显示方式"><button type="button" aria-label="网格视图" aria-pressed={view === 'grid'} onClick={() => setView('grid')}><LayoutGrid size={15} /></button><button type="button" aria-label="列表视图" aria-pressed={view === 'list'} onClick={() => setView('list')}><List size={16} /></button></div>
      </div></div>
      {error ? <Alert type="error" showIcon message={error} action={<Button onClick={refresh}>重新加载</Button>} /> : loading ? <Skeleton active paragraph={{ rows: 6 }} /> : <>
        {!visible.length ? <Empty description={query ? '没有找到这个工程，换个名称试试' : '还没有工程，创建你的第一个作品吧'}>{!query && <Button onClick={openCreate}>新建工程</Button>}</Empty> : <div className={`aicut-library-grid${view === 'list' ? ' is-list' : ''}`}>
          {visible.map((project, index) => <article className="aicut-library-card" key={project.id} style={{ '--project-hue': `${[198, 268, 28, 162][index % 4]}` }}>
            <a href={projectEditorUrl(project.id)} className="aicut-library-cover" aria-label={`打开工程：${project.title}`}>
              <div className="aicut-library-frame" style={{ aspectRatio: (project.aspectRatio || '9:16').replace(':', '/') }}><Film size={18} /></div>
              <span className="aicut-library-ratio">{project.aspectRatio}</span><span className="aicut-library-open">继续剪辑 <ArrowUpRight size={16} /></span>
            </a>
            <div className="aicut-library-card-info"><a href={projectEditorUrl(project.id)} title={project.title}>{project.title}</a><button type="button" aria-label={`重命名：${project.title}`} onClick={() => { setTitle(project.title); setDialog({ type: 'rename', id: project.id }); }}>重命名</button></div>
            <p>修改于 {formatWorkDateTime(project.updatedAt || project.createdAt)}</p>
          </article>)}
        </div>}
      </>}
    </section>
    <Modal title={dialog?.type === 'rename' ? '重命名工程' : '新建剪辑工程'} open={Boolean(dialog)} onCancel={() => !busy && setDialog(null)} onOk={confirm} confirmLoading={busy} okText={dialog?.type === 'rename' ? '保存名称' : '创建并进入剪辑'} cancelText="取消">
      <label className="aicut-library-field">工程名称<Input autoFocus maxLength={120} value={title} placeholder="给你的故事起个名字" onChange={event => setTitle(event.target.value)} onPressEnter={confirm} /></label>
      {dialog?.type === 'create' && <label className="aicut-library-field">画面比例<Select value={ratio} onChange={setRatio} options={[{ value: '9:16', label: '9:16 · 竖屏短片' }, { value: '16:9', label: '16:9 · 横屏视频' }, { value: '1:1', label: '1:1 · 方形画面' }]} /></label>}
    </Modal>
  </main>;
}
