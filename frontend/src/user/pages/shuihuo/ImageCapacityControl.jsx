import React, { useEffect, useRef, useState } from 'react';
import { Alert, Button, Space } from 'antd';
import { getImageCapacity } from '../../../shared/api/imageCapacity';

// 出图并发完全由系统自动判定：起始 20 路、上限 30 路，遇到限流自动降档、
// 恢复后逐级回升。这里只做只读展示，不再提供人工输入。
export default function ImageCapacityControl({ modelId, aspectRatio = '9:16', active = true }) {
  const [state, setState] = useState(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [refresh, setRefresh] = useState(0);
  const request = useRef(null);
  const generation = useRef(0);

  useEffect(() => {
    setState(null);
    setError('');
  }, [modelId, aspectRatio]);

  useEffect(() => {
    const token = ++generation.current;
    if (!active || !modelId) return undefined;
    let disposed = false;
    const load = async () => {
      request.current?.abort();
      const controller = new AbortController();
      request.current = controller;
      setBusy(true);
      try {
        const result = await getImageCapacity(modelId, { aspectRatio, signal: controller.signal });
        if (disposed || controller.signal.aborted || generation.current !== token) return;
        setState(result);
        setError('');
      } catch (cause) {
        if (!disposed && !controller.signal.aborted && generation.current === token) setError((cause.message || '无法读取图片并发状态').replace(/^capability unavailable\s*[:：]\s*/i, ''));
      } finally {
        if (!disposed && !controller.signal.aborted && generation.current === token) setBusy(false);
      }
    };
    load();
    const timer = setInterval(load, 10000);
    return () => { disposed = true; clearInterval(timer); request.current?.abort(); };
  }, [modelId, aspectRatio, active, refresh]);

  const cooling = state?.coolingUntil && new Date(state.coolingUntil).getTime() > Date.now();
  return <section className="batch-factory-engine-field" style={{ gridColumn: '1 / -1', display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 8, minWidth: 0 }}>
    <Space wrap><b>出图并发</b><span>系统自动调节</span><Button size="small" loading={busy} disabled={!modelId || !active} onClick={() => setRefresh(value => value + 1)}>刷新</Button></Space>
    <span style={{ fontSize: 12 }}>并发上限由系统按上游真实成功率自动升降，所有用户和书籍共享同一份额度，无需手动设置。</span>
    {!modelId ? <span>请先选择图片模型</span> : <>
      {error && <Alert type="error" showIcon message={error} />}
      {state && !error && <Space wrap size={[16, 4]}>
        <span><b>总并发 {state.modelTotalRunning} / 上限 {state.policy.modelTotalLimit}</b></span>
        <span>当前规格 {state.modelRunning} / {state.effectiveLimit}</span>
        <span>全站占用 {state.running} / {state.policy.accountLimit}</span>
        <span>排队 {state.waiting}</span>
        <span>待核对占用 {state.uncertain || 0}</span>
        <span>{cooling ? `限流冷却至 ${new Date(state.coolingUntil).toLocaleString()}，已自动降档` : '自动调节中'}</span>
      </Space>}
      {!state && !error && <span>正在读取全站并发状态…</span>}
    </>}
  </section>;
}
