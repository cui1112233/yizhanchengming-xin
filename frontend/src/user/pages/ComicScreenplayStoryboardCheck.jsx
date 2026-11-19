import { Alert, Button, Modal, Select, Spin, Tag, message } from 'antd';
import { useMemo, useState } from 'react';
import { checkStoryboardShots, fillStoryboardCheckShot } from './comic-screenplay-storyboard-api.js';

function episodeOptions(boards, episodes) {
  return (boards || [])
    .filter(board => (board.shots || []).length)
    .map(board => {
      const episode = (episodes || []).find(item => item.no === board.episode_no);
      return {
        value: board.episode_no,
        label: `第 ${board.episode_no} 集${episode?.title ? ` · ${episode.title}` : ''}（${board.shots.length} 个分镜）`,
      };
    });
}

function numberList(values) {
  return (values || []).join(', ');
}

function formatReport(data) {
  const report = data?.report;
  if (!report) return null;
  return {
    total: report.total || 0,
    expectedTotal: report.expected_total || 0,
    minSeq: report.min_seq || 0,
    maxSeq: report.max_seq || 0,
    missing: report.missing || [],
    extra: report.extra || [],
    duplicates: report.duplicates || [],
    invalid: report.invalid || [],
    passed: report.passed === true,
  };
}

