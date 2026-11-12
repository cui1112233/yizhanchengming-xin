import { Alert, Button, Empty, Input, InputNumber, Modal, Select, Space, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { listPersonalTemplates, savePersonalTemplate, deletePersonalTemplate } from '../../../shared/api/aicutPersonalTemplates.js';
import { extractPersonalTemplate, planPersonalTemplate } from './personalTemplates.js';

export default function PersonalTemplateWorkspace({ timeline, assets = [], disabled = false, onApply }) {
  const [templates, setTemplates] = useState([]);
  const [title, setTitle] = useState('');
  const [selectedId, setSelectedId] = useState('');
  const [bindings, setBindings] = useState({});
  const [startMs, setStartMs] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [review, setReview] = useState(null);
  const selected = templates.find(item => item.id === selectedId);
  async function refresh() { setBusy(true); setError(''); try { setTemplates(await listPersonalTemplates()); } catch (e) { setError(e.message); } finally { setBusy(false); } }
  useEffect(() => { let alive = true; listPersonalTemplates().then(items => { if (alive) setTemplates(items); }).catch(e => { if (alive) setError(e.message); }); return () => { alive = false; }; }, []);
  useEffect(() => { setReview(null); }, [timeline, assets, selectedId, bindings, startMs]);
  async function save() {
    setBusy(true); setError(''); setNotice('');
    try { const item = await savePersonalTemplate(title, extractPersonalTemplate(timeline)); setTemplates(old => [item, ...old]); setSelectedId(item.id); setBindings({}); setTitle(''); setNotice('已保存到当前账号，可在其他工程重新选择素材后使用。'); }
    catch(e) { setError(e.message); } finally { setBusy(false); }
  }
  function preview() { setError(''); try { setReview({ revision: timeline.revision, plan: planPersonalTemplate(timeline, selected, assets, { bindings, startMs }) }); } catch(e) { setError(e.message); } }
  async function apply() {
    setError(''); setBusy(true);
    try {
      if (!review || timeline.revision !== review.revision) throw new Error('时间线已变化，请重新检查模板');
      planPersonalTemplate(timeline, selected, assets, { bindings, startMs });
      const result = await onApply({ type: 'timeline.personal_template_apply', payload: { template: selected, assets, options: { bindings: { ...bindings }, startMs } } });
      if (result === false) throw new Error('模板未应用，请检查时间线后重试');
      setReview(null);
    } catch(e) { setError(e.message); } finally { setBusy(false); }
  }
  function remove() {
    Modal.confirm({ title: `删除“${selected.title}”？`, content: '仅删除此个人模板，工程内容不会改变。', okText: '删除模板', cancelText: '取消', onOk: async () => { try { await deletePersonalTemplate(selected.id); setTemplates(old => old.filter(item => item.id !== selected.id)); setSelectedId(''); setReview(null); } catch(e) { setError(e.message); throw e; } } });
  }
  return <section aria-label="个人工程模板" style={{ padding: 12 }}>
    <Typography.Title level={5}>我的模板</Typography.Title>
    <Typography.Paragraph type="secondary">保存当前画面、音频、字幕的排布和样式。模板不保存原素材地址；使用时需逐项选择本工程素材。暂不支持转场、复合片段、联动、人物蒙版、私有字体和变速曲线。</Typography.Paragraph>
    {error && <Alert type="error" showIcon message={error} />}
    {notice && <Alert type="success" message={notice} />}
    <Space.Compact style={{ width: '100%', margin: '8px 0' }}><Input aria-label="个人模板名称" placeholder="输入模板名称" maxLength={80} value={title} disabled={disabled || busy} onChange={event => setTitle(event.target.value)} /><Button loading={busy} disabled={disabled || !title.trim()} onClick={save}>保存当前排布</Button></Space.Compact>
    <Button size="small" disabled={busy} onClick={refresh}>刷新我的模板</Button>
    {!templates.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有个人模板" /> : <>
      <Select aria-label="选择个人模板" style={{ width: '100%', margin: '8px 0' }} placeholder="选择我的模板" value={selectedId || undefined} options={templates.map(t => ({ value: t.id, label: t.title }))} disabled={disabled || busy} onChange={id => { setSelectedId(id); setBindings({}); setError(''); setNotice(''); }} />
      {selected && <Space direction="vertical" style={{ width: '100%' }}>
        {selected.layout.slots.map(slot => <label key={slot.id} style={{ display: 'block' }}>{slot.label}（{slot.kind === 'audio' ? '音频' : slot.kind === 'image' ? '图片' : '视频'}）<Select aria-label={`替换${slot.label}`} style={{ width: '100%' }} placeholder="请明确选择替换素材" value={bindings[slot.id]} disabled={disabled || busy} options={assets.filter(a => a.kind === slot.kind).map(a => ({ value: a.id, label: a.name || a.id }))} onChange={id => setBindings(old => ({ ...old, [slot.id]: id }))} /></label>)}
        <label>插入时间（毫秒） <InputNumber aria-label="个人模板插入时间" min={0} max={86399999} precision={0} value={startMs} disabled={disabled || busy} onChange={setStartMs} /></label>
        <Space><Button type="primary" disabled={disabled || busy} onClick={preview}>检查并预览应用</Button><Button danger disabled={disabled || busy} onClick={remove}>删除模板</Button></Space>
      </Space>}
    </>}
    <Modal title="确认应用个人模板" open={!!review} okText="新增到时间线" cancelText="返回修改" confirmLoading={busy} onCancel={() => setReview(null)} onOk={apply} okButtonProps={{ disabled: disabled || !onApply }}>
      {review && <><p>“{selected?.title}”将新增 {review.plan.summary.trackCount} 条轨道、{review.plan.summary.clipCount} 个片段，持续 {review.plan.summary.durationMs} 毫秒。原工程内容保留，可一次撤销。</p><ul>{review.plan.summary.assets.map(asset => <li key={asset.id}>{asset.name || asset.id}</li>)}</ul><p>素材范围已经检查，时长不足会拒绝应用。</p></>}
    </Modal>
  </section>;
}
