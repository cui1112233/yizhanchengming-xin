import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Modal, Spin, message } from 'antd';
import { CommentaryWorkbench } from './shuihuo/CommentaryWorkbench';
import { BatchFactoryNovelList } from './shuihuo/BatchFactoryNovelList';
import { batchFactoryBatchFromResponse, batchFactoryProjectsFrom, isBatchFactoryV11Project } from './shuihuo/batchFactoryProjects';
import { projectLibraryRefreshResult } from './shuihuo/projectLibraryLoadState';
import { ProjectsView } from './shuihuo/ProjectsView';
import { AssetsView } from './shuihuo/AssetsView';
import { confirmSegmentation, createProject, deleteProject, getProductionHealth, getProject, listModels, listProjects, paragraphSegmentation, replaceProjectSource, smartSegmentation } from '../../shared/api/shuihuoProduction';
import { appendNovelFetchIntake, classifyFetchedBatchMetadata, createBatchFromIntake, createManualIntake, deleteBatchFactoryProject, getBatch, getIntake, listBatchSummaries, startBatchAutomation } from '../../shared/api/batchFactory';
import { BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, novelFetchIntakeBooks, pendingNovelFetchIntakeId } from './shuihuo/batchFactoryNovelFetchHandoff';
import { useLocalExecutorStatus } from '../../shared/hooks/useLocalExecutorStatus';
import { batchExecutorKinds } from './shuihuo/batchFactoryLocalExecutor';
import './shuihuo-production.css';

const modelNames = { text: '文本模型', image: '图片模型', video: '视频模型', audio: '配音模型' };
function readinessItems(health, executorKinds = [], executorOnline = false) {
  const enabled = new Set(health?.enabledModelKinds || []);
  const executorKindSet = new Set(executorKinds);
  return [['数据库', health?.database], ['Redis', health?.redis], ['存储', health?.storage], ...Object.entries(modelNames).map(([kind, name]) => [name, executorKindSet.has(kind) ? { ready: executorOnline, executor: true, reason: executorOnline ? '一战晟铭执行器在线' : '一战晟铭执行器离线' } : { ready: enabled.has(kind), reason: enabled.has(kind) ? '' : `缺少已启用的${name}` }])];
}

