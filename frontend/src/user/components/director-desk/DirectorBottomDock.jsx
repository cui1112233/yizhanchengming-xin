import { Camera, ImagePlus, Move3D, RotateCcw, Send, Upload } from 'lucide-react';
import { useLayoutEffect, useRef } from 'react';

export default function DirectorBottomDock({ prompt, saveStatus, canCapture, transformMode, onTransformModeChange, onPromptChange, onSavePrompt, onOpenAssets, onUploadReference, onResetView, onCapture }) {
  const textarea = useRef(null);
  useLayoutEffect(() => {
    const input = textarea.current;
    if (!input) return;
    input.style.height = 'auto';
    input.style.height = `${Math.min(input.scrollHeight, 112)}px`;
  }, [prompt]);

  const upload = event => {
    const input = event.currentTarget;
    const file = input.files?.[0];
    try {
      if (file) onUploadReference(file);
    } finally {
      input.value = '';
    }
  };

  return <section className="agent-director-bottom-dock" aria-label="导演台工具">
    <div className="agent-director-bottom-dock__tools">
      <button type="button" aria-label="移动" aria-pressed={transformMode === 'translate'} onClick={() => onTransformModeChange('translate')} title="移动"><Move3D size={17} /></button>
      <button type="button" aria-label="重置视角" onClick={onResetView} title="重置视角"><RotateCcw size={17} /></button>
      <button type="button" aria-label="截图到画布" onClick={onCapture} disabled={!canCapture} title="截图到画布"><Camera size={17} /></button>
    </div>
    <div className="agent-director-bottom-dock__composer">
      <button type="button" aria-label="项目图片" onClick={onOpenAssets} title="项目图片"><ImagePlus size={17} /></button>
      <label className="agent-director-bottom-dock__upload" title="上传参考图"><Upload size={17} /><input type="file" accept="image/*" aria-label="上传参考图" onChange={upload} /></label>
      <textarea ref={textarea} rows={1} value={prompt} onChange={event => onPromptChange(event.target.value)} aria-label="场景描述" placeholder="场景描述" />
      <button type="button" aria-label="保存场景描述" onClick={onSavePrompt} disabled={!prompt.trim() || saveStatus === 'saving'} title="保存场景描述"><Send size={17} /></button>
    </div>
    <span className="agent-director-bottom-dock__status" role="status">{saveStatus === 'saving' ? '正在保存' : saveStatus === 'error' ? '尚未同步' : '已保存'}</span>
  </section>;
}
