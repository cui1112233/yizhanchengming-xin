import { FileTextOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Button, Modal, Spin, message } from 'antd';
import { useCallback, useEffect, useRef, useState } from 'react';
import {
  classifyFetchedBatchMetadata,
  createBatchFromIntake,
  createManualIntake,
  getBatch,
  getIntake,
  listBatchSummaries,
  startBatchAutomation
} from '../../shared/api/batchFactory';
import { BatchFactoryNovelList } from './shuihuo/BatchFactoryNovelList';
import { BatchFactoryCreateModal } from './shuihuo/BatchFactoryCreateModal';
import { batchFactoryBatchFromResponse } from './shuihuo/batchFactoryProjects';
import {
  BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY,
  novelFetchIntakeBooks,
  pendingNovelFetchIntakeId
} from './shuihuo/batchFactoryNovelFetchHandoff';
import './shuihuo-production.css';
import './BatchFactoryWorkbenchPage.css';

function batchesFromResult(result) {
  if (Array.isArray(result)) return result;
  if (Array.isArray(result?.batches)) return result.batches;
  if (Array.isArray(result?.items)) return result.items;
  if (Array.isArray(result?.data?.batches)) return result.data.batches;
  if (Array.isArray(result?.data?.items)) return result.data.items;
  return [];
}

function asBatch(result) {
  return batchFactoryBatchFromResponse({ batch: result?.batch || result });
}

function bookCount(batch) {
  return Array.isArray(batch?.books) ? batch.books.length : Number(batch?.bookCount || 0);
}

