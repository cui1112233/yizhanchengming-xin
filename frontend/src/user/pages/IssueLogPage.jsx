import { Button, Empty, Tag, Typography, message } from 'antd';
import { Bug, KeyRound, RefreshCw, ServerCrash } from 'lucide-react';
import { useEffect, useState } from 'react';
import { listMyErrorLogs } from '../../shared/api/client';
import { listBatchRewriteIssues } from '../../shared/api/novelFetchWorkshop';

function formatTime(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false });
}

function readableKind(kind) {
  if (kind === 'client.api-network') return '网络请求失败';
  if (kind === 'client.api-response') return '操作请求失败';
  if (kind === 'client.window-error') return '页面运行异常';
  if (kind === 'client.unhandledrejection') return '操作未完成';
  if (kind === 'batch-rewrite.submit') return '小说获取：网站提交问题';
  if (kind === 'batch-rewrite.sensitive') return '小说获取：敏感词处理问题';
  if (kind === 'batch-rewrite.ai') return '小说获取：AI 文案处理问题';
  if (kind === 'batch-rewrite.processing') return '小说获取：原文处理问题';
  if (kind.startsWith('batch-factory.')) return `批量工厂：${kind.slice('batch-factory.'.length)}`;
  if (kind.startsWith('client.batch-rewrite.submit')) return '小说获取：网站提交问题';
  if (kind.startsWith('client.batch-rewrite.processing')) return '小说获取：批量处理问题';
  if (kind.startsWith('client.batch-rewrite.rules')) return '小说获取：处理规则问题';
  if (kind.startsWith('client.batch-rewrite.api')) return '小说获取：接口问题';
  return '系统记录';
}

function severityForEntry(entry) {
  if (entry.status === 401 || entry.status === 403) return 'warning';
  if (entry.kind === 'client.api-network' || entry.kind.startsWith('batch-rewrite.') || entry.kind.startsWith('batch-factory.') || (Number.isInteger(entry.status) && entry.status >= 400)) return 'error';
  return 'neutral';
}

function entryIcon(entry) {
  const severity = severityForEntry(entry);
  if (severity === 'warning') return <KeyRound size={17} aria-hidden="true" />;
  if (severity === 'error') return <ServerCrash size={17} aria-hidden="true" />;
  return <Bug size={17} aria-hidden="true" />;
}

function summarizeEntries(entries, now = Date.now()) {
  const dayAgo = now - 24 * 60 * 60 * 1000;
  return entries.reduce((summary, entry) => {
    const timestamp = new Date(entry.at).getTime();
    return {
      recent: summary.recent + (Number.isFinite(timestamp) && timestamp >= dayAgo ? 1 : 0),
      api: summary.api + (entry.kind === 'client.api-response' || entry.kind === 'client.api-network' || entry.kind.startsWith('batch-rewrite.') || entry.kind.startsWith('batch-factory.') ? 1 : 0),
      access: summary.access + (entry.status === 401 || entry.status === 403 ? 1 : 0),
    };
  }, { recent: 0, api: 0, access: 0 });
}

export function IssueLogPage() {
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);

  async function loadEntries() {
    setLoading(true);
    try {
      const [errorsResult, batchRewriteResult] = await Promise.allSettled([listMyErrorLogs(), listBatchRewriteIssues()]);
      if (errorsResult.status === 'rejected') throw errorsResult.reason;
      const errors = errorsResult.value;
      const batchRewrite = batchRewriteResult.status === 'fulfilled' ? batchRewriteResult.value : { entries: [] };
      const combined = [
        ...(Array.isArray(errors.entries) ? errors.entries : []),
        ...(Array.isArray(batchRewrite.entries) ? batchRewrite.entries : []),
      ].sort((left, right) => new Date(right.at).getTime() - new Date(left.at).getTime());
      setEntries(combined);
    } catch (error) {
      message.error(error.message || '读取问题日志失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void loadEntries(); }, []);
  const summary = summarizeEntries(entries);

  return <section className="utility-page issue-log-page">
    <div className="issue-log-heading">
      <div><Typography.Title level={3}>问题日志</Typography.Title><Typography.Paragraph>系统自动记录使用过程中的异常，例如接口 404、认证失败、服务器错误和网络中断。</Typography.Paragraph></div>
      <div className="issue-log-actions">
        <Button icon={<RefreshCw size={16} aria-hidden="true" />} loading={loading} onClick={loadEntries}>刷新</Button>
      </div>
    </div>
    <div className="issue-log-summary-grid" aria-label="异常概览">
      <div className="issue-log-summary-card"><span>最近 24 小时异常</span><strong>{summary.recent}</strong></div>
      <div className="issue-log-summary-card"><span>接口失败</span><strong>{summary.api}</strong></div>
      <div className="issue-log-summary-card"><span>认证或配置问题</span><strong>{summary.access}</strong></div>
    </div>
    {entries.length ? <ol className="issue-log-timeline" aria-label="按发生时间排列的问题日志">
      {entries.map(entry => <li className={`issue-log-entry issue-log-entry--${severityForEntry(entry)}`} key={entry.id}>
        <span className="issue-log-marker">{entryIcon(entry)}</span>
        <article className="issue-log-entry-card">
          <div className="issue-log-entry-topline">
            <div><span className="issue-log-kind">{readableKind(entry.kind)}</span><time dateTime={entry.at}>{formatTime(entry.at)}</time></div>
            {Number.isInteger(entry.status) ? <Tag className="issue-log-status">HTTP {entry.status}</Tag> : <Tag className="issue-log-status">{entry.method || '系统记录'}</Tag>}
          </div>
          <p className="issue-log-message">{entry.message}</p>
          <div className="issue-log-meta">
            {entry.path ? <span>{entry.path}</span> : null}
            {entry.method && Number.isInteger(entry.status) ? <span>{entry.method}</span> : null}
          </div>
        </article>
      </li>)}
    </ol> : <Empty description={loading ? '正在读取问题日志' : '暂时没有问题日志'} />}
  </section>;
}
