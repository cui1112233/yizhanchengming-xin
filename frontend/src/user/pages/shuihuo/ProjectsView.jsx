import { AppstoreOutlined, ClockCircleOutlined, CloseOutlined, DeleteOutlined, FileTextOutlined, FolderOpenOutlined, PlusOutlined, SearchOutlined, UnorderedListOutlined, UploadOutlined, UserOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { Alert, Button, Input, Modal, Popconfirm, Select, Upload, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { importProject } from '../../../shared/api/shuihuoProduction';
import { getBatchFactoryMergeCoverBlob, getProductionStatus, isBatchFactoryMergeCoverURL } from '../../../shared/api/batchFactory';
import { ProductionMediaBoundary } from '../batch-factory-v11/ProductionMediaBoundary';
import { BatchFactoryCreateModal } from './BatchFactoryCreateModal';
import { batchFactoryFallbackCoverFromProductionStatus, createCoverHydrationScheduler, isBatchFactoryV11Project } from './batchFactoryProjects';

const scheduleLegacyProjectCoverFallback = createCoverHydrationScheduler();

function readAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('读取文件失败'));
    reader.onload = () => resolve(String(reader.result || ''));
    reader.readAsDataURL(file);
  });
}

function formatTime(value) {
  if (!value) return '尚未开始制作';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '已创建' : date.toLocaleDateString('zh-CN');
}

