import { useMemo, useState } from 'react';
import { Images } from 'lucide-react';
import { buildComicDeliverables } from '../pages/agent-comic-deliverables';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import './agent-comic-deliverable-panel.css';

function PublishRow({ label, entry }) {
  return (
    <div className="agent-comic-panel__publish-row">
      <dt>{label}</dt>
      <dd className={`agent-comic-panel__publish-state${entry.status ? '' : ' is-unsynced'}`}>
        {entry.label}
        {entry.detail ? <span className="agent-comic-panel__publish-detail">{entry.detail}</span> : null}
        {!entry.detail && entry.submittedAt ? <span className="agent-comic-panel__publish-detail">{entry.submittedAt}</span> : null}
      </dd>
    </div>
  );
}

// 只读面板：只消费画布上批量工厂投影出来的漫画节点，不发起任何写入。
export function AgentComicDeliverablePanel({ nodes = [] }) {
  const groups = useMemo(() => buildComicDeliverables(nodes), [nodes]);
  const [collapsed, setCollapsed] = useState(false);
  if (!groups.length) return null;

  return (
    <section className="agent-comic-panel" aria-label="漫画交付物">
      <header className="agent-comic-panel__header">
        <Images size={14} aria-hidden="true" />
        <strong>漫画交付物</strong>
        <span className="agent-comic-panel__readonly">只读</span>
        <button
          type="button"
          className="agent-comic-panel__toggle nodrag nopan"
          aria-expanded={!collapsed}
          onClick={() => setCollapsed(value => !value)}
        >{collapsed ? '展开' : '收起'}</button>
      </header>

      {collapsed ? null : groups.map(group => (
        <article className="agent-comic-panel__book" key={group.key}>
          <header className="agent-comic-panel__book-header">
            <span className="agent-comic-panel__book-title" title={group.bookLabel}>{group.bookLabel}</span>
            <span className="agent-comic-panel__progress">{group.progress.label}</span>
          </header>

          <h4 className="agent-comic-panel__section-title">漫画页</h4>
          {group.pages.length ? (
            <div className="agent-comic-panel__pages">
              {group.pages.map(page => (
                <figure className="agent-comic-panel__page" key={page.id || page.title}>
                  {page.url
                    ? <AuthenticatedAgentImage source={page.url} alt={page.title} loading="lazy" />
                    : <span className="agent-comic-panel__page-placeholder">未同步</span>}
                  <figcaption>
                    <span className="agent-comic-panel__page-title" title={page.title}>{page.title}</span>
                    <span className={`agent-comic-panel__state is-${page.state}`}>{page.statusLabel}</span>
                  </figcaption>
                </figure>
              ))}
            </div>
          ) : <p className="agent-comic-panel__empty">未同步</p>}

          <h4 className="agent-comic-panel__section-title">合集</h4>
          {group.collections.length ? (
            <ul className="agent-comic-panel__collections">
              {group.collections.map(collection => (
                <li key={collection.id || collection.title}>
                  <span className="agent-comic-panel__collection-title" title={collection.title}>{collection.title}</span>
                  <span className="agent-comic-panel__collection-count">{collection.pageCountLabel}</span>
                  <span className={`agent-comic-panel__state${collection.n8Status ? '' : ' is-unsynced'}`}>{collection.statusLabel}</span>
                </li>
              ))}
            </ul>
          ) : <p className="agent-comic-panel__empty">未同步</p>}

          <h4 className="agent-comic-panel__section-title">发布状态</h4>
          <dl className="agent-comic-panel__publish">
            <PublishRow label="N8 漫画" entry={group.publish.n8} />
            <PublishRow label="121" entry={group.publish.provider121} />
            <PublishRow label="YD" entry={group.publish.yd} />
          </dl>
        </article>
      ))}
    </section>
  );
}
