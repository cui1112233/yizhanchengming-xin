import { useEffect, useRef } from 'react';
import { drawTrackingRegion, moveTrackingRegion, normalizeTrackingRegion, resizeTrackingRegion } from '../aicutTrackingRegion.js';

const handles = ['nw', 'ne', 'sw', 'se'];

export default function TrackingRegionOverlay({ active = false, region, onChange, onCommit, onCancel, label = '跟踪目标', ariaLabel = '目标跟踪框', help = '拖动画框 · 拖动框体移动 · 拖四角缩放 · Esc 退出', variant = 'tracking' }) {
  const layerRef = useRef(null);
  const gestureRef = useRef(null);
  const current = normalizeTrackingRegion(region);

  useEffect(() => {
    if (!active) gestureRef.current = null;
  }, [active]);

  useEffect(() => {
    if (!active) return undefined;
    const cancel = (event) => {
      if (event.key !== 'Escape') return;
      if (gestureRef.current) {
        onChange?.(gestureRef.current.origin);
        gestureRef.current = null;
      } else {
        onCancel?.();
      }
      event.preventDefault();
    };
    window.addEventListener('keydown', cancel);
    return () => window.removeEventListener('keydown', cancel);
  }, [active, onCancel, onChange]);

  if (!active) return null;

  function pointForEvent(event) {
    const rect = layerRef.current?.getBoundingClientRect();
    if (!rect?.width || !rect?.height) return { x: 0, y: 0 };
    return {
      x: Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width)),
      y: Math.min(1, Math.max(0, (event.clientY - rect.top) / rect.height))
    };
  }

  function begin(kind, event, handle = '') {
    if (event.button !== 0) return;
    const start = pointForEvent(event);
    gestureRef.current = { kind, handle, start, origin: normalizeTrackingRegion(region) };
    layerRef.current?.setPointerCapture?.(event.pointerId);
    event.preventDefault();
    event.stopPropagation();
  }

  function regionForEvent(event) {
    const gesture = gestureRef.current;
    if (!gesture) return current;
    const point = pointForEvent(event);
    if (gesture.kind === 'draw') return drawTrackingRegion(gesture.start, point);
    const deltaX = point.x - gesture.start.x;
    const deltaY = point.y - gesture.start.y;
    if (gesture.kind === 'resize') return resizeTrackingRegion(gesture.origin, gesture.handle, deltaX, deltaY);
    return moveTrackingRegion(gesture.origin, deltaX, deltaY);
  }

  function move(event) {
    if (!gestureRef.current) return;
    onChange?.(regionForEvent(event));
    event.preventDefault();
    event.stopPropagation();
  }

  function finish(event) {
    if (!gestureRef.current) return;
    const committed = regionForEvent(event);
    gestureRef.current = null;
    layerRef.current?.releasePointerCapture?.(event.pointerId);
    onChange?.(committed);
    onCommit?.(committed);
    event.preventDefault();
    event.stopPropagation();
  }

  function cancelGesture(event) {
    if (!gestureRef.current) return;
    const origin = gestureRef.current.origin;
    gestureRef.current = null;
    onChange?.(origin);
    event.stopPropagation();
  }

  function nudge(event) {
    const keys = { ArrowLeft: [-1, 0], ArrowRight: [1, 0], ArrowUp: [0, -1], ArrowDown: [0, 1] };
    if (!keys[event.key]) return;
    const step = event.shiftKey ? .05 : .01;
    const [x, y] = keys[event.key];
    const next = moveTrackingRegion(current, x * step, y * step);
    onChange?.(next);
    onCommit?.(next);
    event.preventDefault();
    event.stopPropagation();
  }

  const style = {
    left: `${(current.centerX - current.width / 2) * 100}%`,
    top: `${(current.centerY - current.height / 2) * 100}%`,
    width: `${current.width * 100}%`,
    height: `${current.height * 100}%`
  };

  return <div
    ref={layerRef}
    className={`aicut-tracking-region-overlay aicut-region-overlay-${variant}`}
    aria-label={`${ariaLabel}编辑层`}
    onPointerDown={(event) => begin('draw', event)}
    onPointerMove={move}
    onPointerUp={finish}
    onPointerCancel={cancelGesture}
  >
    <div
      className="aicut-tracking-region-box"
      style={style}
      role="application"
      tabIndex={0}
      aria-label={ariaLabel}
      aria-description="拖动框可移动目标，拖动四角可缩放，方向键微调，按 Escape 取消框选。"
      onKeyDown={nudge}
      onPointerDown={(event) => begin('move', event)}
    >
      <span className="aicut-tracking-region-label">{label}</span>
      {handles.map((handle) => <button
        type="button"
        key={handle}
        className={`aicut-tracking-region-handle handle-${handle}`}
        aria-label={`缩放目标框${handle}`}
        onPointerDown={(event) => begin('resize', event, handle)}
      />)}
    </div>
    <div className="aicut-tracking-region-help">{help}</div>
  </div>;
}
