import { Alert, Button, Empty, Input, Select, Spin, Upload, message } from 'antd';
import { ArrowLeft, Image as ImageIcon, RefreshCw, Scissors, Subtitles, Upload as UploadIcon, Wand2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  analyzeAICutCaptionCorrections,
  analyzeAICutFillerCandidates,
  applyAICutCaptionCorrections,
  applyAICutFillerCandidates,
  listAICutProjects,
  loadAICutTimeline,
  smartSplitAICutVideo
} from '../../shared/api/aicut.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { extractDocumentText } from '../../shared/api/pluginTools.js';

const { TextArea } = Input;

const TOOL_META = {
  'ai-subtitle': {
    icon: Subtitles,
    title: 'AI-Subtitle',
    summary: '对照剧本修正草稿字幕里的错别字，只改字幕文本，不动时间轴。',
    resultTitle: 'AI 分析结果',
    resultEmpty: '分析后将会在这里列出需要修改的字幕',
    applyLabel: '确认修改（写回剪映字幕轨）',
    notice: '只改写替换选中时间线里的字幕文本（无删点），不会拆散分句，也不会改其他时间线。确认修改后才会写回草稿。',
    buttonColor: '#22c55e'
  },
  'ai-filler': {
    icon: Wand2,
    title: 'AI-Filler',
    summary: '删除单独成行、语义为空的语气词与长停顿，并顺延后续片段。',
    resultTitle: '要删除的语气词',
    resultEmpty: '分析后将会在这里列出要删除的语气词字幕',
    applyLabel: '删除这些字词（写回剪映字幕轨）',
    notice: '只删除整行都是语气词的字幕（嗯、啊、呃、哦、哎呀、哎哟、唉、嘿、哈、哟），不改其他字幕，也不删视频。写回前请先确认列表。',
    buttonColor: '#fb7185'
  },
  'ai-split': {
    icon: Scissors,
    title: 'AI-Split',
    summary: '按画面变化切分主轨视频，切点最小间隔 500ms。',
    action: '开始切分',
    buttonColor: '#a855f7'
  }
};

