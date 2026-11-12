import { Button, InputNumber, Slider } from 'antd';
import { useEffect, useId, useRef, useState } from 'react';
import { COLOR_CURVE_CHANNELS, COLOR_WHEEL_RANGES, normalizeRGBColorCurves, normalizeColorWheels } from './colorGrade.js';
import './color-grade-editor.css';

const channelLabels = { red: '红', green: '绿', blue: '蓝' };
const wheelLabels = { shadows: '阴影', midtones: '中间调', highlights: '高光' };
const clamp = (value, min, max) => Math.min(max, Math.max(min, value));
const normalize = value => ({ ...value, rgbCurves: normalizeRGBColorCurves(value?.rgbCurves), wheels: normalizeColorWheels(value?.wheels) });

/** value/onChange contain the whole color object, preserving global curve and HSL. */
export default function ColorGradeEditor({ value, onChange, onChangeComplete }) {
  const [draft, setDraft] = useState(() => normalize(value));
  const latest = useRef(draft);
  const signature = JSON.stringify(normalize(value));
  const [channel, setChannel] = useState('red');
  const [selectedPoint, setSelectedPoint] = useState(0);
  const [range, setRange] = useState('shadows');
  const draggingPoint = useRef(null);
  const draggingWheel = useRef(false);
  const curveSVG = useRef(null);
  const gradientId = useId().replace(/:/g, '');
  useEffect(() => { const next = JSON.parse(signature); latest.current = next; setDraft(next); }, [signature]);
  const update = next => { latest.current = next; setDraft(next); onChange?.(next); };
  const complete = () => onChangeComplete?.(latest.current);
  const replaceCurve = points => update({ ...latest.current, rgbCurves: { ...latest.current.rgbCurves, [channel]: points } });
  const movePoint = (index, x, y) => {
    const points = latest.current.rgbCurves[channel];
    const nextX = index === 0 ? 0 : index === points.length - 1 ? 1 : clamp(x, points[index - 1].x + .0001, points[index + 1].x - .0001);
    replaceCurve(points.map((point, i) => i === index ? { x: nextX, y: clamp(y, 0, 1) } : point));
  };
  const curvePoint = event => { const rect = curveSVG.current.getBoundingClientRect(); return { x: clamp((event.clientX - rect.left) / rect.width, 0, 1), y: clamp(1 - (event.clientY - rect.top) / rect.height, 0, 1) }; };
  const addPoint = event => {
    const points = latest.current.rgbCurves[channel];
    if (points.length >= 16) return;
    const point = curvePoint(event);
    if (point.x <= 0 || point.x >= 1 || points.some(existing => Math.abs(existing.x - point.x) < .0002)) return;
    const next = [...points, point].sort((a, b) => a.x - b.x);
    setSelectedPoint(next.indexOf(point)); replaceCurve(next); complete();
  };
  const setWheel = patch => update({ ...latest.current, wheels: { ...latest.current.wheels, [range]: { ...latest.current.wheels[range], ...patch } } });
  const wheelPoint = event => {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = (event.clientX - rect.left) / rect.width * 2 - 1, y = (event.clientY - rect.top) / rect.height * 2 - 1;
    setWheel({ hue: Math.round(Math.atan2(y, x) * 180 / Math.PI * 10) / 10, saturation: Math.round(clamp(Math.hypot(x, y), 0, 1) * 1000) / 10 });
  };
  const points = draft.rgbCurves[channel];
  const index = Math.min(selectedPoint, points.length - 1), selected = points[index];
  const wheel = draft.wheels[range], angle = wheel.hue / 180 * Math.PI;
  return (
    <section className="aicut-color-grade-editor" aria-label="RGB 曲线与三向色轮">
      <div className="aicut-grade-header"><strong>RGB 分通道曲线</strong><Button size="small" type="text" onClick={() => { replaceCurve([{ x: 0, y: 0 }, { x: 1, y: 1 }]); setSelectedPoint(0); complete(); }}>重置{channelLabels[channel]}通道</Button></div>
      <div className="aicut-grade-tabs" role="group" aria-label="曲线通道">
        {COLOR_CURVE_CHANNELS.map(key => <button type="button" key={key} aria-pressed={key === channel} onClick={() => { setChannel(key); setSelectedPoint(0); }}>{channelLabels[key]}通道</button>)}
      </div>
      <svg ref={curveSVG} className={`aicut-rgb-curve is-${channel}`} viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label={`${channelLabels[channel]}通道曲线，双击新增控制点`} onDoubleClick={addPoint} onPointerMove={event => { if (draggingPoint.current !== null) { const point = curvePoint(event); movePoint(draggingPoint.current, point.x, point.y); } }} onPointerUp={() => { if (draggingPoint.current !== null) { draggingPoint.current = null; complete(); } }} onPointerCancel={() => { draggingPoint.current = null; complete(); }}>
        <path className="aicut-curve-grid" d="M25 0V100M50 0V100M75 0V100M0 25H100M0 50H100M0 75H100" />
        <path className="aicut-curve-identity" d="M0 100L100 0" />
        <polyline points={points.map(point => `${point.x * 100},${(1 - point.y) * 100}`).join(' ')} />
        {points.map((point, i) => <circle key={i} cx={point.x * 100} cy={(1 - point.y) * 100} r={i === index ? 2.8 : 2} className={i === index ? 'is-selected' : ''} onPointerDown={event => { event.preventDefault(); event.stopPropagation(); event.currentTarget.setPointerCapture(event.pointerId); setSelectedPoint(i); draggingPoint.current = i; }} />)}
      </svg>
      <div className="aicut-grade-coordinates">
        <label>输入 <InputNumber size="small" aria-label="曲线控制点输入" min={index === 0 ? 0 : points[index - 1].x + .0001} max={index === points.length - 1 ? 1 : points[index + 1].x - .0001} step={.01} precision={4} disabled={index === 0 || index === points.length - 1} value={selected.x} onChange={amount => { if (typeof amount === 'number') movePoint(index, amount, selected.y); }} onBlur={complete} onPressEnter={complete} /></label>
        <label>输出 <InputNumber size="small" aria-label="曲线控制点输出" min={0} max={1} step={.01} precision={4} value={selected.y} onChange={amount => { if (typeof amount === 'number') movePoint(index, selected.x, amount); }} onBlur={complete} onPressEnter={complete} /></label>
        <Button size="small" type="text" disabled={index === 0 || index === points.length - 1} onClick={() => { replaceCurve(points.filter((_, i) => i !== index)); setSelectedPoint(0); complete(); }}>删除点</Button>
      </div>
      <small>双击曲线新增控制点，拖动调整；最多 16 点。</small>
      <div className="aicut-grade-header"><strong>三向色轮</strong><Button size="small" type="text" onClick={() => { setWheel({ hue: 0, saturation: 0, luminance: 0 }); complete(); }}>重置{wheelLabels[range]}</Button></div>
      <div className="aicut-grade-tabs" role="group" aria-label="色轮明暗范围">{COLOR_WHEEL_RANGES.map(key => <button type="button" key={key} aria-pressed={range === key} onClick={() => setRange(key)}>{wheelLabels[key]}</button>)}</div>
      <svg className="aicut-color-wheel" viewBox="-100 -100 200 200" role="img" aria-label={`${wheelLabels[range]}色轮，拖动选择颜色与强度`} onPointerDown={event => { event.currentTarget.setPointerCapture(event.pointerId); draggingWheel.current = true; wheelPoint(event); }} onPointerMove={event => { if (draggingWheel.current) wheelPoint(event); }} onPointerUp={event => { if (draggingWheel.current) { wheelPoint(event); draggingWheel.current = false; complete(); } }} onPointerCancel={() => { draggingWheel.current = false; complete(); }}>
        <defs><radialGradient id={gradientId}><stop offset="0" stopColor="white" /><stop offset="1" stopColor="white" stopOpacity="0" /></radialGradient></defs>
        <circle cx="0" cy="0" r="100" fill={`url(#${gradientId})`} />
        <path d="M-100 0H100M0-100V100" className="aicut-wheel-axis" />
        <circle cx={Math.cos(angle) * wheel.saturation} cy={Math.sin(angle) * wheel.saturation} r="5" className="aicut-wheel-marker" />
      </svg>
      {[['hue', '色相', -180, 180], ['saturation', '强度', 0, 100], ['luminance', '明度', -100, 100]].map(([key, label, min, max]) => <div className="aicut-wheel-field" key={key}><label>{label}</label><Slider ariaLabelForHandle={`${wheelLabels[range]}${label}`} min={min} max={max} step={.1} value={wheel[key]} onChange={amount => setWheel({ [key]: amount })} onChangeComplete={complete} /><InputNumber size="small" aria-label={`${wheelLabels[range]}${label}数值`} min={min} max={max} step={.1} precision={1} value={wheel[key]} onChange={amount => { if (typeof amount === 'number') setWheel({ [key]: amount }); }} onBlur={complete} onPressEnter={complete} /></div>)}
      <Button size="small" type="text" onClick={() => { update({ ...latest.current, rgbCurves: normalizeRGBColorCurves(), wheels: normalizeColorWheels() }); setSelectedPoint(0); complete(); }}>重置 RGB 曲线与全部色轮</Button>
    </section>
  );
}
