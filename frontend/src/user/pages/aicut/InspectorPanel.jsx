import { Input, InputNumber, Slider } from 'antd';
import { PanelRightClose, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { catalogById } from './editorCatalog.js';
import { normalizeTransform } from './canvasTransform.js';

function kindTitle(clip) {
  if (!clip) return '未选择对象';
  if (clip.kind === 'audio') return '音频';
  if (clip.kind === 'subtitle') return '文字与字幕';
  if (clip.kind === 'effect') return '特效';
  if (clip.transition) return '转场';
  return '画面';
}

export default function InspectorPanel({ selection, drawerMode = false, onClose, onCollapse, onPatchClip, onPatchEffect, onPatchTransition }) {
  const clip = selection?.clip;
  const [draft, setDraft] = useState(() => ({ ...(clip || {}), canvasTransform: normalizeTransform(clip?.canvasTransform) }));
  useEffect(() => setDraft({ ...(clip || {}), canvasTransform: normalizeTransform(clip?.canvasTransform) }), [clip]);
  const resourceId = clip?.catalogResourceId || clip?.effectResourceId || clip?.transition || '';
  const resource = resourceId ? catalogById(resourceId) : null;
  const commitNumber = (key, value) => onPatchClip({ [key]: Number(value || 0) });
  const patchTransform = (patch, commit = false) => {
    const canvasTransform = normalizeTransform({ ...draft.canvasTransform, ...patch });
    setDraft(current => ({ ...current, canvasTransform }));
    if (commit) onPatchClip({ canvasTransform });
  };

  return <aside className="aicut-context-inspector" aria-label="属性栏">
    <header><strong>{kindTitle(clip)}</strong>{drawerMode ? <button type="button" aria-label="关闭属性栏" onClick={onClose}><X size={15} /></button> : <button type="button" aria-label="收起属性栏" onClick={onCollapse}><PanelRightClose size={15} /></button>}</header>
    {!clip ? <div className="aicut-inspector-placeholder"><strong>未选择对象</strong><span>在时间线或播放器中选择片段后，这里会显示可编辑参数。</span></div> : <div className="aicut-inspector-sections">
      <section><h3>时间</h3><label>开始时间<InputNumber value={draft.startMs} min={0} onChange={(value) => setDraft(current => ({ ...current, startMs: value }))} onBlur={() => commitNumber('startMs', draft.startMs)} onPressEnter={() => commitNumber('startMs', draft.startMs)} addonAfter="ms" /></label><label>片段时长<InputNumber value={draft.durationMs} min={1} onChange={(value) => setDraft(current => ({ ...current, durationMs: value }))} onBlur={() => commitNumber('durationMs', draft.durationMs)} onPressEnter={() => commitNumber('durationMs', draft.durationMs)} addonAfter="ms" /></label></section>
      {['video', 'image'].includes(clip.kind) ? <section><h3>位置与大小</h3><div className="aicut-inspector-grid"><label>水平位置<InputNumber value={draft.canvasTransform.x} min={-1} max={1} step={.01} onChange={(value) => patchTransform({ x: value })} onBlur={() => patchTransform({}, true)} /></label><label>垂直位置<InputNumber value={draft.canvasTransform.y} min={-1} max={1} step={.01} onChange={(value) => patchTransform({ y: value })} onBlur={() => patchTransform({}, true)} /></label><label>横向缩放<InputNumber value={draft.canvasTransform.scaleX} min={.05} max={20} step={.05} onChange={(value) => patchTransform({ scaleX: value })} onBlur={() => patchTransform({}, true)} /></label><label>纵向缩放<InputNumber value={draft.canvasTransform.scaleY} min={.05} max={20} step={.05} onChange={(value) => patchTransform({ scaleY: value })} onBlur={() => patchTransform({}, true)} /></label><label>旋转<InputNumber value={draft.canvasTransform.rotation} min={-180} max={180} onChange={(value) => patchTransform({ rotation: value })} onBlur={() => patchTransform({}, true)} addonAfter="°" /></label></div><label>不透明度<Slider min={0} max={1} step={.01} value={draft.canvasTransform.opacity} onChange={(value) => patchTransform({ opacity: value })} onPointerUp={() => patchTransform({}, true)} /></label></section> : null}
      {clip.kind === 'subtitle' ? <section><h3>文字与字幕</h3><label>文字内容<Input.TextArea value={draft.text || ''} rows={4} onChange={(event) => setDraft(current => ({ ...current, text: event.target.value }))} onBlur={() => onPatchClip({ text: draft.text || '' })} /></label></section> : null}
      {clip.kind === 'audio' ? <section><h3>音频</h3><label>音量<Slider min={0} max={2} step={.05} value={Number(draft.volume ?? 1)} onChange={(volume) => setDraft(current => ({ ...current, volume }))} onPointerUp={() => onPatchClip({ volume: Number(draft.volume ?? 1) })} /></label></section> : null}
      {clip.transition ? <section><h3>转场</h3><label>持续时间<InputNumber min={100} max={3000} value={draft.transitionMs || 500} onChange={(transitionMs) => setDraft(current => ({ ...current, transitionMs }))} onBlur={() => onPatchTransition({ transitionMs: Number(draft.transitionMs || 500) })} addonAfter="ms" /></label></section> : null}
      {clip.kind === 'effect' ? <section><h3>特效</h3><label>强度<Slider min={0} max={1} step={.05} value={Number(draft.parameters?.intensity ?? .5)} onChange={(intensity) => setDraft(current => ({ ...current, parameters: { ...(current.parameters || {}), intensity } }))} onPointerUp={() => onPatchEffect({ parameters: draft.parameters })} /></label></section> : null}
      {resourceId && !resource ? <section><h3>未知资源</h3><p>资源 ID：{resourceId}</p><small>当前版本只读保留该资源，不会擅自删除工程数据。</small></section> : null}
    </div>}
  </aside>;
}
