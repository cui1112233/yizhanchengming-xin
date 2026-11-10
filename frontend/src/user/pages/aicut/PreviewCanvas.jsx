import { useEffect, useRef, useState } from 'react';
import { dragTransform, normalizeTransform, resizeTransform, rotateTransform } from './canvasTransform.js';

export default function PreviewCanvas({ enabled, transform, onCommitTransform }) {
  const layerRef = useRef(null);
  const gestureRef = useRef(null);
  const [draft, setDraft] = useState(() => normalizeTransform(transform));

  useEffect(() => setDraft(normalizeTransform(transform)), [transform]);

  useEffect(() => {
    const cancel = (event) => {
      if (event.key !== 'Escape' || !gestureRef.current) return;
      setDraft(gestureRef.current.origin);
      gestureRef.current = null;
    };
    window.addEventListener('keydown', cancel);
    return () => window.removeEventListener('keydown', cancel);
  }, []);

  function begin(kind, event) {
    if (!enabled || event.button !== 0) return;
    const rect = layerRef.current?.getBoundingClientRect();
    if (!rect) return;
    gestureRef.current = {
      kind,
      origin: normalizeTransform(draft),
      start: { x: event.clientX, y: event.clientY },
      center: { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 },
      canvas: { width: rect.width, height: rect.height },
    };
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.preventDefault();
  }

  function transformForEvent(event) {
    const gesture = gestureRef.current;
    if (!gesture) return draft;
    const point = { x: event.clientX, y: event.clientY };
    const delta = { x: point.x - gesture.start.x, y: point.y - gesture.start.y };
    if (gesture.kind === 'drag') return dragTransform(gesture.origin, delta, gesture.canvas);
    if (gesture.kind === 'resize') return resizeTransform(gesture.origin, delta, gesture.canvas, { lockRatio: !event.shiftKey });
    if (gesture.kind === 'rotate') return rotateTransform(gesture.origin, gesture.start, point, gesture.center);
    return gesture.origin;
  }

  function move(event) {
    if (!gestureRef.current) return;
    setDraft(transformForEvent(event));
  }

  function finish(event) {
    if (!gestureRef.current) return;
    const committed = transformForEvent(event);
    setDraft(committed);
    gestureRef.current = null;
    onCommitTransform(normalizeTransform(committed));
  }

  const style = {
    opacity: draft.opacity,
    transform: `translate(${draft.x * 100}%, ${draft.y * 100}%) rotate(${draft.rotation}deg) scale(${draft.scaleX}, ${draft.scaleY})`,
  };

  return <div ref={layerRef} className={`aicut-canvas-direct-layer${enabled ? ' enabled' : ''}`} aria-label="画布直接操作层">
    {enabled ? <div className="aicut-canvas-selection" style={style} onPointerDown={(event) => begin('drag', event)} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => { setDraft(gestureRef.current?.origin || draft); gestureRef.current = null; }}>
      <button type="button" className="handle corner-nw" aria-label="缩放左上角" onPointerDown={(event) => { event.stopPropagation(); begin('resize', event); }} />
      <button type="button" className="handle corner-ne" aria-label="缩放右上角" onPointerDown={(event) => { event.stopPropagation(); begin('resize', event); }} />
      <button type="button" className="handle corner-sw" aria-label="缩放左下角" onPointerDown={(event) => { event.stopPropagation(); begin('resize', event); }} />
      <button type="button" className="handle corner-se" aria-label="缩放右下角" onPointerDown={(event) => { event.stopPropagation(); begin('resize', event); }} />
      <button type="button" className="handle rotate" aria-label="旋转画面" onPointerDown={(event) => { event.stopPropagation(); begin('rotate', event); }} />
    </div> : null}
  </div>;
}
