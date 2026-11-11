import { Alert, Button, Select, message } from 'antd';
import { ArrowLeft, ArrowUpCircle, CheckCircle2, ExternalLink, FileVideo, FolderOpen, RotateCcw, X, XCircle } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelUpscaleJob, getUpscaleConfig, getUpscaleJob, submitUpscaleJob } from '../../shared/api/upscale.js';

const TERMINAL_STATUS = new Set(['completed', 'failed', 'cancelled']);
const POLL_INTERVAL_MS = 8000;
const VIDEO_PATTERN = /\.(mp4|mov|mkv|avi|webm|m4v)$/i;

const STAGE_TEXT = {
  queued: '排队中',
  uploading: '上传中',
  submitted: '已提交',
  processing: '排队 / 处理中',
  completed: '完成',
  failed: '失败',
  cancelled: '已取消'
};

function fileKey(file) {
  return `${file.name}:${file.size}:${file.lastModified}`;
}

function formatBytes(value) {
  const size = Number(value || 0);
  if (!size) return '';
  if (size >= 1024 ** 3) return `${(size / 1024 ** 3).toFixed(2)} GB`;
  if (size >= 1024 ** 2) return `${(size / 1024 ** 2).toFixed(1)} MB`;
  return `${Math.max(1, Math.round(size / 1024))} KB`;
}

function formatElapsed(seconds) {
  const total = Math.max(0, Math.floor(seconds));
  const minutes = Math.floor(total / 60);
  const rest = total % 60;
  return minutes ? `${minutes} 分 ${rest} 秒` : `${rest} 秒`;
}