export function ShuihuoProductionPage({ openBatchOnLoad = false }) {
  const [projects, setProjects] = useState([]);
  const [activeProject, setActiveProject] = useState(null);
  const [activeBatchProject, setActiveBatchProject] = useState(null);
  const [importNotice, setImportNotice] = useState(null);
  const [view, setView] = useState('projects');
  const [loading, setLoading] = useState(true);
  const [assetsOpen, setAssetsOpen] = useState(false);
  const [health, setHealth] = useState(null);
  const [healthError, setHealthError] = useState('');
  const [projectsLoadError, setProjectsLoadError] = useState('');
  const mountedRef = useRef(true);
  const projectRequestRef = useRef(0);
  const refreshRequestRef = useRef(0);
  const projectsRef = useRef([]);
  const novelFetchHandoffRef = useRef('');
  const requestedBatchIdRef = useRef(typeof window !== 'undefined'
    ? new URLSearchParams(window.location.search).get('batchId') || ''
    : '');
  const requestedBatchId = requestedBatchIdRef.current;
  const activeBatchExecutorKinds = useMemo(() => batchExecutorKinds(activeBatchProject), [activeBatchProject]);
  const { status: localExecutorStatus } = useLocalExecutorStatus({ enabled: activeBatchExecutorKinds.length > 0, intervalMs: 5000 });
  projectsRef.current = projects;

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      projectRequestRef.current += 1;
      refreshRequestRef.current += 1;
    };
  }, []);

  useEffect(() => {
    document.body.classList.add('shuihuo-theme-active');
    document.body.classList.add('shuihuo-reference-shell');
    return () => {
      document.body.classList.remove('shuihuo-theme-active');
      document.body.classList.remove('shuihuo-reference-shell');
    };
  }, []);

  const refreshProjects = useCallback(async () => {
    const requestId = refreshRequestRef.current + 1;
    refreshRequestRef.current = requestId;
    setLoading(true);
    try {
      const [water, batch] = await Promise.allSettled([
        listProjects({ silent: true }),
        listBatchSummaries()
      ]);

      if (batch.status === 'rejected') {
        if (requestId === refreshRequestRef.current) {
          console.error('[共享作品库] 批量工厂读取失败', batch.reason);
        }
      }

      if (water.status === 'rejected') {
        console.error('[共享作品库] 漫剧作品读取失败', water.reason);
      }

      const refreshResult = projectLibraryRefreshResult({
        previousProjects: projectsRef.current,
        water,
        batch,
        batchProjectsFrom: batchFactoryProjectsFrom
      });
      if (!mountedRef.current || requestId !== refreshRequestRef.current) return;
      if (refreshResult.error) {
        setProjectsLoadError(refreshResult.error);
        setLoading(false);
        return;
      }

      projectsRef.current = refreshResult.projects;
      setProjects(refreshResult.projects);
      setProjectsLoadError('');
      setLoading(false);
    } catch (error) {
      if (mountedRef.current && requestId === refreshRequestRef.current) {
        message.error(error.message || '读取项目库失败');
        setLoading(false);
      }
    }
  }, []);
  useEffect(() => { refreshProjects(); }, [refreshProjects]);
  useEffect(() => {
    if (!requestedBatchId || pendingNovelFetchIntakeId(window.location.search)) return undefined;
    const requestId = projectRequestRef.current + 1;
    projectRequestRef.current = requestId;
    getBatch(requestedBatchId).then(result => {
      if (!mountedRef.current || requestId !== projectRequestRef.current) return;
      const batch = batchFactoryBatchFromResponse(result);
      localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
      setActiveBatchProject(batch);
      setView('batch-novels');
    }).catch(error => {
      if (mountedRef.current && requestId === projectRequestRef.current) {
        message.error(error.message || '读取批量工程失败');
      }
    });
    return () => {
      if (requestId === projectRequestRef.current) projectRequestRef.current += 1;
    };
  }, [requestedBatchId]);
  useEffect(() => {
    let active = true;
    getProductionHealth({ silent: true }).then(result => { if (active) { setHealth(result); setHealthError(''); } }).catch(error => { if (active) setHealthError(error.message || '无法读取运行依赖状态'); });
    return () => { active = false; };
  }, []);

  const openProject = useCallback(async project => {
    const requestId = projectRequestRef.current + 1;
    projectRequestRef.current = requestId;
    try {
      if (isBatchFactoryV11Project(project)) {
        const batch = batchFactoryBatchFromResponse(await getBatch(project.batchId));
        if (!mountedRef.current || requestId !== projectRequestRef.current) return;
        localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
        window.history.replaceState({}, '', `/batch-factory?batchId=${encodeURIComponent(batch.id)}`);
        setActiveBatchProject(batch);
        setView('batch-novels');
        return;
      }
      const next = await getProject(project.id);
      if (!mountedRef.current || requestId !== projectRequestRef.current) return;
      setActiveProject(next); setView('studio');
    } catch (error) { message.error(error.message || '读取项目工作台失败'); }
  }, []);

  async function segmentAndOpenProject(readModel, segmentationMode) {
    const projectId = readModel?.project?.id;
    if (!projectId) throw new Error('项目创建后无法读取原文');

    let candidates;
    if (segmentationMode === 'smart') {
      const modelResult = await listModels();
      const textModels = modelResult.models || [];
      const textModel = textModels.find(model => model.kind === 'text');
      if (!textModel) throw new Error('当前没有可用的文本分析模型，请联系管理员配置后重试');
      const result = await smartSegmentation(projectId, { modelId: textModel.id });
      candidates = result.candidates || [];
    } else {
      const result = await paragraphSegmentation(projectId, {});
      candidates = result.candidates || [];
    }

    if (!candidates.length) throw new Error('未识别到可导入的文本段落');
    const confirmed = await confirmSegmentation(projectId, candidates);
    setActiveProject(confirmed?.project ? confirmed : await getProject(projectId));
    setImportNotice(candidates.length);
    setView('studio');
    message.success(`成功导入 ${candidates.length} 条文本`);
  }

  async function handleCreate({ name, sourceText, segmentationMode, productionMode }) {
    const created = await createProject({ name, productionMode });
    let readModel;
    try {
      readModel = sourceText ? await replaceProjectSource(created.id, { sourceText }) : await getProject(created.id);
    } catch (error) {
      const sourceError = error.message || '保存原文失败';
      try {
        await deleteProject(created.id);
      } catch (_) {
        await refreshProjects();
        throw new Error(`${sourceError}。项目已创建，但自动清理失败，请在项目库手动删除“${name}”。`);
      }
      await refreshProjects();
      throw new Error(`${sourceError}。已自动删除未保存原文的项目。`);
    }
    try {
      await segmentAndOpenProject(readModel, segmentationMode);
    } finally {
      await refreshProjects();
    }
  }
  async function handleCreateBatch(input) {
    // 巨量素材弹窗已提前建好 intake（含 N 本占位书）：直接从 intake 建批量。
    // 旧路径 createManualIntake 的 Go 端 ManualIntakeInput 不认识 intakeId，会按空 InputText 解析报错。
    const created = input?.intakeId
      ? await createBatchFromIntake(input.intakeId, { title: input.title, initialBatchSettings: input.initialBatchSettings, giantAutomationPlan: input.giantAutomationPlan })
      : await createManualIntake(input);
    const batch = created?.batch || created?.data?.batch || created;
    localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
    setActiveBatchProject(batch);
    setView('batch-novels');
    await refreshProjects();
    if (batch?.id) {
      try {
        const result = await classifyFetchedBatchMetadata(batch.id);
        const rows = Array.isArray(result?.results) ? result.results : [];
        const classified = rows.filter(item => item?.status === 'classified').length;
        const failed = rows.filter(item => item?.status === 'failed').length;
        if (classified) message.success(`已识别 ${classified} 本小说的男女频与风格`);
        if (failed) message.warning(`${failed} 本小说的男女频与风格暂未识别，可在单书中重试；不影响生产。`);
        const refreshed = batchFactoryBatchFromResponse(await getBatch(batch.id));
        if (mountedRef.current) setActiveBatchProject(current => String(current?.id) === String(batch.id) ? refreshed : current);
        await refreshProjects();
      } catch (error) {
        if (mountedRef.current) message.warning(`男女频与风格暂未识别：${error?.message || '可在单书中重试；不影响生产。'}`);
      }
    }
    if (input?.automationEnabled === true && batch?.id) {
      try {
        await startBatchAutomation(batch.id, {
          scheduledAt: input.scheduledAt || '',
          scheduleMode: input.scheduleMode || 'execute',
          presetId: input.presetId || '',
          runMode: input.runMode || (input.autoPublishEnabled === true ? 'full_submit' : 'video_no_submit'),
          autoPublish: input.autoPublishEnabled === true,
          concurrency: input.automationConcurrency
        });
        message.success(input.scheduledAt ? `批量工程已创建，服务端自动生产将在设定时间启动${input.autoPublishEnabled ? '，并自动上传视频管理系统' : ''}。` : `批量工程已创建，服务端自动生产已启动；关闭页面不影响执行${input.autoPublishEnabled ? '，完成后将自动上传视频管理系统' : ''}。`);
      } catch (error) {
        message.warning(`批量工程已创建，但自动生产启动失败：${error?.message || '请进入工程后手动启动'}`);
      }
    }
    if (input?.intakeId && batch?.id) {
      // intakeId 路径下 classifyFetchedBatchMetadata 会逐本写分类 metadata 并 bump 书的
      // revision：重新拉取最新快照返回，避免弹窗按旧 revision 写 executorJobId 乐观锁冲突。
      try {
        return batchFactoryBatchFromResponse(await getBatch(batch.id));
      } catch (_) {
        return batch;
      }
    }
    return batch;
  }

  useEffect(() => {
    const intakeId = pendingNovelFetchIntakeId(window.location.search);
    if (!intakeId || novelFetchHandoffRef.current === intakeId) return;
    novelFetchHandoffRef.current = intakeId;
    let cancelled = false;
    const clearHandoffQuery = () => window.history.replaceState({}, '', '/batch-factory');
    const openImportedBatch = async batch => {
      if (cancelled || !batch?.id) return;
      localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
      setActiveBatchProject(batchFactoryBatchFromResponse(batch));
      setView('batch-novels');
      clearHandoffQuery();
      await refreshProjects();
    };
    const join = async () => {
      try {
        const [intakeResult, currentBatchId] = await Promise.all([
          getIntake(intakeId),
          Promise.resolve(localStorage.getItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY) || '')
        ]);
        if (cancelled) return;
        const intake = intakeResult?.intake || intakeResult;
        const incoming = novelFetchIntakeBooks(intake);
        if (!currentBatchId) {
          const created = await createBatchFromIntake(intakeId, {
            title: incoming.length === 1 ? `小说获取·${incoming[0]?.title || '新批量'}` : `小说获取·${incoming.length || 0}本`
          });
          await openImportedBatch(created?.batch || created);
          message.success('小说获取内容已登记到新批量');
          return;
        }
        const current = await getBatch(currentBatchId);
        if (cancelled) return;
        const batch = current?.batch || current;
        const sourceKey = book => `${String(book?.platform || '').trim()}\u0000${String(book?.bookId || book?.id || '').trim()}`;
        const existingSourceKeys = new Set((batch?.books || []).map(sourceKey).filter(key => !key.endsWith('\u0000')));
        const duplicates = incoming.filter(book => existingSourceKeys.has(sourceKey(book)));
        const append = async () => {
          const updated = await appendNovelFetchIntake(batch.id, intakeId);
          await openImportedBatch(updated?.batch || updated);
          message.success(duplicates.length ? `小说获取内容已登记；同书城同 Book ID 的 ${duplicates.length} 本内容已跳过` : '小说获取内容已登记到当前批量');
        };
        await append();
      } catch (error) {
        if (!cancelled) {
          novelFetchHandoffRef.current = '';
          message.error(error?.message || '小说获取交接失败');
        }
      }
    };
    join();
    return () => { cancelled = true; };
  }, [refreshProjects]);
  async function handleImported(readModel, segmentationMode) {
    try {
      await segmentAndOpenProject(readModel, segmentationMode);
    } finally {
      await refreshProjects();
    }
  }
  async function handleDelete(project) {
    // A delayed project read must not restore an item after it has been deleted.
    projectRequestRef.current += 1;
    try {
      if (isBatchFactoryV11Project(project)) await deleteBatchFactoryProject(project.batchId);
      else await deleteProject(project.id);
      if (activeProject?.project?.id === project.id) { setActiveProject(null); setView('projects'); }
      await refreshProjects();
      message.success('项目已删除');
    } catch (error) { message.error(error.message || '删除项目失败'); }
  }
  function applyReadModel(readModel) { setActiveProject(readModel); }
  async function refreshActive() {
    if (!activeProject?.project?.id) return;
    try { setActiveProject(await getProject(activeProject.project.id)); } catch (error) { message.error(error.message || '刷新项目失败'); }
  }

  return <div className={`shuihuo-production ${view === 'studio' || view === 'batch-novels' ? 'is-workbench' : ''}`}>
    {view !== 'projects' && health ? <div className="shuihuo-readiness-strip" role="status" aria-live="polite"><strong className="shuihuo-readiness-title">运行依赖</strong>{readinessItems(health, activeBatchExecutorKinds, localExecutorStatus.online).map(([name, dependency]) => <span className={`${dependency?.ready ? 'ready' : 'missing'}${dependency?.executor ? ' is-local-executor' : ''}`} key={name} title={dependency?.reason || `${name}已配置`}>{dependency?.executor ? <><i aria-hidden="true" />{name}：一战晟铭执行器{dependency.ready ? '在线' : '离线'}</> : <>{name}：{dependency?.ready ? '已配置' : '未配置'}{dependency?.ready || !dependency?.reason ? '' : `（${dependency.reason}）`}</>}</span>)}</div> : null}
    {view === 'studio' && healthError ? <div className="shuihuo-readiness-strip" role="status" aria-live="polite"><span className="missing">状态读取失败：{healthError}</span></div> : null}
    {loading && view === 'projects' ? <div className="shuihuo-loading"><Spin /></div> : null}
    {!loading && view === 'projects' ? <ProjectsView projects={projects} loadError={projectsLoadError} health={health} onCreate={handleCreate} onImported={handleImported} onCreateBatch={handleCreateBatch} onOpen={openProject} onDelete={handleDelete} onRefresh={refreshProjects} openCreateOnLoad={openBatchOnLoad} /> : null}
    {view === 'batch-novels' && activeBatchProject ? <BatchFactoryNovelList batch={activeBatchProject} onBack={() => { projectRequestRef.current += 1; window.history.replaceState({}, '', '/batch-factory'); setActiveBatchProject(null); setView('projects'); refreshProjects(); }} onBatchChanged={async () => { const current = batchFactoryBatchFromResponse(await getBatch(activeBatchProject.id)); setActiveBatchProject(current); await refreshProjects(); return current; }} /> : null}
    {view === 'studio' && activeProject ? <CommentaryWorkbench data={activeProject} readiness={health} importNotice={importNotice} onBackToProjects={() => { setImportNotice(null); setView('projects'); refreshProjects(); }} onOpenAssets={() => setAssetsOpen(true)} onDataChange={applyReadModel} /> : null}
    <Modal title="人物场景预设" open={assetsOpen} onCancel={() => setAssetsOpen(false)} footer={null} width="min(1360px, calc(100vw - 48px))" className="shuihuo-assets-modal" destroyOnClose={false}>
      {activeProject ? <AssetsView data={activeProject} onRefresh={refreshActive} embedded /> : null}
    </Modal>
  </div>;
}

export default ShuihuoProductionPage;
