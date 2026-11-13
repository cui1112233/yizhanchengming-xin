import React, { useEffect, useRef, useState } from 'react';
import { Alert, Button, Input, InputNumber, Space, Switch } from 'antd';
import { getImageCapacity, saveImageCapacity } from '../../../shared/api/imageCapacity';

export default function ImageCapacityControl({ modelId, aspectRatio = '9:16', active = true }) {
  const [state, setState] = useState(null);
  const [draft, setDraft] = useState(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [refresh, setRefresh] = useState(0);
  const [editing, setEditing] = useState(false);
  const request = useRef(null);
  const generation = useRef(0);
  const saving = useRef(false);

  useEffect(() => {
    setState(null);
    setDraft(null);
    setEditing(false);
    setError('');
  }, [modelId, aspectRatio]);

  useEffect(() => {
    const token = ++generation.current;
    if (!active || !modelId) return undefined;
    let disposed = false;
    const load = async () => {
      if (saving.current) return;
      // A refresh supersedes the previous snapshot, including slow requests.
      request.current?.abort();
      const controller = new AbortController();
      request.current = controller;
      setBusy(true);
      try {
        const result = await getImageCapacity(modelId, { aspectRatio, signal: controller.signal });
        if (disposed || controller.signal.aborted || generation.current !== token) return;
        setState(result);
        setDraft(previous => previous || result.policy);
        setError('');
      } catch (cause) {
        if (!disposed && !controller.signal.aborted && generation.current === token) setError(cause.message || '无法读取图片并发状态');
      } finally {
        if (!disposed && !controller.signal.aborted && generation.current === token) setBusy(false);
      }
    };
    load();
    const timer = setInterval(load, 10000);
    return () => { disposed = true; clearInterval(timer); request.current?.abort(); };
  }, [modelId, aspectRatio, active, refresh]);

  const save = async () => {
    const token = generation.current;
    saving.current = true;
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    setBusy(true);
    try {
      const result = await saveImageCapacity(modelId, {
        aspectRatio, accountLimit: draft.accountLimit, modelLimit: draft.modelLimit,
        initialLimit: draft.initialLimit, groupKey: draft.groupKey || '', verified: draft.verified === true
      }, { signal: controller.signal });
      if (controller.signal.aborted || generation.current !== token) return;
      setState(result);
      setDraft(result.policy);
      setError('');
      setEditing(false);
    } catch (cause) {
      if (!controller.signal.aborted && generation.current === token) setError(cause.message || '保存失败');
    } finally {
      saving.current = false;
      if (!controller.signal.aborted && generation.current === token) setBusy(false);
    }
  };
  const patch = value => setDraft(previous => ({ ...previous, ...value }));
  const cooling = state?.coolingUntil && new Date(state.coolingUntil).getTime() > Date.now();
  return <section className="batch-factory-engine-field" style={{ gridColumn: '1 / -1', display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 8, minWidth: 0 }}>
    <Space wrap><b>出图并发</b><span>自动分配</span><Button size="small" loading={busy} disabled={!modelId || !active} onClick={() => setRefresh(value => value + 1)}>刷新</Button></Space>
    <span style={{ fontSize: 12 }}>所有用户和书籍共享通道额度，空闲时自动使用，繁忙时排队。</span>
    {!modelId ? <span>请先选择图片模型</span> : <>
      {error && <Alert type="error" showIcon message={error} description="状态读取或保存失败，请刷新后重试。" />}
      {state && !error && <Space wrap size={[16, 4]}>
        <span>全站占用 {state.running} / {state.policy.accountLimit}</span>
        <span>模型总占用 {state.modelTotalRunning} / {state.policy.modelTotalLimit}</span>
        <span>当前规格占用 {state.modelRunning} / {state.policy.modelLimit}</span>
        <span>当前规格并发 / 上限 {state.effectiveLimit} / {state.policy.modelLimit}</span>
        <span>图片尺寸 {state.policy.size || '读取中'} · {state.policy.references ? '带参考图' : '无参考图'} · 每次 {state.policy.count} 张</span>
        <span>排队 {state.waiting}</span>
        <span title="请求已发出但未确认完成，暂时保留名额以防超额；不会自动再生成扣费">待核对占用 {state.uncertain || 0}</span>
        <span>{cooling ? `冷却至 ${new Date(state.coolingUntil).toLocaleString()}` : '未冷却'}</span>
        {!state.policy.verified && <span>并发未实测</span>}
      </Space>}
      {!state && !error && <span>正在读取全站并发状态…</span>}
      {state?.canManage && draft && <details open={editing} onToggle={event => setEditing(event.currentTarget.open)}>
        <summary style={{ cursor: 'pointer' }}>管理员并发设置</summary>
        <Space direction="vertical" size={8} style={{ width: '100%', paddingTop: 8 }}>
          <span>账号分组：<span title={state.policy.accountKey}>{String(state.policy.accountKey || '').slice(0, 16)}{String(state.policy.accountKey || '').length > 16 ? '…' : ''}</span> · 模型：{state.policy.model}</span>
          <Space wrap>{[['accountLimit', '账号总上限'], ['modelLimit', '当前规格上限'], ['initialLimit', '初始并发']].map(([key, label]) => <label key={key}>{label} <InputNumber min={1} max={20} precision={0} value={draft[key]} onChange={value => patch({ [key]: value })} /></label>)}</Space>
          <label>共享账号分组 <Input value={draft.groupKey || ''} onChange={event => patch({ groupKey: event.target.value })} placeholder="留空使用自动账号分组" /></label>
          <span style={{ fontSize: 12 }}>同一上游账号的不同密钥可填写相同分组，共享账号总上限；修改会影响全站任务。降低上限后，已提交的图片会继续完成。</span>
          <Space><span>该请求规格已实测</span><Switch checked={draft.verified === true} onChange={verified => patch({ verified })} /></Space>
          <Button size="small" type="primary" loading={busy} disabled={!active || ![draft.accountLimit, draft.modelLimit, draft.initialLimit].every(value => Number.isInteger(value) && value > 0 && value <= 20) || draft.initialLimit > Math.min(draft.accountLimit, draft.modelLimit)} onClick={save}>保存并发设置</Button>
        </Space>
      </details>}
    </>}
  </section>;
}
