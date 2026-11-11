import { Alert, Button, Input, InputNumber, Progress, Radio, Select, Upload, message } from 'antd';
import { ArrowLeft, Check, FileJson, Film, FolderOpen, ListTree, RefreshCw, Scissors, Upload as UploadIcon, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { analyzeAISegment, exportAISegment } from '../../shared/api/pluginTools.js';
import { aiSegmentModeContract, aiSegmentModeHas, defaultAISegmentMode, visibleAISegmentModes } from '../pages/ai-segment-workbench.js';
import './ai-segment-workbench.css';

const { TextArea } = Input;

function readTextFile(file, setter, success) {
  file.text().then(value => {
    setter(value);
    message.success(success);
  }).catch(() => message.error('文件读取失败'));
  return Upload.LIST_IGNORE;
}

function draftTextTracks(value) {
  try {
    const draft = JSON.parse(value);
    return (Array.isArray(draft?.tracks) ? draft.tracks : [])
      .filter(track => ['text', 'subtitle'].includes(String(track?.type || '').toLowerCase()))
      .map((track, index) => ({ value: String(track.id || `text-track-${index + 1}`), label: track.name || `文本轨道 ${index + 1}（${track.segments?.length || 0} 条）` }));
  } catch {
    return [];
  }
}

function saveBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

function displayTime(value) {
  const total = Math.max(0, Number(value || 0));
  const minutes = Math.floor(total / 60);
  const seconds = Math.floor(total % 60);
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

export default function AISegmentWorkbench({ onBack }) {
  const [mode, setMode] = useState(defaultAISegmentMode);
  const [video, setVideo] = useState(null);
  const [srt, setSRT] = useState('');
  const [draftContent, setDraftContent] = useState('');
  const [script, setScript] = useState('');
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState();
  const [textTrackId, setTextTrackId] = useState();
  const [frontCount, setFrontCount] = useState(2);
  const [frontMin, setFrontMin] = useState(90);
  const [frontMax, setFrontMax] = useState(180);
  const [backMin, setBackMin] = useState(45);
  const [backMax, setBackMax] = useState(90);
  const [resolution, setResolution] = useState('1080p');
  const [fps, setFPS] = useState(30);
  const [result, setResult] = useState(null);
  const [running, setRunning] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [progress, setProgress] = useState(0);
  const [logs, setLogs] = useState([]);
  const [outputFolder, setOutputFolder] = useState(null);
  const controllerRef = useRef(null);
  const trackOptions = useMemo(() => draftTextTracks(draftContent), [draftContent]);
  const contract = aiSegmentModeContract[mode];
  const busy = running || exporting;

  useEffect(() => {
    let active = true;
    listAvailableModels('text').then(items => {
      if (!active) return;
      setModels(items);
      setModelId(current => current || items[0]?.id);
    }).catch(error => message.error(error?.message || '文本模型加载失败'));
    return () => {
      active = false;
      controllerRef.current?.abort();
    };
  }, []);

  function appendLog(text) {
    const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
    setLogs(items => [...items, `[${time}] ${text}`]);
  }

  function switchMode(value) {
    if (busy) return;
    setMode(value);
    setResult(null);
    setProgress(0);
    setLogs([]);
    const quality = aiSegmentModeContract[value].quality;
    setResolution(quality.resolution);
    setFPS(quality.fps);
  }

  async function chooseOutputFolder() {
    if (!window.showDirectoryPicker) {
      message.info('当前浏览器不支持选择目录，完成后将下载 ZIP 文件');
      return;
    }
    try {
      const handle = await window.showDirectoryPicker({ mode: 'readwrite' });
      setOutputFolder(handle);
      appendLog(`输出目录：${handle.name}`);
    } catch (error) {
      if (error?.name !== 'AbortError') message.error('输出目录选择失败');
    }
  }

  async function runAnalysis() {
    if (mode === 'subtitle' && !srt.trim()) { message.warning('请上传 SRT 字幕文件'); return null; }
    if (mode !== 'subtitle' && !draftContent.trim()) { message.warning('请选择剪映 draft_content.json'); return null; }
    if (mode === 'text-track' && !textTrackId) { message.warning('请选择包含字幕的文本轨道'); return null; }
    if (mode === 'ai' && !script.trim()) { message.warning('请先提供剧本文稿'); return null; }
    if (mode === 'ai' && !modelId) { message.warning('请选择模型'); return null; }
    controllerRef.current = new AbortController();
    setRunning(true);
    setResult(null);
    setProgress(18);
    appendLog(mode === 'ai' ? '开始读取主轨道、字幕和剧本文稿' : '开始读取切点数据');
    try {
      const analyzed = await analyzeAISegment({
        mode, modelId, script, srt, draftContent, textTrackId,
        constraints: { frontCount, frontMin, frontMax, backMin, backTarget: Math.min(60, backMax), backMax }
      }, undefined, controllerRef.current.signal);
      setResult(analyzed);
      setProgress(100);
      appendLog(`分析完成：主轨道 ${analyzed.clipCount || 0} 个片段，字幕 ${analyzed.subtitleCount || 0} 条，共 ${analyzed.episodes?.length || 0} 集`);
      message.success(`分析完成，共 ${analyzed.episodes?.length || 0} 集`);
      return analyzed;
    } catch (error) {
      if (error?.name !== 'AbortError') {
        appendLog(`分析失败：${error?.message || '未知错误'}`);
        message.error(error?.message || '分段分析失败');
      }
      return null;
    } finally {
      controllerRef.current = null;
      setRunning(false);
    }
  }

  async function persistArchive(archive) {
    if (!outputFolder) {
      saveBlob(archive, 'AI分段结果.zip');
      return '已开始下载 ZIP 文件';
    }
    const fileHandle = await outputFolder.getFileHandle('AI分段结果.zip', { create: true });
    const writable = await fileHandle.createWritable();
    await writable.write(archive);
    await writable.close();
    return `已保存到 ${outputFolder.name}`;
  }

  async function runExport(source) {
    const episodes = source?.episodes || result?.episodes || [];
    if (!video) { message.warning('请选择已导出的成片'); return; }
    if (!episodes.length) { message.warning(mode === 'ai' ? '请先点「开始分段」做 AI 分析' : '请先分析切点'); return; }
    controllerRef.current = new AbortController();
    setExporting(true);
    setProgress(12);
    appendLog(`开始切割 ${episodes.length} 集视频`);
    try {
      const archive = await exportAISegment({ video, episodes, resolution, fps }, undefined, controllerRef.current.signal);
      setProgress(86);
      const notice = await persistArchive(archive);
      setProgress(100);
      appendLog(`视频切割完成，${notice}`);
      message.success(`视频切割完成，${notice}`);
    } catch (error) {
      if (error?.name !== 'AbortError') {
        appendLog(`视频切割失败：${error?.message || '未知错误'}`);
        message.error(error?.message || '视频切割失败');
      }
    } finally {
      controllerRef.current = null;
      setExporting(false);
    }
  }

  async function startSegment() {
    if (busy) { message.info('正在处理，请勿重复提交'); return; }
    if (!video) { message.warning('请选择已导出的成片'); return; }
    const analyzed = result?.episodes?.length ? result : await runAnalysis();
    if (!analyzed?.episodes?.length) return;
    await runExport(analyzed);
  }

  function cancelCurrent() {
    controllerRef.current?.abort();
    controllerRef.current = null;
    setRunning(false);
    setExporting(false);
    setProgress(0);
    appendLog('用户取消了本次操作');
    message.info('已取消本次操作');
  }

  return (
    <section className="ai-segment-page">
      <header className="ai-segment-page-head">
        <Button size="small" icon={<ArrowLeft size={14} />} onClick={onBack}>返回插件</Button>
        <div className="ai-segment-title-copy">
          <div>
            <span className="ai-segment-title-icon"><Scissors size={16} aria-hidden="true" /></span>
            <h1>AI-分段</h1>
            <span className="ai-segment-authorized"><Check size={13} /> 已授权</span>
          </div>
        </div>
        <Button type="text" shape="circle" aria-label="关闭AI分段" icon={<X size={18} />} disabled={busy} onClick={onBack} />
      </header>
      <p className="ai-segment-version-note">剪映草稿分集 · 读取草稿主轨道片段边界切割。仅支持剪映 11.6 及以下版本，网页重写版通过 draft_content.json 读取草稿数据。</p>

      <div className="ai-segment-grid">
        <div className="ai-segment-form">
          <Radio.Group className="ai-segment-mode-buttons" value={mode} onChange={event => switchMode(event.target.value)} disabled={busy}>
            {visibleAISegmentModes.map(item => <Radio.Button key={item.value} value={item.value}>{item.label}</Radio.Button>)}
          </Radio.Group>

          {aiSegmentModeHas(mode, 'model') ? <div className="ai-segment-section">
            <label className="ai-segment-label">选择模型</label>
            <Select value={modelId} onChange={setModelId} options={modelSelectOptions(models, 'text')} placeholder="选择用于分集的模型" disabled={busy} />
          </div> : null}

          <Upload className="ai-segment-upload" accept="video/*,.mp4,.mov,.mkv,.avi" showUploadList={false} beforeUpload={file => { setVideo(file); appendLog(`已选择视频：${file.name}`); return Upload.LIST_IGNORE; }}>
            <div className="ai-segment-picker"><Film size={25} /><div><strong>{video?.name || '点击选择剪映已导出的成片'}</strong><span>请先在剪映导出带字幕和特效的成片，再拿到这里切割</span></div><Button type="primary" size="small" style={{ background: '#8b5cf6', borderColor: '#8b5cf6' }}>选择</Button></div>
          </Upload>

          {aiSegmentModeHas(mode, 'srt') ? <Upload className="ai-segment-upload" accept=".srt" showUploadList={false} beforeUpload={file => readTextFile(file, setSRT, 'SRT 字幕读取成功')}><div className="ai-segment-picker"><Film size={23} /><div><strong>{srt ? '已读取 SRT 字幕' : '点击选择SRT字幕文件'}</strong><span>按下一句开始时间落刀，每段只多不少</span></div><Button size="small">选择</Button></div></Upload> : null}

          {aiSegmentModeHas(mode, 'draft') ? <div className="ai-segment-section">
            <div className="ai-segment-section-title">
              <strong>选择剪映草稿</strong>
              <Upload accept=".json" showUploadList={false} beforeUpload={file => readTextFile(file, value => { setDraftContent(value); setTextTrackId(undefined); }, '剪映草稿读取成功')}>
                <Button size="small" icon={<RefreshCw size={13} />} disabled={busy}>{draftContent ? '刷新' : '选择草稿'}</Button>
              </Upload>
            </div>
            <div className="ai-segment-draft-status"><FileJson size={18} /><span>{draftContent ? '已读取 draft_content.json' : '尚未选择 draft_content.json'}</span></div>
            <p className="ai-segment-help">{mode === 'text-track' ? '读取所选文本轨道作为切点' : mode === 'ai' ? '读取主轨道片段交界，并叠加字幕内容交给 AI 判断' : '只读取主轨道片段时长和交界'}</p>
            {aiSegmentModeHas(mode, 'text-track') ? <><label className="ai-segment-label ai-segment-label-gap">选择文本轨道</label><Select value={textTrackId} onChange={setTextTrackId} options={trackOptions} placeholder={trackOptions.length ? '选择一条文本轨道作为切点' : '请先读取包含文本轨道的草稿'} disabled={!trackOptions.length || busy} /></> : null}
          </div> : null}

          {aiSegmentModeHas(mode, 'script') ? <div className="ai-segment-section">
            <div className="ai-segment-section-title"><strong>剧本文稿</strong><Upload accept=".txt" showUploadList={false} beforeUpload={file => readTextFile(file, setScript, '剧本文稿读取成功')}><Button size="small" icon={<UploadIcon size={14} />}>上传 TXT</Button></Upload></div>
            <TextArea rows={5} value={script} onChange={event => setScript(event.target.value)} placeholder="粘贴剧本文稿，或上传 txt 文档。AI 只建议剧情切点，最终会校正到主轨片段交界。" disabled={busy} />
            <p className="ai-segment-help">前 2 集 1:30～3:00，后续每集 45～60 秒（最长 1:30）。只在主轨道片段交界切，不从中间切开。</p>
          </div> : null}

          <button type="button" className="ai-segment-picker ai-segment-output-picker" onClick={chooseOutputFolder} disabled={busy}>
            <FolderOpen size={25} /><div><strong>{outputFolder?.name || '点击选择输出文件夹'}</strong><span>{outputFolder ? '切割后的 ZIP 将保存到此目录' : '未选择时由浏览器下载 ZIP 文件'}</span></div><span className="ai-segment-picker-button">选择</span>
          </button>

          <div className="ai-segment-section">
            <strong>参数设置</strong>
            {aiSegmentModeHas(mode, 'duration') ? <div className="ai-segment-duration-grid">
              <label>前 <InputNumber min={1} max={20} value={frontCount} onChange={value => setFrontCount(Number(value || 2))} disabled={busy} /> 集，每集 <InputNumber min={30} value={frontMin} onChange={value => setFrontMin(Number(value || 90))} disabled={busy} /> 至 <InputNumber min={30} value={frontMax} onChange={value => setFrontMax(Number(value || 180))} disabled={busy} /> 秒</label>
              <label>后续各集时长 <InputNumber min={15} value={backMin} onChange={value => setBackMin(Number(value || 45))} disabled={busy} /> 至 <InputNumber min={15} value={backMax} onChange={value => setBackMax(Number(value || 90))} disabled={busy} /> 秒</label>
            </div> : <p className="ai-segment-help">{mode === 'ai' ? '时长由 AI 按剧本转折决定，并校正到主轨道片段交界。' : '按所选文本轨道每一段的时长直接切片。'}</p>}
            <div className="ai-segment-quality"><label>分辨率<Select value={resolution} onChange={setResolution} disabled={busy || contract.quality.locked} options={[{ value: '', label: '保持原始' }, { value: '1080p', label: '1080p' }, { value: '720p', label: '720p' }]} /></label><label>帧率<Select value={fps} onChange={setFPS} disabled={busy || contract.quality.locked} options={[{ value: 0, label: '保持原始' }, { value: 24, label: '24 FPS' }, { value: 25, label: '25 FPS' }, { value: 30, label: '30 FPS' }]} /></label></div>
            <p className="ai-segment-help">{mode === 'ai' ? 'AI 模式固定导出 1080p / 30 帧。' : '时长只多不少；只在已有片段交界处落刀。'}</p>
          </div>

          <Button
            type="primary"
            size="large"
            block
            loading={running || exporting}
            disabled={!video || busy && !running && !exporting}
            style={{ background: '#8b5cf6', borderColor: '#8b5cf6' }}
            onClick={startSegment}
          >
            {running ? '分析中…' : exporting ? '处理中…' : '开始分段'}
          </Button>
          {busy ? <Button danger block icon={<X size={15} />} onClick={cancelCurrent}>{running ? '取消分析' : '取消切割'}</Button> : null}
        </div>

        <aside className="ai-segment-log-panel">
          <div className="ai-segment-log-title"><ListTree size={17} /><strong>处理日志</strong></div>
          <Progress percent={progress} size="small" strokeColor="#8b5cf6" />
          {result?.episodes?.length ? <div className="ai-segment-result-block">
            <div className="ai-segment-section-title"><strong>分集预览</strong><span>主轨道 {result.clipCount || 0} 个片段，字幕 {result.subtitleCount || 0} 条</span></div>
            {result.warnings?.length ? <Alert type="warning" showIcon message={result.warnings.join('；')} /> : null}
            <div className="ai-segment-episode-list">{result.episodes.map(item => <div key={item.index} className="ai-segment-episode"><b>第{item.index}集</b><span>{displayTime(item.startSeconds)}-{displayTime(item.endSeconds)}</span><em>{Number(item.durationSeconds || 0).toFixed(1)}秒</em><small>{item.reason || '按主轨片段边界规划'}</small></div>)}</div>
          </div> : null}
          <div className="ai-segment-log-list">{logs.length ? logs.map((item, index) => <p key={`${index}-${item}`}>{item}</p>) : <div className="ai-segment-log-empty"><ListTree size={28} /><span>暂无日志</span></div>}</div>
        </aside>
      </div>
    </section>
  );
}
