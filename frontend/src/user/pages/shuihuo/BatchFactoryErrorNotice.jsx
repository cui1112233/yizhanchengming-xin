import { Button, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
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
// lineMode 只用于批次数表底部的持久失败记录：折叠态必须是一条 28px 细线，
// 点“详情”才展开原始信息与重试入口；其它位置保持原来的面板样式。
export function BatchFactoryErrorNotice({
  title = '自动生产失败',
  stageText = '',
  error = '',
  stale = false,
  staleText = '',
  actions = null,
  compact = false,
  lineMode = false,
  focusKey = '',
  anchorId = '',
  className = ''
}) {
  const parts = useMemo(() => splitBatchFactoryError(error), [error]);
  const [expanded, setExpanded] = useState(false);
  const hasRaw = Boolean(parts.code || parts.message || parts.requestId || parts.raw);
  const summary = stale ? (staleText || '该失败记录已超过 7 天，不再作为当前状态展示。') : parts.summary;
  const providerSuffix = !stale && parts.upstreamRejected && parts.provider ? ` · ${parts.provider}` : '';
  const heading = `${title}${stageText ? `：${stageText}` : ''}${providerSuffix}`;
  const canExpand = lineMode ? Boolean(hasRaw || actions) : (!stale && hasRaw);
  // 折叠态的一句话：阶段失败原因 + 上游 + 原因；上游原文里的技术串压成短标签。
  const lineReason = !stale && parts.upstreamRejected && /ref_image_0/i.test(parts.raw)
    ? 'ref_image_0 未包含'
    : summary;
  const lineText = [stageText ? `${stageText}失败` : title, lineMode && !stale ? parts.provider : '', lineReason]
    .filter(Boolean)
    .join(' · ');

  useEffect(() => {
    // 顶部 toast 的“查看详情”通过 focusKey 把对应的持久记录展开。
    if (lineMode && focusKey) setExpanded(true);
  }, [focusKey, lineMode]);

  if (lineMode) {
    return (
      <section id={anchorId || undefined} className={`batch-factory-error-notice is-line${expanded ? ' is-expanded' : ''}${className ? ` ${className}` : ''}`}>
        <div className="batch-factory-error-notice-line">
          <span className="batch-factory-error-notice-line-text" title={lineText}>{lineText}</span>
          {canExpand ? (
            <button type="button" className="batch-factory-error-notice-line-btn" onClick={() => setExpanded(current => !current)}>
              {expanded ? '收起' : '详情'}
            </button>
          ) : null}
        </div>
        {expanded && canExpand ? (
          <>
            <div className="batch-factory-error-notice-raw">
              {parts.code ? <div className="batch-factory-error-notice-raw-row"><span>code</span><code>{parts.code}</code></div> : null}
              {parts.message ? <div className="batch-factory-error-notice-raw-row"><span>message</span><code>{parts.message}</code></div> : null}
              {parts.requestId ? <div className="batch-factory-error-notice-raw-row"><span>request_id</span><code>{parts.requestId}</code></div> : null}
              {parts.raw ? <pre className="batch-factory-error-notice-raw-full">{parts.raw}</pre> : null}
            </div>
            <div className="batch-factory-error-notice-footer">
              <Button size="small" onClick={() => copyText(parts.raw || summary)}>复制错误信息</Button>
              {actions}
            </div>
          </>
        ) : null}
      </section>
    );
  }

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
