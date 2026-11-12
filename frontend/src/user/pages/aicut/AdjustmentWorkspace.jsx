import { Button, Empty, InputNumber, Slider } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { normalizeVisualProcessing } from '../aicutVisualProcessing.js';
import HSLColorEditor from './HSLColorEditor.jsx';
import ColorGradeEditor from './ColorGradeEditor.jsx';
import './workspace-panels.css';

const colorControls = [
  ['exposure', '曝光', -2, 2, .1],
  ['temperature', '色温', -1, 1, .05],
  ['tint', '色调', -1, 1, .05],
  ['highlights', '高光', -1, 1, .05],
  ['shadows', '阴影', -1, 1, .05],
  ['gamma', '伽马', .5, 2, .05],
];
const detailControls = [['sharpen', '锐化', 0, 1, .05], ['vignette', '暗角', 0, 1, .05], ['grain', '颗粒', 0, 1, .05]];

export default function AdjustmentWorkspace({ selection, onChange, onPickVisual }) {
  const clip = selection?.clip;
  const canAdjust = ['video', 'image'].includes(clip?.kind) && !selection?.track?.locked;
  const signature = JSON.stringify(normalizeVisualProcessing(clip?.visualProcessing));
  const selectionKey = `${selection?.track?.id || ''}:${clip?.id || ''}`;
  const [draft, setDraft] = useState(() => JSON.parse(signature));
  const latest = useRef(draft);
  const committed = useRef(signature);
  useEffect(() => {
    const next = JSON.parse(signature);
    latest.current = next;
    committed.current = signature;
    setDraft(next);
  }, [selectionKey, signature]);

  const update = (next, complete = false) => {
    const normalized = normalizeVisualProcessing(next);
    latest.current = normalized;
    setDraft(normalized);
    const nextSignature = JSON.stringify(normalized);
    if (complete && canAdjust && nextSignature !== committed.current) {
      committed.current = nextSignature;
      onChange?.(normalized);
    }
  };
  const patch = (section, value, complete = false) => update({ ...latest.current, [section]: { ...latest.current[section], ...value } }, complete);
  const complete = () => update(latest.current, true);
  const reset = () => {
    const defaults = normalizeVisualProcessing({});
    update({ ...latest.current, color: defaults.color, details: defaults.details }, true);
  };
  const renderControl = (section, [key, label, min, max, step]) => <div className="aicut-adjustment-control" key={key}>
    <label htmlFor={`aicut-adjustment-${key}`}>{label}</label>
    <InputNumber id={`aicut-adjustment-${key}`} aria-label={`${label}数值`} size="small" min={min} max={max} step={step} precision={2} value={draft[section][key]} onChange={value => {
      if (typeof value === 'number' && Number.isFinite(value)) patch(section, { [key]: value });
    }} onBlur={complete} onPressEnter={complete} />
    <Slider min={min} max={max} step={step} value={draft[section][key]} ariaLabelForHandle={label} onChange={value => patch(section, { [key]: value })} onChangeComplete={value => patch(section, { [key]: value }, true)} />
  </div>;

  return <section className="aicut-workspace-panel aicut-adjustment-workspace" aria-label="画面调节">
    <header className="aicut-workspace-heading"><strong>调节</strong>{canAdjust ? <Button type="text" size="small" onClick={reset}>重置调节</Button> : null}</header>
    {!canAdjust ? <div className="aicut-adjustment-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="选择时间线中的视频或图片，开始调节画面。" /><Button block onClick={onPickVisual} disabled={!onPickVisual}>选择画面片段</Button></div> : <>
      <p className="aicut-adjustment-clip" title={clip.name || clip.label || clip.id}>{clip.name || clip.label || '已选画面片段'}</p>
      <section className="aicut-adjustment-section"><h3>基础调色</h3>{colorControls.map(control => renderControl('color', control))}</section>
      <section className="aicut-adjustment-section"><ColorGradeEditor key={`${selectionKey}:grade`} value={draft.color} onChange={color => patch('color', color)} onChangeComplete={color => patch('color', color, true)} /></section>
      <section className="aicut-adjustment-section"><h3>HSL 分色</h3><HSLColorEditor key={selectionKey} value={draft.color.hsl} onChange={hsl => patch('color', { hsl })} onChangeComplete={hsl => patch('color', { hsl }, true)} /></section>
      <section className="aicut-adjustment-section"><h3>画面细节</h3>{detailControls.map(control => renderControl('details', control))}</section>
    </>}
  </section>;
}