function AuthenticatedProjectCover({ src, alt, batchId }) {
  const imageRef = useRef(null);
  const protectedCover = isBatchFactoryMergeCoverURL(src);
  const [visible, setVisible] = useState(!protectedCover);
  const [blobURL, setBlobURL] = useState('');
  const [fallbackVideoURL, setFallbackVideoURL] = useState('');
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    setVisible(!protectedCover);
    setBlobURL('');
    setFallbackVideoURL('');
    setFailed(false);
    if (!protectedCover || typeof IntersectionObserver === 'undefined') {
      setVisible(true);
      return undefined;
    }
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) {
        setVisible(true);
        observer.disconnect();
      }
    }, { rootMargin: '240px' });
    if (imageRef.current) observer.observe(imageRef.current);
    return () => observer.disconnect();
  }, [protectedCover, src]);

  useEffect(() => {
    if (!protectedCover || !visible) return undefined;
    let active = true;
    let objectURL = '';
    const recoverLegacyCover = () => scheduleLegacyProjectCoverFallback(async () => {
      if (!active || !batchId) return null;
      try {
        return batchFactoryFallbackCoverFromProductionStatus(await getProductionStatus(batchId, { silent: true, suppressGlobalError: true }));
      } catch (_) {
        return null;
      }
    }).then(media => {
      if (!active) return;
      if (media?.kind === 'video' && media.url) setFallbackVideoURL(media.url);
      else setFailed(true);
    });
    getBatchFactoryMergeCoverBlob(src)
      .then(blob => {
        objectURL = URL.createObjectURL(blob);
        if (active) setBlobURL(objectURL);
      })
      .catch(recoverLegacyCover);
    return () => {
      active = false;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [protectedCover, src, visible]);

  if (fallbackVideoURL) return <ProductionMediaBoundary showDownload={false}><video className="shuihuo-project-card-cover-media" src={fallbackVideoURL} muted playsInline preload="metadata" onLoadedMetadata={event => { event.currentTarget.currentTime = 0.001; }} onError={() => { setFallbackVideoURL(''); setFailed(true); }} aria-label={alt} /></ProductionMediaBoundary>;
  if (failed) return <span className="shuihuo-project-card-cover-fallback"><strong>批量工厂</strong><small>历史成片已归档</small></span>;
  return <img ref={imageRef} className="shuihuo-project-card-cover-media" src={protectedCover ? (blobURL || undefined) : src} alt={alt} loading="lazy" onError={() => { if (protectedCover) setFailed(true); }} />;
}

export function ProjectsView({ projects, loadError, health, onCreate, onImported, onCreateBatch, onOpen, onDelete, onRefresh, openCreateOnLoad = false }) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');
  const [viewMode, setViewMode] = useState('grid');
  const [name, setName] = useState('');
  const [sourceText, setSourceText] = useState('');
  const [file, setFile] = useState(null);
  const [segmentationMode, setSegmentationMode] = useState('paragraph');
  const [busy, setBusy] = useState(false);
  const [batchOpen, setBatchOpen] = useState(false);

  useEffect(() => {
    if (openCreateOnLoad) setBatchOpen(true);
  }, [openCreateOnLoad]);

  const hasSource = Boolean(file || sourceText.trim());
  const smartReady = Boolean(health?.database?.ready && health?.enabledModelKinds?.includes('text'));

  function reset() {
    setName('');
    setSourceText('');
    setFile(null);
    setSegmentationMode('paragraph');
  }

  function selectSourceFile(nextFile) {
    setFile(nextFile);
    setSourceText('');
    return false;
  }

  function pasteSourceText() {
    setFile(null);
  }

  async function submit() {
    const normalizedName = name.trim();
    if (!normalizedName) { message.warning('请填写作品名称'); return; }
    if (!sourceText.trim() && !file) { message.warning('请粘贴原文或选择 TXT、SRT、DOCX 文件'); return; }
    setBusy(true);
    try {
      if (file) {
        const dataUrl = await readAsDataURL(file);
        const result = await importProject({ name: normalizedName, filename: file.name, dataUrl });
        await onImported(result, segmentationMode);
      } else {
        await onCreate({ name: normalizedName, sourceText: sourceText.trim(), segmentationMode });
      }
      setOpen(false);
      reset();
    } catch (error) {
      message.error(error.message || '创建项目失败');
    } finally {
      setBusy(false);
    }
  }

  const visibleProjects = useMemo(() => {
    const keyword = search.trim().toLocaleLowerCase();
    return [...(projects || [])]
      .filter(project => !keyword || project.name.toLocaleLowerCase().includes(keyword))
      .sort((left, right) => {
        const leftTime = new Date(left.updatedAt || left.createdAt || 0).getTime();
        const rightTime = new Date(right.updatedAt || right.createdAt || 0).getTime();
        return sortOrder === 'desc' ? rightTime - leftTime : leftTime - rightTime;
      });
  }, [projects, search, sortOrder]);

  function projectCount(project) {
    if (Number.isFinite(project.segmentCount)) return `${project.segmentCount} 个分镜`;
    return project.segmentationStatus === 'confirmed' ? '已确认分镜' : '待生成分镜';
  }

  function projectCover(project) {
    const media = project?.coverMedia;
    if (!media?.url) return null;
    return media.kind === 'video'
      ? <ProductionMediaBoundary showDownload={false}><video className="shuihuo-project-card-cover-media" src={media.url} muted playsInline preload="auto" onLoadedMetadata={event => { event.currentTarget.currentTime = 0.001; }} aria-label={`${project.name} 封面视频`} /></ProductionMediaBoundary>
      : <AuthenticatedProjectCover src={media.url} alt={`${project.name} 封面`} batchId={project.batchId} />;
  }

  return <section className="shuihuo-project-library">
    <div className="shuihuo-project-library-heading">
      <div className="shuihuo-project-library-title"><h1>漫剧解说</h1><p>管理和创建您的漫剧解说作品</p></div>
      <div className="shuihuo-create-actions"><Button className="shuihuo-create-project" type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>创作漫剧</Button><Button className="shuihuo-create-batch" icon={<ThunderboltOutlined />} onClick={() => setBatchOpen(true)}>批量工厂</Button></div>
    </div>
    <div className="shuihuo-project-library-section-title"><UserOutlined /> <strong>个人作品</strong></div>
    {loadError ? <Alert className="shuihuo-project-load-error" type="warning" showIcon message="作品读取暂时失败" description={loadError} action={<Button size="small" onClick={() => onRefresh?.()}>重试</Button>} /> : null}
    <div className="shuihuo-project-library-toolbar">
      <Input className="shuihuo-project-search" prefix={<SearchOutlined />} value={search} onChange={event => setSearch(event.target.value)} placeholder="搜索作品..." allowClear />
      <span className="shuihuo-project-filter-label">全部作品</span>
      <button className="shuihuo-project-filter" type="button" onClick={() => setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc')}><ClockCircleOutlined /><span>{sortOrder === 'desc' ? '按时间降序' : '按时间升序'}</span></button>
      <Button className="shuihuo-project-view-toggle" type="text" icon={viewMode === 'grid' ? <UnorderedListOutlined /> : <AppstoreOutlined />} onClick={() => setViewMode(viewMode === 'grid' ? 'list' : 'grid')} aria-label={viewMode === 'grid' ? '切换为列表视图' : '切换为网格视图'} title={viewMode === 'grid' ? '切换为列表视图' : '切换为网格视图'} />
    </div>
    <div className={`shuihuo-project-grid${viewMode === 'list' ? ' is-list' : ''}`}>
      {visibleProjects.map(project => {
        const cover = projectCover(project);
        return <article className={`shuihuo-project-card${cover ? ' has-cover' : ''}`} key={project.id}>
        <button type="button" className="shuihuo-project-card-open" onClick={() => onOpen(project)} aria-label={`打开${project.name}`}>
          <div className="shuihuo-project-card-cover">{cover || <><span className={`shuihuo-project-card-mode ${isBatchFactoryV11Project(project) ? 'is-batch-factory' : ''}`}>{isBatchFactoryV11Project(project) ? '批量工厂' : '水货生产'}</span><span>{isBatchFactoryV11Project(project) ? '小说列表' : projectCount(project)}</span></>}</div>
          <div className="shuihuo-project-card-meta"><strong>{project.name}</strong><span>{formatTime(project.updatedAt || project.createdAt)}</span></div>
        </button>
        <div className="shuihuo-project-card-actions"><Button type="text" icon={<FolderOpenOutlined />} onClick={() => onOpen(project)} aria-label={`打开工作台 ${project.name}`} /><Popconfirm title={isBatchFactoryV11Project(project) ? '删除批量项目？' : '删除项目？此操作不会撤销。'} description={isBatchFactoryV11Project(project) ? '将删除其中全部小说和本地生产记录；不会影响已提交到 121 的内容。' : undefined} okText="删除" cancelText="取消" onConfirm={() => onDelete(project)}><Button type="text" danger icon={<DeleteOutlined />} aria-label={`删除${project.name}`} /></Popconfirm></div>
      </article>;
      })}
      {!visibleProjects.length && !loadError ? <div className="shuihuo-empty"><FileTextOutlined /><p>{projects?.length ? '没有匹配的作品' : '还没有项目'}</p><Button onClick={() => setOpen(true)}>从原文或字幕开始</Button><Button className="shuihuo-create-batch" onClick={() => setBatchOpen(true)}>批量工厂</Button></div> : null}
    </div>
    <Modal className="shuihuo-create-project-modal" title="新建漫剧" open={open} onCancel={() => { setOpen(false); reset(); }} footer={<><Button onClick={() => { setOpen(false); reset(); }}>取消</Button><Button type="primary" loading={busy} onClick={submit}>确定创建</Button></>} width={500}>
      <label className="shuihuo-form-label" htmlFor="shuihuo-project-name">作品名称 <em>*</em></label>
      <Input id="shuihuo-project-name" value={name} onChange={event => setName(event.target.value)} placeholder="请输入作品名称" maxLength={255} />
      <label className="shuihuo-form-label" htmlFor="shuihuo-project-source">字幕内容 <em>*</em></label>
      <div className={`shuihuo-source-composer${file || sourceText ? ' has-value' : ''}`}>
        {file ? <div className="shuihuo-source-file"><FileTextOutlined /><strong>{file.name}</strong><Button type="text" danger icon={<CloseOutlined />} aria-label="移除字幕文件" onClick={() => setFile(null)} /></div> : <Input.TextArea id="shuihuo-project-source" value={sourceText} onPaste={pasteSourceText} onChange={event => { setSourceText(event.target.value); if (event.target.value) setFile(null); }} rows={sourceText ? 5 : 4} placeholder="直接粘贴字幕内容" />}
        {!file ? <Upload accept=".txt,.srt,.docx,text/plain,application/x-subrip,application/vnd.openxmlformats-officedocument.wordprocessingml.document" showUploadList={false} maxCount={1} beforeUpload={selectSourceFile}><Button className="shuihuo-source-file-trigger" type="text" icon={<UploadOutlined />}>点击选择字幕文件 (SRT/TXT/DOCX)</Button></Upload> : null}
        {!file && !sourceText ? <span className="shuihuo-source-composer-hint">也可直接在此输入或粘贴内容</span> : null}
      </div>
      <p className="shuihuo-modal-note">文件与文本二选一；支持 TXT、SRT、DOCX，文件会保留原始内容并解析正文。</p>
      {hasSource ? <><label className="shuihuo-form-label" htmlFor="shuihuo-segmentation-mode">分段方式</label><Select id="shuihuo-segmentation-mode" value={segmentationMode} onChange={setSegmentationMode} options={[{ value: 'paragraph', label: '自动识别' }, { value: 'smart', label: '智能识别', disabled: !smartReady }]} /></> : null}
    </Modal>
    <BatchFactoryCreateModal open={batchOpen} onCancel={() => setBatchOpen(false)} onCreated={async payload => { await onCreateBatch(payload); setBatchOpen(false); await onRefresh?.(); }} />
  </section>;
}
