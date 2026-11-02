import { Button, Space, Tooltip, Typography } from 'antd';
import { ArrowLeft } from 'lucide-react';
import { useEffect, useRef } from 'react';
import { cylinderVisualState, nextCylinderIndex } from './novel-fetch-cylinder.js';

export function NovelFetchSubpageShell({ title, description, onBack, actions, children, className = '' }) {
  return <section className={`novel-fetch-subpage ${className}`.trim()}>
    <header className="novel-fetch-subpage-header">
      <Tooltip title="返回处理页">
        <Button className="novel-fetch-subpage-back" aria-label="返回处理页" icon={<ArrowLeft size={17} />} onClick={onBack} />
      </Tooltip>
      <div className="novel-fetch-subpage-heading">
        <Typography.Title level={3}>{title}</Typography.Title>
        {description ? <Typography.Text type="secondary">{description}</Typography.Text> : null}
      </div>
      {actions ? <Space className="novel-fetch-subpage-actions" wrap>{actions}</Space> : null}
    </header>
    <div className="novel-fetch-subpage-content">{children}</div>
  </section>;
}

export function NovelFetchMetricStrip({ items }) {
  return <div className="novel-fetch-metric-strip">
    {(items || []).map(item => {
      const Tag = item.onClick ? 'button' : 'div';
      return <Tag type={item.onClick ? 'button' : undefined} className={`novel-fetch-metric${item.active ? ' is-active' : ''}`} key={item.key || item.label} onClick={item.onClick}>
      <Typography.Text type="secondary">{item.label}</Typography.Text>
      <strong>{item.value ?? 0}</strong>
      {item.description ? <span>{item.description}</span> : null}
    </Tag>;
    })}
  </div>;
}

export function NovelFetchCylinderList({
  items = [],
  selectedId,
  onSelect,
  getId = (item, index) => item?.id ?? index,
  getTitle = item => item?.title || item?.name || item?.id || '未命名',
  getMeta = item => item?.category || item?.style || '',
  ariaLabel = '知识库条目'
}) {
  const viewportRef = useRef(null);
  const itemRefs = useRef(new Map());
  const rafRef = useRef(0);
  const scrollCommitTimerRef = useRef(0);
  const nearestIdRef = useRef(null);
  const selectedIdRef = useRef(selectedId);
  const onSelectRef = useRef(onSelect);

  useEffect(() => {
    selectedIdRef.current = selectedId;
    onSelectRef.current = onSelect;
  }, [selectedId, onSelect]);

  function updateVisuals() {
    rafRef.current = 0;
    const viewport = viewportRef.current;
    if (!viewport) return;
    const viewportRect = viewport.getBoundingClientRect();
    const center = viewportRect.top + viewportRect.height / 2;
    let nearestItem = null;
    let nearestDistance = Number.POSITIVE_INFINITY;

    items.forEach((item, index) => {
      const id = String(getId(item, index));
      const element = itemRefs.current.get(id);
      if (!element) return;
      const rect = element.getBoundingClientRect();
      const distance = (rect.top + rect.height / 2 - center) / Math.max(rect.height + 4, 1);
      const visual = cylinderVisualState(distance);
      element.style.setProperty('--cylinder-rotate', `${visual.rotateX}deg`);
      element.style.setProperty('--cylinder-depth', `${visual.translateZ}px`);
      element.style.setProperty('--cylinder-scale', String(visual.scale));
      element.style.setProperty('--cylinder-opacity', String(visual.opacity));
      element.style.setProperty('--cylinder-brightness', String(visual.brightness));
      if (Math.abs(distance) < nearestDistance) {
        nearestDistance = Math.abs(distance);
        nearestItem = item;
      }
    });

    if (nearestItem) {
      const nearestIndex = items.indexOf(nearestItem);
      const nearestId = getId(nearestItem, nearestIndex);
      nearestIdRef.current = nearestId;
    }
  }

  function scheduleVisualUpdate() {
    if (!rafRef.current) rafRef.current = window.requestAnimationFrame(updateVisuals);
  }

  function commitCenteredItem() {
    scrollCommitTimerRef.current = 0;
    updateVisuals();
    const nearestId = nearestIdRef.current;
    if (nearestId !== null && String(nearestId) !== String(selectedIdRef.current ?? '')) {
      onSelectRef.current?.(nearestId);
    }
  }

  function handleScroll() {
    scheduleVisualUpdate();
    if (scrollCommitTimerRef.current) window.clearTimeout(scrollCommitTimerRef.current);
    scrollCommitTimerRef.current = window.setTimeout(commitCenteredItem, 100);
  }

  function scrollItemIntoView(id, behavior = 'smooth') {
    itemRefs.current.get(String(id))?.scrollIntoView({ block: 'center', behavior });
  }

  useEffect(() => {
    scheduleVisualUpdate();
    const handleResize = () => scheduleVisualUpdate();
    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
      if (rafRef.current) window.cancelAnimationFrame(rafRef.current);
      if (scrollCommitTimerRef.current) window.clearTimeout(scrollCommitTimerRef.current);
    };
  }, [items]);

  useEffect(() => {
    if (selectedId !== undefined && selectedId !== null && selectedId !== '') scrollItemIntoView(selectedId);
  }, [selectedId]);

  function handleKeyDown(event) {
    if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
    event.preventDefault();
    const currentIndex = Math.max(0, items.findIndex((item, index) => String(getId(item, index)) === String(selectedId ?? '')));
    const nextIndex = nextCylinderIndex(items.length, currentIndex, event.key === 'ArrowUp' ? -1 : 1);
    const nextItem = items[nextIndex];
    if (!nextItem) return;
    const nextId = getId(nextItem, nextIndex);
    onSelectRef.current?.(nextId);
    scrollItemIntoView(nextId);
  }

  return <div className="novel-fetch-cylinder-shell">
    <div className="novel-fetch-cylinder-guide" aria-hidden="true" />
    <div
      ref={viewportRef}
      className="novel-fetch-cylinder-list"
      role="listbox"
      tabIndex={0}
      aria-label={ariaLabel}
      onScroll={handleScroll}
      onKeyDown={handleKeyDown}
    >
      {items.map((item, index) => {
        const rawId = getId(item, index);
        const id = String(rawId);
        const selected = id === String(selectedId ?? '');
        return <button
          type="button"
          role="option"
          aria-selected={selected}
          className={`novel-fetch-cylinder-item${selected ? ' is-selected' : ''}`}
          key={id}
          ref={element => {
            if (element) itemRefs.current.set(id, element);
            else itemRefs.current.delete(id);
          }}
          onClick={() => {
            onSelectRef.current?.(rawId);
            scrollItemIntoView(rawId);
          }}
        >
          <strong>{getTitle(item)}</strong>
          <span>{getMeta(item) || '未分类'}</span>
        </button>;
      })}
    </div>
  </div>;
}
