import { Button, message } from 'antd';
import { useMemo, useState } from 'react';
import { splitBatchFactoryError } from './batchFactoryErrorParse';

export { splitBatchFactoryError } from './batchFactoryErrorParse';

async function copyText(text) {
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
    }
    message.success('错误信息已复制');
  } catch (_) {
    message.error('复制失败，请手动选择错误信息');
  }
}

// 统一的批工厂错误展示：标题 + 人话摘要 + 可折叠原始信息 + 一键复制。
// 卡片、单书详情、任务日志都复用它，避免再出现“能力门控 / 上游 code / request_id”
// 被平铺成一坨、又被单行 ellipsis 切断的展示。
export function BatchFactoryErrorNotice({
  title = '自动生产失败',
  stageText = '',
  error = '',
  stale = false,
  staleText = '',
  actions = null,
  compact = false,
  className = ''
}) {
  const parts = useMemo(() => splitBatchFactoryError(error), [error]);
  const [expanded, setExpanded] = useState(false);
  const hasRaw = Boolean(parts.code || parts.message || parts.requestId || parts.raw);
  const summary = stale ? (staleText || '该失败记录已超过 7 天，不再作为当前状态展示。') : parts.summary;
  const providerSuffix = !stale && parts.upstreamRejected && parts.provider ? ` · ${parts.provider}` : '';
  const heading = `${title}${stageText ? `：${stageText}` : ''}${providerSuffix}`;
  const canExpand = !stale && hasRaw;

  return (
    <section className={`batch-factory-error-notice${compact ? ' is-compact' : ''}${className ? ` ${className}` : ''}`}>
      <div className="batch-factory-error-notice-title">{heading}</div>
      <p className={`batch-factory-error-notice-summary${expanded ? '' : ' is-clamped'}`}>{summary}</p>
      {canExpand ? (
        <button type="button" className="batch-factory-error-notice-toggle" onClick={() => setExpanded(current => !current)}>
          {expanded ? '收起原始信息' : '展开原始信息'}
        </button>
      ) : null}
      {expanded && canExpand ? (
        <div className="batch-factory-error-notice-raw">
          {parts.code ? <div className="batch-factory-error-notice-raw-row"><span>code</span><code>{parts.code}</code></div> : null}
          {parts.message ? <div className="batch-factory-error-notice-raw-row"><span>message</span><code>{parts.message}</code></div> : null}
          {parts.requestId ? <div className="batch-factory-error-notice-raw-row"><span>request_id</span><code>{parts.requestId}</code></div> : null}
          {parts.raw ? <pre className="batch-factory-error-notice-raw-full">{parts.raw}</pre> : null}
        </div>
      ) : null}
      <div className="batch-factory-error-notice-footer">
        <Button size="small" onClick={() => copyText(parts.raw || summary)}>复制错误信息</Button>
        {actions}
      </div>
    </section>
  );
}

export default BatchFactoryErrorNotice;
