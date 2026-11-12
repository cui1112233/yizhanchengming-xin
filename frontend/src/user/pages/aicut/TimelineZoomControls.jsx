import { ZoomIn, ZoomOut, Scan } from 'lucide-react';
export default function TimelineZoomControls({ zoom, onChange, durationMs, viewportWidth }) {
  const change = value => onChange(Math.max(.5, Math.min(4, Math.round(value * 10) / 10)));
  return <div className="aicut-timeline-zoom-controls" role="group" aria-label="时间线缩放设置">
    <button type="button" aria-label="缩小时间线" title="缩小时间线" disabled={zoom <= .5} onClick={() => change(zoom - .1)}><ZoomOut size={14} /></button>
    <input type="range" aria-label="时间线缩放" min="0.5" max="4" step="0.1" value={zoom} onChange={event => change(Number(event.target.value))} />
    <button type="button" aria-label="放大时间线" title="放大时间线" disabled={zoom >= 4} onClick={() => change(zoom + .1)}><ZoomIn size={14} /></button>
    <span>{Math.round(zoom * 100)}%</span>
    <button type="button" aria-label="适应时间线" title="将成片时长适应时间线窗口（50%～400%）" onClick={() => change(Math.max(1, viewportWidth - 24) / (Math.max(1000, durationMs + 1000) / 1000 * 72))}><Scan size={14} /></button>
  </div>;
}