export default function UpscaleWorkbench({ onBack }) {
  const [config, setConfig] = useState({ configured: false, resolutions: [] });
  const [loadingConfig, setLoadingConfig] = useState(true);
  const [files, setFiles] = useState([]);
  const [resolution, setResolution] = useState('1080p');
  const [items, setItems] = useState([]);
  const [folderName, setFolderName] = useState('');
  const [tick, setTick] = useState(0);
  const folderRef = useRef(null);
  const folderInputRef = useRef(null);
  const videoFolderRef = useRef(null);
  const videoFilesRef = useRef(null);
  const itemsRef = useRef(items);

  useEffect(() => {
    itemsRef.current = items;
  }, [items]);

  useEffect(() => {
    let alive = true;
    getUpscaleConfig()
      .then(value => {
        if (!alive) return;
        setConfig(value);
        const first = value.resolutions[0]?.value;
        if (first) setResolution(current => current || first);
      })
      .catch(error => message.error(error?.message || '读取超分配置失败'))
      .finally(() => alive && setLoadingConfig(false));
    return () => {
      alive = false;
    };
  }, []);

  const busy = items.some(item => !TERMINAL_STATUS.has(item.state));

  useEffect(() => {
    if (!busy) return undefined;
    const timer = window.setInterval(() => setTick(value => value + 1), 1000);
    return () => window.clearInterval(timer);
  }, [busy]);

  function patchItem(key, patch) {
    setItems(prev => prev.map(item => (item.key === key ? { ...item, ...patch } : item)));
  }

  function addFiles(list) {
    const incoming = Array.from(list || [])
      .filter(file => VIDEO_PATTERN.test(file.name) || String(file.type || '').startsWith('video/'))
      .map(file => ({ key: fileKey(file), file }));
    if (!incoming.length) {
      message.warning('没有找到可用的视频文件');
      return;
    }
    setFiles(prev => {
      const known = new Set(prev.map(item => item.key));
      return [...prev, ...incoming.filter(item => !known.has(item.key))];
    });
  }

  async function chooseFolder() {
    if (!window.showDirectoryPicker) {
      message.info('当前浏览器不支持选择目录；完成后请用结果链接下载');
      return;
    }
    try {
      const handle = await window.showDirectoryPicker({ mode: 'readwrite' });
      folderRef.current = handle;
      setFolderName(handle.name);
    } catch (error) {
      if (error?.name !== 'AbortError') message.error('保存文件夹选择失败');
    }
  }

  async function persistResult(item, job) {
    const folder = folderRef.current;
    if (!folder || !job?.resultUrl) return;
    patchItem(item.key, { saving: true });
    try {
      const response = await fetch(job.resultUrl);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const blob = await response.blob();
      const baseName = String(job.fileName || item.file.name || 'video').replace(/\.[^.]+$/, '');
      const name = `${baseName}-${String(job.resolution || resolution).toUpperCase()}.mp4`;
      const handle = await folder.getFileHandle(name, { create: true });
      const writable = await handle.createWritable();
      await writable.write(blob);
      await writable.close();
      patchItem(item.key, { saving: false, savedTo: `${folderName || folder.name}/${name}` });
      message.success(`已保存到 ${folder.name}/${name}`);
    } catch {
      patchItem(item.key, { saving: false });
      message.warning('视频已超分完成，但浏览器无法直接写入文件夹（跨域限制），请用结果链接下载');
    }
  }

  useEffect(() => {
    const timer = window.setInterval(async () => {
      const running = itemsRef.current.filter(item => item.job?.id && !TERMINAL_STATUS.has(item.state));
      for (const item of running) {
        try {
          const result = await getUpscaleJob(item.job.id);
          if (!result?.job) continue;
          patchItem(item.key, { job: result.job, state: result.job.status, pollHits: (item.pollHits || 0) + 1 });
          if (result.job.status === 'completed') {
            message.success(`${item.file.name} 超分完成`);
            await persistResult(item, result.job);
          }
        } catch {
          // 轮询失败不打断使用，保留上一次状态，下一轮继续。
        }
      }
    }, POLL_INTERVAL_MS);
    return () => window.clearInterval(timer);
  }, [folderName]);

  async function start() {
    if (!config.configured) return message.warning('未配置火山凭证');
    if (!files.length) return message.warning('请先选择要超分的视频');
    const runningKeys = new Set(itemsRef.current.filter(item => !TERMINAL_STATUS.has(item.state)).map(item => item.key));
    const queue = files.filter(file => !runningKeys.has(file.key));
    if (!queue.length) return message.warning('选中的视频都在处理中');
    setItems(prev => [
      ...prev,
      ...queue.map(file => ({ key: file.key, file: file.file, state: 'queued', job: null, pollHits: 0, startedAt: 0, savedTo: '' }))
    ]);
    for (const file of queue) {
      patchItem(file.key, { state: 'uploading', startedAt: Date.now() });
      try {
        const job = await submitUpscaleJob({ file: file.file, resolution });
        patchItem(file.key, {
          job,
          state: job?.status || 'submitted',
          startedAt: job?.createdAt ? new Date(job.createdAt).getTime() : Date.now()
        });
        message.success(`${file.file.name} 已提交`);
      } catch (error) {
        patchItem(file.key, { state: 'failed', error: error?.message || '提交超分任务失败' });
        message.error(`${file.file.name}：${error?.message || '提交超分任务失败'}`);
      }
    }
  }

  async function cancelItem(item) {
    if (!item.job?.id) return;
    try {
      const updated = await cancelUpscaleJob(item.job.id);
      patchItem(item.key, { job: updated, state: updated?.status || 'cancelled' });
      message.success(updated?.status === 'cancelled' ? '已取消' : '任务已结束');
    } catch (error) {
      message.error(error?.message || '取消失败');
    }
  }

  const resolutions = config.resolutions.length
    ? config.resolutions
    : [{ value: '1080p', label: '1080p' }, { value: '2k', label: '2K' }];

  const summary = useMemo(() => {
    if (!items.length) return '';
    const done = items.filter(item => item.state === 'completed').length;
    return `共 ${items.length} 条，完成 ${done} 条`;
  }, [items]);

  return (
    <section className="draft-tool draft-tool-page">
      <header className="draft-tool-head">
        <Button size="small" icon={<ArrowLeft size={14} />} onClick={onBack}>返回插件</Button>
        <span className="draft-tool-title"><ArrowUpCircle size={16} aria-hidden="true" />AI-Upscale</span>
      </header>

      {!loadingConfig && !config.configured ? (
        <Alert
          type="warning"
          showIcon
          message="未配置火山凭证"
          description="服务端没有设置 QIANTIE_VOLCANO_MEDIAKIT_API_KEY，AI-Upscale 无法提交任务。请管理员配置后重启服务再试。"
        />
      ) : null}

      <div className="draft-tool-split">
        <div className="draft-tool-params">
          <label className="draft-tool-field">
            <span>模型</span>
            <Select
              value={resolution}
              onChange={setResolution}
              disabled={busy}
              options={resolutions.map(item => ({ value: item.value, label: item.label }))}
            />
          </label>

          <div className="draft-tool-field">
            <span className="draft-tool-label-row">
              视频
              <span className="draft-tool-file-actions">
                <button type="button" disabled={busy} onClick={() => videoFolderRef.current?.click()}>选择文件夹</button>
                <button type="button" disabled={busy} onClick={() => videoFilesRef.current?.click()}>多选视频</button>
              </span>
            </span>
            <div className="upscale-picker">
              <p>文件夹会带上里面全部的源视频，也可以只多选其中几个。</p>
              {files.length ? (
                <ul className="upscale-file-list">
                  {files.map(item => (
                    <li key={item.key}>
                      <FileVideo size={14} aria-hidden="true" />
                      <span className="upscale-file-name">{item.file.name}</span>
                      <small>{formatBytes(item.file.size)}</small>
                      <button type="button" aria-label={`移除 ${item.file.name}`} disabled={busy} onClick={() => setFiles(prev => prev.filter(file => file.key !== item.key))}>
                        <X size={13} />
                      </button>
                    </li>
                  ))}
                </ul>
              ) : null}
            </div>
            <input
              ref={videoFolderRef}
              type="file"
              accept="video/*"
              multiple
              webkitdirectory="true"
              style={{ display: 'none' }}
              onChange={event => { addFiles(event.target.files); event.target.value = ''; }}
            />
            <input
              ref={videoFilesRef}
              type="file"
              accept="video/*"
              multiple
              style={{ display: 'none' }}
              onChange={event => { addFiles(event.target.files); event.target.value = ''; }}
            />
          </div>

          <div className="draft-tool-field">
            <span>保存文件夹</span>
            <div className="upscale-folder">
              <input readOnly value={folderName} placeholder="超分后的视频放在这里" />
              <Button size="small" icon={<FolderOpen size={13} />} disabled={busy} onClick={chooseFolder}>选择</Button>
            </div>
            <small className="draft-tool-hint">浏览器受跨域限制时无法直接写入该目录，会提示改用结果链接下载。</small>
          </div>

          <Button
            block
            className="draft-tool-analyze"
            style={{ background: '#1a1a1a', borderColor: '#1a1a1a' }}
            loading={busy && items.some(item => item.state === 'uploading')}
            disabled={!config.configured || !files.length || items.length > 0 && busy}
            onClick={start}
          >
            开始超分
          </Button>
        </div>

        <aside className="draft-tool-result-panel">
          <div className="draft-tool-result-title">
            进度和日志
            {summary ? <small>{summary}</small> : null}
          </div>
          <div className="draft-tool-result-body">
            {items.length ? (
              <ul className="upscale-job-list">
                {items.map(item => {
                  const elapsed = item.startedAt ? Math.max(0, (Date.now() - item.startedAt) / 1000) : 0;
                  const running = !TERMINAL_STATUS.has(item.state);
                  return (
                    <li key={item.key} data-state={item.state}>
                      <div className="upscale-job-head">
                        <strong>
                          {item.state === 'completed' ? <CheckCircle2 size={14} aria-hidden="true" /> : null}
                          {item.file.name}
                        </strong>
                        <span>{STAGE_TEXT[item.state] || item.state}</span>
                      </div>
                      <div className="upscale-job-meta">
                        <span>{String(item.job?.resolution || resolution).toUpperCase()} · {formatBytes(item.file.size)}</span>
                        {running ? <span>已等待 {formatElapsed(elapsed)}</span> : null}
                        {item.savedTo ? <span>已保存：{item.savedTo}</span> : null}
                      </div>
                      {running ? <div className="upscale-progress" data-stage={item.state}><span /></div> : null}
                      {item.state === 'completed' ? (
                        <div className="upscale-job-actions">
                          {item.job?.resultUrl ? (
                            <Button type="link" size="small" icon={<ExternalLink size={13} />} href={item.job.resultUrl} target="_blank" rel="noreferrer">打开结果视频</Button>
                          ) : <span className="draft-tool-hint">未返回结果链接，可重新提交。</span>}
                        </div>
                      ) : null}
                      {item.state === 'failed' ? (
                        <div className="upscale-job-actions">
                          <span className="upscale-job-error">{item.error || item.job?.error || item.job?.message || '超分失败'}</span>
                          <Button size="small" icon={<RotateCcw size={13} />} onClick={start}>重新提交</Button>
                        </div>
                      ) : null}
                      {item.state === 'cancelled' ? (
                        <div className="upscale-job-actions">
                          <span className="draft-tool-hint">任务已取消，可重新提交。</span>
                          <Button size="small" icon={<RotateCcw size={13} />} onClick={start}>重新提交</Button>
                        </div>
                      ) : null}
                      {running && item.job?.id ? (
                        <div className="upscale-job-actions">
                          <Button size="small" danger icon={<XCircle size={13} />} onClick={() => cancelItem(item)}>取消任务</Button>
                        </div>
                      ) : null}
                    </li>
                  );
                })}
              </ul>
            ) : (
              <div className="draft-tool-result-empty">
                <FileVideo size={34} aria-hidden="true" />
                <strong>等待开始</strong>
                <span>开始后，每一条视频的结果会在这里。</span>
              </div>
            )}
          </div>
        </aside>
      </div>
    </section>
  );
}
