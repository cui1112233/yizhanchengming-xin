import { Alert, Button, Empty, Input, InputNumber, Modal, Progress, Select, Spin, Tag, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import { applyReshoot, extractReshootFrames, listReshootCandidates, loadReshootStatus, startReshoot } from './comic-screenplay-storyboard-api.js';

// 洛水 2026-10-29 版「片段重拍」：在 20 帧时间轴上框选一段，抽首尾帧重新生成，
// 再把原前段、新片段、原后段拼回完整视频。这里不伪造任何没有后端承接的能力。
const MIN_CLIP = 0.1;
const MIN_SECONDS = 1;
const MAX_SECONDS = 30;

function timeText(value) {
  const number = Math.max(0, Number(value) || 0);
  const minutes = Math.floor(number / 60);
  return `${minutes}:${(number % 60).toFixed(1).padStart(4, '0')}`;
}

function episodeOptions(boards, episodes) {
  return (boards || [])
    .filter(board => (board.shots || []).length)
    .map(board => {
      const episode = (episodes || []).find(item => item.no === board.episode_no);
      return { value: board.episode_no, label: `第 ${board.episode_no} 集${episode?.title ? ` · ${episode.title}` : ''}` };
    });
}

function capability(view, key) {
  const caps = view?.capabilities || {};
  return caps[key] || {};
}

export default function ComicScreenplayReshoot({ projectId, boards, episodes, engine, readOnly, defaultEpisodeNo, onClose, onFinished }) {
  const episodeChoices = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (episodeChoices.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : episodeChoices[0]?.value ?? null));
  const [candidates, setCandidates] = useState([]);
  const [candidatesLoading, setCandidatesLoading] = useState(false);
  const [candidatesError, setCandidatesError] = useState('');
  const [capabilities, setCapabilities] = useState({});
  const [shotId, setShotId] = useState('');
  const [frames, setFrames] = useState([]);
  const [framesLoading, setFramesLoading] = useState(false);
  const [framesError, setFramesError] = useState('');
  const [videoDuration, setVideoDuration] = useState(0);
  const [start, setStart] = useState(0);
  const [end, setEnd] = useState(0);
  const [genSeconds, setGenSeconds] = useState(null);
  const [prompt, setPrompt] = useState('');
  const [task, setTask] = useState(null);
  const [playhead, setPlayhead] = useState(0);
  const [busy, setBusy] = useState('');
  const [error, setError] = useState('');
  const videoRef = useRef(null);
  const timelineRef = useRef(null);
  const ticket = useRef(0);
  const rangeRef = useRef({ start: 0, end: 0 });
  const durationTouched = useRef(false);

  const selected = candidates.find(item => item.shot_id === shotId) || null;
  const clipSeconds = Math.max(0, end - start);
  const canGenerate = capability({ capabilities }, 'generate').available === true;
  const canApply = capability({ capabilities }, 'apply').available === true;
  const editing = !task || task.status === 'failed';
  const generating = task?.status === 'generating';
  const ready = task?.status === 'done' || task?.status === 'applied';

  useEffect(() => {
    if (!projectId || !episodeNo) {
      setCandidates([]);
      return undefined;
    }
    const token = ++ticket.current;
    setCandidatesLoading(true);
    setCandidatesError('');
    listReshootCandidates(projectId, { episode: episodeNo, engine })
      .then(view => {
        if (token !== ticket.current) return;
        setCandidates(view.candidates || []);
        setCapabilities(view.capabilities || {});
        setShotId(current => (view.candidates || []).some(item => item.shot_id === current) ? current : (view.candidates?.[0]?.shot_id || ''));
      })
      .catch(err => {
        if (token === ticket.current) setCandidatesError(err.message || '重拍对象读取失败');
      })
      .finally(() => {
        if (token === ticket.current) setCandidatesLoading(false);
      });
    return () => { ticket.current += 1; };
  }, [projectId, episodeNo, engine]);

  useEffect(() => {
    if (!projectId || !shotId) {
      setFrames([]);
      setTask(null);
      return undefined;
    }
    const token = ++ticket.current;
    setFramesLoading(true);
    setFramesError('');
    setFrames([]);
    setTask(null);
    setError('');
    setPrompt('');
    setGenSeconds(null);
    setStart(0);
    setEnd(0);
    rangeRef.current = { start: 0, end: 0 };
    durationTouched.current = false;
    setPlayhead(0);
    setVideoDuration(0);
    let cancelled = false;
    loadReshootStatus(projectId, { shotId })
      .then(view => { if (!cancelled && token === ticket.current) setTask(view); })
      .catch(() => {});
    extractReshootFrames(projectId, { shotId, start: 0, end: 0 })
      .then(view => {
        if (cancelled || token !== ticket.current) return;
        const duration = Number(view.duration) || 0;
        setFrames(view.frames || []);
        setVideoDuration(duration);
        const initialEnd = Math.min(duration, 3);
        setEnd(initialEnd);
        rangeRef.current = { start: 0, end: initialEnd };
        setGenSeconds(Math.max(MIN_SECONDS, Math.round(initialEnd)));
      })
      .catch(err => {
        if (!cancelled && token === ticket.current) setFramesError(err.message || '帧预览加载失败');
      })
      .finally(() => {
        if (!cancelled && token === ticket.current) setFramesLoading(false);
      });
    return () => { cancelled = true; ticket.current += 1; };
  }, [projectId, shotId]);

  useEffect(() => {
    if (!generating || !projectId || !shotId) return undefined;
    const timer = setInterval(() => {
      loadReshootStatus(projectId, { shotId })
        .then(view => {
          setTask(view);
          if (view.status === 'done') message.success('片段重拍视频已生成，可预览后点击更新');
          if (view.status === 'failed') message.error(`片段重拍失败：${view.error || '生成失败'}`);
        })
        .catch(() => {});
    }, 3000);
    return () => clearInterval(timer);
  }, [generating, projectId, shotId]);

  function seekFromEvent(event) {
    const rect = timelineRef.current?.getBoundingClientRect();
    const video = videoRef.current;
    if (!rect || !video || videoDuration <= 0) return;
    const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
    const at = ratio * videoDuration;
    video.currentTime = at;
    setPlayhead(at);
  }

  function beginDrag(which, event) {
    event.preventDefault();
    event.stopPropagation();
    const rect = timelineRef.current?.getBoundingClientRect();
    if (!rect || videoDuration <= 0) return;
    const move = moveEvent => {
      const ratio = Math.max(0, Math.min(1, (moveEvent.clientX - rect.left) / rect.width));
      const at = ratio * videoDuration;
      if (which === 'start') {
        const next = Math.max(0, Math.min(at, rangeRef.current.end - MIN_CLIP));
        rangeRef.current = { ...rangeRef.current, start: next };
        setStart(next);
        if (videoRef.current) videoRef.current.currentTime = next;
      } else {
        const next = Math.min(videoDuration, Math.max(at, rangeRef.current.start + MIN_CLIP));
        rangeRef.current = { ...rangeRef.current, end: next };
        setEnd(next);
        if (videoRef.current) videoRef.current.currentTime = next;
      }
      setPlayhead(at);
    };
    const up = () => {
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
      if (!durationTouched.current) {
        const seconds = Math.max(MIN_SECONDS, Math.round(rangeRef.current.end - rangeRef.current.start));
        setGenSeconds(seconds);
      }
    };
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  }

  async function submit() {
    if (!shotId || clipSeconds < MIN_CLIP || busy) return;
    setBusy('generate');
    setError('');
    try {
      const view = await startReshoot(projectId, {
        shotId,
        episode: episodeNo,
        engine,
        start,
        end,
        duration: genSeconds || Math.max(MIN_SECONDS, Math.round(clipSeconds)),
        prompt,
        mode: 'frames',
      });
      setTask(view);
      message.success('已提交重拍任务');
    } catch (err) {
      setError(err.message || '提交重拍任务失败');
    } finally {
      setBusy('');
    }
  }

  async function apply() {
    if (!shotId || busy) return;
    setBusy('apply');
    setError('');
    try {
      const view = await applyReshoot(projectId, { shotId });
      setTask(view);
      message.success('片段已更新，完整视频已保存到素材库');
      await onFinished?.();
      try {
        const refreshed = await listReshootCandidates(projectId, { episode: episodeNo, engine });
        setCandidates(refreshed.candidates || []);
      } catch {
        // 候选刷新失败不影响已经保存的成片。
      }
    } catch (err) {
      setError(err.message || '更新失败');
    } finally {
      setBusy('');
    }
  }

  function backToEdit() {
    setTask(null);
    setError('');
    if (videoRef.current) videoRef.current.currentTime = start;
  }

  const generateReason = capability({ capabilities }, 'generate').reason;
  const applyReason = capability({ capabilities }, 'apply').reason;
  const stageText = ready ? '预览新片段' : '框选片段';

  return <Modal title={null} footer={null} width={880} style={{ maxWidth: '96vw' }} open onCancel={onClose} destroyOnClose className="comic-reshoot-modal">
    <div className="comic-rs-head">
      <div className="comic-rs-title"><OriginalIcon name="FilmOutline" size={16}/><b>片段重拍</b>{selected ? <span>分镜 {selected.seq}</span> : null}<Tag>{stageText}</Tag></div>
      <Button type="text" aria-label="关闭片段重拍" icon={<OriginalIcon name="CloseOutline" size={16}/>} onClick={onClose}/>
    </div>
    <div className="comic-rs-pickers">
      <label>剧集<Select style={{ minWidth: 180 }} value={episodeNo ?? undefined} options={episodeChoices} onChange={value => setEpisodeNo(value)} placeholder="选择剧集"/></label>
      <label>分镜<Select style={{ minWidth: 260 }} value={shotId || undefined} loading={candidatesLoading} options={candidates.map(item => ({ value: item.shot_id, label: `第 ${item.seq} 镜${item.duration ? `（${Number(item.duration).toFixed(1)} 秒）` : ''}` }))} onChange={value => setShotId(value)} placeholder={candidates.length ? '选择要重拍的分镜' : '本集还没有生成好的镜头视频'}/></label>
      <span className="comic-rs-hint">只重拍框选的一段，原视频与其它镜头不受影响。</span>
    </div>
    {candidatesError ? <Alert type="error" showIcon message={candidatesError}/> : null}
    {!candidatesLoading && !candidates.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本集还没有生成好的镜头视频；请先在制作任务里生成镜头视频，再来重拍。"/> : null}
    {shotId ? <>
      <div className="comic-rs-player">
        {ready ? (task.segment_url ? <video key={task.segment_url} src={task.segment_url} controls preload="metadata"/> : <div className="comic-rs-player-empty">新片段地址读取中…</div>)
          : selected?.video_url ? <video ref={videoRef} src={selected.video_url} preload="metadata" onClick={() => { const video = videoRef.current; if (!video) return; if (video.paused) video.play().catch(() => {}); else video.pause(); }} onTimeUpdate={event => setPlayhead(event.currentTarget.currentTime)} onLoadedMetadata={event => { const duration = Number(event.currentTarget.duration) || 0; if (duration > 0) { setVideoDuration(duration); const nextEnd = rangeRef.current.end > 0 ? Math.min(rangeRef.current.end, duration) : Math.min(duration, 3); rangeRef.current = { start: 0, end: nextEnd }; setEnd(nextEnd); } }}/> : <div className="comic-rs-player-empty">视频读取中…</div>}
      </div>
      {!ready ? <>
        <div className="comic-rs-timeline" ref={timelineRef} onClick={seekFromEvent}>
          {framesLoading ? <div className="comic-rs-frames is-loading"><Spin size="small"/></div> : frames.length ? <div className="comic-rs-frames">{frames.map((src, index) => <img key={index} src={src} alt="" draggable="false"/>)}</div> : <div className="comic-rs-frames is-empty">暂无帧预览</div>}
          {videoDuration > 0 ? <>
            <div className="comic-rs-shade" style={{ left: 0, width: `${(start / videoDuration) * 100}%` }}/>
            <div className="comic-rs-shade" style={{ right: 0, width: `${(1 - end / videoDuration) * 100}%` }}/>
            <div className="comic-rs-range" style={{ left: `${(start / videoDuration) * 100}%`, right: `${(1 - end / videoDuration) * 100}%` }}/>
            <div className="comic-rs-handle is-start" style={{ left: `calc(${(start / videoDuration) * 100}% - 7px)` }} onMouseDown={event => beginDrag('start', event)}><i/></div>
            <div className="comic-rs-handle is-end" style={{ left: `calc(${(end / videoDuration) * 100}% - 7px)` }} onMouseDown={event => beginDrag('end', event)}><i/></div>
            <div className="comic-rs-playhead" style={{ left: `calc(${(playhead / videoDuration) * 100}% - 1px)` }}/>
          </> : null}
        </div>
        <div className="comic-rs-meta">
          <span>{timeText(start)} - {timeText(end)}</span>
          <span>选中片段 {clipSeconds.toFixed(1)} 秒</span>
          <span className="comic-rs-flex"/>
          <span className="comic-rs-help">生成时长</span>
          <InputNumber size="small" min={MIN_SECONDS} max={MAX_SECONDS} step={1} value={genSeconds ?? undefined} onChange={value => { durationTouched.current = true; setGenSeconds(value == null ? null : Math.max(MIN_SECONDS, Math.min(MAX_SECONDS, Math.round(Number(value) || MIN_SECONDS)))); }} addonAfter="秒"/>
        </div>
        {framesError ? <Alert type="warning" showIcon message={framesError}/> : null}
        <Input.TextArea rows={3} value={prompt} onChange={event => setPrompt(event.target.value)} placeholder="描述重拍内容（可选）。首尾帧会作为图片1、图片2交给视频模型。"/>
        <div className="comic-rs-mode">
          <Tag color="blue">首尾帧参考</Tag>
          <span className="comic-rs-help">抽首帧 / 尾帧作为图片1、图片2；当前视频通道不支持把片段作为参考视频。</span>
        </div>
        {!canGenerate && generateReason ? <Alert type="warning" showIcon message={`暂不能提交重拍：${generateReason}`}/> : null}
      </> : <div className="comic-rs-result">
        <div className="comic-rs-result-title"><OriginalIcon name="CheckmarkCircleOutline" size={18}/>新片段已生成</div>
        <p>将替换原视频 {timeText(task.start)} – {timeText(task.end)}（{(task.end - task.start).toFixed(1)} 秒）。点击「更新」拼接为完整视频，并保存到素材库。</p>
        {task.status === 'applied' && task.applied_url ? <div className="comic-rs-applied"><p>已更新为完整视频：</p><video src={task.applied_url} controls preload="metadata"/></div> : null}
      </div>}
      {generating ? <Progress percent={Number(task.progress) || 8} size="small" status="active"/> : null}
      {task?.status === 'failed' ? <Alert type="error" showIcon message={`生成失败：${task.error || '未知原因'}`}/> : null}
      {error ? <Alert type="error" showIcon message={error}/> : null}
      <div className="comic-rs-footer">
        {!ready ? <>
          <Button icon={<OriginalIcon name="RefreshOutline" size={14}/>} disabled={!shotId || busy === 'generate'} onClick={() => { const nextEnd = Math.min(videoDuration, 3); rangeRef.current = { start: 0, end: nextEnd }; durationTouched.current = false; setStart(0); setEnd(nextEnd); setGenSeconds(Math.max(MIN_SECONDS, Math.round(nextEnd))); setPrompt(''); setError(''); }}>重置框选</Button>
          <span className="comic-rs-flex"/>
          <Button type="primary" disabled={!canGenerate || generating || clipSeconds < MIN_CLIP || !!busy} loading={busy === 'generate' || generating} onClick={submit}>{generating ? `${task.message || '生成中'} ${task.progress || 0}%` : '重拍'}</Button>
        </> : <>
          <Button icon={<OriginalIcon name="RefreshOutline" size={13}/>} disabled={busy === 'apply'} onClick={backToEdit}>重新生成</Button>
          <span className="comic-rs-flex"/>
          <Button type="primary" disabled={task.status === 'applied' || !canApply || !!busy} loading={busy === 'apply'} title={canApply ? undefined : applyReason} onClick={apply}>{task.status === 'applied' ? '已更新' : busy === 'apply' ? '更新中' : '更新'}</Button>
        </>}
      </div>
    </> : null}
  </Modal>;
}
