import { Alert, Button, Empty, Input, Modal, Progress, Select, Spin, Tabs, Tag, message } from 'antd';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { combineLongVideo, loadLongVideoStatus, planLongVideo, startLongVideo } from './comic-screenplay-storyboard-api.js';

const TARGET_OPTIONS = [
  { value: 100, label: '100 秒' },
  { value: 200, label: '200 秒' },
  { value: 250, label: '250 秒' },
  { value: 300, label: '300 秒' },
];
const FILTERS = [
  { key: 'all', label: '全部' },
  { key: 'idle', label: '待生成' },
  { key: 'generating', label: '生成中' },
  { key: 'done', label: '已完成' },
  { key: 'failed', label: '失败' },
];
const STATUS_TEXT = { idle: '待生成', generating: '生成中', done: '已完成', failed: '失败' };
const STATUS_COLOR = { idle: 'default', generating: 'processing', done: 'success', failed: 'error' };

function secondsText(value) {
  const number = Number(value || 0);
  return `${Number.isInteger(number) ? number : number.toFixed(1)} 秒`;
}

function sectionKey(section) {
  return `${section.index}:${(section.shot_ids || []).join(',')}`;
}

function episodeOptions(boards, episodes) {
  return (boards || [])
    .filter(board => (board.shots || []).length)
    .map(board => {
      const episode = (episodes || []).find(item => item.no === board.episode_no);
      return { value: board.episode_no, label: `第 ${board.episode_no} 集${episode?.title ? ` · ${episode.title}` : ''}（${board.shots.length} 个分镜）` };
    });
}

