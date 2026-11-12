import { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, Checkbox, Drawer, Empty, Input, InputNumber, Select, Space, Spin, Switch, Tag, Typography, message } from 'antd';
import { Bot, CheckCircle2, Clock3, Sparkles, XCircle } from 'lucide-react';
import { applyAICutEditPlan, cancelAICutEditPlan, createAICutEditPlan, previewAICutEditPlan } from '../../shared/api/aicutEditPlan.js';

import { operationReviewDetails, selectedOperationIds } from './aicutEditPlanReview.js';

const operationNames = {
  place_asset: '放置真实素材',
  trim_clip: '裁剪片段', remove_range: '删除区间', move_clip: '移动片段', add_caption: '添加字幕', add_title: '添加文字',
  set_transition: '设置转场', set_filter: '设置滤镜', add_effect_layer: '添加特效', add_bgm: '添加背景音乐',
  set_volume_envelope: '调整音量', set_project_format: '设置画面比例'
};

function durationText(milliseconds) {
  const seconds = Math.max(0, Math.round(Number(milliseconds || 0) / 1000));
  return `${Math.floor(seconds / 60)}分${String(seconds % 60).padStart(2, '0')}秒`;
}

function modelOption(model) {
  const value = String(model?.id || model?.modelId || model?.key || '').trim();
  return value ? { value, label: model?.displayName || model?.name || model?.label || value } : null;
}

function assetOption(asset) {
  return { value: asset.id, label: `${asset.name || '未命名素材'} · ${asset.kind || '未知类型'} · ${durationText(asset.durationMs)}` };
}

