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
} from '../../shared/api/batchFactoryV11';
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
  const mountedRef = useRef(true);

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
    setLoading(true);
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
    const activeBatchId = localStorage.getItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY) || '';
    if (!activeBatchId) return undefined;
    void openBatch(activeBatchId);
    return undefined;
  }, [openBatch]);

  useEffect(() => {
    const intakeId = pendingNovelFetchIntakeId(window.location.search);
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
      ? await createBatchFromIntake(input.intakeId, { title: input.title, giantAutomation: input.giantAutomation })
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
      if (failed) message.warning(`${failed} 本小说的男女频与风格暂未识别，可在单书中重试；不影响生产。`);
      const refreshed = asBatch(await getBatch(batch.id));
      if (mountedRef.current) setActiveBatch(current => String(current?.id) === String(batch.id) ? refreshed : current);
      await refreshBatches({ quiet: true });
    }).catch(error => {
      if (mountedRef.current) message.warning(`男女频与风格暂未识别：${error?.message || '可在单书中重试；不影响生产。'}`);
    });
    if (!input?.automationEnabled) return batch;
    try {
      await startBatchAutomation(batch.id, {
        scheduledAt: input.scheduledAt || '',
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

  if (activeBatch) {
    return <div className="shuihuo-production is-workbench batch-factory-canonical-entry">
      <BatchFactoryNovelList batch={activeBatch} openAutomationOnLoad={openAutomationOnLoad} onBack={() => { setActiveBatch(null); refreshBatches({ quiet: true }); }} onBatchChanged={async () => {
        setActiveBatch(asBatch(await getBatch(activeBatch.id)));
        await refreshBatches({ quiet: true });
      }} />
    </div>;
  }

  return <main className="shuihuo-production batch-factory-canonical-entry batch-factory-workbench-home">
    <section className="batch-factory-workbench-home-heading">
      <div><h1>批量工厂</h1><p>统一管理批次、单书配置、分镜生产与视频提交。</p></div>
      <div className="batch-factory-workbench-home-actions"><Button onClick={() => window.location.assign('/shuihuo-production/creative')}>创作漫剧</Button><Button icon={<ReloadOutlined />} loading={loading} onClick={() => refreshBatches()}>刷新</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>新建批量</Button></div>
    </section>
    {loading ? <div className="batch-factory-workbench-home-state"><Spin /><span>正在读取批量工程</span></div> : null}
    {!loading && error ? <div className="batch-factory-workbench-home-state is-error"><strong>批量工程暂时无法读取</strong><span>{error}</span><Button onClick={() => refreshBatches()}>重试</Button></div> : null}
    {!loading && !error ? <section className="batch-factory-workbench-grid">
      {batches.map(batch => <button type="button" className="batch-factory-workbench-card" key={batch.id} onClick={() => openBatch(batch.id)}><FileTextOutlined /><strong>{batch.title || '未命名批量'}</strong><span>{bookCount(batch)} 本小说</span><small>{batch.updatedAt || batch.createdAt || '刚刚创建'}</small></button>)}
      {!batches.length ? <div className="batch-factory-workbench-home-state"><FileTextOutlined /><strong>还没有批量工程</strong><span>新建批量后即可进入每本小说的完整生产工作台。</span><Button type="primary" onClick={() => setCreateOpen(true)}>新建批量</Button></div> : null}
    </section> : null}
    <BatchFactoryCreateModal open={createOpen} onCancel={() => setCreateOpen(false)} onCreated={async input => { const batch = await createBatch(input); setCreateOpen(false); return batch; }} onBatchUpdated={async batchId => { const refreshed = asBatch(await getBatch(batchId)); if (mountedRef.current) setActiveBatch(refreshed); await refreshBatches({ quiet: true }); }} />
  </main>;
}
