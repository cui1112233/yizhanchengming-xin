import { Button, Select, Upload } from 'antd';

const stateLabels = { queued: '排队中', running: '识别中', cancel_requested: '正在取消', succeeded: '已写入字幕', failed: '识别失败', canceled: '已取消' };

export default function CaptionWorkspace({ assets = [], models = [], jobs = [], assetId, modelId, language = 'zh', busy, onAssetChange, onModelChange, onLanguageChange, onStart, onCancel, onRetry, onImportSRT, onExportSRT, onAddManualCaption }) {
  const activeJob = jobs.find(job => ['queued', 'running', 'cancel_requested'].includes(job.state));
  const canStart = Boolean(assetId && modelId && language && !activeJob && !busy);
  return <section className="aicut-caption-workspace">
    <header><strong>识别字幕</strong><span>识别结果会新建字幕轨，不覆盖已有字幕。</span></header>
    {!models.length ? <div className="aicut-caption-config"><p>还没有配置可用的语音识别模型。</p><Button onClick={() => window.location.assign('/settings?tab=api')}>前往 API 配置</Button></div> : <>
      <label>音视频素材<Select value={assetId || undefined} onChange={onAssetChange} placeholder="选择要识别的素材" options={assets.filter(asset => ['audio', 'video'].includes(asset.kind)).map(asset => ({ value: asset.id, label: asset.name }))} /></label>
      <label>识别模型<Select value={modelId || undefined} onChange={onModelChange} options={models.map(model => ({ value: model.id, label: model.name || model.id }))} /></label>
      <label>识别语言<Select value={language} onChange={onLanguageChange} options={[{ value: 'zh', label: '中文' }, { value: 'en', label: '英文' }, { value: 'auto', label: '自动识别' }]} /></label>
      <Button type="primary" disabled={!canStart} loading={busy} onClick={onStart}>开始识别</Button>
    </>}
    <div className="aicut-caption-jobs">{jobs.map(job => <article key={job.id}><span><strong>{stateLabels[job.state] || job.state}</strong><small>{job.progress || 0}%{job.error ? ` · ${job.error}` : ''}</small></span>{['queued', 'running', 'cancel_requested'].includes(job.state) ? <Button size="small" disabled={job.state === 'cancel_requested'} onClick={() => onCancel(job)}>取消</Button> : ['failed', 'canceled'].includes(job.state) ? <Button size="small" onClick={() => onRetry(job)}>重新识别</Button> : null}</article>)}</div>
    <footer><Upload showUploadList={false} beforeUpload={onImportSRT} accept=".srt"><Button>导入 SRT</Button></Upload><Button onClick={onExportSRT}>导出 SRT</Button><Button onClick={onAddManualCaption}>添加手动字幕</Button></footer>
  </section>;
}
