import { Button, Empty, Input, Switch, Upload as AntUpload, message } from 'antd';
import { ImagePlus, Trash2, Upload } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

import { getComicReferencePreviewBlob, uploadComicGlobalReference } from '../../../shared/api/batchFactoryV11';
import { comicStylePreviewReference, comicStyleSelection, normalizeComicReference } from './batchFactoryComicStyle';

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
  const styles = useMemo(() => (Array.isArray(catalog) ? catalog : []).filter(item => item.slot === 'batch.comic-style'), [catalog]);
  const comic = value && typeof value === 'object' ? value : {};
  const reference = normalizeComicReference(comic.globalReferenceImage);
  const selectedId = String(comic.style?.presetId || '');
  const patch = next => onChange?.({ ...comic, ...next, enabled: true });
  const selectStyle = preset => patch({ style: comicStyleSelection(preset) });
  const usePreview = preset => {
    setLocalReferencePreview('');
    patch({ globalReferenceImage: comicStylePreviewReference(preset) });
  };
  const uploadReference = async file => {
    if (!batchId) return false;
    setUploading(true);
    try {
      const dataUrl = await fileAsDataURL(file);
      const result = await uploadComicGlobalReference(batchId, dataUrl);
      setLocalReferencePreview(dataUrl);
      patch({ globalReferenceImage: normalizeComicReference(result?.reference) });
      message.success('漫画全局参考图已上传');
    } catch (error) {
      message.error(error?.message || '漫画全局参考图上传失败');
    } finally {
      setUploading(false);
    }
    return false;
  };

  return <div className="batch-factory-comic-style-picker">
    <div className="batch-factory-comic-style-heading"><div><b>画风库</b><span>{inherited ? '继承统一配置' : '当前配置'}</span></div><small>画风提示词只在后端参与生成；预览图不会自动发给 AI。</small></div>
    {styles.length ? <div className="batch-factory-comic-style-grid" role="radiogroup" aria-label="漫画画风库">
      {styles.map(preset => {
        const selected = selectedId === preset.id;
        return <article className={`batch-factory-comic-style-card${selected ? ' is-selected' : ''}`} key={preset.id}>
          <button type="button" role="radio" aria-checked={selected} disabled={disabled} onClick={() => selectStyle(preset)}>
            <span className="batch-factory-comic-style-image"><AuthenticatedPreview src={preset.previewUrl} alt={`${preset.name}预览`} /></span>
            <span className="batch-factory-comic-style-copy"><b>{preset.name}</b><small>{preset.category || '漫画画风'}</small></span>
          </button>
          {selected ? <Button type="link" size="small" disabled={disabled} onClick={() => usePreview(preset)}>设为漫画全局参考图</Button> : null}
        </article>;
      })}
    </div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无已发布画风" />}

    <div className="batch-factory-comic-reference-panel">
      <div className="batch-factory-comic-reference-preview">
        {localReferencePreview || reference.previewUrl ? <AuthenticatedPreview src={localReferencePreview || reference.previewUrl} alt="漫画全局参考图" /> : <ImagePlus size={28} />}
      </div>
      <div className="batch-factory-comic-reference-controls">
        <div className="batch-factory-comic-reference-title"><b>漫画全局参考图</b><small>可来自画风预览图，也可单独上传。</small></div>
        <div className="batch-factory-comic-reference-actions">
          <AntUpload accept="image/png,image/jpeg,image/webp" showUploadList={false} beforeUpload={uploadReference} disabled={disabled || uploading}><Button icon={<Upload size={15} />} loading={uploading}>上传参考图</Button></AntUpload>
          {reference.assetId ? <Button aria-label="移除漫画全局参考图" icon={<Trash2 size={15} />} disabled={disabled} onClick={() => { setLocalReferencePreview(''); patch({ globalReferenceImage: null }); }} /> : null}
        </div>
        <div className="batch-factory-comic-reference-switches">
          <label><span>启用参考图</span><Switch size="small" checked={reference.enabled && Boolean(reference.assetId)} disabled={disabled || !reference.assetId} onChange={enabled => patch({ globalReferenceImage: { ...reference, enabled } })} /></label>
          <label><span>用于资产图</span><Switch size="small" checked={reference.applyToAssets} disabled={disabled || !reference.assetId} onChange={applyToAssets => patch({ globalReferenceImage: { ...reference, applyToAssets } })} /></label>
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
