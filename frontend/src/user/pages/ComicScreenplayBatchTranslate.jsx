import { Alert, Button, InputNumber, Modal, Select, Spin, Tag, message } from 'antd';
import { useMemo, useState } from 'react';
import { prepareBatchTranslate, translateStoryboardItem } from './comic-screenplay-storyboard-api.js';

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

const FIELD_OPTIONS = [{ value: 'image', label: '画面提示词' }, { value: 'video', label: '视频提示词' }];
// 洛水方向选项原文：中文 → 英文 / 英文 → 中文。
const DIRECTION_OPTIONS = [{ value: 'zh2en', label: '中文 → 英文' }, { value: 'en2zh', label: '英文 → 中文' }];
// 洛水范围含「未锁定 / 已锁定」；V88 分镜没有锁定概念，不做假选项。
const SCOPE_OPTIONS = [{ value: 'all', label: '全部' }, { value: 'incomplete', label: '未完成' }, { value: 'custom', label: '自定义范围' }];
const FIELD_NAME = { image: '画面提示词', video: '视频提示词' };
const DIRECTION_NAME = { zh2en: '中→英', en2zh: '英→中' };

// 批量翻译：洛水「批量操作 · 翻译」页——先按类型/方向/范围筛出待翻译分镜（空提示词回退台词），
// 再逐条发给当前作品文本模型，翻译完写回原字段，进度与「翻译完成：成功 X/Y」提示与洛水一致。
export default function ComicScreenplayBatchTranslate({ projectId, boards, episodes, engine, modelId, readOnly, defaultEpisodeNo, onClose, onChanged }) {
  const options = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (options.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : options[0]?.value ?? null));
  const [field, setField] = useState('image');
  const [direction, setDirection] = useState('zh2en');
  const [scope, setScope] = useState('all');
  const [shotStart, setShotStart] = useState(1);
  const [shotEnd, setShotEnd] = useState(() => options.find(item => item.value === episodeNo)?.count ?? 1);
  const [busy, setBusy] = useState(false);
  const [progress, setProgress] = useState(null); // {current,total,label}
  const [failures, setFailures] = useState([]); // [{seq,error}]
  const [summary, setSummary] = useState(null); // {translated,total,failed}
  const [error, setError] = useState('');

  const shotCount = options.find(item => item.value === episodeNo)?.count || 0;
  const running = progress != null;
  const canStart = Boolean(projectId && episodeNo && engine && !readOnly && !running && (modelId || '').trim());
  const label = `${FIELD_NAME[field]}（${DIRECTION_NAME[direction]}）`;

  async function start() {
    if (!canStart) return;
    if (scope === 'custom' && shotEnd < shotStart) {
      setError('自定义范围无效：结束编号不能小于起始编号');
      return;
    }
    setBusy(true); setError(''); setFailures([]); setSummary(null);
    try {
      const input = { episode: episodeNo, engine, field, direction, scope, shotStart, shotEnd };
      const data = await prepareBatchTranslate(projectId, input);
      const items = Array.isArray(data?.items) ? data.items : [];
      if (!items.length) {
        setSummary({ translated: 0, total: 0, failed: 0 });
        setError(data?.message || '没有需要翻译的分镜');
        return;
      }
      setProgress({ current: 0, total: items.length, label });
      let translated = 0;
      const failed = [];
      for (let index = 0; index < items.length; index++) {
        setProgress({ current: index, total: items.length, label });
        try {
          await translateStoryboardItem(projectId, { ...input, shotId: items[index].shot_id, model: modelId });
          translated += 1;
        } catch (err) {
          failed.push({ seq: items[index].seq, error: err.message || '翻译失败' });
          setFailures([...failed]);
        }
        setProgress({ current: index + 1, total: items.length, label });
      }
      setSummary({ translated, total: items.length, failed: failed.length });
      if (failed.length) message.warning(`翻译完成：成功 ${translated}/${items.length}，失败 ${failed.length}`);
      else message.success(`翻译完成：成功 ${translated}/${items.length}`);
      await onChanged?.(episodeNo);
    } catch (err) {
      setError(err.message || '批量翻译失败');
    } finally {
      setBusy(false);
      setProgress(null);
    }
  }

  return <Modal
    className="comic-sb-translate-modal"
    title="批量翻译"
    open
    width={640}
    style={{ maxWidth: '94vw' }}
    onCancel={() => { if (!busy) onClose?.(); }}
    maskClosable={!busy}
    footer={<>
      <Button disabled={busy} onClick={() => onClose?.()}>{running ? '关闭窗口（翻译继续按条提交，停止再点取消）' : '取消'}</Button>
      <Button type="primary" loading={busy} disabled={!canStart} onClick={start}>开始翻译</Button>
    </>}
  >
    <p className="comic-sb-translate-desc">把分镜提示词逐条发给当前作品文本模型翻译后写回原字段；提示词为空时回退台词。只改文字，不动已生成的图片和视频。</p>
    <div className="comic-sb-translate-controls">
      <label>
        <span>剧集</span>
        <Select aria-label="选择要翻译的剧集" style={{ width: '100%' }} value={episodeNo ?? undefined}
          placeholder={options.length ? '选择剧集' : '还没有已生成分镜的剧集'} disabled={busy || !options.length}
          options={options.map(({ value, label: text }) => ({ value, label: text }))} onChange={value => { setEpisodeNo(value); setShotStart(1); setShotEnd(options.find(item => item.value === value)?.count || 1); setSummary(null); setFailures([]); setError(''); }}
          notFoundContent="还没有已生成分镜的剧集" />
      </label>
      <label>
        <span>类型</span>
        <Select aria-label="翻译类型" style={{ width: '100%' }} value={field} options={FIELD_OPTIONS} disabled={busy} onChange={value => { setField(value); setSummary(null); }} />
      </label>
      <label>
        <span>翻译方向</span>
        <Select aria-label="翻译方向" style={{ width: '100%' }} value={direction} options={DIRECTION_OPTIONS} disabled={busy} onChange={value => { setDirection(value); setSummary(null); }} />
      </label>
      <label>
        <span>范围</span>
        <Select aria-label="翻译范围" style={{ width: '100%' }} value={scope} options={SCOPE_OPTIONS} disabled={busy} onChange={value => { setScope(value); setSummary(null); }} />
      </label>
      {scope === 'custom' ? <label>
        <span>分镜范围</span>
        <span className="comic-sb-translate-range">
          <InputNumber aria-label="起始分镜" min={1} max={shotCount || 1} value={shotStart} disabled={busy} onChange={value => setShotStart(Math.max(1, Math.min(shotCount || 1, Number(value) || 1)))} />
          <em>至</em>
          <InputNumber aria-label="结束分镜" min={1} max={shotCount || 1} value={shotEnd} disabled={busy} onChange={value => setShotEnd(Math.max(1, Math.min(shotCount || 1, Number(value) || 1)))} />
          <em className="comic-sb-translate-total">共 {shotCount} 个分镜</em>
        </span>
      </label> : null}
    </div>
    {!modelId ? <Alert type="warning" showIcon message="还没有选择作品文本模型，在页头选好模型后再翻译。" /> : null}
    {error ? <Alert type={summary ? 'warning' : 'error'} showIcon message={error} /> : null}
    <div className="comic-sb-translate-progress" aria-live="polite">
      {running ? <div className="comic-sb-translate-loading"><Spin size="small" /> <span>正在翻译 {progress.label}：{progress.current}/{progress.total}</span></div> : null}
      {summary && !running ? <div className="comic-sb-translate-summary">
        <Tag color={summary.failed ? 'orange' : 'green'}>翻译完成：成功 {summary.translated}/{summary.total}{summary.failed ? `，失败 ${summary.failed}` : ''}</Tag>
      </div> : null}
      {failures.length ? <ul className="comic-sb-translate-failures">{failures.slice(0, 8).map(item => <li key={item.seq}>分镜 {item.seq}：{item.error}</li>)}</ul> : null}
    </div>
    <p className="comic-sb-translate-hint">「未完成」= 该字段提示词非空的分镜（有文本才可翻译）。洛水的范围还包括按分镜锁定状态筛选；本系统分镜暂无锁定概念，未提供该选项。</p>
  </Modal>;
}
