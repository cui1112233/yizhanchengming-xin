import { Alert, Button, Empty, Modal, Select, Tag } from 'antd';
import { useEffect, useState } from 'react';
import { reviewScreenplayDoctor, reviewStoryboardDoctor } from './comic-screenplay-storyboard-api.js';

const severityMeta = {
  blocker: { color: 'red', label: '阻断' },
  major: { color: 'orange', label: '重大' },
  minor: { color: 'blue', label: '优化' },
};

const doctorMeta = {
  screenplay: {
    title: '剧本医生',
    description: '审核当前剧本的剧情逻辑、人物动机、资产引用、节奏推进、镜头缺失与台词违和，只出报告不改稿。',
    empty: '还没有剧本内容，先 AI 生成或导入后再来',
  },
  storyboard: {
    title: '分镜医生',
    description: '对照剧本审核分镜提示词的人物错位、场景不连贯、资产引用错误与 video_desc 结构缺失，只出报告不改稿。',
    empty: '本集还没有分镜，先生成分镜后再来',
  },
};

function severityOf(value) {
  return severityMeta[String(value || '').toLowerCase()] || severityMeta.minor;
}

// 医生审核报告（洛水 wsvp / luoshui_doctor_shared review 规范对齐）：
// pass 横幅 + issues（severity 着色）+ suggested_actions + summary。
export function DoctorReviewReport({ review, emptyText }) {
  if (!review) return <Empty description={emptyText || '还没有审核报告，点击「开始审核」运行'} />;
  const actions = Array.isArray(review.suggested_actions) ? review.suggested_actions.filter(Boolean) : [];
  const issues = Array.isArray(review.issues) ? review.issues : [];
  return <div className="doctor-review-report">
    <Alert className="doctor-review-banner" type={review.pass ? 'success' : 'error'} showIcon
      message={review.pass ? '审核通过' : '审核未通过'}
      description={review.pass ? (review.summary || '没有发现阻断问题；以下仍列出可优化项。') : (review.summary || '存在阻断问题，建议修复后再进入下一步。')} />
    <div className="doctor-review-stats">
      <div><b>{issues.length}</b><span>问题</span></div>
      <div><b>{issues.filter(item => item.severity === 'blocker').length}</b><span>阻断</span></div>
      <div><b>{issues.filter(item => item.severity === 'major').length}</b><span>重大</span></div>
      <div><b>{issues.filter(item => item.severity === 'minor').length}</b><span>优化</span></div>
    </div>
    {issues.length ? <ul className="doctor-review-issues">
      {issues.map(issue => {
        const meta = severityOf(issue.severity);
        return <li key={issue.id} className={`is-${String(issue.severity || 'minor').toLowerCase()}`}>
          <div className="doctor-review-issue-head">
            <Tag color={meta.color} bordered={false}>{meta.label}</Tag>
            {issue.dimension ? <span className="doctor-review-dimension">{issue.dimension}</span> : null}
            {issue.location ? <span className="doctor-review-location">{issue.location}</span> : null}
            {issue.episode ? <span className="doctor-review-location">第{issue.episode}集</span> : null}
          </div>
          <div className="doctor-review-problem">{issue.problem}</div>
          {issue.suggestion ? <div className="doctor-review-suggestion">建议：{issue.suggestion}</div> : null}
          {issue.evidence ? <div className="doctor-review-evidence">原文：{issue.evidence}</div> : null}
        </li>;
      })}
    </ul> : null}
    {actions.length ? <div className="doctor-review-actions">建议动作：{actions.map(action => <Tag key={action}>{action}</Tag>)}</div> : null}
    <div className="doctor-review-scope">审核时间 {String(review.reviewed_at || '').slice(0, 19).replace('T', ' ')}</div>
  </div>;
}

// 医生审核弹窗：kind=screenplay（当前剧本全文）/ storyboard（指定集分镜对照剧本）。
export default function ComicScreenplayDoctor({ projectId, kind = 'screenplay', engine, episodes = [], defaultEpisodeNo, modelId, readOnly = false, hasContent = true, open = false, onClose = () => {} }) {
  const [running, setRunning] = useState(false);
  const [review, setReview] = useState(null);
  const [error, setError] = useState('');
  const [episodeNo, setEpisodeNo] = useState(Number(defaultEpisodeNo) || (episodes[0]?.no) || 1);
  const meta = doctorMeta[kind] || doctorMeta.screenplay;

  useEffect(() => {
    if (open) {
      setReview(null);
      setError('');
      setRunning(false);
      setEpisodeNo(Number(defaultEpisodeNo) || (episodes[0]?.no) || 1);
    }
  }, [open]); // eslint-disable-line react-hooks/exhaustive-deps

  async function run() {
    if (running) return;
    setError('');
    setRunning(true);
    try {
      const request = kind === 'storyboard'
        ? reviewStoryboardDoctor(projectId, { episode: episodeNo, engine, model: modelId })
        : reviewScreenplayDoctor(projectId, { model: modelId });
      const response = await request;
      setReview(response?.review || null);
    } catch (err) {
      setError(err.message || '审核失败');
    } finally {
      setRunning(false);
    }
  }

  return <Modal className="comic-doctor-modal" width="min(860px, 96vw)"
    title={<span className="doctor-title">{meta.title}<span className="doctor-sub">{meta.description}</span></span>}
    open={open} onCancel={onClose}
    footer={<><Button onClick={onClose}>关闭</Button><Button type="primary" loading={running} disabled={readOnly || !modelId || (!hasContent && kind === 'storyboard')} onClick={run}>开始审核</Button></>}>
    {readOnly ? <Alert type="info" showIcon message="当前作品只有查看权限，不能运行医生审核" /> : null}
    {!modelId ? <Alert type="info" showIcon message="请先在页头选择并保存文本模型" /> : null}
    {!hasContent && kind === 'storyboard' ? <Alert type="info" showIcon message={meta.empty} /> : null}
    {kind === 'storyboard' ? <label className="doctor-episode-row"><span>审核剧集</span><Select value={episodeNo} onChange={setEpisodeNo} disabled={readOnly || running} options={episodes.map(item => ({ value: item.no, label: `第 ${item.no} 集` }))} /></label> : null}
    {error ? <Alert type="error" showIcon message={error} /> : null}
    <DoctorReviewReport review={review} emptyText={meta.empty} />
  </Modal>;
}