function formatClock(ms) {
  const total = Math.max(0, Math.round(Number(ms || 0) / 1000));
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

function collectStats(timeline) {
  const tracks = Array.isArray(timeline?.tracks) ? timeline.tracks : [];
  let captions = 0;
  const videos = [];
  tracks.forEach(track => {
    (track.clips || []).forEach(clip => {
      if (clip.kind === 'subtitle') captions += 1;
      if (clip.kind === 'video' && clip.assetId) {
        videos.push({ id: clip.assetId, name: clip.name || clip.assetId });
      }
    });
  });
  const uniqueVideos = videos.filter((item, index) => videos.findIndex(other => other.id === item.id) === index);
  return { captions, videos: uniqueVideos };
}

function subtitleClips(track) {
  return (track?.clips || []).filter(clip => clip.kind === 'subtitle');
}

function describeResult(tool, result) {
  if (!result) return null;
  if (tool === 'ai-subtitle') return `修正字幕 ${Number(result.applied || 0)} 条`;
  if (tool === 'ai-filler') {
    const summary = result.summary || {};
    return `删除语气词 ${Number(summary.fillerCount || 0)} 处，处理停顿 ${Number(summary.pauseCount || 0)} 处，共缩短 ${Number(summary.removedMs || 0)} ms`;
  }
  return `切出 ${Number(result.splitCount || 0)} 个新片段`;
}

export default function DraftToolWorkbench({ tool, onBack }) {
  const meta = TOOL_META[tool] || TOOL_META['ai-subtitle'];
  const Icon = meta.icon;
  const twoColumn = tool === 'ai-subtitle' || tool === 'ai-filler';
  const [projects, setProjects] = useState([]);
  const [projectId, setProjectId] = useState('');
  const [timeline, setTimeline] = useState(null);
  const [loadingProjects, setLoadingProjects] = useState(true);
  const [loadingTimeline, setLoadingTimeline] = useState(false);
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState();
  const [script, setScript] = useState('');
  const [assetId, setAssetId] = useState('');
  const [threshold, setThreshold] = useState(0.4);
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState(null);
  // 两阶段：开始分析只拿预览，确认修改才写回。
  const [trackId, setTrackId] = useState('');
  const [captionsText, setCaptionsText] = useState('');
  const [readCount, setReadCount] = useState(0);
  const [preview, setPreview] = useState([]);
  const [analyzing, setAnalyzing] = useState(false);
  const [applying, setApplying] = useState(false);
  const controllerRef = useRef(null);

  const stats = useMemo(() => collectStats(timeline), [timeline]);
  const subtitleTracks = useMemo(
    () => (Array.isArray(timeline?.tracks) ? timeline.tracks.filter(track => track.kind === 'subtitle') : []),
    [timeline]
  );
  const busy = running || analyzing || applying;

  useEffect(() => () => { controllerRef.current?.abort(); }, []);

  function loadProjects() {
    setLoadingProjects(true);
    return listAICutProjects()
      .then(list => {
        setProjects(list);
        setProjectId(current => current || list[0]?.id || '');
      })
      .catch(error => message.error(error?.message || '读取草稿列表失败'))
      .finally(() => setLoadingProjects(false));
  }

  useEffect(() => {
    let alive = true;
    listAICutProjects()
      .then(list => {
        if (!alive) return;
        setProjects(list);
        setProjectId(current => current || list[0]?.id || '');
      })
      .catch(error => message.error(error?.message || '读取草稿列表失败'))
      .finally(() => alive && setLoadingProjects(false));
    return () => {
      alive = false;
    };
  }, []);

  useEffect(() => {
    if (tool === 'ai-split') return undefined;
    let alive = true;
    listAvailableModels('text')
      .then(list => {
        if (!alive) return;
        setModels(list);
        setModelId(current => current || list[0]?.id);
      })
      .catch(() => alive && setModels([]));
    return () => {
      alive = false;
    };
  }, [tool]);

  useEffect(() => {
    if (!projectId) {
      setTimeline(null);
      return undefined;
    }
    let alive = true;
    setLoadingTimeline(true);
    setResult(null);
    setPreview([]);
    loadAICutTimeline(projectId)
      .then(value => {
        if (!alive) return;
        setTimeline(value);
        setScript(value?.script || '');
      })
      .catch(error => alive && message.error(error?.message || '读取草稿失败'))
      .finally(() => alive && setLoadingTimeline(false));
    return () => {
      alive = false;
    };
  }, [projectId]);

  // 草稿或轨道变化时自动读取字幕，对齐洛水「选择时间线后自动读取」。
  useEffect(() => {
    if (!twoColumn) return;
    setTrackId(current => (subtitleTracks.some(track => track.id === current) ? current : subtitleTracks[0]?.id || ''));
  }, [subtitleTracks, twoColumn]);

  useEffect(() => {
    if (!twoColumn) return;
    const track = subtitleTracks.find(item => item.id === trackId);
    const items = subtitleClips(track);
    setReadCount(items.length);
    setCaptionsText(items.map((clip, index) => (
      `${index + 1} · ${formatClock(clip.startMs)} - ${formatClock(clip.startMs + clip.durationMs)} · ${clip.text || ''}`
    )).join('\n'));
    setPreview([]);
  }, [subtitleTracks, trackId, twoColumn]);

  useEffect(() => {
    if (tool !== 'ai-split') return;
    setAssetId(current => (stats.videos.some(item => item.id === current) ? current : stats.videos[0]?.id || ''));
  }, [stats.videos, tool]);

  async function readScriptFile(file) {
    const lowerName = String(file?.name || '').toLowerCase();
    try {
      if (lowerName.endsWith('.txt') || lowerName.endsWith('.md')) {
        setScript(await file.text());
      } else if (lowerName.endsWith('.doc') || lowerName.endsWith('.docx')) {
        setScript(await extractDocumentText(file));
      } else {
        message.warning('仅支持 TXT / Word 文档');
        return Upload.LIST_IGNORE;
      }
      message.success('剧本文件读取成功');
    } catch (error) {
      message.error(error?.message || '剧本文件读取失败');
    }
    return Upload.LIST_IGNORE;
  }

  async function analyze() {
    if (busy) return message.info('正在处理，请勿重复提交');
    if (!projectId) return message.warning('请先选择剪映草稿');
    if (!modelId) return message.warning('请选择模型');
    if (tool === 'ai-subtitle' && !script.trim()) return message.warning('请先填入剧本文本');
    if (!trackId) return message.warning('请先选择字幕轨道');
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setAnalyzing(true);
    setPreview([]);
    try {
      const expectedRevision = Number(timeline?.revision || 0);
      const response = tool === 'ai-subtitle'
        ? await analyzeAICutCaptionCorrections(projectId, { script, modelId, expectedRevision, signal: controller.signal })
        : await analyzeAICutFillerCandidates(projectId, { modelId, expectedRevision, signal: controller.signal });
      const items = Array.isArray(response?.preview) ? response.preview : [];
      setPreview(items);
      if (items.length) message.success(`分析完成，共 ${items.length} 条`);
      else message.info(response?.message || '没有需要处理的字幕');
    } catch (error) {
      if (error?.name === 'AbortError' || error?.aborted) return;
      message.error(error?.message || '分析失败');
    } finally {
      if (controllerRef.current === controller) controllerRef.current = null;
      setAnalyzing(false);
    }
  }

  async function applyPreview() {
    if (busy) return message.info('正在处理，请勿重复提交');
    if (!preview.length) return message.warning('请先开始分析');
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setApplying(true);
    try {
      const expectedRevision = Number(timeline?.revision || 0);
      const response = tool === 'ai-subtitle'
        ? await applyAICutCaptionCorrections(projectId, {
          corrections: preview.map(item => ({ clipId: item.clipId, text: item.newText, reason: item.reason })),
          expectedRevision,
          signal: controller.signal
        })
        : await applyAICutFillerCandidates(projectId, {
          fillers: preview.map(item => ({ clipId: item.clipId, reason: item.reason })),
          expectedRevision,
          signal: controller.signal
        });
      setPreview([]);
      if (response?.timeline) setTimeline(response.timeline);
      setResult(response);
      message.success(response?.message || '已写回草稿');
    } catch (error) {
      if (error?.name === 'AbortError' || error?.aborted) return;
      message.error(error?.message || '写回失败');
    } finally {
      if (controllerRef.current === controller) controllerRef.current = null;
      setApplying(false);
    }
  }

  async function runSplit() {
    if (busy) return message.info('正在处理，请勿重复提交');
    if (!projectId) return message.warning('请先选择草稿');
    if (!assetId) return message.warning('请选择要切分的视频');
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setRunning(true);
    setResult(null);
    try {
      const expectedRevision = Number(timeline?.revision || 0);
      const response = await smartSplitAICutVideo(projectId, { assetId, threshold, expectedRevision, signal: controller.signal });
      setResult(response);
      if (response?.timeline) setTimeline(response.timeline);
      message.success(response?.message || '执行完成');
    } catch (error) {
      if (error?.name === 'AbortError' || error?.aborted) return;
      message.error(error?.message || '执行失败');
    } finally {
      if (controllerRef.current === controller) controllerRef.current = null;
      setRunning(false);
    }
  }

  const projectSelect = (
    <label className="draft-tool-field">
      <span>目标草稿</span>
      <Select
        loading={loadingProjects}
        value={projectId || undefined}
        onChange={setProjectId}
        placeholder={projects.length ? '选择草稿' : '暂无草稿'}
        options={projects.map(item => ({ value: item.id, label: item.title || item.id }))}
        notFoundContent={loadingProjects ? <Spin size="small" /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无草稿" />}
      />
    </label>
  );

  if (!twoColumn) {
    return (
      <section className="draft-tool">
        <header className="draft-tool-head">
          <Button size="small" icon={<ArrowLeft size={14} />} onClick={onBack}>返回插件</Button>
          <span className="draft-tool-title"><Icon size={16} aria-hidden="true" />{meta.title}</span>
        </header>
        <p className="draft-tool-summary">{meta.summary}</p>

        <div className="draft-tool-panel">
          {projectSelect}
          {projectId ? (
            <div className="draft-tool-stats">
              <span>字幕 <strong>{stats.captions}</strong> 条</span>
              <span>视频轨 <strong>{stats.videos.length}</strong> 段</span>
              <span>版本 <strong>r{Number(timeline?.revision || 0)}</strong></span>
              {loadingTimeline ? <span className="draft-tool-loading">读取中…</span> : null}
            </div>
          ) : null}

          <label className="draft-tool-field">
            <span>切分视频</span>
            <Select
              value={assetId || undefined}
              onChange={setAssetId}
              disabled={!stats.videos.length}
              placeholder={stats.videos.length ? '选择视频' : '该草稿没有可切分的视频'}
              options={stats.videos.map(item => ({ value: item.id, label: item.name }))}
            />
          </label>
          <label className="draft-tool-field">
            <span>画面变化阈值：{threshold.toFixed(2)}</span>
            <input className="draft-tool-range" type="range" min="0.1" max="0.9" step="0.05" value={threshold} onChange={event => setThreshold(Number(event.target.value))} />
          </label>

          <div className="draft-tool-actions">
            <Button type="primary" loading={running} disabled={!projectId} onClick={runSplit}>{running ? '执行中…' : meta.action}</Button>
            <span>结果会直接写回该草稿时间轴，可在 AI-CUT 工作台继续编辑。</span>
          </div>
        </div>

        {result ? (
          <Alert
            className="draft-tool-result"
            type="success"
            showIcon
            message={result.message || '执行完成'}
            description={describeResult(tool, result)}
          />
        ) : null}
      </section>
    );
  }

  const selectedTrack = subtitleTracks.find(track => track.id === trackId);
  const isSubtitle = tool === 'ai-subtitle';

  return (
    <section className="draft-tool draft-tool-page">
      <header className="draft-tool-head">
        <Button size="small" icon={<ArrowLeft size={14} />} onClick={onBack}>返回插件</Button>
        <span className="draft-tool-title"><Icon size={16} aria-hidden="true" />{meta.title}</span>
      </header>
      <p className="draft-tool-summary">{meta.summary}</p>

      <div className="draft-tool-split">
        <div className="draft-tool-params">
          <label className="draft-tool-field">
            <span>选择模型</span>
            <Select
              value={modelId}
              onChange={setModelId}
              options={modelSelectOptions(models, 'text')}
              placeholder={models.length ? '选择模型' : '未配置可用模型'}
              disabled={busy}
            />
          </label>

          <div className="draft-tool-field">
            <span className="draft-tool-label-row">
              选择剪映草稿
              <Button size="small" type="text" icon={<RefreshCw size={13} />} disabled={busy} onClick={loadProjects}>刷新</Button>
            </span>
            <Select
              loading={loadingProjects}
              value={projectId || undefined}
              onChange={setProjectId}
              placeholder={projects.length ? '选择草稿' : '暂无草稿'}
              options={projects.map(item => ({ value: item.id, label: item.title || item.id }))}
              disabled={busy}
              notFoundContent={loadingProjects ? <Spin size="small" /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无草稿" />}
            />
          </div>

          <label className="draft-tool-field">
            <span>时间线</span>
            <Select
              value={projectId ? String(timeline?.revision ?? '') : undefined}
              options={projectId ? [{ value: String(timeline?.revision ?? ''), label: `主时间线 · r${Number(timeline?.revision || 0)}` }] : []}
              placeholder="选择草稿后自动读取"
              disabled={!projectId || busy}
            />
          </label>

          <div className="draft-tool-field">
            <span className="draft-tool-label-row">
              字幕轨道
              <Button
                size="small"
                type="text"
                icon={<Subtitles size={13} />}
                disabled={!selectedTrack || busy}
                onClick={() => {
                  const items = subtitleClips(selectedTrack);
                  setReadCount(items.length);
                  setCaptionsText(items.map((clip, index) => (
                    `${index + 1} · ${formatClock(clip.startMs)} - ${formatClock(clip.startMs + clip.durationMs)} · ${clip.text || ''}`
                  )).join('\n'));
                  setPreview([]);
                  message.success(`已读取 ${items.length} 条字幕`);
                }}
              >
                读取字幕
              </Button>
            </span>
            <Select
              value={trackId || undefined}
              onChange={setTrackId}
              disabled={!subtitleTracks.length || busy}
              placeholder={subtitleTracks.length ? '选择字幕轨道' : '该草稿没有字幕轨道'}
              options={subtitleTracks.map(track => ({
                value: track.id,
                label: `${track.name || track.id}（${subtitleClips(track).length} 条${track.locked ? ' · 已锁定' : ''}）`
              }))}
            />
          </div>

          <p className="draft-tool-hint">
            {isSubtitle
              ? `已读取 ${readCount} 条字幕；只改选中时间线里的 subtitle，不改「AI生成」本段`
              : `已读取 ${readCount} 条字幕；只删整行都是语气词的字幕，字幕的语序不动`}
          </p>

          <div className="draft-tool-field">
            <span className="draft-tool-label-row">
              剧本文本
              <Upload accept=".txt,.md,.doc,.docx" showUploadList={false} beforeUpload={readScriptFile} disabled={busy}>
                <span className="draft-tool-file-actions">
                  <button type="button" disabled={busy}>选择文件</button>
                  <button type="button" disabled={busy}><UploadIcon size={12} /> 上传</button>
                </span>
              </Upload>
            </span>
            <TextArea
              rows={6}
              value={script}
              onChange={event => setScript(event.target.value)}
              placeholder="粘贴剧本文本，或上传 txt / doc / docx..."
              disabled={busy}
            />
            {isSubtitle ? null : <small className="draft-tool-hint">AI-Filler 只按字幕判断，剧本文本仅作对照参考。</small>}
          </div>

          <label className="draft-tool-field">
            <span>已读字幕（序号 · 时间戳 · 文本）</span>
            <TextArea rows={7} value={captionsText} readOnly placeholder="选择时间线后自动读取" />
          </label>

          <Button
            block
            className="draft-tool-analyze"
            style={{ background: meta.buttonColor, borderColor: meta.buttonColor }}
            loading={analyzing}
            disabled={!projectId || !trackId || busy && !analyzing}
            onClick={analyze}
          >
            {analyzing ? '分析中…' : '开始分析'}
          </Button>

          <p className="draft-tool-notice">{meta.notice}</p>
        </div>

        <aside className="draft-tool-result-panel">
          <div className="draft-tool-result-title">{meta.resultTitle}</div>
          <div className="draft-tool-result-body">
            {preview.length ? (
              <ol className="draft-tool-preview-list">
                {preview.map((item, index) => (
                  <li key={`${item.clipId}-${index}`}>
                    <span className="draft-tool-preview-index">{index + 1}</span>
                    <span className="draft-tool-preview-time">{formatClock(item.startMs)} - {formatClock(item.endMs)}</span>
                    <span className="draft-tool-preview-copy">
                      {isSubtitle ? (
                        <><del>{item.oldText}</del><ins>{item.newText}</ins></>
                      ) : (
                        <strong>{item.text}</strong>
                      )}
                      {item.reason ? <small>{item.reason}</small> : null}
                    </span>
                  </li>
                ))}
              </ol>
            ) : (
              <div className="draft-tool-result-empty">
                <ImageIcon size={34} aria-hidden="true" />
                <span>{meta.resultEmpty}</span>
              </div>
            )}
          </div>
          <Button
            block
            className="draft-tool-apply"
            loading={applying}
            disabled={!preview.length || busy && !applying}
            onClick={applyPreview}
          >
            {applying ? '写回中…' : meta.applyLabel}
          </Button>
        </aside>
      </div>
    </section>
  );
}
