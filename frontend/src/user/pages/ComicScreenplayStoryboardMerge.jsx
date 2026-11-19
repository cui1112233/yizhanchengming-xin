import { Alert, Button, Modal, Select, Spin, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { mergeStoryboardShots } from './comic-screenplay-storyboard-api.js';

const DURATION_OPTIONS = [
  { value: 10, label: '10 秒' },
  { value: 15, label: '15 秒' },
  { value: 30, label: '30 秒' },
];

function secondsText(value) {
  const number = Number(value || 0);
  return `${Number.isInteger(number) ? number : number.toFixed(1)} 秒`;
}

function episodeOptions(boards, episodes) {
  return (boards || [])
    .filter(board => (board.shots || []).length)
    .map(board => {
      const episode = (episodes || []).find(item => item.no === board.episode_no);
      return {
        value: board.episode_no,
        label: `第 ${board.episode_no} 集${episode?.title ? ` · ${episode.title}` : ''}（${board.shots.length} 个分镜）`,
        count: board.shots.length,
      };
    });
}

export default function ComicScreenplayStoryboardMerge({ projectId, boards, episodes, engine, readOnly, defaultEpisodeNo, onClose, onMerged }) {
  const options = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (options.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : options[0]?.value ?? null));
  const [maxSeconds, setMaxSeconds] = useState(30);
  const [plan, setPlan] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const ticket = useRef(0);

  useEffect(() => {
    if (!projectId || !episodeNo || !engine) {
      setPlan(null);
      setError('');
      setLoading(false);
      return undefined;
    }
    const token = ++ticket.current;
    setLoading(true);
    setError('');
    mergeStoryboardShots(projectId, episodeNo, engine, { maxSeconds, dryRun: true })
      .then(data => {
        if (token !== ticket.current) return;
        if (!data?.plan) throw Error('合并预览读取失败');
        setPlan(data.plan);
      })
      .catch(err => {
        if (token !== ticket.current) return;
        setPlan(null);
        setError(err.message || '合并预览读取失败');
      })
      .finally(() => { if (token === ticket.current) setLoading(false); });
    return () => { ticket.current++; };
  }, [projectId, episodeNo, engine, maxSeconds]);

  const pendingGroups = plan?.pending_groups || 0;
  const canMerge = !readOnly && !busy && !loading && !error && pendingGroups > 0;
  const previewGroups = (plan?.groups || []).slice(0, 8);

  async function submit() {
    if (!canMerge) return;
    setBusy(true);
    try {
      const data = await mergeStoryboardShots(projectId, episodeNo, engine, { maxSeconds });
      const after = data?.plan?.after_count;
      message.success(Number.isFinite(after)
        ? `第 ${episodeNo} 集已把 ${data.plan.before_count} 个分镜合并为 ${after} 个`
        : `第 ${episodeNo} 集分镜合并完成`);
      await onMerged?.(episodeNo);
      onClose?.();
    } catch (err) {
      message.error(err.message || '分镜合并失败');
    } finally {
      setBusy(false);
    }
  }

  return <Modal
    className="comic-sb-merge-modal"
    title="分镜合并"
    open
    width={620}
    style={{ maxWidth: '94vw' }}
    onCancel={() => { if (!busy) onClose?.(); }}
    maskClosable={!busy}
    footer={<>
      <Button disabled={busy} onClick={() => onClose?.()}>取消</Button>
      <Button type="primary" loading={busy} disabled={!canMerge} onClick={submit}>开始合并</Button>
    </>}
  >
    <div className="comic-sb-merge-controls">
      <label>
        <span>剧集</span>
        <Select
          aria-label="选择要合并的剧集"
          style={{ width: '100%' }}
          value={episodeNo ?? undefined}
          placeholder={options.length ? '选择剧集' : '还没有已生成分镜的剧集'}
          disabled={busy || !options.length}
          options={options.map(({ value, label }) => ({ value, label }))}
          onChange={setEpisodeNo}
          notFoundContent="还没有已生成分镜的剧集"
        />
      </label>
      <label>
        <span>最长合并时长</span>
        <Select
          aria-label="最长合并时长"
          style={{ width: '100%' }}
          value={maxSeconds}
          disabled={busy}
          options={DURATION_OPTIONS}
          onChange={value => setMaxSeconds(value)}
        />
      </label>
    </div>

    {error ? <Alert type="error" showIcon message={error} /> : null}
    {!options.length ? <Alert type="info" showIcon message="当前还没有已生成的分镜，先生成分镜后再合并。" /> : null}

    <div className="comic-sb-merge-preview" aria-live="polite">
      {loading ? <div className="comic-sb-merge-loading"><Spin size="small" /> <span>正在计算合并结果…</span></div> : null}
      {!loading && plan && !error ? <>
        <div className="comic-sb-merge-stats">
          <div><span>当前分镜</span><b>{plan.before_count}</b></div>
          <div><span>合并后</span><b>{plan.after_count}</b></div>
          <div><span>减少</span><b>{plan.removed_count}</b></div>
          <div><span>待合并</span><b>{plan.pending_groups} 组</b></div>
        </div>
        {plan.pending_groups > 0
          ? <ul className="comic-sb-merge-groups">{previewGroups.map((group, index) => <li key={group.keep_shot_id || index}>
            <span>第 {index + 1} 组</span>
            <b>{group.shot_ids?.length || 0} 个分镜 → 保留 1 个</b>
            <em>{secondsText(group.seconds)}</em>
          </li>)}</ul>
          : <Alert type="warning" showIcon message={`按最长 ${secondsText(maxSeconds)} 没有可合并的分镜，可换成更长的时长再试。`} />}
        {plan.pending_groups > previewGroups.length ? <p className="comic-sb-merge-more">还有 {plan.pending_groups - previewGroups.length} 组未列出，合并会一次处理全部。</p> : null}
        {plan.untimed_count > 0 ? <p className="comic-sb-merge-more">有 {plan.untimed_count} 个分镜没有推荐时长，不参与合并，也不会被跨过合并。</p> : null}
      </> : null}
    </div>

    <p className="comic-sb-merge-hint">每组只保留第一个分镜，其余并入后删除。合并后保留分镜的推荐时长为各分镜之和，视频提示词与台词按顺序拼接，角色、场景、道具取并集，已生成的图片和视频会保留到合并后的分镜。只整理本地分镜，不调用模型。</p>
  </Modal>;
}
