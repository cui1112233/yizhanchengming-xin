import { Alert, Button, InputNumber, Modal, Select, Spin, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { applyCostumeChange, loadCostumeOptions, previewCostumeChange } from './comic-screenplay-storyboard-api.js';

const UNLIMITED = '__unlimited__';

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

// 换装：洛水 2026-08-18 版「按场景筛选分镜，把绑定的人物 A 批量换成 A-睡衣或其他人物资产」，
// 2026-09-05 版补上分镜范围与「不限制」场景。整集一次性保存，命中规则与外观提示都由后端给出。
export default function ComicScreenplayCostumeChange({ projectId, boards, episodes, engine, readOnly, defaultEpisodeNo, onClose, onChanged }) {
  const options = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (options.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : options[0]?.value ?? null));
  const [shotStart, setShotStart] = useState(1);
  const [shotEnd, setShotEnd] = useState(() => options.find(item => item.value === episodeNo)?.count ?? 1);
  const [scene, setScene] = useState('');
  const [from, setFrom] = useState(null);
  const [to, setTo] = useState(null);
  const [optionsData, setOptionsData] = useState({ characters: [], scenes: [], shot_count: 0 });
  const [optionsLoading, setOptionsLoading] = useState(false);
  const [optionsError, setOptionsError] = useState('');
  const [plan, setPlan] = useState(null);
  const [previewError, setPreviewError] = useState('');
  const [previewMessage, setPreviewMessage] = useState('');
  const [previewing, setPreviewing] = useState(false);
  const [busy, setBusy] = useState(false);
  const ticket = useRef(0);

  const shotCount = optionsData.shot_count || options.find(item => item.value === episodeNo)?.count || 0;

  useEffect(() => {
    if (!projectId || !episodeNo || !engine) {
      setOptionsData({ characters: [], scenes: [], shot_count: 0 });
      return undefined;
    }
    let alive = true;
    setOptionsLoading(true);
    setOptionsError('');
    loadCostumeOptions(projectId, { episode: episodeNo, engine })
      .then(data => {
        if (!alive) return;
        const next = data?.options || { characters: [], scenes: [], shot_count: 0 };
        setOptionsData(next);
        setShotStart(1);
        setShotEnd(next.shot_count || 1);
        setScene('');
        setFrom(null);
        setTo(null);
      })
      .catch(err => { if (alive) setOptionsError(err.message || '换装选项读取失败'); })
      .finally(() => { if (alive) setOptionsLoading(false); });
    return () => { alive = false; };
  }, [projectId, episodeNo, engine]);

  const fromMissing = Boolean(from) && !(optionsData.characters || []).includes(from);
  const previewReady = Boolean(projectId && episodeNo && engine && from && to && from !== to && !readOnly);
  const previewKey = `${episodeNo}|${shotStart}|${shotEnd}|${scene}|${from}|${to}`;

  useEffect(() => {
    if (!previewReady) {
      setPlan(null);
      setPreviewError('');
      setPreviewing(false);
      return undefined;
    }
    const token = ++ticket.current;
    setPreviewing(true);
    setPreviewError('');
    setPreviewMessage('');
    const timer = setTimeout(() => {
      previewCostumeChange(projectId, { episode: episodeNo, engine, shotStart, shotEnd, scene: scene === UNLIMITED ? '' : scene, from, to })
        .then(data => {
          if (token !== ticket.current) return;
          setPlan(data?.plan || null);
          setPreviewMessage(data?.message || '');
        })
        .catch(err => {
          if (token !== ticket.current) return;
          setPlan(null);
          setPreviewError(err.message || '换装预览读取失败');
        })
        .finally(() => { if (token === ticket.current) setPreviewing(false); });
    }, 260);
    return () => { clearTimeout(timer); ticket.current++; };
    // previewKey 覆盖全部预览输入，避免把对象依赖写成不稳定引用。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [previewKey, previewReady, projectId, engine]);

  const matched = plan?.matched || 0;
  const canApply = previewReady && !busy && !previewing && !previewError && matched > 0;
  const previewTargets = (plan?.targets || []).slice(0, 6);

  async function submit() {
    if (!canApply) return;
    setBusy(true);
    try {
      const data = await applyCostumeChange(projectId, { episode: episodeNo, engine, shotStart, shotEnd, scene: scene === UNLIMITED ? '' : scene, from, to });
      message.success(data?.message || `已在第 ${episodeNo} 集完成换装`);
      await onChanged?.(episodeNo);
      onClose?.();
    } catch (err) {
      message.error(err.message || '换装失败');
    } finally {
      setBusy(false);
    }
  }

  return <Modal
    className="comic-sb-costume-modal"
    title="换装"
    open
    width={640}
    style={{ maxWidth: '94vw' }}
    onCancel={() => { if (!busy) onClose?.(); }}
    maskClosable={!busy}
    footer={<>
      <Button disabled={busy} onClick={() => onClose?.()}>取消</Button>
      <Button type="primary" loading={busy} disabled={!canApply} onClick={submit}>开始换装</Button>
    </>}
  >
    <p className="comic-sb-costume-desc">先选分镜范围；筛选场景可选「不限制」。再把范围内绑定的人物 A 换成 A-睡衣或其他人物。已是目标人物、未绑定原人物、或不在范围内的分镜不会改。</p>

    <div className="comic-sb-costume-controls">
      <label>
        <span>剧集</span>
        <Select
          aria-label="选择要换装的剧集"
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
        <span>分镜范围</span>
        <span className="comic-sb-costume-range">
          <InputNumber aria-label="起始分镜" min={1} max={shotCount || 1} value={shotStart} disabled={busy || optionsLoading} onChange={value => setShotStart(Math.max(1, Math.min(shotCount || 1, Number(value) || 1)))} />
          <em>到</em>
          <InputNumber aria-label="结束分镜" min={1} max={shotCount || 1} value={shotEnd} disabled={busy || optionsLoading} onChange={value => setShotEnd(Math.max(1, Math.min(shotCount || 1, Number(value) || 1)))} />
          <em className="comic-sb-costume-total">共 {shotCount || 0} 个分镜</em>
        </span>
      </label>
      <label>
        <span>筛选场景</span>
        <Select
          aria-label="筛选场景"
          style={{ width: '100%' }}
          value={scene || UNLIMITED}
          disabled={busy || optionsLoading}
          onChange={value => setScene(value === UNLIMITED ? '' : value)}
          options={[{ value: UNLIMITED, label: '不限制场景' }, ...(optionsData.scenes || []).map(name => ({ value: name, label: name }))]}
        />
      </label>
      <label>
        <span>原人物</span>
        <Select
          aria-label="原人物"
          style={{ width: '100%' }}
          showSearch
          value={from || undefined}
          placeholder="选择要被换掉的人物"
          disabled={busy || optionsLoading}
          onChange={value => setFrom(value)}
          options={(optionsData.characters || []).map(name => ({ value: name, label: name }))}
          notFoundContent="本集分镜与人物资产里还没有可选人物"
        />
      </label>
      <label>
        <span>替换为</span>
        <Select
          aria-label="替换为"
          style={{ width: '100%' }}
          showSearch
          value={to || undefined}
          placeholder="选择目标人物资产"
          disabled={busy || optionsLoading}
          onChange={value => setTo(value)}
          options={(optionsData.characters || []).map(name => ({ value: name, label: name }))}
          notFoundContent="人物资产里还没有可选人物"
        />
      </label>
    </div>

    {optionsError ? <Alert type="error" showIcon message={optionsError} /> : null}
    {!options.length ? <Alert type="info" showIcon message="当前还没有已生成的分镜，先生成分镜后再换装。" /> : null}
    {fromMissing ? <Alert type="warning" showIcon message={`「${from}」不在本集人物里，可能是资产被删掉了，重新选一个原人物。`} /> : null}

    <div className="comic-sb-costume-preview" aria-live="polite">
      {optionsLoading || previewing ? <div className="comic-sb-costume-loading"><Spin size="small" /> <span>正在统计命中的分镜…</span></div> : null}
      {previewError && !previewing ? <Alert type="error" showIcon message={previewError} /> : null}
      {!optionsLoading && !previewing && plan && !previewError ? <>
        <div className="comic-sb-costume-stats">
          <div><span>范围内分镜</span><b>{plan.shot_end - plan.shot_start + 1}</b></div>
          <div><span>场景</span><b>{plan.scene || '不限制'}</b></div>
          <div><span>命中</span><b>{plan.matched}</b></div>
        </div>
        {matched > 0
          ? <ul className="comic-sb-costume-list">{previewTargets.map(target => <li key={target.shot_id}>
            <span>分镜 {target.seq}{target.scene ? ` · ${target.scene}` : ''}</span>
            <b>{target.before || '无'}</b>
            <em>→</em>
            <b>{target.after || '无'}</b>
          </li>)}</ul>
          : <Alert type="warning" showIcon message={previewMessage || `没有可换装的分镜：分镜 ${plan.shot_start}-${plan.shot_end} 中需绑定「${plan.from}」，并且还不是「${plan.to}」`} />}
        {matched > previewTargets.length ? <p className="comic-sb-costume-more">还有 {matched - previewTargets.length} 个分镜未列出，换装会一次处理全部命中分镜。</p> : null}
      </> : null}
      {busy ? <div className="comic-sb-costume-progress"><Spin size="small" /> <span>正在更新 {matched} 个分镜…</span></div> : null}
    </div>

    <p className="comic-sb-costume-hint">替换时保留人物在分镜中的顺序，@图片N 仍对得上。只改人物名，不动已生成的图片和视频。</p>
  </Modal>;
}
