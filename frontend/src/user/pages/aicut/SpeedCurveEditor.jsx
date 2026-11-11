import { Button, InputNumber } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { insertSpeedCurvePoint, moveSpeedCurvePoint, normalizeSpeedCurve, removeSpeedCurvePoint } from '../aicutSpeedCurve.js';

function sourceCurve(points, rate) {
  return normalizeSpeedCurve(Array.isArray(points) && points.length >= 2 ? points : [{ position: 0, rate }, { position: 1, rate }]);
}

export default function SpeedCurveEditor({ points, rate = 1, onChange }) {
  const signature = JSON.stringify(points || []);
  const [draft, setDraft] = useState(() => sourceCurve(points, rate));
  const [selectedIndex, setSelectedIndex] = useState(0);
  const svgRef = useRef(null);
  const draggingRef = useRef(null);

  useEffect(() => {
    const next = sourceCurve(points, rate);
    setDraft(next);
    setSelectedIndex(current => Math.min(current, next.length - 1));
  }, [signature, rate]);

  const eventPoint = event => {
    const rect = svgRef.current.getBoundingClientRect();
    return {
      position: Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))),
      rate: .25 + (1 - Math.min(1, Math.max(0, (event.clientY - rect.top) / Math.max(1, rect.height)))) * 3.75
    };
  };
  const moveSelected = (position, nextRate, commit = false, pointIndex = selectedIndex) => {
    const next = moveSpeedCurvePoint(draft, pointIndex, position, nextRate);
    setDraft(next);
    if (commit) onChange(next);
  };
  const handlePointerMove = event => {
    if (draggingRef.current === null) return;
    const next = eventPoint(event);
    moveSelected(next.position, next.rate, false, draggingRef.current);
  };
  const finishDrag = event => {
    if (draggingRef.current === null) return;
    const pointIndex = draggingRef.current;
    draggingRef.current = null;
    const point = eventPoint(event);
    const next = moveSpeedCurvePoint(draft, pointIndex, point.position, point.rate);
    setDraft(next);
    onChange(next);
  };
  const addPoint = () => {
    const next = insertSpeedCurvePoint(draft);
    const oldPositions = new Set(draft.map(point => point.position));
    setDraft(next);
    setSelectedIndex(Math.max(0, next.findIndex(point => !oldPositions.has(point.position))));
    onChange(next);
  };
  const removePoint = () => {
    const next = removeSpeedCurvePoint(draft, selectedIndex);
    setDraft(next);
    setSelectedIndex(Math.min(Math.max(0, selectedIndex - 1), next.length - 1));
    onChange(next);
  };
  const coordinates = draft.map(point => ({ x: point.position * 100, y: 100 - ((point.rate - .25) / 3.75 * 100) }));
  const selected = draft[selectedIndex];

  return (
    <div className="aicut-speed-curve-editor">
      <div className="aicut-speed-curve-chart is-editable" aria-label="自由变速曲线编辑器">
        <svg ref={svgRef} viewBox="0 0 100 100" preserveAspectRatio="none" onPointerMove={handlePointerMove} onPointerUp={finishDrag} onPointerCancel={finishDrag}>
          <path d={`M ${coordinates.map(point => `${point.x} ${point.y}`).join(' L ')}`} />
          {coordinates.map((point, index) => (
            <circle className={index === selectedIndex ? 'is-selected' : ''} key={`${index}:${draft[index].position}`} cx={point.x} cy={point.y} r={index === selectedIndex ? 4 : 3} onPointerDown={event => {
              event.preventDefault();
              event.currentTarget.setPointerCapture?.(event.pointerId);
              setSelectedIndex(index);
              draggingRef.current = index;
            }} />
          ))}
        </svg>
        <span>0.25x</span><span>4x</span>
      </div>
      <div className="aicut-speed-curve-actions">
        <Button size="small" disabled={draft.length >= 12} onClick={addPoint}>新增控制点</Button>
        <Button size="small" danger disabled={draft.length <= 2 || selectedIndex === 0 || selectedIndex === draft.length - 1} onClick={removePoint}>删除控制点</Button>
      </div>
      <div className="aicut-speed-curve-fields">
        <label>位置（%）<InputNumber size="small" min={selectedIndex === 0 ? 0 : draft[selectedIndex - 1].position * 100 + .1} max={selectedIndex === draft.length - 1 ? 100 : draft[selectedIndex + 1].position * 100 - .1} step={.1} precision={1} disabled={selectedIndex === 0 || selectedIndex === draft.length - 1} value={selected.position * 100} onChange={value => moveSelected(Number(value) / 100, selected.rate, true)} /></label>
        <label>倍率（x）<InputNumber size="small" min={.25} max={4} step={.05} precision={2} value={selected.rate} onChange={value => moveSelected(selected.position, Number(value), true)} /></label>
      </div>
      <small>拖动控制点自由调整速度；端点只能上下调倍率，中间点可同时调整时间位置。</small>
    </div>
  );
}
