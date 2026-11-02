import { Button, Empty, Input, Switch, Upload as AntUpload, message } from 'antd';
import { Check, ChevronLeft, ChevronRight, ImagePlus, Trash2, Upload } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { getComicReferencePreviewBlob, uploadComicGlobalReference } from '../../../shared/api/batchFactory';
import {
  comicReferenceForSelectedStyle,
  comicReferenceNeedsStyleDefault,
  comicStyleCoverFlowStyle,
  comicStyleIndexFromDrag,
  comicStyleRelativeOffset,
  comicStyleSelection,
  comicStyleWheelDirection,
  normalizeComicReference,
  wrapComicStyleIndex
} from './batchFactoryComicStyle';

function fileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(new Error('参考图读取失败'));
    reader.readAsDataURL(file);
  });
}

function AuthenticatedPreview({ src, alt }) {
  const [displayURL, setDisplayURL] = useState('');
  useEffect(() => {
    let alive = true;
    let objectURL = '';
    setDisplayURL('');
    if (!src) return undefined;
    if (String(src).startsWith('data:') || /^https?:\/\//i.test(String(src))) {
      setDisplayURL(src);
      return undefined;
    }
    getComicReferencePreviewBlob(src).then(blob => {
      if (!alive) return;
      objectURL = URL.createObjectURL(blob);
      setDisplayURL(objectURL);
    }).catch(() => alive && setDisplayURL(''));
    return () => {
      alive = false;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [src]);
  return displayURL ? <img src={displayURL} alt={alt} loading="lazy" /> : <span className="batch-factory-comic-preview-loading"><ImagePlus size={22} /></span>;
}

export function BatchFactoryComicStylePicker({ batchId, catalog, value, onChange, inherited = false, disabled = false }) {
  const [uploading, setUploading] = useState(false);
  const [localReferencePreview, setLocalReferencePreview] = useState('');
  const [activeIndex, setActiveIndex] = useState(0);
  const dragStartX = useRef(null);
  const dragged = useRef(false);
  const wheelLocked = useRef(false);
  const wheelTimer = useRef(null);
  const styles = useMemo(() => (Array.isArray(catalog) ? catalog : []).filter(item => item.slot === 'batch.comic-style'), [catalog]);
  const comic = value && typeof value === 'object' ? value : {};
  const reference = normalizeComicReference(comic.globalReferenceImage);
  const selectedId = String(comic.style?.presetId || '');
  const selectedIndex = styles.findIndex(item => item.id === selectedId);
  const selectedPreset = selectedIndex >= 0 ? styles[selectedIndex] : null;
  const activePreset = styles[activeIndex] || selectedPreset || null;
  const patch = next => onChange?.({ ...comic, ...next, enabled: true });
  const selectStyle = preset => {
    setLocalReferencePreview('');
    patch({ style: comicStyleSelection(preset), globalReferenceImage: comicReferenceForSelectedStyle(comic.globalReferenceImage, preset) });
  };
  const uploadReference = async file => {
    if (!batchId) return false;
    setUploading(true);
    try {
      const dataUrl = await fileAsDataURL(file);
      const result = await uploadComicGlobalReference(batchId, dataUrl);
      setLocalReferencePreview(dataUrl);
      patch({ globalReferenceImage: normalizeComicReference({ ...result?.reference, source: 'manual-upload' }) });
      message.success('漫画全局参考图已上传');
    } catch (error) {
      message.error(error?.message || '漫画全局参考图上传失败');
    } finally {
      setUploading(false);
    }
    return false;
  };

  useEffect(() => {
    if (!styles.length) {
      setActiveIndex(0);
      return;
    }
    setActiveIndex(current => selectedIndex >= 0 ? selectedIndex : wrapComicStyleIndex(current, styles.length));
  }, [selectedIndex, styles.length]);

  useEffect(() => {
    if (inherited || !selectedPreset || !comicReferenceNeedsStyleDefault(comic.globalReferenceImage, selectedPreset)) return;
    patch({ globalReferenceImage: comicReferenceForSelectedStyle(comic.globalReferenceImage, selectedPreset) });
  }, [inherited, selectedPreset?.id, comic.globalReferenceImage?.assetId, comic.globalReferenceImage?.source]);

  useEffect(() => () => {
    if (wheelTimer.current) window.clearTimeout(wheelTimer.current);
  }, []);

  const selectIndex = index => {
    if (!styles.length || disabled) return;
    const nextIndex = wrapComicStyleIndex(index, styles.length);
    setActiveIndex(nextIndex);
    selectStyle(styles[nextIndex]);
  };
  const move = direction => selectIndex(activeIndex + direction);
  const handlePointerDown = event => {
    if (disabled) return;
    dragStartX.current = event.clientX;
    dragged.current = false;
    event.currentTarget.setPointerCapture?.(event.pointerId);
  };
  const finishDrag = event => {
    if (dragStartX.current == null) return;
    const deltaX = event.clientX - dragStartX.current;
    const nextIndex = comicStyleIndexFromDrag(activeIndex, deltaX, styles.length);
    dragged.current = nextIndex !== activeIndex;
    dragStartX.current = null;
    if (event.currentTarget.hasPointerCapture?.(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
    if (nextIndex !== activeIndex) selectIndex(nextIndex);
    window.setTimeout(() => { dragged.current = false; }, 0);
  };
  const cancelDrag = event => {
    dragStartX.current = null;
    dragged.current = false;
    if (event.currentTarget.hasPointerCapture?.(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
  };
  const handleWheel = event => {
    const direction = comicStyleWheelDirection(event.deltaX, event.deltaY, event.shiftKey);
    if (disabled || wheelLocked.current || !direction) return;
    event.preventDefault();
    wheelLocked.current = true;
    move(direction);
    wheelTimer.current = window.setTimeout(() => { wheelLocked.current = false; }, 260);
  };
  const handleKeyDown = event => {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
    event.preventDefault();
    move(event.key === 'ArrowRight' ? 1 : -1);
  };

  return <div className="batch-factory-comic-style-picker">
    <div className="batch-factory-comic-style-heading"><div><b>画风库</b><span>{selectedPreset ? `当前画风：${selectedPreset.name}` : (inherited ? '继承统一配置 · 尚未选择画风' : '尚未选择画风')}</span></div><small>居中的画风即为当前选择；画风库图片会默认作为参考图发给 AI。</small></div>
    {styles.length ? <div
      className="batch-factory-comic-coverflow"
      role="radiogroup"
      aria-label="漫画画风库"
      tabIndex={disabled ? -1 : 0}
      onKeyDown={handleKeyDown}
      onWheel={handleWheel}
    >
      <div className="batch-factory-comic-coverflow-backdrop" aria-hidden="true">
        {activePreset ? <AuthenticatedPreview src={activePreset.previewUrl} alt="" /> : null}
      </div>
      <button type="button" className="batch-factory-comic-coverflow-nav is-previous" aria-label="上一画风" disabled={disabled} onClick={() => move(-1)}><ChevronLeft size={20} /></button>
      <div className="batch-factory-comic-coverflow-stage" onPointerDown={handlePointerDown} onPointerUp={finishDrag} onPointerCancel={cancelDrag}>
        {styles.map((preset, index) => {
          const selected = selectedId === preset.id;
          const offset = comicStyleRelativeOffset(index, activeIndex, styles.length);
          return <article className={`batch-factory-comic-style-card${selected ? ' is-selected' : ''}`} style={comicStyleCoverFlowStyle(offset)} key={preset.id}>
            <button type="button" role="radio" aria-checked={selected} disabled={disabled} onClick={() => { if (!dragged.current) selectIndex(index); }}>
              <span className="batch-factory-comic-style-image"><AuthenticatedPreview src={preset.previewUrl} alt={`${preset.name}预览`} /></span>
              <span className="batch-factory-comic-style-copy"><b>{preset.name}</b><small>{preset.category || '漫画画风'}</small></span>
            </button>
            {selected ? <span className="batch-factory-comic-style-selected"><Check size={13} />已选择</span> : null}
          </article>;
        })}
      </div>
      <button type="button" className="batch-factory-comic-coverflow-nav is-next" aria-label="下一画风" disabled={disabled} onClick={() => move(1)}><ChevronRight size={20} /></button>
      <div className="batch-factory-comic-coverflow-caption" aria-live="polite">
        <div><b>{activePreset?.name || '尚未选择画风'}</b><small>{activePreset?.description || '选择一个漫画画风'}</small></div>
      </div>
    </div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无已发布画风" />}

    <div className="batch-factory-comic-reference-panel">
      <div className="batch-factory-comic-reference-preview">
        {localReferencePreview || reference.previewUrl ? <AuthenticatedPreview src={localReferencePreview || reference.previewUrl} alt="漫画全局参考图" /> : <ImagePlus size={28} />}
      </div>
      <div className="batch-factory-comic-reference-controls">
        <div className="batch-factory-comic-reference-title"><b>漫画全局参考图</b><small>默认使用当前画风库图片，也可以手动上传替换。</small></div>
        <div className="batch-factory-comic-reference-actions">
          <AntUpload accept="image/png,image/jpeg,image/webp" showUploadList={false} beforeUpload={uploadReference} disabled={disabled || uploading}><Button icon={<Upload size={15} />} loading={uploading}>上传参考图</Button></AntUpload>
          {reference.source === 'manual-upload' ? <Button aria-label="恢复画风库参考图" icon={<Trash2 size={15} />} disabled={disabled} onClick={() => { setLocalReferencePreview(''); patch({ globalReferenceImage: comicReferenceForSelectedStyle(null, selectedPreset) }); }} /> : null}
        </div>
        <div className="batch-factory-comic-reference-switches">
          <label><span>启用参考图</span><Switch size="small" checked={reference.enabled && Boolean(reference.assetId)} disabled={disabled || !reference.assetId} onChange={enabled => patch({ globalReferenceImage: { ...reference, enabled } })} /></label>
          <label><span>用于资产图（仅人物）</span><Switch size="small" checked={reference.applyToAssets} disabled={disabled || !reference.assetId} onChange={applyToAssets => patch({ globalReferenceImage: { ...reference, applyToAssets } })} /></label>
          <label><span>用于漫画页</span><Switch size="small" checked={reference.applyToPages} disabled={disabled || !reference.assetId} onChange={applyToPages => patch({ globalReferenceImage: { ...reference, applyToPages } })} /></label>
        </div>
      </div>
    </div>

    <div className="batch-factory-comic-prompt-affixes">
      <label><span><b>图片描述词全局前缀</b></span><Input.TextArea rows={3} value={comic.imagePromptPrefix || ''} disabled={disabled} onChange={event => patch({ imagePromptPrefix: event.target.value })} /></label>
      <label><span><b>图片描述词全局后缀</b></span><Input.TextArea rows={3} value={comic.imagePromptSuffix || ''} disabled={disabled} onChange={event => patch({ imagePromptSuffix: event.target.value })} /></label>
    </div>
  </div>;
}
