import { useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { fetchLatestProductUpdate } from '../api/productUpdates';
import { markProductUpdateSeen, normalizeProductUpdate, resolveProductUpdateStorage, shouldShowProductUpdate } from '../productUpdate';
import '../styles/product-update-modal.css';

// Wait for the first paint before opening: the changelog must not compete with
// the page the user came for.
const SHOW_DELAY_MS = 1500;

export function ProductUpdateModal({ enabled = true }) {
  const [note, setNote] = useState(null);
  const [open, setOpen] = useState(false);
  const confirmRef = useRef(null);

  useEffect(() => {
    if (!enabled) return undefined;
    let cancelled = false;
    const timer = window.setTimeout(() => {
      fetchLatestProductUpdate()
        .then(payload => {
          if (cancelled) return;
          const normalized = normalizeProductUpdate(payload);
          if (!normalized) return;
          if (!shouldShowProductUpdate(normalized.releaseId, resolveProductUpdateStorage())) return;
          setNote(normalized);
          setOpen(true);
        })
        .catch(() => {
          // A missing or failed endpoint is silent by design.
        });
    }, SHOW_DELAY_MS);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [enabled]);

  const close = useCallback(() => {
    setOpen(false);
    if (note?.releaseId) markProductUpdateSeen(resolveProductUpdateStorage(), note.releaseId);
  }, [note]);

  useEffect(() => {
    if (!open) return undefined;
    confirmRef.current?.focus();
    const handleKeyDown = event => {
      if (event.key === 'Escape') close();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, close]);

  if (!open || !note || typeof document === 'undefined') return null;

  return createPortal(
    <div
      className="product-update-overlay"
      onMouseDown={event => {
        if (event.target === event.currentTarget) close();
      }}
    >
      <div className="product-update-modal" role="dialog" aria-modal="true" aria-labelledby="product-update-title">
        <header className="product-update-head">
          <div className="product-update-head-main">
            {note.version ? <span className="product-update-version">v{note.version}</span> : null}
            <h2 className="product-update-title" id="product-update-title">{note.title}</h2>
          </div>
          <div className="product-update-head-meta">
            {note.releasedAt ? <time className="product-update-date" dateTime={note.releasedAt}>{note.releasedAt}</time> : null}
            <button type="button" className="product-update-close" aria-label="关闭更新说明" onClick={close}>
              <span aria-hidden="true">×</span>
            </button>
          </div>
        </header>
        {note.summary ? <p className="product-update-summary">{note.summary}</p> : null}
        <div className="product-update-body">
          {note.sections.map(section => (
            <section className="product-update-section" data-type={section.type} key={`${section.type}-${section.label}`}>
              <h3 className="product-update-section-title">
                <span className="product-update-dot" aria-hidden="true" />
                {section.label}
              </h3>
              <ul className="product-update-list">
                {section.items.map(item => (
                  <li className="product-update-item" key={`${section.type}-${item.title}`}>
                    <div className="product-update-item-main">
                      <h4 className="product-update-item-title">{item.title}</h4>
                      {item.desc ? <p className="product-update-item-desc">{item.desc}</p> : null}
                      {item.tags.length > 0 ? (
                        <div className="product-update-tags">
                          {item.tags.map(tag => <span className="product-update-tag" key={tag}>{tag}</span>)}
                        </div>
                      ) : null}
                    </div>
                    {item.link ? (
                      <a className="product-update-item-link" href={item.link} target="_blank" rel="noreferrer" onClick={close}>
                        体验 →
                      </a>
                    ) : null}
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
        <footer className="product-update-foot">
          <button ref={confirmRef} type="button" className="product-update-confirm" onClick={close}>
            我知道了
          </button>
        </footer>
      </div>
    </div>,
    document.body
  );
}