export default function BatchFactoryWorkbenchPage() {
  const [batches, setBatches] = useState([]);
  const [activeBatch, setActiveBatch] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [openAutomationOnLoad, setOpenAutomationOnLoad] = useState(false);
  const handoffRef = useRef('');
  // When arriving from Novel Fetch, the intake must win over the previously
  // remembered batch. Otherwise the two async loads can race and put the old
  // project back on screen after the new intake has already been created.
  const pendingHandoffRef = useRef(
    pendingNovelFetchIntakeId(typeof window !== 'undefined' ? window.location.search : '')
  );
  const pendingBatchIdRef = useRef(typeof window !== 'undefined'
    ? (() => {
      const params = new URLSearchParams(window.location.search);
      return params.get('batchId') || params.get('batch') || params.get('project') || '';
    })()
    : '');
  const pendingBatchId = pendingBatchIdRef.current;
  const mountedRef = useRef(true);
  const refreshInFlightRef = useRef(false);

  useEffect(() => {
    mountedRef.current = true;
    document.body.classList.add('shuihuo-theme-active');
    document.body.classList.add('shuihuo-reference-shell');
    return () => {
      mountedRef.current = false;
      document.body.classList.remove('shuihuo-theme-active');
      document.body.classList.remove('shuihuo-reference-shell');
    };
  }, []);

  const refreshBatches = useCallback(async ({ quiet = false } = {}) => {
    if (refreshInFlightRef.current) return undefined;
    refreshInFlightRef.current = true;
    if (!quiet) setLoading(true);
    setError('');
    try {
      const next = batchesFromResult(await listBatchSummaries());
      if (mountedRef.current) setBatches(next);
      return next;
    } catch (requestError) {
      const reason = requestError?.message || '批量工程读取失败';
      if (mountedRef.current) setError(reason);
      if (!quiet) message.error(reason);
      return [];
    } finally {
      refreshInFlightRef.current = false;
      if (mountedRef.current) setLoading(false);
    }
  }, []);

  const openBatch = useCallback(async batchId => {
    try {
      const batch = asBatch(await getBatch(batchId));
      localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
      if (mountedRef.current) setActiveBatch(batch);
    } catch (requestError) {
      message.error(requestError?.message || '读取批量工程失败');
    }
  }, []);

  useEffect(() => { refreshBatches({ quiet: true }); }, [refreshBatches]);

  useEffect(() => {
    if (activeBatch) return undefined;
    const refreshWhenVisible = () => {
      if (!document.hidden) void refreshBatches({ quiet: true });
    };
    const timer = window.setInterval(refreshWhenVisible, 15000);
    document.addEventListener('visibilitychange', refreshWhenVisible);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener('visibilitychange', refreshWhenVisible);
    };
  }, [activeBatch, refreshBatches]);

  useEffect(() => {
    if (pendingHandoffRef.current || pendingBatchId) return undefined;
    const activeBatchId = localStorage.getItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY) || '';
    if (!activeBatchId) return undefined;
    void openBatch(activeBatchId);
    return undefined;
  }, [openBatch]);

  useEffect(() => {
    if (!pendingBatchId || pendingHandoffRef.current) return undefined;
    let cancelled = false;
    getBatch(pendingBatchId).then(result => {
      if (cancelled) return;
      const batch = asBatch(result);
      localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
      setActiveBatch(batch);
      window.history.replaceState({}, '', '/batch-factory');
    }).catch(requestError => {
      if (!cancelled) message.error(requestError?.message || '读取批量工程失败');
    });
    return () => { cancelled = true; };
  }, [pendingBatchId]);

  useEffect(() => {
    const intakeId = pendingHandoffRef.current;
    if (!intakeId || handoffRef.current === intakeId) return undefined;
    handoffRef.current = intakeId;
    let cancelled = false;
    const clearHandoffQuery = () => window.history.replaceState({}, '', '/batch-factory');
    const finish = async result => {
      const batch = asBatch(result);
      if (cancelled) return;
      localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
      setActiveBatch(batch);
      setOpenAutomationOnLoad(true);
      clearHandoffQuery();
      await refreshBatches({ quiet: true });
    };
    const join = async () => {
      try {
        const intakeResult = await getIntake(intakeId);
        const intake = intakeResult?.intake || intakeResult;
        const incoming = novelFetchIntakeBooks(intake);
        const result = await createBatchFromIntake(intakeId, {
          title: incoming.length === 1 ? `小说获取·${incoming[0]?.title || '新批量'}` : `小说获取·${incoming.length || 0}本`,
          initialBatchSettings: {
            contentRangeLines: Number(intake?.metadata?.contentRangeLines || 5),
            contentCaptureCharacters: Number(intake?.metadata?.contentCaptureCharacters || 4000),
          }
        });
        await finish(result);
        message.success('小说获取内容已登记到新批量；统一配置、内容范围、内容截取和自动生产入口已保留');
      } catch (requestError) {
        if (!cancelled) {
          handoffRef.current = '';
          message.error(requestError?.message || '小说获取交接失败');
        }
      }
    };
    join();
    return () => { cancelled = true; };
  }, [refreshBatches]);

  async function createBatch(input) {
    const created = input?.intakeId
      ? await createBatchFromIntake(input.intakeId, { title: input.title, initialBatchSettings: input.initialBatchSettings, giantAutomationPlan: input.giantAutomationPlan })
      : await createManualIntake(input);
    const batch = asBatch(created);
    localStorage.setItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, batch.id);
    setActiveBatch(batch);
    await refreshBatches({ quiet: true });
    // Text-model classification is deliberately detached from batch creation:
    // a provider failure must never prevent a fetched source from entering
    // production.  The server handles each book independently and persists
    // successful gender/style results before this refresh.
    void classifyFetchedBatchMetadata(batch.id).then(async result => {
      const rows = Array.isArray(result?.results) ? result.results : [];
      const classified = rows.filter(item => item?.status === 'classified').length;
      const failed = rows.filter(item => item?.status === 'failed').length;
      if (classified) message.success(`已识别 ${classified} 本小说的男女频与风格`);
      if (failed) message.warning(`${failed} 本小说原文已获取，等待分类确认；确认分类后可继续改写。`);
      const refreshed = asBatch(await getBatch(batch.id));
      if (mountedRef.current) setActiveBatch(current => String(current?.id) === String(batch.id) ? refreshed : current);
      await refreshBatches({ quiet: true });
    }).catch(error => {
      if (mountedRef.current) message.warning(`原文已获取，等待分类确认；确认分类后可继续改写。${error?.message ? `原因：${error.message}` : ''}`);
    });
    if (!input?.automationEnabled) return batch;
    try {
      await startBatchAutomation(batch.id, {
        scheduledAt: input.scheduledAt || '',
        scheduleMode: input.scheduleMode || 'execute',
        presetId: input.presetId || '',
        runMode: input.runMode || (input.autoPublishEnabled ? 'full_submit' : 'video_no_submit'),
        autoPublish: input.autoPublishEnabled === true
      });
      message.success(input.scheduledAt ? '批量工程已创建，自动生产将在设定时间启动。' : '批量工程已创建并启动自动生产。');
    } catch (requestError) {
      message.warning(`批量工程已创建，但自动生产启动失败：${requestError?.message || '请进入工程后手动启动'}`);
    }
    return batch;
  }

  const refreshActiveBatch = useCallback(async () => {
    if (!activeBatch?.id) return undefined;
    const latest = asBatch(await getBatch(activeBatch.id));
    setActiveBatch(current => current?.updatedAt === latest?.updatedAt && current?.revision === latest?.revision ? current : latest);
    return latest;
  }, [activeBatch?.id]);

  if (activeBatch) {
    return <div className="shuihuo-production is-workbench batch-factory-canonical-entry">
      <BatchFactoryNovelList batch={activeBatch} openAutomationOnLoad={openAutomationOnLoad} onBack={() => { setActiveBatch(null); refreshBatches({ quiet: true }); }} onBatchChanged={refreshActiveBatch} />
    </div>;
  }

  return <main className="shuihuo-production batch-factory-canonical-entry batch-factory-workbench-home">
    <section className="batch-factory-workbench-home-heading">
      <div><h1>批量工厂</h1><p>统一管理批次、单书配置、分镜生产与视频提交。</p></div>
      <div className="batch-factory-workbench-home-actions"><Button onClick={() => window.location.assign('/shuihuo-production/creative')}>创作漫剧</Button><Button icon={<ReloadOutlined />} loading={loading} onClick={() => refreshBatches()}>刷新</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>新建批量</Button></div>
    </section>
    {loading ? <div className="batch-factory-workbench-home-state"><Spin /><span>正在读取批量工程</span></div> : null}
    {!loading && error ? <div className="batch-factory-workbench-home-state is-error"><strong>批量工程暂时无法读取</strong><span>{error}{batches.length ? '，当前仍显示上次读取的数据。' : ''}</span><Button onClick={() => refreshBatches()}>重试</Button></div> : null}
    {!loading && (!error || batches.length > 0) ? <section className="batch-factory-workbench-grid">
      {batches.map(batch => <button type="button" className="batch-factory-workbench-card" key={batch.id} onClick={() => openBatch(batch.id)}><FileTextOutlined /><strong>{batch.title || '未命名批量'}</strong><span>{bookCount(batch)} 本小说</span><small>{batch.updatedAt || batch.createdAt || '刚刚创建'}</small></button>)}
      {!batches.length ? <div className="batch-factory-workbench-home-state"><FileTextOutlined /><strong>还没有批量工程</strong><span>新建批量后即可进入每本小说的完整生产工作台。</span><Button type="primary" onClick={() => setCreateOpen(true)}>新建批量</Button></div> : null}
    </section> : null}
    <BatchFactoryCreateModal open={createOpen} onCancel={() => setCreateOpen(false)} onCreated={async input => { const batch = await createBatch(input); setCreateOpen(false); return batch; }} onBatchUpdated={async batchId => { const refreshed = asBatch(await getBatch(batchId)); if (mountedRef.current) setActiveBatch(refreshed); await refreshBatches({ quiet: true }); }} />
  </main>;
}
