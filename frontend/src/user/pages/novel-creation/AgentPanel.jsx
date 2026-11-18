import { Button, Form, InputNumber, Select, Switch, message } from 'antd';
import { useEffect, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { PHASES, THINKING, availableModel, modelOptions } from './contracts.js';

export default function AgentPanel({ projectId, models }) {
  const [form] = Form.useForm(); const [settings, setSettings] = useState(null); const [busy, setBusy] = useState(false);
  const scope = useWorkspaceScope(projectId);
  useEffect(() => {
    scope.call(novelProjectRequest(projectId, '/settings', undefined, { signal: scope.signal }), r => {
      setSettings(r.settings || {}); const a = r.settings?.agent || {};
      form.setFieldsValue({ ...a, targetChapterWords: a.targetChapterWords ?? 2000, maxRevisionRounds: a.maxRevisionRounds ?? 3,
        passScore: a.passScore ?? 80, baseTemperature: a.baseTemperature ?? 0.65, temperatureStep: a.temperatureStep ?? 0.15,
        thinkingLevel: a.thinkingLevel ?? 'standard', enableLengthNorm: a.enableLengthNorm ?? true,
        enablePostWrite: a.enablePostWrite ?? true, enableAiTells: a.enableAiTells ?? true, requireSecondaryConfirm: a.requireSecondaryConfirm ?? false });
    }).catch(e => message.error(e.message));
  }, [projectId]);
  async function save() {
    if (busy || !scope.isCurrent()) return;
    try {
      const agent = await form.validateFields();
      if (Object.values(agent.modelRoutes || {}).some(id => id && !availableModel(models, id))) { message.warning('分步骤原模型已不可用，请重新选择或清空'); return; }
      setBusy(true);
      await scope.call(novelProjectRequest(projectId, '/settings', { ...settings, agent }, { method: 'PUT', signal: scope.signal }), r => { setSettings(r.settings); message.success('Agent设置已保存'); });
    } catch (e) { if (!e.errorFields) message.error(e.message); } finally { if (scope.isCurrent()) setBusy(false); }
  }
  return <Form form={form} layout="vertical" className="novel-panel">
    <p>分步骤留空使用本次选择的模型。保存后影响下一次新任务，历史任务保留原设置。</p>
    <div className="novel-form-grid">{[
      ['targetChapterWords', '每章目标字数', 800, 10000, 100], ['maxRevisionRounds', '最多修订轮数', 0, 5, 1],
      ['passScore', '审核通过分', 0, 100, 1], ['baseTemperature', '基础温度', 0, 2, 0.05], ['temperatureStep', '修订温度增量', 0, 0.5, 0.05]
    ].map(([key, label, min, max, step]) => <Form.Item key={key} name={key} label={label}><InputNumber aria-label={label} min={min} max={max} step={step} /></Form.Item>)}</div>
    <Form.Item name="thinkingLevel" label="默认思考强度"><Select options={THINKING} /></Form.Item>
    {[
      ['enableLengthNorm', '字数偏离超过45%时整理（调用模型）'], ['enablePostWrite', '后写校验（本地，不改正文）'], ['enableAiTells', '表达指标（本地统计，供人工判断）'], ['requireSecondaryConfirm', '每次后续模型调用前确认']
    ].map(([key, label]) => <Form.Item key={key} name={key} label={label} valuePropName="checked"><Switch aria-label={label} /></Form.Item>)}
    <p>后写校验与表达指标不调用模型。表达指标尚未验证与原版AI味算法一致。</p><h3>分步骤模型</h3>{PHASES.filter(([key])=>!['post-write','ai-tells'].includes(key)).map(([key, label]) => <Form.Item key={key} name={['modelRoutes', key]} label={label}>
      <Select aria-label={`${label}模型`} allowClear placeholder="使用本次选中模型" options={modelOptions(models, settings?.agent?.modelRoutes?.[key])} />
    </Form.Item>)}
    <Button type="primary" disabled={!settings} loading={busy} onClick={save}>保存Agent设置</Button>
  </Form>;
}
