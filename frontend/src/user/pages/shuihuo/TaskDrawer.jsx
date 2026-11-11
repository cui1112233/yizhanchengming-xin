import { Alert, Button, Drawer, Empty, Popconfirm, Select, Tag, message } from 'antd';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { apiRequest } from '../../../shared/api/client';
import { cancelTask, createTask, listModels, listTasks, retryTask } from '../../../shared/api/shuihuoProduction';
import { BatchFactoryErrorNotice } from './BatchFactoryErrorNotice';
import { taskReadiness } from './taskReadiness';

const labels = { draft: '草稿', queued: '排队中', running: '生成中', succeeded: '已完成', failed: '失败', cancelled: '已取消' };
const colors = { queued: 'blue', running: 'gold', succeeded: 'green', failed: 'red', cancelled: 'default' };

function completedMediaFor(task, media) {
  return media.map(item => item?.media || item).find(item => item?.taskId === task.id) || null;
}

export function TaskDrawer({ open, taskFilter = 'all', project, segments, media = [], readiness, onClose, onCompleted }) {
  const [models, setModels] = useState([]); const [tasks, setTasks] = useState([]); const [modelId, setModelId] = useState(); const [segmentId, setSegmentId] = useState(); const [kind, setKind] = useState('image'); const [busy, setBusy] = useState(false);
  const [loadError, setLoadError] = useState('');
  const requestGeneration = useRef(0);
  const availableModels = useMemo(() => models.filter(model => model.kind === kind), [models, kind]);
  const eligibleSegments = useMemo(() => segments.filter(segment => segment.confirmed), [segments]);
  const eligibleSegmentIds = useMemo(() => new Set(eligibleSegments.map(segment => segment.id)), [eligibleSegments]);
  const selectedSegment = useMemo(() => segments.find(segment => segment.id === segmentId), [segmentId, segments]);
  const segmentOrder = segment => segment?.orderIndex ?? segment?.id ?? '-';
  const kindAvailability = useMemo(() => taskReadiness(readiness, kind, {
    confirmed: selectedSegment ? selectedSegment.confirmed : true
  }), [kind, readiness, selectedSegment]);
  const kindReady = kindAvailability.ready;
  const visibleTasks = useMemo(() => taskFilter === 'active'
    ? tasks.filter(task => task.status === 'queued' || task.status === 'running')
    : tasks, [taskFilter, tasks]);
  const refresh = useCallback(async () => {
    const projectID = project?.id;
    const generation = ++requestGeneration.current;
    if (!projectID) return false;
    const [modelResult, taskResult] = await Promise.all([listModels(), listTasks(projectID)]);
    if (generation !== requestGeneration.current) return false;
    setModels(modelResult.models || []);
    setTasks(taskResult.tasks || []);
    return true;
  }, [project?.id]);
  useEffect(() => {
    requestGeneration.current += 1;
    if (!open) return undefined;
    let active = true;
    setModels([]); setTasks([]); setSegmentId(undefined); setLoadError('');
    refresh().catch(error => {
      if (!active) return;
      // 读取失败要单独标记，不能和「还没有任务」共用同一个空状态。
      setLoadError(error.message || '读取任务中心失败');
      message.error(error.message || '读取任务中心失败');
    });
    return () => { active = false; requestGeneration.current += 1; };
  }, [open, refresh]);
  useEffect(() => {
    if (!open || !tasks.some(task => task.status === 'queued' || task.status === 'running')) return undefined;
    const timer = window.setInterval(() => refresh().then(refreshed => { if (refreshed) onCompleted?.(); }).catch(() => {}), 2500);
    return () => window.clearInterval(timer);
  }, [open, refresh, tasks, onCompleted]);
  useEffect(() => { setModelId(undefined); }, [kind]);
  useEffect(() => {
    if (segmentId && !eligibleSegmentIds.has(segmentId)) setSegmentId(undefined);
  }, [eligibleSegmentIds, segmentId]);
  async function submit() {
    if (!kindReady) { message.warning(kindAvailability.reason || '当前任务依赖未就绪'); return; }
    if (!segmentId || modelId === undefined || modelId === null) { message.warning(availableModels.length ? '请选择分段与模型' : '当前没有可用模型'); return; }
    setBusy(true);
    try {
      await createTask(project.id, { segmentId, modelId, kind });
      await refresh();
      message.success(`已提交第 ${segmentOrder(selectedSegment)} 个分镜${kind === 'image' ? '图片' : '视频'}任务`);
    } catch (error) { message.error(error.message || '提交任务失败'); } finally { setBusy(false); }
  }
  async function operate(task, action) { setBusy(true); try { if (action === 'cancel') await cancelTask(task.id); else await retryTask(task.id); await refresh(); } catch (error) { message.error(error.message || '任务操作失败'); } finally { setBusy(false); } }
  async function manuallyRefresh() { setBusy(true); try { await refresh(); setLoadError(''); } catch (error) { setLoadError(error.message || '读取任务中心失败'); message.error(error.message || '读取任务中心失败'); } finally { setBusy(false); } }
  async function openCompletedMedia(completedMedia) {
    try {
      const blob = await apiRequest(`/api/shuihuo-production/media/${completedMedia.id}/download`, { responseType: 'blob' });
      const url = URL.createObjectURL(blob);
      window.open(url, '_blank', 'noopener,noreferrer');
      window.setTimeout(() => URL.revokeObjectURL(url), 60000);
    } catch (error) { message.error(error.message || '读取完成素材失败'); }
  }
  return <Drawer title={taskFilter === 'active' ? '生成任务中心 · 可取消任务' : '生成任务中心'} open={open} onClose={onClose} width={520} extra={<><Button type="text" onClick={manuallyRefresh} disabled={busy}>刷新</Button><Button type="primary" onClick={submit} loading={busy} disabled={!kindReady}>提交任务</Button></>}>
    <div className="shuihuo-task-form">
      <Select value={kind} onChange={setKind} options={[{ value:'image', label:'生成图片' }, { value:'video', label:'图生视频' }]} />
      <Select placeholder="选择已确认分段" value={segmentId} onChange={setSegmentId} options={eligibleSegments.map(segment => ({ value:segment.id, label:`#${segment.orderIndex} ${segment.sourceText.slice(0, 24)}` }))} />
      <Select placeholder={availableModels.length ? '选择生成模型' : '当前没有此类可用模型'} value={modelId} onChange={setModelId} options={availableModels.map(model => ({ value:model.id, label:model.name }))} disabled={!availableModels.length || !kindReady} />
    </div>
    {kind === 'video' ? <Alert type="info" showIcon message="图生视频" description="YD 优先使用绑定场景预设图作为场景图，没有时使用当前分镜主图片；人物和道具预设图会作为参考图。" /> : null}
    {!kindReady ? <Alert className="shuihuo-inline-alert" type="warning" showIcon message="当前任务依赖未就绪" description={kindAvailability.reason || '请检查运行配置'} /> : null}
    {!availableModels.length ? <p className="shuihuo-task-empty-note">当前没有启用的{kind === 'image' ? '图片' : kind === 'video' ? '视频' : '配音'}模型，不能提交生成任务。</p> : null}
    {loadError ? <Alert type="error" showIcon message="任务列表读取失败" description={loadError} action={<Button size="small" onClick={manuallyRefresh} disabled={busy}>重试</Button>} style={{ marginBottom: 12 }} /> : null}
    <div className="shuihuo-task-list">{visibleTasks.map(task => {
      const completedMedia = task.status === 'succeeded' ? completedMediaFor(task, media) : null;
      const retryable = task.status === 'failed' || task.status === 'cancelled';
      const taskSegment = segments.find(segment => segment.id === task.segmentId);
      return <article key={task.id} className="shuihuo-task-item"><div><strong>{task.kind === 'image' ? '图片生成' : task.kind === 'video' ? '视频生成' : task.kind === 'audio' ? '配音生成' : task.kind}</strong><span>#{task.id} · 第 {segmentOrder(taskSegment)} 个分镜</span>{task.provider ? <span>提供方：{task.provider}{task.providerTaskId ? ` · 上游任务 ${task.providerTaskId}` : ''}</span> : null}{task.errorCode ? <span>错误代码：{task.errorCode}</span> : null}{task.errorMessage ? <BatchFactoryErrorNotice compact title="生成任务失败" error={task.errorMessage} /> : null}{retryable ? <span>此任务可重试</span> : null}{completedMedia ? <Button type="link" size="small" onClick={() => openCompletedMedia(completedMedia)}>查看完成素材</Button> : null}</div><Tag color={colors[task.status]}>{labels[task.status] || task.status}</Tag>{task.status === 'queued' || task.status === 'running' ? <Popconfirm title="取消该任务？" onConfirm={() => operate(task, 'cancel')}><Button size="small">取消</Button></Popconfirm> : null}{retryable ? <Button size="small" onClick={() => operate(task, 'retry')}>重试</Button> : null}</article>;
    })}{!visibleTasks.length && !loadError ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={taskFilter === 'active' ? '暂无可取消任务' : '暂未提交任务'} /> : null}</div>
  </Drawer>;
}
