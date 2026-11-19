import { Button, Form, InputNumber, Select, Switch, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
import AgentRequirements from './AgentRequirements.jsx';
import ReviewDimensions from './ReviewDimensions.jsx';
import './AgentPanel.css';
import useWorkspaceScope from './useWorkspaceScope.js';
import { PHASES, THINKING, availableModel, modelOptions } from './contracts.js';

export default function AgentPanel({ projectId, models, mode }) {
  const [form] = Form.useForm(); const [settings, setSettings] = useState(null); const [busy, setBusy] = useState(false);
  const [dimensionCatalog, setDimensionCatalog] = useState([]);
  const saving = useRef(false), pauseTouched = useRef(false);
  const scope = useWorkspaceScope(projectId);
  useEffect(() => {
    scope.call(novelProjectRequest(projectId, '/settings', undefined, { signal: scope.signal }), r => {
      pauseTouched.current = false; setSettings(r.settings || {}); const a = r.settings?.agent || {};
      const catalog = Array.isArray(r.reviewDimensionCatalog) ? r.reviewDimensionCatalog : [];
      setDimensionCatalog(catalog);
      form.setFieldValue('reviewDimensionIds', a.reviewDimensionIds ?? catalog.filter(d => d.defaultEnabled).map(d => d.id));
      form.setFieldsValue({ ...a, reviewPrompts: a.reviewPrompts || [], customWritePrompts: a.customWritePrompts || [], targetChapterWords: a.targetChapterWords ?? 2000, maxRevisionRounds: a.maxRevisionRounds ?? 3,
        passScore: a.passScore ?? 80, baseTemperature: a.baseTemperature ?? 0.65, temperatureStep: a.temperatureStep ?? 0.15,
        thinkingLevel: a.thinkingLevel ?? 'standard', enableLengthNorm: a.enableLengthNorm ?? true,
        enablePostWrite: a.enablePostWrite ?? true, enableAiTells: a.enableAiTells ?? true, requireSecondaryConfirm: a.requireSecondaryConfirm ?? false, pauseAfterBlueprint: a.pauseAfterBlueprint ?? (mode ? ['long_original','long_imitation'].includes(mode) : true) });
    }).catch(e => message.error(e.message));
  }, [projectId]);
  async function save() {
    if (saving.current || !scope.isCurrent()) return;
    saving.current = true;
    try {
      const agent = await form.validateFields();
      if (!dimensionCatalog.length) delete agent.reviewDimensionIds;
      if (settings?.agent?.pauseAfterBlueprint == null && !pauseTouched.current) delete agent.pauseAfterBlueprint;
      if (Object.values(agent.modelRoutes || {}).some(id => id && !availableModel(models, id))) { message.warning('分步骤原模型已不可用，请重新选择或清空'); return; }
      setBusy(true);
      await scope.call(novelProjectRequest(projectId, '/settings', { styleProfile: settings?.styleProfile || '', rules: settings?.rules || [], agent }, { method: 'PUT', signal: scope.signal }), r => { setSettings(r.settings); message.success('Agent设置已保存'); });
    } catch (e) { if (!e.errorFields) message.error(e.message); } finally { saving.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  return <Form form={form} layout="vertical" className="novel-panel novel-agent-page" disabled={busy}>
    <div className="novel-agent-intro"><h2>Agent 设置</h2><p>为创作步骤指定自己的模型，设置审稿与写作要求。</p></div>
    <section className="novel-agent-section"><h3>模型路由</h3><p>留空则使用本次对话选择的模型。这里仅列出当前已接通的创作步骤。</p>
      <div className="novel-agent-routes">{PHASES.filter(([key])=>!['post-write','ai-tells'].includes(key)).map(([key, label]) => <Form.Item key={key} name={['modelRoutes', key]} label={label}>
        <Select aria-label={`${label}模型`} allowClear placeholder="使用本次选中模型" options={modelOptions(models, settings?.agent?.modelRoutes?.[key])} />
      </Form.Item>)}</div>
    </section>
    <section className="novel-agent-section"><h3>写章审稿循环</h3><p>写章后自动审稿。未达通过分将尝试修订，并保留审核结果较好的版本。</p>
      <div className="novel-form-grid">{[
        ['maxRevisionRounds', '最多修订轮数', 0, 5, 1], ['passScore', '审核通过分', 0, 100, 1],
        ['baseTemperature', '基础温度', 0, 2, 0.05], ['temperatureStep', '修订温度增量', 0, 0.5, 0.05]
      ].map(([key, label, min, max, step]) => <Form.Item key={key} name={key} label={label}><InputNumber aria-label={label} min={min} max={max} step={step} /></Form.Item>)}</div>
      <div className="novel-agent-switches">{[
        ['enableLengthNorm', '字数偏离超过45%时整理（调用模型）'], ['enablePostWrite', '后写校验（本地，不改正文）'],
        ['enableAiTells', '表达指标（本地统计，供人工判断）'], ['pauseAfterBlueprint', '大纲/蓝图生成后暂停']
      ].map(([key, label]) => <Form.Item key={key} name={key} label={label} valuePropName="checked"><Switch aria-label={label} onChange={() => { if (key === 'pauseAfterBlueprint') pauseTouched.current = true; }} /></Form.Item>)}</div>
      <p>后写校验与表达指标不调用模型。表达指标尚未验证与原版AI味算法一致。</p>
    </section>
    <p>大纲、蓝图或基础设定生成后暂停，可修改已保存内容后继续。旧作品未明确设置时保留原有暂停规则。</p>
    <section className="novel-agent-section"><h3>模型调用确认</h3><Form.Item name="requireSecondaryConfirm" label="每次后续模型调用前确认" valuePropName="checked"><Switch aria-label="每次后续模型调用前确认" /></Form.Item><p>此项保留旧设置的逐次调用保护，与上面的内容暂停分别生效。</p></section>
    <Form.Item name="reviewDimensionIds" noStyle><ReviewDimensions catalog={dimensionCatalog} disabled={busy || !settings} /></Form.Item>
    <AgentRequirements />
    <AgentRequirements writing />
    <section className="novel-agent-section"><h3>当前作品设置</h3><div className="novel-form-grid">
      <Form.Item name="targetChapterWords" label="每章目标字数"><InputNumber aria-label="每章目标字数" min={800} max={10000} step={100} /></Form.Item>
      <Form.Item name="thinkingLevel" label="默认思考强度"><Select options={THINKING} /></Form.Item>
    </div></section>
    <footer className="novel-agent-footer"><span>保存到当前作品，下一次新任务生效；历史任务保留原设置。</span><Button type="primary" disabled={!settings} loading={busy} onClick={save}>保存Agent设置</Button></footer>
  </Form>;
}