export default function AICutAutoEditDrawer({ open, projectId, revision = 0, assets = [], models = [], onClose, onApplied, embedded = false, selectedModelId, onModelChange, onBusyChange }) {
  const [localModelId, setLocalModelId] = useState('');
  const modelId = selectedModelId ?? localModelId;
  const setModelId = onModelChange || setLocalModelId;
  const [assetIds, setAssetIds] = useState([]);
  const [goal, setGoal] = useState('剪成节奏紧凑、信息清晰的成片');
  const [targetSeconds, setTargetSeconds] = useState(60);
  const [aspectRatio, setAspectRatio] = useState('9:16');
  const [options, setOptions] = useState({ captions: true, transitions: true, filters: true, bgm: false });
  const [extraRequirement, setExtraRequirement] = useState('');
  const [plan, setPlan] = useState(null);
  const [enabledIds, setEnabledIds] = useState([]);
  const [busy, setBusy] = useState('');
  const [reviewSummary, setReviewSummary] = useState(null);
  const [reviewError, setReviewError] = useState('');
  const controllerRef = useRef(null);
  const modelOptions = useMemo(() => models.map(modelOption).filter(Boolean), [models]);
  const assetOptions = useMemo(() => assets.filter(item => item?.id).map(assetOption), [assets]);

  useEffect(() => {
    if (!open) return;
    setPlan(null);
    setEnabledIds([]);
    setBusy('');
    setAssetIds(current => current.filter(id => assets.some(asset => asset.id === id)));
    if (!modelId && modelOptions[0]) setModelId(modelOptions[0].value);
  }, [open, projectId]);

  useEffect(() => {
    if (!modelId && modelOptions[0]) setModelId(modelOptions[0].value);
  }, [modelId, modelOptions]);

  useEffect(() => () => controllerRef.current?.abort(), []);

  useEffect(() => { onBusyChange?.(Boolean(busy)); }, [busy, onBusyChange]);

  useEffect(() => {
    if (plan?.state !== 'ready') return;
    let active = true;
    setReviewSummary(null);
    setReviewError('');
    previewAICutEditPlan(projectId, plan.id, selectedOperationIds(plan, enabledIds)).then(summary => {
      if (active) setReviewSummary(summary);
    }).catch(error => { if (active) setReviewError(error?.message || '这些操作无法一起执行，请调整勾选'); });
    return () => { active = false; };
  }, [projectId, plan, enabledIds]);

  function resetDraft() {
    if (busy) return;
    setPlan(null);
    setEnabledIds([]);
    setReviewSummary(null);
    setReviewError('');
  }

  async function generateDraft() {
    const controller = new AbortController();
    controllerRef.current = controller;
    setBusy('planning');
    try {
      const created = await createAICutEditPlan(projectId, {
        clientRequestId: `edit-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        modelId, expectedRevision: revision, assetIds, goal, targetDurationMs: Math.round(Number(targetSeconds || 0) * 1000),
        aspectRatio, includeCaptions: options.captions, includeTransitions: options.transitions,
        includeFilters: options.filters, includeBgm: options.bgm, extraRequirement, signal: controller.signal
      });
      setPlan(created);
      setEnabledIds((created?.operations || []).filter(item => item.enabled !== false).map(item => item.id));
    } catch (error) {
      message.error(/[一-鿿]/.test(error?.message || '') ? error.message : 'AI 剪辑草案生成失败');
    } finally {
      controllerRef.current = null;
      setBusy('');
    }
  }

  async function cancelDraft() {
    if (!plan?.id) {
      controllerRef.current?.abort();
      setBusy('');
      return;
    }
    setBusy('canceling');
    try {
      const canceled = await cancelAICutEditPlan(projectId, plan.id);
      setPlan(canceled);
      message.info('草案已取消，时间线没有改动');
    } catch (error) {
      message.error(error?.message || '取消草案失败');
    } finally {
      setBusy('');
    }
  }

  async function applyDraft() {
    if (plan?.state !== 'ready') return;
    setBusy('applying');
    try {
      const enabledOperationIds = selectedOperationIds(plan, enabledIds);
      const result = await applyAICutEditPlan(projectId, plan.id, enabledOperationIds);
      // 应用成功后立刻锁成已应用，避免同一份草案被重复写入时间线。
      setPlan(current => (result?.plan || (current ? { ...current, state: 'applied' } : current)));
      setReviewSummary(result?.summary || null);
      onApplied?.(result);
    } catch (error) {
      const text = error?.message || '剪辑草案应用失败';
      if (text.includes('过期') || text.includes('时间线已更新')) setPlan(current => current ? { ...current, state: 'stale' } : current);
      message.error(/[一-鿿]/.test(text) ? text : '剪辑草案应用失败');
    } finally {
      setBusy('');
    }
  }

  const planning = busy === 'planning';
  const content = <>
    {!plan ? <div className="aicut-auto-edit-form">
      <label><strong>选择素材</strong><Select mode="multiple" value={assetIds} options={assetOptions} onChange={setAssetIds} placeholder="选择要参与剪辑的素材" optionFilterProp="label" /></label>
      <label><strong>文本模型</strong><Select value={modelId || undefined} options={modelOptions} onChange={setModelId} placeholder="选择文本模型" optionFilterProp="label" /></label>
      <label><strong>剪辑目标</strong><Input.TextArea value={goal} onChange={event => setGoal(event.target.value)} autoSize={{ minRows: 3, maxRows: 6 }} maxLength={2000} /></label>
      <div className="aicut-auto-edit-grid"><label><strong>目标时长（尽量接近）</strong><InputNumber min={0} max={7200} value={targetSeconds} onChange={value => setTargetSeconds(value ?? 0)} addonAfter="秒" /></label><label><strong>画面比例</strong><Select value={aspectRatio} onChange={setAspectRatio} options={['9:16', '16:9', '1:1'].map(value => ({ value, label: value }))} /></label></div>
      <div className="aicut-auto-edit-switches">
        {[['captions', '字幕'], ['transitions', '转场'], ['filters', '滤镜'], ['bgm', '背景音乐']].map(([key, label]) => <label key={key}><span>{label}</span><Switch checked={options[key]} onChange={checked => setOptions(current => ({ ...current, [key]: checked }))} /></label>)}
      </div>
      <label><strong>补充要求</strong><Input.TextArea value={extraRequirement} onChange={event => setExtraRequirement(event.target.value)} placeholder="例如：保留人物对白，前 3 秒快速进入主题" maxLength={4000} /></label>
      {!assets.length ? <Alert type="warning" showIcon message="素材库为空，请先上传视频、图片或音频" /> : null}
      <Button type="primary" block size="large" icon={<Bot size={16} />} loading={planning} disabled={!projectId || !assetIds.length || !modelId || !goal.trim()} onClick={generateDraft}>生成剪辑草案</Button>
    </div> : <div className="aicut-auto-edit-review">
      {planning ? <div className="aicut-auto-edit-planning"><Spin /><Typography.Title level={4}>正在根据素材信息和时间线生成草案</Typography.Title><Typography.Text type="secondary">这里只生成草案，不会直接修改时间线。</Typography.Text></div> : null}
      {plan.state === 'stale' ? <Alert type="error" showIcon message="草案已过期，请基于最新工程重新生成" /> : null}
      {plan.state === 'canceled' ? <Alert type="info" showIcon message="草案已取消，时间线没有改动" /> : null}
      {plan.state === 'failed' ? <Alert type="error" showIcon message={plan.error || 'AI 剪辑草案生成失败'} /> : null}
      {plan.state === 'ready' || plan.state === 'applied' ? <>
        <section className="aicut-auto-edit-summary"><div><Clock3 size={17} /><span>预计成片时长</span><strong>{reviewSummary ? durationText(reviewSummary.estimatedDurationMs) : plan.state === 'applied' ? durationText(plan.summary?.estimatedDurationMs) : '正在计算'}</strong></div><div><CheckCircle2 size={17} /><span>准备应用</span><strong>{enabledIds.length} 项</strong></div></section>
        {(plan.warnings || []).map((warning,index) => <Alert key={index} type="warning" showIcon message={warning} />)}
        {reviewError ? <Alert type="error" showIcon message={reviewError} /> : null}
        <Typography.Paragraph type="secondary">时长按当前勾选的操作试算。时间单位为毫秒，sourceStartMs 表示源素材起点，startMs 表示时间线位置。</Typography.Paragraph>
        <div className="aicut-auto-edit-operations">{(plan.operations || []).length ? plan.operations.map(operation => <article key={operation.id} className={enabledIds.includes(operation.id) ? 'is-enabled' : ''}><Checkbox checked={enabledIds.includes(operation.id)} disabled={plan.state !== 'ready'} onChange={event => setEnabledIds(current => event.target.checked ? [...current, operation.id] : current.filter(id => id !== operation.id))} /><div><strong>{operationNames[operation.type] || operation.type}</strong><p><b>操作理由：</b>{operation.reason || '未提供'}</p><p><b>时间线起点：</b>{operation.startMs ?? 0} 毫秒 · <b>持续：</b>{operation.durationMs ?? 0} 毫秒 · <b>源起点：</b>{operation.sourceStartMs ?? 0} 毫秒{operation.endMs != null ? ` · 结束：${operation.endMs} 毫秒` : ''}</p>{operation.text ? <p><b>文字内容：</b>{operation.text}</p> : null}<details><summary>查看全部操作参数</summary><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{operationReviewDetails(operation)}</pre></details></div></article>) : <Empty description="草案没有需要执行的操作" />}</div>
      </> : null}
      <Space className="aicut-auto-edit-actions">
        <Button disabled={Boolean(busy)} onClick={resetDraft}>新建草案</Button>
        <Button icon={<XCircle size={15} />} disabled={busy === 'applying' || ['applied', 'canceled', 'failed', 'stale'].includes(plan.state)} loading={busy === 'canceling'} onClick={cancelDraft}>取消草案</Button>
        <Button type="primary" disabled={plan?.state !== 'ready' || !enabledIds.length || !reviewSummary || Boolean(reviewError)} loading={busy === 'applying'} onClick={applyDraft}>{plan?.state === 'applied' ? '已应用到时间线' : '应用到时间线'}</Button>
      </Space>
    </div>}
  </>;
  if (embedded) return <div className="aicut-auto-edit-embedded">{content}</div>;
  return <Drawer className="aicut-auto-edit-drawer" width={520} open={open} onClose={busy ? undefined : onClose} title="AI 自动剪辑 · 自动成片">{content}</Drawer>;
}