export default function ComicScreenplayLongVideo({ projectId, boards, episodes, engine, readOnly, modelName, defaultEpisodeNo, onClose, onFinished }) {
  const options = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (options.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : options[0]?.value ?? null));
  const [targetSeconds, setTargetSeconds] = useState(300);
  const [planView, setPlanView] = useState(null);
  const [savedView, setSavedView] = useState(null);
  const [prompts, setPrompts] = useState({});
  const [filter, setFilter] = useState('all');
  const [selectedIndex, setSelectedIndex] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState('');
  const ticket = useRef(0);
  const pollTicket = useRef(0);

  const refreshSaved = useCallback(async () => {
    if (!projectId || !episodeNo) return null;
    try {
      const view = await loadLongVideoStatus(projectId, { episode: episodeNo });
      setSavedView(view);
      return view;
    } catch {
      setSavedView(null);
      return null;
    }
  }, [projectId, episodeNo]);

  useEffect(() => {
    if (!projectId || !episodeNo) {
      setPlanView(null);
      setError('');
      return undefined;
    }
    const token = ++ticket.current;
    setLoading(true);
    setError('');
    Promise.all([
      planLongVideo(projectId, { episode: episodeNo, engine, targetSeconds }),
      loadLongVideoStatus(projectId, { episode: episodeNo }).catch(() => null),
    ])
      .then(([plan, saved]) => {
        if (token !== ticket.current) return;
        setPlanView(plan);
        setSavedView(saved);
        setPrompts(current => {
          const next = { ...current };
          (plan?.sections || []).forEach(section => { if (!next[sectionKey(section)]) next[sectionKey(section)] = section.prompt || ''; });
          return next;
        });
        setSelectedIndex(plan?.sections?.[0]?.index ?? null);
      })
      .catch(err => {
        if (token !== ticket.current) return;
        setPlanView(null);
        setError(err.message || '超长视频编排读取失败');
      })
      .finally(() => { if (token === ticket.current) setLoading(false); });
    return () => { ticket.current++; };
  }, [projectId, episodeNo, engine, targetSeconds]);

  const sections = useMemo(() => {
    const planned = planView?.sections || [];
    const saved = new Map((savedView?.sections || []).map(section => [sectionKey(section), section]));
    return planned.map(section => {
      const record = saved.get(sectionKey(section));
      return {
        ...section,
        status: record?.status || 'idle',
        progress: record?.progress || 0,
        video_url: record?.video_url || '',
        error: record?.error || '',
        prompt: prompts[sectionKey(section)] ?? section.prompt ?? '',
      };
    });
  }, [planView, savedView, prompts]);

  const running = savedView?.running === true || sections.some(section => section.status === 'generating') || savedView?.final_status === 'combining';

  useEffect(() => {
    if (!running || !projectId || !episodeNo) return undefined;
    const token = ++pollTicket.current;
    const timer = setInterval(() => {
      if (token !== pollTicket.current) return;
      refreshSaved().then(view => {
        if (view?.final_status === 'done' && token === pollTicket.current) onFinished?.();
      });
    }, 2500);
    return () => { pollTicket.current++; clearInterval(timer); };
  }, [running, projectId, episodeNo, refreshSaved, onFinished]);

  const visible = sections.filter(section => (filter === 'all' ? true : section.status === filter));
  const doneCount = sections.filter(section => section.status === 'done').length;
  const plannedSeconds = sections.reduce((sum, section) => sum + Number(section.total_seconds || 0), 0);
  const selected = sections.find(section => section.index === selectedIndex) || visible[0] || null;
  const allDone = sections.length > 0 && doneCount === sections.length;
  const finalStatus = savedView?.final_status || 'idle';
  const capabilities = savedView?.capabilities || planView?.capabilities || {};
  const canGenerate = !readOnly && !loading && !error && sections.length > 0 && capabilities.generate?.available !== false;
  const generateBlockedReason = capabilities.generate?.reason || '需要先在设置中保存视频模型';

  async function runGenerate(indexes) {
    if (!canGenerate || busy) return;
    setBusy(indexes?.length ? 'group' : 'all');
    try {
      const payloadSections = sections.map(section => ({ index: section.index, prompt: section.prompt, shot_ids: section.shot_ids, total_seconds: section.total_seconds, label: section.label }));
      const view = await startLongVideo(projectId, { episode: episodeNo, engine, targetSeconds, indexes, sections: payloadSections });
      setSavedView(view);
      message.success(indexes?.length ? `已提交第 ${indexes.join('、')} 组生成` : '已提交全部待生成的组');
    } catch (err) {
      message.error(err.message || '提交生成失败');
    } finally {
      setBusy('');
    }
  }

  async function runCombine() {
    if (busy) return;
    setBusy('combine');
    try {
      const view = await combineLongVideo(projectId, { episode: episodeNo });
      setSavedView(view);
      message.success('已开始合并成片');
    } catch (err) {
      message.error(err.message || '合并成片失败');
    } finally {
      setBusy('');
    }
  }

  const tabs = FILTERS.map(item => ({ key: item.key, label: `${item.label}${item.key === 'all' ? `（${sections.length}）` : ''}` }));

  return <Modal
    className="comic-lv-modal"
    title="超长视频"
    open
    width={1080}
    style={{ maxWidth: '96vw', top: 24 }}
    onCancel={() => onClose?.()}
    footer={<div className="comic-lv-footer"><span>单组不超过 30 秒，全部组累计不超过所选目标；选 300 秒也可能实际生成 280 秒，不会超。</span><div><Button onClick={() => onClose?.()}>关闭</Button><Button type="primary" disabled={!allDone || busy === 'combine' || finalStatus === 'combining'} loading={busy === 'combine' || finalStatus === 'combining'} onClick={runCombine}>合并成片</Button></div></div>}
  >
    <div className="comic-lv-controls">
      <label><span>剧集</span>
        <Select aria-label="选择剧集" style={{ width: '100%' }} value={episodeNo ?? undefined} placeholder={options.length ? '选择剧集' : '还没有已生成分镜的剧集'} disabled={!!busy} options={options} onChange={setEpisodeNo} notFoundContent="还没有已生成分镜的剧集" />
      </label>
      <label><span>目标总时长</span>
        <Select aria-label="目标总时长" style={{ width: '100%' }} value={targetSeconds} disabled={!!busy} options={TARGET_OPTIONS} onChange={setTargetSeconds} />
      </label>
      <div className="comic-lv-summary">
        <div><span>分组</span><b>{sections.length} 组</b></div>
        <div><span>已完成</span><b>{doneCount}/{sections.length}</b></div>
        <div><span>预计时长</span><b>{secondsText(plannedSeconds)}</b></div>
        <div><span>生成模型</span><b>{modelName || '跟随设置'}</b></div>
      </div>
      <div className="comic-lv-actions">
        <Button type="primary" loading={busy === 'all'} disabled={!canGenerate || !!busy || allDone} title={canGenerate ? '' : generateBlockedReason} onClick={() => runGenerate()}>批量生成</Button>
      </div>
    </div>

    {error ? <Alert type="error" showIcon message={error} /> : null}
    {!options.length ? <Alert type="info" showIcon message="当前还没有已生成的分镜，先生成分镜后再编排超长视频。" /> : null}
    {savedView && savedView.capabilities?.generate && savedView.capabilities.generate.available === false ? <Alert type="warning" showIcon message={`还不能生成视频：${savedView.capabilities.generate.reason}`} /> : null}
    {!savedView && planView?.capabilities?.generate && planView.capabilities.generate.available === false ? <Alert type="warning" showIcon message={`还不能生成视频：${planView.capabilities.generate.reason}`} /> : null}

    {loading ? <div className="comic-lv-loading"><Spin size="small" /> <span>正在按推荐时长编排分组…</span></div> : null}

    {!loading && sections.length ? <div className="comic-lv-body">
      <div className="comic-lv-side">
        <Tabs size="small" activeKey={filter} onChange={setFilter} items={tabs} />
        <ul className="comic-lv-list">
          {visible.map(section => <li key={section.index}>
            <button type="button" className={`comic-lv-row${selected?.index === section.index ? ' is-active' : ''}`} onClick={() => setSelectedIndex(section.index)}>
              <span className="comic-lv-row-title">{section.label}</span>
              <span className="comic-lv-row-sub">{section.shot_numbers || `${(section.shot_ids || []).length} 个分镜`} · {secondsText(section.total_seconds)}</span>
              <span className="comic-lv-row-state"><Tag color={STATUS_COLOR[section.status] || 'default'}>{STATUS_TEXT[section.status] || section.status}</Tag></span>
              {section.status === 'generating' ? <Progress percent={section.progress || 0} size="small" showInfo={false} /> : null}
              {section.status === 'failed' && section.error ? <span className="comic-lv-row-error">{section.error}</span> : null}
            </button>
          </li>)}
        </ul>
        {!visible.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该状态下没有分组" /> : null}
      </div>

      <div className="comic-lv-main">
        {selected ? <>
          <div className="comic-lv-main-head">
            <div>
              <h3>{selected.label}</h3>
              <p>{selected.shot_numbers || `${(selected.shot_ids || []).length} 个分镜`} · 推荐时长 {secondsText(selected.total_seconds)}</p>
            </div>
            <div className="comic-lv-main-actions">
              <Button size="small" loading={busy === 'group'} disabled={!canGenerate || !!busy || selected.status === 'generating' || running} onClick={() => runGenerate([selected.index])}>{selected.status === 'done' || selected.status === 'failed' ? '重新生成该组' : '生成该组'}</Button>
            </div>
          </div>
          {selected.status === 'generating' ? <Progress percent={selected.progress || 0} size="small" /> : null}
          {selected.error ? <Alert type="error" showIcon message={selected.error} /> : null}
          <div className="comic-lv-video">
            {selected.video_url ? <video src={selected.video_url} controls preload="none" /> : <div className="comic-lv-video-empty">过程视频：生成完成后在这里播放</div>}
          </div>
          <label className="comic-lv-prompt">
            <span>本组视频提示词（可单独修改）</span>
            <Input.TextArea rows={7} value={selected.prompt} disabled={readOnly || !!busy || selected.status === 'generating'} onChange={event => setPrompts(current => ({ ...current, [sectionKey(selected)]: event.target.value }))} />
          </label>
          {selected.characters || selected.scene || selected.props ? <div className="comic-lv-merged">
            {selected.characters ? <p>人物：{selected.characters}</p> : null}
            {selected.scene ? <p>场景：{selected.scene}</p> : null}
            {selected.props ? <p>道具：{selected.props}</p> : null}
          </div> : null}
        </> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="选择左侧分组查看提示词与过程视频" />}
      </div>
    </div> : null}

    {savedView?.final_url ? <div className="comic-lv-final">
      <h4>成片预览</h4>
      <video src={savedView.final_url} controls preload="none" />
      {savedView.final_status === 'done' ? <Button size="small" onClick={() => window.open(savedView.final_url, '_blank', 'noopener')}>下载成片</Button> : null}
    </div> : null}
    {finalStatus === 'failed' && savedView?.final_error ? <Alert type="error" showIcon message={`合并成片失败：${savedView.final_error}`} /> : null}
  </Modal>;
}
