import { Alert, Button, Input, InputNumber, Modal, Select } from 'antd';
import { useMemo, useState } from 'react';
import { planProjectTemplate } from './projectTemplates.js';

export default function ProjectTemplateDialog({ template, timeline, assets, startMs = 0, onApply, onClose }) {
  const [options, setOptions] = useState({ visualAssetId: '', audioAssetId: '', title: '', startMs, durationMs: 5000 });
  const review = useMemo(() => {
    try { return { plan: planProjectTemplate(timeline, template.id, assets, options) }; }
    catch (error) { return { error: error.message }; }
  }, [timeline, template.id, assets, options]);
  const patch = (value) => setOptions(current => ({ ...current, ...value }));
  const assetOptions = (kinds) => assets.filter(asset => kinds.includes(asset.kind)).map(asset => ({ value: asset.id, label: asset.name }));
  return <Modal open title={`应用工程模板 · ${template.labelZh}`} width={580} onCancel={onClose} footer={<><Button onClick={onClose}>取消</Button><Button type="primary" disabled={!review.plan} onClick={() => onApply({ templateId: template.id, options, reviewedTimeline: timeline })}>确认应用模板</Button></>}>
    <div className="aicut-template-form">
      <p>{template.description}。新内容会加入独立轨道，保留工程中现有的片段。</p>
      <label>画面素材<Select aria-label="模板画面素材" placeholder="选择本工程的视频或图片" value={options.visualAssetId || undefined} showSearch optionFilterProp="label" options={assetOptions(['video', 'image'])} onChange={visualAssetId => patch({ visualAssetId })} /></label>
      <label>背景音频<Select aria-label="模板背景音频" allowClear placeholder="可选，使用本工程音频" value={options.audioAssetId || undefined} options={assetOptions(['audio'])} onChange={audioAssetId => patch({ audioAssetId: audioAssetId || '' })} /></label>
      <label>标题或字幕<Input.TextArea aria-label="模板标题" value={options.title} placeholder="输入这段内容要显示的文字" maxLength={500} autoSize={{ minRows: 2, maxRows: 4 }} onChange={event => patch({ title: event.target.value })} /></label>
      <div className="aicut-template-timing"><label>开始位置<InputNumber aria-label="模板开始位置" min={0} max={86400} step={.1} value={options.startMs / 1000} addonAfter="秒" onChange={seconds => patch({ startMs: Math.round(Number(seconds || 0) * 1000) })} /></label><label>模板时长<InputNumber aria-label="模板时长" min={.1} max={600} step={.1} value={options.durationMs / 1000} addonAfter="秒" onChange={seconds => patch({ durationMs: Math.round(Number(seconds || 0) * 1000) })} /></label></div>
      {review.error ? <Alert type="info" showIcon message={review.error} /> : <div className="aicut-template-review" aria-label="模板变更预览"><strong>将新增 {review.plan.summary.trackCount} 条轨道、{review.plan.summary.clipCount} 个片段</strong><p>{review.plan.summary.trackNames.join('、')}</p><p>开始 {(review.plan.summary.startMs / 1000).toFixed(2)} 秒 · 时长 {(review.plan.summary.durationMs / 1000).toFixed(2)} 秒</p>{review.plan.summary.warnings?.map((warning, index) => <p key={index}>{warning}</p>)}<small>应用后可调整素材、文字和参数，也可以撤销整次操作。</small></div>}
    </div>
  </Modal>;
}
