import { Button, Image, Spin, Typography } from 'antd';
import { ImagePlus, Sparkles, Trash2, Upload } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { generateReferenceAssetImage, getReferenceAssetImageTask, loadReferenceAssetImage, uploadReferenceAssetImage } from '../../shared/api/referenceAssets';
import { appendImageCandidateAndSelectSingle } from '../pages/scriptEntityImages';
import { createEntityImagePreviewLoader } from './entityImagePreviewLoader';
import { createEntityImageRequestGuard } from './entityImageRequestGuard';

function responseUrl(response) { return String(response?.url || response?.data?.url || response?.result?.url || '').trim(); }

export default function EntityImagePanel({ assetType, assetId, generationPayload, imageModelId, imageProvider, imageAspectRatio, imageUrls = [], mainImageUrl = '', onChange, requestKey = assetId, disabled = false }) {
  const [uploading, setUploading] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');
  const [generationStatus, setGenerationStatus] = useState('');
  const [previewUrls, setPreviewUrls] = useState({});
  const [failedPreviewUrls, setFailedPreviewUrls] = useState([]);
  const inputRef = useRef(null);
  const previewGeneration = useRef(0);
  const guard = useRef(createEntityImageRequestGuard(requestKey));
  guard.current.activate(requestKey);
  const busy = uploading || generating;

  useEffect(() => () => guard.current.invalidate(), []);
  useEffect(() => { setUploading(false); setGenerating(false); setError(''); setGenerationStatus(''); }, [requestKey]);
  useEffect(() => {
    const generation = previewGeneration.current + 1;
    previewGeneration.current = generation;
    const loader = createEntityImagePreviewLoader(imageUrls, { loadImage: loadReferenceAssetImage, createObjectUrl: URL.createObjectURL, revokeObjectUrl: URL.revokeObjectURL });
    loader.promise.then(result => {
      if (generation !== previewGeneration.current) return;
      setPreviewUrls(result.previews);
      setFailedPreviewUrls(result.failed);
    });
    return () => { previewGeneration.current += 1; loader.cancel(); };
  }, [imageUrls.join('\u0000')]);

  function update(next, token = guard.current.begin()) { guard.current.commit(token, () => onChange?.(next, token.key)); }
  function select(url) { update({ imageUrls, mainImageUrl: url }); }
  function remove(event, url) {
    event.stopPropagation();
    const nextUrls = imageUrls.filter(item => item !== url);
    update({ imageUrls: nextUrls, mainImageUrl: mainImageUrl === url ? nextUrls[0] || '' : mainImageUrl });
  }
  async function upload(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || disabled) return;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type) || file.size > 10 * 1024 * 1024) { setError('请选择不超过 10MB 的 PNG、JPG 或 WebP 图片。'); return; }
    const token = guard.current.begin();
    guard.current.commit(token, () => { setUploading(true); setError(''); });
    try {
      const dataUrl = await new Promise((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(reader.result); reader.onerror = () => reject(new Error('读取图片失败')); reader.readAsDataURL(file); });
      if (!guard.current.isCurrent(token)) return;
      const response = await uploadReferenceAssetImage({ asset_type: assetType, asset_id: assetId, variant: 'source', data_url: dataUrl });
      const url = responseUrl(response);
      if (!url) throw new Error('图片上传接口未返回图片地址');
      const next = appendImageCandidateAndSelectSingle(imageUrls, mainImageUrl, url);
      update(next, token);
    } catch (caught) { guard.current.commit(token, () => setError(caught.message || '图片上传失败，请稍后重试')); }
    finally { guard.current.commit(token, () => setUploading(false)); }
  }
  async function generate() {
    if (disabled) return;
    const token = guard.current.begin();
    guard.current.commit(token, () => { setGenerating(true); setError(''); setGenerationStatus('正在提交图片任务...'); });
    try {
      const response = await generateReferenceAssetImage({ ...generationPayload, ...(imageProvider ? { imageProvider } : {}), ...(imageModelId ? { imageModelId } : {}), ...(imageAspectRatio ? { aspect_ratio: imageAspectRatio } : {}) });
      let resolved = response;
      const taskId = String(response?.taskId || response?.data?.taskId || '').trim();
      if (taskId) {
        guard.current.commit(token, () => setGenerationStatus('图片正在后台生成...'));
        resolved = await waitForReferenceAssetImageTask(taskId, () => guard.current.isCurrent(token), status => {
          guard.current.commit(token, () => setGenerationStatus(status));
        });
      }
      const url = responseUrl(resolved);
      if (!url) throw new Error('图片生成接口未返回图片地址');
      update(appendImageCandidateAndSelectSingle(imageUrls, mainImageUrl, url), token);
      guard.current.commit(token, () => setGenerationStatus('图片已生成'));
    } catch (caught) { guard.current.commit(token, () => setError(caught.message || '人物/场景图片生成失败，请稍后重试')); }
    finally { guard.current.commit(token, () => setGenerating(false)); }
  }
  const mainPreview = previewUrls[mainImageUrl];
  const mainPreviewFailed = failedPreviewUrls.includes(mainImageUrl);
  return <div className="entity-editor-image-panel">
    <div className="entity-editor-image-panel-heading"><Typography.Text strong>主图预览</Typography.Text>{busy ? <Spin size="small" /> : null}</div>
    {mainImageUrl && mainPreview ? <div className="entity-editor-main-image"><Image src={mainPreview} alt="主图预览" preview={{ mask: '点击放大' }} /></div>
      : mainImageUrl && mainPreviewFailed ? <div className="entity-editor-image-empty"><ImagePlus size={34} /><span>图片读取失败，请重新上传</span></div>
        : mainImageUrl ? <div className="entity-editor-image-empty"><Spin /><span>主图加载中</span></div>
        : <div className="entity-editor-image-empty"><ImagePlus size={34} /><span>暂无主图</span></div>}
    <div className="entity-editor-image-actions"><input ref={inputRef} type="file" accept="image/png,image/jpeg,image/webp" hidden onChange={upload} disabled={disabled || busy} /><Button icon={<Upload size={15} />} onClick={() => inputRef.current?.click()} disabled={disabled || busy}>上传图片</Button><Button type="primary" icon={<Sparkles size={15} />} onClick={generate} loading={generating} disabled={disabled || busy}>AI生成</Button></div>
    <div className="entity-editor-image-grid" aria-label="图片缩略图">{imageUrls.map(url => <div className={`entity-editor-image-thumbnail${url === mainImageUrl ? ' is-main' : ''}`} key={url}><button className="entity-editor-image-select" type="button" onClick={() => select(url)} disabled={disabled || busy} aria-label="选择图片">{previewUrls[url] ? <img src={previewUrls[url]} alt="实体图片缩略图" /> : failedPreviewUrls.includes(url) ? <ImagePlus size={16} /> : <Spin size="small" />}</button><button className="entity-editor-image-delete-control" type="button" disabled={disabled || busy} aria-label="删除图片" onClick={event => remove(event, url)}><Trash2 size={14} /></button></div>)}</div>
    {generationStatus ? <Typography.Paragraph type="secondary" style={{ margin: '10px 0 0' }}>{generationStatus}</Typography.Paragraph> : null}
    {error ? <Typography.Paragraph type="danger" style={{ margin: '10px 0 0' }}>{error}</Typography.Paragraph> : null}
  </div>;
}

function wait(ms) { return new Promise(resolve => window.setTimeout(resolve, ms)); }

async function waitForReferenceAssetImageTask(taskId, isCurrent, onStatus) {
  for (;;) {
    if (!isCurrent()) throw new Error('图片任务已取消');
    const task = await getReferenceAssetImageTask(taskId);
    if (task?.status === 'succeeded') return task;
    if (task?.status === 'failed') throw new Error(task.error || '图片生成失败');
    onStatus('图片正在后台生成...');
    await wait(2500);
  }
}