// 分镜检查（洛水 storyboard_check 技能）：检查分镜序号的缺失 / 重复 / 异常ID，
// 生成检查报告（可下载为 storyboard_check_report.json），缺失分镜逐条调用 AI 补齐插入。
// 洛水按「原文序号行」对照出缺失与多余（越界）；本系统分镜由规划链生成、seq 即镜头序号，
// 检查基准是「序号 1..最大序号 连续完整」，报告里的多余（extra）字段恒为空。
export default function ComicScreenplayStoryboardCheck({ projectId, boards, episodes, engine, modelId, readOnly, defaultEpisodeNo, onClose, onChanged }) {
  const options = useMemo(() => episodeOptions(boards, episodes), [boards, episodes]);
  const [episodeNo, setEpisodeNo] = useState(() => (options.some(item => item.value === defaultEpisodeNo) ? defaultEpisodeNo : options[0]?.value ?? null));
  const [busy, setBusy] = useState(false);
  const [checking, setChecking] = useState(false);
  const [report, setReport] = useState(null);
  const [progress, setProgress] = useState(null); // {current,total}
  const [failures, setFailures] = useState([]); // [{seq,error}]
  const [summary, setSummary] = useState(null); // {filled,total,failed}
  const [error, setError] = useState('');

  const running = progress != null;
  const canFill = Boolean(projectId && engine && !readOnly && !running && (modelId || '').trim());

  async function runCheck() {
    if (!projectId || !episodeNo || !engine) return;
    setBusy(true); setChecking(true); setError(''); setReport(null); setSummary(null); setFailures([]);
    try {
      const data = await checkStoryboardShots(projectId, { episode: episodeNo, engine });
      const next = formatReport(data);
      if (!next) throw Error(data?.message || '检查报告读取失败');
      setReport(next);
      if (next.passed) message.success('检查通过：序号连续，无缺失、重复或异常');
    } catch (err) {
      setError(err.message || '分镜检查失败');
    } finally {
      setBusy(false);
      setChecking(false);
    }
  }

  async function fillMissing() {
    if (!canFill || !report?.missing.length) return;
    setBusy(true); setError(''); setSummary(null); setFailures([]);
    const targets = [...report.missing];
    setProgress({ current: 0, total: targets.length });
    let filled = 0;
    const failed = [];
    for (let index = 0; index < targets.length; index++) {
      setProgress({ current: index, total: targets.length });
      try {
        await fillStoryboardCheckShot(projectId, { episode: episodeNo, engine, seq: targets[index], model: modelId });
        filled += 1;
      } catch (err) {
        failed.push({ seq: targets[index], error: err.message || '补全失败' });
        setFailures([...failed]);
      }
      setProgress({ current: index + 1, total: targets.length });
    }
    setProgress(null);
    setSummary({ filled, total: targets.length, failed: failed.length });
    if (failed.length) message.warning(`分镜补全完成：已补全 ${filled}/${targets.length}，失败 ${failed.length}`);
    else message.success(`分镜补全完成：已补全 ${filled} 个缺失分镜`);
    setBusy(false);
    try {
      const data = await checkStoryboardShots(projectId, { episode: episodeNo, engine });
      const next = formatReport(data);
      if (next) setReport(next);
    } catch { /* 保留原报告，失败明细已展示 */ }
    await onChanged?.(episodeNo);
  }

  function downloadReport() {
    if (!report) return;
    const blob = new Blob([JSON.stringify({ storyboard_check_report: report }, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = 'storyboard_check_report.json';
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
  }

  return <Modal
    className="comic-sb-check-modal"
    title="分镜检查"
    open
    width={640}
    style={{ maxWidth: '94vw' }}
    onCancel={() => { if (!busy) onClose?.(); }}
    maskClosable={!busy}
    footer={<>
      {report ? <Button disabled={busy} onClick={downloadReport}>下载报告 storyboard_check_report.json</Button> : null}
      <Button disabled={busy} onClick={() => onClose?.()}>关闭</Button>
      <Button type="primary" loading={checking} disabled={!projectId || !episodeNo || !engine || busy} onClick={runCheck}>{report ? '重新检查' : '开始检查'}</Button>
    </>}
  >
    <p className="comic-sb-check-desc">检查本集分镜序号的缺失、重复与异常ID，生成检查报告；缺失分镜可由当前作品文本模型逐条补齐并插入。只改分镜列表，不动已生成的图片和视频。</p>
    <div className="comic-sb-check-controls">
      <label>
        <span>剧集</span>
        <Select aria-label="选择要检查的剧集" style={{ width: '100%' }} value={episodeNo ?? undefined}
          placeholder={options.length ? '选择剧集' : '还没有已生成分镜的剧集'} disabled={busy || !options.length}
          options={options.map(({ value, label: text }) => ({ value, label: text }))} onChange={value => { setEpisodeNo(value); setReport(null); setSummary(null); setFailures([]); setError(''); }}
          notFoundContent="还没有已生成分镜的剧集" />
      </label>
    </div>
    {!modelId ? <Alert type="warning" showIcon message="还没有选择作品文本模型，在页头选好模型后再补齐缺失分镜。" /> : null}
    {error ? <Alert type={report ? 'warning' : 'error'} showIcon message={error} /> : null}
    {report ? <div className="comic-sb-check-report" aria-live="polite">
      <div className="comic-sb-check-stats">
        <div><b>{report.total}</b><span>个分镜</span></div>
        <div><b>{report.minSeq}-{report.maxSeq}</b><span>序号范围</span></div>
        <div><b>{report.missing.length}</b><span>缺失</span></div>
        <div><b>{report.duplicates.length}</b><span>重复</span></div>
        <div><b>{report.invalid.length}</b><span>异常ID</span></div>
      </div>
      {report.passed ? <Tag color="green">检查通过：序号连续，无缺失、重复或异常</Tag> : null}
      {report.missing.length ? <div className="comic-sb-check-section is-missing">
        <div><b>【缺失的分镜】</b><span>序号范围内缺席</span></div>
        <div>缺失编号: {numberList(report.missing)}</div>
        <div className="comic-sb-check-notice">发现 {report.missing.length} 个缺失分镜，建议重新处理对应批次，或由 AI 补齐</div>
        <Button type="primary" size="small" loading={running} disabled={!canFill} onClick={fillMissing}>AI 补齐缺失分镜</Button>
      </div> : null}
      {report.duplicates.length ? <div className="comic-sb-check-section is-warn">
        <div><b>【重复的分镜】</b><span>同一编号出现多次</span></div>
        <div>重复编号: {numberList(report.duplicates)}</div>
      </div> : null}
      {report.invalid.length ? <div className="comic-sb-check-section is-warn">
        <div><b>【异常ID】</b><span>序号异常（小于等于 0）</span></div>
        <div>异常编号: {numberList(report.invalid)}</div>
      </div> : null}
    </div> : null}
    <div className="comic-sb-check-progress" aria-live="polite">
      {running ? <div className="comic-sb-check-loading"><Spin size="small" /> <span>正在补全缺失分镜：第 {progress.current + 1}/{progress.total} 条（序号 {report?.missing?.[progress.current]}）</span></div> : null}
      {summary && !running ? <Tag color={summary.failed ? 'orange' : 'green'}>分镜补全完成：已补全 {summary.filled}/{summary.total}{summary.failed ? `，失败 ${summary.failed}` : ''}</Tag> : null}
      {failures.length ? <ul className="comic-sb-check-failures">{failures.slice(0, 8).map(item => <li key={item.seq}>序号 {item.seq}：{item.error}</li>)}</ul> : null}
    </div>
    <p className="comic-sb-check-hint">洛水的分镜检查按「原文序号行」对照 AI 结果，还包含「多余（越界）」项；本系统分镜由规划链生成、序号即镜头序号，因此按 1..最大序号的连续性检查，报告中的多余（extra）字段恒为空。</p>
  </Modal>;
}
