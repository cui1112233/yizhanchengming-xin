import { Camera, ImagePlus, Maximize2, Move3D, Rotate3D, RotateCcw, Send, Upload } from 'lucide-react';
import { useLayoutEffect, useRef } from 'react';

function resizePrompt(input) {
  input.style.height = 'auto';
  input.style.height = `${Math.min(input.scrollHeight, 112)}px`;
}

export default function DirectorBottomDock({ prompt, shotPrompt = '', saveStatus, canCapture, transformMode, onTransformModeChange, onPromptChange, onShotPromptChange, onSavePrompt, onOpenAssets, onUploadReference, onResetView, onCapture }) {
  const textarea = useRef(null);
  useLayoutEffect(() => {
    const input = textarea.current;
    if (!input) return;
    resizePrompt(input);
  }, [prompt]);
  useLayoutEffect(() => {
    const input = textarea.current;
    if (!input) return;
    let width = input.getBoundingClientRect().width;
    const observer = new ResizeObserver(() => {
      const nextWidth = input.getBoundingClientRect().width;
      if (nextWidth === width) return;
      width = nextWidth;
      resizePrompt(input);
    });
    observer.observe(input);
    const fonts = document.fonts;
    const onFontsLoaded = () => resizePrompt(input);
    fonts?.addEventListener('loadingdone', onFontsLoaded);
    let active = true;
    fonts?.ready.then(() => { if (active) onFontsLoaded(); });
    return () => {
      active = false;
      observer.disconnect();
      fonts?.removeEventListener('loadingdone', onFontsLoaded);
    };
  }, []);

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
      <button type="button" aria-label="旋转" aria-pressed={transformMode === 'rotate'} onClick={() => onTransformModeChange('rotate')} title="旋转"><Rotate3D size={17} /></button>
      <button type="button" aria-label="缩放" aria-pressed={transformMode === 'scale'} onClick={() => onTransformModeChange('scale')} title="缩放"><Maximize2 size={17} /></button>
      <button type="button" aria-label="重置视角" onClick={onResetView} title="重置视角"><RotateCcw size={17} /></button>
      <button type="button" aria-label="截图到画布" onClick={onCapture} disabled={!canCapture} title="截图到画布"><Camera size={17} /></button>
    </div>
    <div className="agent-director-bottom-dock__composer">
      <button type="button" aria-label="项目图片" onClick={onOpenAssets} title="项目图片"><ImagePlus size={17} /></button>
      <label className="agent-director-bottom-dock__upload" title="上传参考图"><Upload size={17} /><input type="file" accept="image/*" aria-label="上传参考图" onChange={upload} /></label>
      <textarea ref={textarea} rows={1} value={prompt} onChange={event => onPromptChange(event.target.value)} aria-label="场景描述" placeholder="场景描述" />
      <textarea rows={1} value={shotPrompt} onChange={event => onShotPromptChange?.(event.target.value)} aria-label="分镜视频提示词" placeholder="分镜视频提示词" />
      <button type="button" aria-label="保存场景描述和分镜视频提示词" onClick={onSavePrompt} disabled={(!prompt.trim() && !shotPrompt.trim()) || saveStatus === 'saving'} title="保存场景描述和分镜视频提示词"><Send size={17} /></button>
    </div>
    <span className="agent-director-bottom-dock__status" role="status">{saveStatus === 'saving' ? '正在保存' : saveStatus === 'error' ? '尚未同步' : '已保存'}</span>
  </section>;
}
