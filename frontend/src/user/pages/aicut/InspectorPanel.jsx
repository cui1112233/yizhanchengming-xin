import { Button, Checkbox, Input, InputNumber, Select, Slider } from 'antd';
import { PanelRightClose, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { catalogById } from './editorCatalog.js';
import { normalizeTransform } from './canvasTransform.js';
import { speedCurvePresets } from '../aicutSpeedCurve.js';
import { audioEnvelopePoints } from '../aicutVolumeEnvelope.js';
import { subtitleAnimationCatalog } from '../aicutEffects.js';
import { aicutFontOptions, normalizeAICutTextStyle } from '../aicutText.js';
import { audioEffectOptions, normalizeAudioProcessing } from '../aicutAudioProcessing.js';
import { defaultBezierMaskPoints, defaultPolygonMaskPoints, normalizeBezierMaskPoints, normalizeColorCurvePoints, normalizePolygonMaskPoints, normalizeVisualProcessing, polygonArea, sampleColorCurve } from '../aicutVisualProcessing.js';
import { keyframeEasingPresets, keyframeEasingSamples, normalizeKeyframeEasing } from '../aicutKeyframeEasing.js';
import SpeedCurveEditor from './SpeedCurveEditor.jsx';
import { isPersonSegmentationCurrent, normalizePersonSegmentation } from '../aicutPersonSegmentation.js';
import { normalizeTrackingRegion } from '../aicutTrackingRegion.js';

function kindTitle(clip) {
  if (!clip) return '未选择对象';
  if (clip.kind === 'audio') return '音频';
  if (clip.kind === 'subtitle') return '文字与字幕';
  if (clip.kind === 'effect') return '特效';
  if (clip.transition) return '转场';
  return '画面';
}

function KeyframeEasingEditor({ keyframes, onSet, onBatchEditKeyframes, onCopyKeyframes, onPasteKeyframes, keyframeClipboardCount = 0 }) {
  const frames = [...(keyframes || [])].sort((left, right) => Number(left.offsetMs) - Number(right.offsetMs));
  const [selectedId, setSelectedId] = useState(frames[0]?.id || '');
  const [batchMode, setBatchMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState([]);
  const [batchEasing, setBatchEasing] = useState('linear');
  const [shiftMs, setShiftMs] = useState(0);
  const svgRef = useRef(null);
  const dragRef = useRef(null);
  const selected = frames.find(frame => frame.id === selectedId) || frames[0];
  const sourceEasing = normalizeKeyframeEasing(selected?.easing);
  const [draftEasing, setDraftEasing] = useState(sourceEasing);
  useEffect(() => {
    if (!frames.some(frame => frame.id === selectedId)) setSelectedId(frames[0]?.id || '');
  }, [frames.map(frame => frame.id).join('|'), selectedId]);
  useEffect(() => {
    const available = new Set(frames.map(frame => frame.id));
    setSelectedIds(current => current.filter(id => available.has(id)));
  }, [frames.map(frame => frame.id).join('|')]);
  useEffect(() => { setDraftEasing(sourceEasing); }, [selected?.id, JSON.stringify(sourceEasing)]);
  if (!frames.length) return <div className="aicut-keyframe-empty"><small>播放头移到片段内，点击时间线工具栏“＋关键帧”。</small><Button size="small" disabled={!keyframeClipboardCount} onClick={onPasteKeyframes}>粘贴到播放头</Button><small>关键帧剪贴板：{keyframeClipboardCount} 个</small></div>;
  const easing = draftEasing;
  const samples = keyframeEasingSamples(easing);
  const selectedIndex = frames.findIndex(frame => frame.id === selected.id);
  const isEndpoint = selectedIndex === frames.length - 1;
  const publish = (next, commit = false) => {
    const normalized = normalizeKeyframeEasing(next);
    setDraftEasing(normalized);
    if (commit && !isEndpoint) onSet(selected.id, normalized);
  };
  const setPreset = type => publish(type === 'custom' ? { type: 'custom' } : { type }, true);
  const eventPoint = event => {
    const rect = svgRef.current.getBoundingClientRect();
    return {
      x: Math.round(Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))) * 100) / 100,
      y: Math.round(Math.min(1, Math.max(0, 1 - (event.clientY - rect.top) / Math.max(1, rect.height))) * 100) / 100
    };
  };
  const moveHandle = (event, done = false) => {
    if (!dragRef.current || easing.type !== 'custom') return;
    const point = eventPoint(event);
    publish({ ...easing, [`x${dragRef.current}`]: point.x, [`y${dragRef.current}`]: point.y }, done);
    if (done) dragRef.current = null;
  };
  const setCoordinate = (handle, axis, value, commit = false) => publish({ ...easing, [`${axis}${handle}`]: Math.min(1, Math.max(0, Number(value || 0) / 100)) }, commit);
  const toggleFrame = (frameId) => {
    if (!batchMode) return setSelectedId(frameId);
    setSelectedIds(current => current.includes(frameId) ? current.filter(id => id !== frameId) : [...current, frameId]);
  };
  const runBatch = (action, payload = {}) => {
    onBatchEditKeyframes(selectedIds, action, payload);
    if (action === 'remove') setSelectedIds([]);
  };
  return <div className="aicut-keyframe-easing">
    <div className="aicut-keyframe-batch-heading"><Button size="small" type={batchMode ? 'primary' : 'default'} onClick={() => { setBatchMode(value => !value); setSelectedIds([]); }}>批量选择</Button><Button size="small" disabled={!keyframeClipboardCount} onClick={onPasteKeyframes}>粘贴到播放头</Button>{batchMode ? <><Button size="small" onClick={() => setSelectedIds(frames.map(frame => frame.id))}>全选</Button><Button size="small" onClick={() => setSelectedIds([])}>取消选择</Button><small>已选 {selectedIds.length} 个</small></> : <small>关键帧剪贴板：{keyframeClipboardCount} 个</small>}</div>
    <div className="aicut-keyframe-list">{frames.map((frame, index) => <button type="button" key={frame.id} aria-pressed={batchMode ? selectedIds.includes(frame.id) : frame.id === selected.id} className={(batchMode ? selectedIds.includes(frame.id) : frame.id === selected.id) ? 'selected' : ''} onClick={() => toggleFrame(frame.id)}>{batchMode ? (selectedIds.includes(frame.id) ? '☑' : '☐') : '◆'} {(Number(frame.offsetMs) / 1000).toFixed(2)}s{index === frames.length - 1 ? ' 终点' : ''}</button>)}</div>
    {batchMode ? <div className="aicut-keyframe-batch-tools">
      <label>统一缓动<Select aria-label="统一缓动" value={batchEasing} options={keyframeEasingPresets.map(item => ({ value: item.id, label: item.label }))} onChange={setBatchEasing} /></label>
      <Button size="small" disabled={!selectedIds.length} onClick={() => runBatch('easing', { easing: batchEasing === 'custom' ? { type: 'custom' } : { type: batchEasing } })}>应用统一缓动</Button>
      <label>时间偏移<InputNumber aria-label="时间偏移" value={shiftMs} step={10} onChange={value => setShiftMs(Number(value || 0))} addonAfter="ms" /></label>
      <Button size="small" disabled={!selectedIds.length || !shiftMs} onClick={() => runBatch('shift', { deltaMs: shiftMs })}>应用偏移</Button>
      <Button size="small" disabled={!selectedIds.length} onClick={() => onCopyKeyframes(selectedIds)}>复制所选</Button>
      <Button danger size="small" disabled={!selectedIds.length} onClick={() => runBatch('remove')}>删除所选</Button>
      <small>批量操作只生成一条撤销记录；偏移越界或与其他关键帧重叠时不会修改。</small>
    </div> : null}
    <label>关键帧缓动<Select aria-label="关键帧缓动" disabled={isEndpoint} value={easing.type} options={keyframeEasingPresets.map(item => ({ value: item.id, label: item.label }))} onChange={setPreset} /></label>
    <small>线性、缓入、缓出、缓入缓出、定格或自定义</small>
    <svg ref={svgRef} className="aicut-keyframe-curve" viewBox="0 0 100 100" preserveAspectRatio="none" aria-label="关键帧缓动曲线" onPointerMove={moveHandle} onPointerUp={event => moveHandle(event, true)} onPointerCancel={() => { dragRef.current = null; }}>
      <path className="grid" d="M0 25H100 M0 50H100 M0 75H100 M25 0V100 M50 0V100 M75 0V100" />
      <polyline points={samples.map(point => `${point.x * 100},${100 - point.y * 100}`).join(' ')} />
      {easing.type === 'custom' ? <>
        <path className="handles" d={`M0 100L${easing.x1 * 100} ${100 - easing.y1 * 100} M100 0L${easing.x2 * 100} ${100 - easing.y2 * 100}`} />
        <circle aria-label="控制柄1" cx={easing.x1 * 100} cy={100 - easing.y1 * 100} r="4" onPointerDown={event => { event.preventDefault(); dragRef.current = 1; event.currentTarget.setPointerCapture?.(event.pointerId); }} />
        <circle aria-label="控制柄2" cx={easing.x2 * 100} cy={100 - easing.y2 * 100} r="4" onPointerDown={event => { event.preventDefault(); dragRef.current = 2; event.currentTarget.setPointerCapture?.(event.pointerId); }} />
      </> : null}
    </svg>
    {easing.type === 'custom' ? <div className="aicut-keyframe-handle-values">{[1, 2].flatMap(handle => ['x', 'y'].map(axis => <label key={`${handle}-${axis}`}>控制柄{handle}{axis.toUpperCase()}<InputNumber min={0} max={100} disabled={isEndpoint} value={Math.round(easing[`${axis}${handle}`] * 100)} onChange={value => setCoordinate(handle, axis, value)} onBlur={() => setCoordinate(handle, axis, easing[`${axis}${handle}`] * 100, true)} onPressEnter={() => setCoordinate(handle, axis, easing[`${axis}${handle}`] * 100, true)} addonAfter="%" /></label>))}</div> : null}
    <small>曲线作用于当前关键帧到下一个关键帧的插值；最后一个关键帧只作为终点。</small>
  </div>;
}

function ColorCurveEditor({ curvePoints, onChange, onCommit }) {
  const normalized = normalizeColorCurvePoints({ curvePoints });
  const [points, setPoints] = useState(normalized);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const svgRef = useRef(null);
  const dragRef = useRef(null);
  const curveKey = JSON.stringify(normalized);
  useEffect(() => { setPoints(normalized); setSelectedIndex(index => Math.min(index, normalized.length - 1)); }, [curveKey]);
  const eventPoint = (event) => {
    const rect = svgRef.current.getBoundingClientRect();
    return {
      x: Math.round(Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))) * 100) / 100,
      y: Math.round(Math.min(1, Math.max(0, 1 - (event.clientY - rect.top) / Math.max(1, rect.height))) * 100) / 100
    };
  };
  const publish = (next, commit = false) => {
    setPoints(next);
    onChange(next);
    if (commit) onCommit(next);
  };
  const updatePoint = (event, commit = false) => {
    const drag = dragRef.current;
    if (!drag) return;
    const requested = eventPoint(event);
    const minimumX = drag.index === 0 ? 0 : points[drag.index - 1].x + .01;
    const maximumX = drag.index === points.length - 1 ? 1 : points[drag.index + 1].x - .01;
    const point = { x: drag.lockedX ?? Math.min(maximumX, Math.max(minimumX, requested.x)), y: requested.y };
    const next = points.map((item, index) => index === drag.index ? point : item);
    publish(next, commit);
    if (commit) dragRef.current = null;
  };
  const addPoint = (requested = null) => {
    if (points.length >= 16) return;
    let point = requested;
    if (!point) {
      let gapIndex = 0;
      for (let index = 1; index < points.length - 1; index += 1) if (points[index + 1].x - points[index].x > points[gapIndex + 1].x - points[gapIndex].x) gapIndex = index;
      const x = Math.round(((points[gapIndex].x + points[gapIndex + 1].x) / 2) * 100) / 100;
      point = { x, y: Math.round(sampleColorCurve(points, x) * 100) / 100 };
    }
    if (point.x <= 0 || point.x >= 1 || points.some(item => Math.abs(item.x - point.x) < .01)) return;
    const next = [...points, point].sort((left, right) => left.x - right.x);
    setSelectedIndex(next.findIndex(item => item === point));
    publish(next, true);
  };
  const deleteSelected = () => {
    if (selectedIndex <= 0 || selectedIndex >= points.length - 1 || points.length <= 2) return;
    const next = points.filter((_, index) => index !== selectedIndex);
    setSelectedIndex(Math.min(selectedIndex, next.length - 1));
    publish(next, true);
  };
  const reset = () => {
    const next = normalizeColorCurvePoints();
    setSelectedIndex(0);
    publish(next, true);
  };
  const setSelectedCoordinate = (axis, percent, commit = false) => {
    if (!points[selectedIndex]) return;
    const value = Math.min(1, Math.max(0, Number(percent || 0) / 100));
    const previous = points[selectedIndex - 1];
    const following = points[selectedIndex + 1];
    const nextPoint = { ...points[selectedIndex], [axis]: axis === 'x' ? (selectedIndex === 0 ? 0 : selectedIndex === points.length - 1 ? 1 : Math.min(following.x - .01, Math.max(previous.x + .01, value))) : value };
    publish(points.map((point, index) => index === selectedIndex ? nextPoint : point), commit);
  };
  const selected = points[selectedIndex];
  return <div className="aicut-color-curve">
    <div className="aicut-color-curve-heading"><strong>自由调色曲线</strong><div><Button size="small" disabled={points.length >= 16} onClick={() => addPoint()}>新增控制点</Button><Button size="small" danger disabled={selectedIndex <= 0 || selectedIndex >= points.length - 1} onClick={deleteSelected}>删除选中点</Button><Button size="small" onClick={reset}>恢复直线</Button></div></div>
    <svg ref={svgRef} viewBox="0 0 100 100" preserveAspectRatio="none" aria-label="自由调色曲线" onDoubleClick={(event) => addPoint(eventPoint(event))} onPointerMove={updatePoint} onPointerUp={(event) => updatePoint(event, true)} onPointerCancel={() => { dragRef.current = null; }}>
      <path className="aicut-color-curve-grid" d="M0 25H100 M0 50H100 M0 75H100 M25 0V100 M50 0V100 M75 0V100" />
      <polyline points={points.map(point => `${point.x * 100},${100 - point.y * 100}`).join(' ')} />
      {points.map((point, index) => <circle key={`${point.x}-${index}`} className={selectedIndex === index ? 'selected' : ''} cx={point.x * 100} cy={100 - point.y * 100} r="3.5" aria-label={`曲线控制点${index + 1}`} onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); setSelectedIndex(index); dragRef.current = { index, lockedX: index === 0 || index === points.length - 1 ? point.x : null }; event.currentTarget.setPointerCapture?.(event.pointerId); }} />)}
    </svg>
    <div className="aicut-color-curve-scale"><span>暗部</span><small>双击曲线区域新增</small><span>高光</span></div>
    {selected ? <div className="aicut-color-curve-values"><label>输入<InputNumber aria-label="曲线点横坐标" min={0} max={100} disabled={selectedIndex === 0 || selectedIndex === points.length - 1} value={Math.round(selected.x * 100)} onChange={(value) => setSelectedCoordinate('x', value)} onBlur={(event) => setSelectedCoordinate('x', event.target.value, true)} onPressEnter={(event) => setSelectedCoordinate('x', event.target.value, true)} addonAfter="%" /></label><label>输出<InputNumber aria-label="曲线点纵坐标" min={0} max={100} value={Math.round(selected.y * 100)} onChange={(value) => setSelectedCoordinate('y', value)} onBlur={(event) => setSelectedCoordinate('y', event.target.value, true)} onPressEnter={(event) => setSelectedCoordinate('y', event.target.value, true)} addonAfter="%" /></label></div> : null}
  </div>;
}

function PolygonMaskEditor({ maskPoints, onChange, onCommit }) {
  const normalized = normalizePolygonMaskPoints(maskPoints);
  const [points, setPoints] = useState(normalized);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const svgRef = useRef(null);
  const dragRef = useRef(null);
  const pointsKey = JSON.stringify(normalized);
  useEffect(() => { setPoints(normalized); setSelectedIndex(index => Math.min(index, normalized.length - 1)); }, [pointsKey]);
  const eventPoint = (event) => {
    const rect = svgRef.current.getBoundingClientRect();
    return {
      x: Math.round(Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))) * 100) / 100,
      y: Math.round(Math.min(1, Math.max(0, (event.clientY - rect.top) / Math.max(1, rect.height))) * 100) / 100
    };
  };
  const publish = (next, commit = false) => {
    if (next.length < 3 || next.length > 12 || polygonArea(next) < .0001) return;
    setPoints(next);
    onChange(next);
    if (commit) onCommit(next);
  };
  const movePoint = (event, commit = false) => {
    if (dragRef.current === null) return;
    const next = points.map((point, index) => index === dragRef.current ? eventPoint(event) : point);
    publish(next, commit);
    if (commit) dragRef.current = null;
  };
  const nearestEdge = (requested) => points.reduce((best, point, index) => {
    const next = points[(index + 1) % points.length];
    const middle = { x: (point.x + next.x) / 2, y: (point.y + next.y) / 2 };
    const score = requested ? (middle.x - requested.x) ** 2 + (middle.y - requested.y) ** 2 : -((point.x - next.x) ** 2 + (point.y - next.y) ** 2);
    return score < best.score ? { index, score } : best;
  }, { index: 0, score: Number.POSITIVE_INFINITY }).index;
  const addPoint = (requested = null) => {
    if (points.length >= 12) return;
    const edgeIndex = nearestEdge(requested);
    const following = points[(edgeIndex + 1) % points.length];
    const point = requested || { x: Math.round((points[edgeIndex].x + following.x) * 50) / 100, y: Math.round((points[edgeIndex].y + following.y) * 50) / 100 };
    const next = [...points.slice(0, edgeIndex + 1), point, ...points.slice(edgeIndex + 1)];
    setSelectedIndex(edgeIndex + 1);
    publish(next, true);
  };
  const deleteSelected = () => {
    if (points.length <= 3) return;
    const next = points.filter((_, index) => index !== selectedIndex);
    setSelectedIndex(Math.min(selectedIndex, next.length - 1));
    publish(next, true);
  };
  const reset = () => { const next = defaultPolygonMaskPoints.map(point => ({ ...point })); setSelectedIndex(0); publish(next, true); };
  const setSelectedCoordinate = (axis, percent, commit = false) => {
    const value = Math.min(1, Math.max(0, Number(percent || 0) / 100));
    const next = points.map((point, index) => index === selectedIndex ? { ...point, [axis]: value } : point);
    publish(next, commit);
  };
  const selected = points[selectedIndex];
  return <div className="aicut-polygon-mask">
    <div className="aicut-polygon-mask-heading"><strong>手绘多边形</strong><div><Button size="small" disabled={points.length >= 12} onClick={() => addPoint()}>新增蒙版点</Button><Button size="small" danger disabled={points.length <= 3} onClick={deleteSelected}>删除蒙版点</Button><Button size="small" onClick={reset}>重置多边形</Button></div></div>
    <svg ref={svgRef} viewBox="0 0 100 100" preserveAspectRatio="none" aria-label="手绘多边形蒙版" onDoubleClick={(event) => addPoint(eventPoint(event))} onPointerMove={movePoint} onPointerUp={(event) => movePoint(event, true)} onPointerCancel={() => { dragRef.current = null; }}>
      <path className="aicut-polygon-mask-grid" d="M0 25H100 M0 50H100 M0 75H100 M25 0V100 M50 0V100 M75 0V100" />
      <polygon points={points.map(point => `${point.x * 100},${point.y * 100}`).join(' ')} />
      {points.map((point, index) => <circle key={`${point.x}-${point.y}-${index}`} className={selectedIndex === index ? 'selected' : ''} cx={point.x * 100} cy={point.y * 100} r="3.5" aria-label={`蒙版控制点${index + 1}`} onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); setSelectedIndex(index); dragRef.current = index; event.currentTarget.setPointerCapture?.(event.pointerId); }} />)}
    </svg>
    <div className="aicut-polygon-mask-note"><small>双击画布新增，拖动控制点改变轮廓</small><span>{points.length}/12 点</span></div>
    {selected ? <div className="aicut-polygon-mask-values"><label>X<InputNumber min={0} max={100} value={Math.round(selected.x * 100)} onChange={(value) => setSelectedCoordinate('x', value)} onBlur={(event) => setSelectedCoordinate('x', event.target.value, true)} onPressEnter={(event) => setSelectedCoordinate('x', event.target.value, true)} addonAfter="%" /></label><label>Y<InputNumber min={0} max={100} value={Math.round(selected.y * 100)} onChange={(value) => setSelectedCoordinate('y', value)} onBlur={(event) => setSelectedCoordinate('y', event.target.value, true)} onPressEnter={(event) => setSelectedCoordinate('y', event.target.value, true)} addonAfter="%" /></label></div> : null}
  </div>;
}

function BezierMaskEditor({ bezierPoints, onChange, onCommit }) {
  const normalized = normalizeBezierMaskPoints(bezierPoints);
  const [points, setPoints] = useState(normalized);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const svgRef = useRef(null);
  const dragRef = useRef(null);
  const pointsKey = JSON.stringify(normalized);
  useEffect(() => { setPoints(normalized); setSelectedIndex(index => Math.min(index, normalized.length - 1)); }, [pointsKey]);
  const eventPoint = (event) => {
    const rect = svgRef.current.getBoundingClientRect();
    return {
      x: Math.round(Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))) * 100) / 100,
      y: Math.round(Math.min(1, Math.max(0, (event.clientY - rect.top) / Math.max(1, rect.height))) * 100) / 100
    };
  };
  const publish = (next, commit = false) => { setPoints(next); onChange(next); if (commit) onCommit(next); };
  const movePoint = (event, commit = false) => {
    if (!dragRef.current) return;
    const requested = eventPoint(event);
    const { index, kind } = dragRef.current;
    const next = points.map((point, pointIndex) => {
      if (pointIndex !== index) return point;
      if (kind === 'anchor') {
        const dx = requested.x - point.x;
        const dy = requested.y - point.y;
        return { ...point, x: requested.x, y: requested.y, inX: Math.min(1, Math.max(0, point.inX + dx)), inY: Math.min(1, Math.max(0, point.inY + dy)), outX: Math.min(1, Math.max(0, point.outX + dx)), outY: Math.min(1, Math.max(0, point.outY + dy)) };
      }
      return { ...point, [`${kind}X`]: requested.x, [`${kind}Y`]: requested.y };
    });
    publish(next, commit);
    if (commit) dragRef.current = null;
  };
  const midpoint = (left, right) => ({ x: (left.x + right.x) / 2, y: (left.y + right.y) / 2 });
  const addAnchor = () => {
    if (points.length >= 8) return;
    const index = selectedIndex;
    const nextIndex = (index + 1) % points.length;
    const current = points[index];
    const following = points[nextIndex];
    const p0 = { x: current.x, y: current.y };
    const p1 = { x: current.outX, y: current.outY };
    const p2 = { x: following.inX, y: following.inY };
    const p3 = { x: following.x, y: following.y };
    const p01 = midpoint(p0, p1), p12 = midpoint(p1, p2), p23 = midpoint(p2, p3);
    const p012 = midpoint(p01, p12), p123 = midpoint(p12, p23), anchor = midpoint(p012, p123);
    let next = points.map((point, pointIndex) => pointIndex === index ? { ...point, outX: p01.x, outY: p01.y } : pointIndex === nextIndex ? { ...point, inX: p23.x, inY: p23.y } : point);
    const inserted = { x: anchor.x, y: anchor.y, inX: p012.x, inY: p012.y, outX: p123.x, outY: p123.y };
    next = index === points.length - 1 ? [...next, inserted] : [...next.slice(0, index + 1), inserted, ...next.slice(index + 1)];
    setSelectedIndex(index + 1);
    publish(next, true);
  };
  const deleteAnchor = () => {
    if (points.length <= 3) return;
    const next = points.filter((_, index) => index !== selectedIndex);
    setSelectedIndex(Math.min(selectedIndex, next.length - 1));
    publish(next, true);
  };
  const reset = () => { const next = defaultBezierMaskPoints.map(point => ({ ...point })); setSelectedIndex(0); publish(next, true); };
  const setCoordinate = (key, percent, commit = false) => {
    const value = Math.min(1, Math.max(0, Number(percent || 0) / 100));
    const next = points.map((point, index) => {
      if (index !== selectedIndex) return point;
      if (key === 'x') { const delta = value - point.x; return { ...point, x: value, inX: Math.min(1, Math.max(0, point.inX + delta)), outX: Math.min(1, Math.max(0, point.outX + delta)) }; }
      if (key === 'y') { const delta = value - point.y; return { ...point, y: value, inY: Math.min(1, Math.max(0, point.inY + delta)), outY: Math.min(1, Math.max(0, point.outY + delta)) }; }
      return { ...point, [key]: value };
    });
    publish(next, commit);
  };
  const first = points[0];
  const path = first ? `M ${first.x * 100} ${first.y * 100} ${points.map((point, index) => { const next = points[(index + 1) % points.length]; return `C ${point.outX * 100} ${point.outY * 100} ${next.inX * 100} ${next.inY * 100} ${next.x * 100} ${next.y * 100}`; }).join(' ')} Z` : '';
  const selected = points[selectedIndex];
  const input = (label, key) => <label>{label}<InputNumber min={0} max={100} value={Math.round(selected[key] * 100)} onChange={(value) => setCoordinate(key, value)} onBlur={(event) => setCoordinate(key, event.target.value, true)} onPressEnter={(event) => setCoordinate(key, event.target.value, true)} addonAfter="%" /></label>;
  return <div className="aicut-bezier-mask">
    <div className="aicut-bezier-mask-heading"><strong>贝塞尔曲线</strong><div><Button size="small" disabled={points.length >= 8} onClick={addAnchor}>新增锚点</Button><Button size="small" danger disabled={points.length <= 3} onClick={deleteAnchor}>删除锚点</Button><Button size="small" onClick={reset}>重置曲线</Button></div></div>
    <svg ref={svgRef} viewBox="0 0 100 100" preserveAspectRatio="none" aria-label="贝塞尔曲线蒙版" onPointerMove={movePoint} onPointerUp={(event) => movePoint(event, true)} onPointerCancel={() => { dragRef.current = null; }}>
      <path className="aicut-bezier-mask-grid" d="M0 25H100 M0 50H100 M0 75H100 M25 0V100 M50 0V100 M75 0V100" />
      <path className="aicut-bezier-mask-shape" d={path} />
      {selected ? <><line x1={selected.x * 100} y1={selected.y * 100} x2={selected.inX * 100} y2={selected.inY * 100} /><line x1={selected.x * 100} y1={selected.y * 100} x2={selected.outX * 100} y2={selected.outY * 100} /><circle className="handle in" aria-label="入控制柄" cx={selected.inX * 100} cy={selected.inY * 100} r="2.5" onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); dragRef.current = { index: selectedIndex, kind: 'in' }; }} /><circle className="handle out" aria-label="出控制柄" cx={selected.outX * 100} cy={selected.outY * 100} r="2.5" onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); dragRef.current = { index: selectedIndex, kind: 'out' }; }} /></> : null}
      {points.map((point, index) => <circle key={`${point.x}-${point.y}-${index}`} className={`anchor ${selectedIndex === index ? 'selected' : ''}`} cx={point.x * 100} cy={point.y * 100} r="3.5" aria-label={`贝塞尔锚点${index + 1}`} onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); setSelectedIndex(index); dragRef.current = { index, kind: 'anchor' }; }} />)}
    </svg>
    <div className="aicut-bezier-mask-note"><small>拖动锚点改变轮廓，拖动入/出控制柄调节弧度</small><span>{points.length}/8 锚点</span></div>
    {selected ? <div className="aicut-bezier-mask-values">{input('锚点X', 'x')}{input('锚点Y', 'y')}{input('入控制柄X', 'inX')}{input('入控制柄Y', 'inY')}{input('出控制柄X', 'outX')}{input('出控制柄Y', 'outY')}</div> : null}
  </div>;
}

function VolumeEnvelopeEditor({ clip, onUpsert, onMove, onRemove }) {
  const [points, setPoints] = useState(() => audioEnvelopePoints(clip));
  const [selectedId, setSelectedId] = useState('');
  const svgRef = useRef(null);
  const dragRef = useRef(null);
  useEffect(() => setPoints(audioEnvelopePoints(clip)), [clip]);
  const positionFromEvent = (event, lockedOffset = null) => {
    const rect = svgRef.current.getBoundingClientRect();
    const position = lockedOffset === null ? Math.min(1, Math.max(0, (event.clientX - rect.left) / Math.max(1, rect.width))) : lockedOffset / Math.max(1, Number(clip.durationMs));
    const volume = Math.min(2, Math.max(0, 2 * (1 - (event.clientY - rect.top) / Math.max(1, rect.height))));
    return { offsetMs: Math.round(position * Number(clip.durationMs)), position, volume: Math.round(volume * 100) / 100 };
  };
  const moveDraft = (event) => {
    if (!dragRef.current) return;
    const current = dragRef.current;
    const next = positionFromEvent(event, current.lockedOffset);
    setPoints(items => items.map((item, index) => index === current.index ? { ...item, ...next } : item).sort((a, b) => a.offsetMs - b.offsetMs));
  };
  const finishDrag = (event) => {
    if (!dragRef.current) return;
    const current = dragRef.current;
    const next = positionFromEvent(event, current.lockedOffset);
    dragRef.current = null;
    if (current.id) onMove(current.id, next.offsetMs, next.volume);
    else onUpsert(next.offsetMs, next.volume);
  };
  const path = points.map(point => `${point.position * 100},${100 - point.volume * 50}`).join(' ');
  return <div className="aicut-volume-envelope">
    <div className="aicut-volume-envelope-heading"><strong>音量包络</strong><small>双击新增，拖动控制点改变时间和音量</small></div>
    <svg ref={svgRef} viewBox="0 0 100 100" preserveAspectRatio="none" onPointerMove={moveDraft} onPointerUp={finishDrag} onDoubleClick={(event) => { const point = positionFromEvent(event); onUpsert(point.offsetMs, point.volume); }}>
      <polyline points={path} />
      {points.map((point, index) => <circle key={point.id || `base-${index}`} cx={point.position * 100} cy={100 - point.volume * 50} r="3" className={selectedId && selectedId === point.id ? 'selected' : ''} onPointerDown={(event) => { event.preventDefault(); event.stopPropagation(); setSelectedId(point.id); dragRef.current = { index, id: point.id, lockedOffset: point.offsetMs === 0 || point.offsetMs === Number(clip.durationMs) ? point.offsetMs : null }; }} />)}
    </svg>
    <div className="aicut-volume-envelope-scale"><span>0%</span><span>100%</span><span>0–200%</span></div>
    <Button danger size="small" disabled={!selectedId} onClick={() => { onRemove(selectedId); setSelectedId(''); }}>删除选中音量点</Button>
  </div>;
}

export default function InspectorPanel({ selection, selectionCount = 0, projectSummary, drawerMode = false, onClose, onCollapse, onPatchClip, onPatchTrack, onPatchEffect, onPatchTransition, onSetPlaybackRate, onSetSpeedCurve, onSetAudioFades, onSetAudioProcessing, onSetVisualProcessing, onBatchApplyVisualProcessing, onDetectChroma, onStartChromaPicker, chromaPickerActive = false, regionTrackingBusy = false, trackingRegion, trackingSelectionActive = false, onTrackingRegionChange, onToggleTrackingSelection, onRunRegionTracking, onCancelRegionTracking, objectRemovalRegion, objectRemovalSelectionActive = false, onObjectRemovalRegionChange, onToggleObjectRemovalSelection, onCommitObjectRemovalRegion, onSetObjectRemovalEnabled, personSegmentationBusy = false, personSegmentationProgress = 0, onRunPersonSegmentation, onCancelPersonSegmentation, onSetReverse, onSetFreezeFrame, onSetBlendMode, onSetKeyframeEasing, onBatchEditKeyframes, onCopyKeyframes, onPasteKeyframes, keyframeClipboardCount, onVolumePointUpsert, onVolumePointMove, onVolumePointRemove }) {
  const clip = selection?.clip;
  const [draft, setDraft] = useState(() => ({ ...(clip || {}), canvasTransform: normalizeTransform(clip?.canvasTransform) }));
  const normalizedTrackingRegion = normalizeTrackingRegion(trackingRegion);
  const normalizedObjectRemovalRegion = normalizeTrackingRegion(objectRemovalRegion);
  useEffect(() => setDraft({ ...(clip || {}), canvasTransform: normalizeTransform(clip?.canvasTransform) }), [clip]);
  const resourceId = clip?.catalogResourceId || clip?.effectResourceId || clip?.transition || '';
  const resource = resourceId ? catalogById(resourceId) : null;
  const textStyle = normalizeAICutTextStyle(selection?.track?.settings);
  const audioProcessing = normalizeAudioProcessing(draft.audioProcessing);
  const visualProcessing = normalizeVisualProcessing(draft.visualProcessing);
  const personSegmentation = normalizePersonSegmentation(visualProcessing.personSegmentation);
  const personSegmentationCurrent = clip?.kind === 'video' && isPersonSegmentationCurrent(clip, personSegmentation);
  const personSegmentationStale = Boolean(personSegmentation.matteAssetId) && !personSegmentationCurrent;
  const commitNumber = (key, value) => onPatchClip({ [key]: Number(value || 0) });
  const patchTransform = (patch, commit = false) => {
    const canvasTransform = normalizeTransform({ ...draft.canvasTransform, ...patch });
    setDraft(current => ({ ...current, canvasTransform }));
    if (commit) onPatchClip({ canvasTransform });
  };
  const patchVisual = (section, patch, commit = false) => {
    const current = normalizeVisualProcessing(draft.visualProcessing);
    const visualProcessing = { ...current, [section]: { ...current[section], ...patch } };
    setDraft(value => ({ ...value, visualProcessing }));
    if (commit) onSetVisualProcessing(visualProcessing);
  };
  const updateAudioFade = (key, value) => {
    const next = {
      fadeInMs: Number(draft.fadeInMs) || 0,
      fadeOutMs: Number(draft.fadeOutMs) || 0,
      [key]: Number(value) || 0
    };
    setDraft(current => ({ ...current, ...next }));
    onSetAudioFades(next);
  };

  return <aside className="aicut-context-inspector" aria-label="属性栏">
    <header><strong>{clip ? kindTitle(clip) : '草稿参数'}</strong>{drawerMode ? <button type="button" aria-label="关闭属性栏" onClick={onClose}><X size={15} /></button> : <button type="button" aria-label="收起属性栏" onClick={onCollapse}><PanelRightClose size={15} /></button>}</header>
    {!clip ? (projectSummary || <div className="aicut-inspector-placeholder"><strong>未选择对象</strong><span>在时间线或播放器中选择片段后，这里会显示可编辑参数。</span></div>) : <div className="aicut-inspector-sections">
      <section><h3>时间</h3><label>开始时间<InputNumber value={draft.startMs} min={0} onChange={(value) => setDraft(current => ({ ...current, startMs: value }))} onBlur={() => commitNumber('startMs', draft.startMs)} onPressEnter={() => commitNumber('startMs', draft.startMs)} addonAfter="ms" /></label><label>片段时长<InputNumber value={draft.durationMs} min={1} onChange={(value) => setDraft(current => ({ ...current, durationMs: value }))} onBlur={() => commitNumber('durationMs', draft.durationMs)} onPressEnter={() => commitNumber('durationMs', draft.durationMs)} addonAfter="ms" /></label></section>
	  {clip.kind === 'video' ? <section className="aicut-playback-tools"><h3>片段播放</h3><Checkbox checked={Boolean(clip.reversed)} disabled={clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null} onChange={(event) => onSetReverse(event.target.checked)}>倒放</Checkbox><Button size="small" onClick={() => onSetFreezeFrame(clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null)}>{clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null ? '取消定格' : '冻结当前帧'}</Button>{clip.freezeSourceMs !== undefined && clip.freezeSourceMs !== null ? <small>定格素材位置：{Number(clip.freezeSourceMs)} ms</small> : null}<label>视频补帧<Select aria-label="视频补帧" value={visualProcessing.frameInterpolation} options={[{ value: 'off', label: '关闭' }, { value: 'blend', label: '帧混合（快速）' }, { value: 'optical-flow', label: '光流补帧（高质量）' }]} onChange={(frameInterpolation) => { const next = { ...visualProcessing, frameInterpolation }; setDraft(current => ({ ...current, visualProcessing: next })); onSetVisualProcessing(next); }} /></label><small>编辑器保持原素材轻量播放；云端成片执行真实补帧，最终帧率跟随导出设置。</small><div className="aicut-stabilization"><strong>视频防抖</strong><Checkbox checked={visualProcessing.stabilization.enabled} onChange={(event) => { const next = { ...visualProcessing, stabilization: { ...visualProcessing.stabilization, enabled: event.target.checked } }; setDraft(current => ({ ...current, visualProcessing: next })); onSetVisualProcessing(next); }}>启用防抖</Checkbox><label>平滑度<InputNumber min={1} max={100} value={visualProcessing.stabilization.smoothness} disabled={!visualProcessing.stabilization.enabled} onChange={(smoothness) => { const next = { ...visualProcessing, stabilization: { ...visualProcessing.stabilization, smoothness: Number(smoothness ?? 15) } }; setDraft(current => ({ ...current, visualProcessing: next })); }} onBlur={() => onSetVisualProcessing(normalizeVisualProcessing(draft.visualProcessing))} addonAfter="帧" /></label><label>边缘缩放<InputNumber min={0} max={30} step={1} value={visualProcessing.stabilization.zoom} disabled={!visualProcessing.stabilization.enabled} onChange={(zoom) => { const next = { ...visualProcessing, stabilization: { ...visualProcessing.stabilization, zoom: Number(zoom ?? 0) } }; setDraft(current => ({ ...current, visualProcessing: next })); }} onBlur={() => onSetVisualProcessing(normalizeVisualProcessing(draft.visualProcessing))} addonAfter="%" /></label><Checkbox checked={visualProcessing.stabilization.tripod} disabled={!visualProcessing.stabilization.enabled} onChange={(event) => { const next = { ...visualProcessing, stabilization: { ...visualProcessing.stabilization, tripod: event.target.checked } }; setDraft(current => ({ ...current, visualProcessing: next })); onSetVisualProcessing(next); }}>三脚架模式</Checkbox><small>编辑器使用低耗平移防抖预览；云端导出再做两遍分析与精确稳定变换。</small></div></section> : null}
      {clip.kind === 'video' ? <section className="aicut-region-tracking">
        <h3>画面区域跟踪</h3>
        <small>先在播放器框住人物或物体，再分析移动轨迹并生成真实位置关键帧。</small>
        <Button type={trackingSelectionActive ? 'primary' : 'default'} disabled={regionTrackingBusy} onClick={onToggleTrackingSelection}>
          {trackingSelectionActive ? '完成框选' : '播放器框选目标'}
        </Button>
        <div className="aicut-inspector-grid">
          <label>目标中心X<InputNumber min={0} max={100} step={1} value={Math.round(normalizedTrackingRegion.centerX * 100)} disabled={regionTrackingBusy} onChange={(centerX) => onTrackingRegionChange?.({ ...normalizedTrackingRegion, centerX: Number(centerX ?? 50) / 100 })} addonAfter="%" /></label>
          <label>目标中心Y<InputNumber min={0} max={100} step={1} value={Math.round(normalizedTrackingRegion.centerY * 100)} disabled={regionTrackingBusy} onChange={(centerY) => onTrackingRegionChange?.({ ...normalizedTrackingRegion, centerY: Number(centerY ?? 50) / 100 })} addonAfter="%" /></label>
          <label>目标宽度<InputNumber min={5} max={100} step={1} value={Math.round(normalizedTrackingRegion.width * 100)} disabled={regionTrackingBusy} onChange={(width) => onTrackingRegionChange?.({ ...normalizedTrackingRegion, width: Number(width ?? 40) / 100 })} addonAfter="%" /></label>
          <label>目标高度<InputNumber min={5} max={100} step={1} value={Math.round(normalizedTrackingRegion.height * 100)} disabled={regionTrackingBusy} onChange={(height) => onTrackingRegionChange?.({ ...normalizedTrackingRegion, height: Number(height ?? 40) / 100 })} addonAfter="%" /></label>
        </div>
        <div className="aicut-region-tracking-actions">
          {regionTrackingBusy ? <Button danger block onClick={onCancelRegionTracking}>取消跟踪</Button> : <Button type="primary" block onClick={() => onRunRegionTracking?.(normalizedTrackingRegion)}>开始区域跟踪</Button>}
        </div>
        <small>最多抽取 60 张低清帧，结果进入统一关键帧链路，可预览、撤销、保存并随云端成片导出。倒放、定格和曲线变速片段会明确拒绝，不会写入半成品。</small>
      </section> : null}
      {['video', 'image'].includes(clip.kind) ? <section className="aicut-object-removal">
        <h3>区域物体消除</h3>
        <small>播放器中框住固定水印或静态物体；预览使用轻量周边采样，云端成片使用 FFmpeg 周边像素修复。</small>
        <Checkbox checked={visualProcessing.objectRemoval.enabled} onChange={(event) => onSetObjectRemovalEnabled?.(event.target.checked)}>启用区域消除</Checkbox>
        <Button type={objectRemovalSelectionActive ? 'primary' : 'default'} onClick={onToggleObjectRemovalSelection}>
          {objectRemovalSelectionActive ? '完成消除框选' : '播放器框选消除区域'}
        </Button>
        <div className="aicut-inspector-grid">
          <label>区域中心X<InputNumber min={0} max={100} step={1} value={Math.round(normalizedObjectRemovalRegion.centerX * 100)} onChange={(centerX) => onObjectRemovalRegionChange?.({ ...normalizedObjectRemovalRegion, centerX: Number(centerX ?? 50) / 100 })} onBlur={() => onCommitObjectRemovalRegion?.(normalizedObjectRemovalRegion, true)} addonAfter="%" /></label>
          <label>区域中心Y<InputNumber min={0} max={100} step={1} value={Math.round(normalizedObjectRemovalRegion.centerY * 100)} onChange={(centerY) => onObjectRemovalRegionChange?.({ ...normalizedObjectRemovalRegion, centerY: Number(centerY ?? 50) / 100 })} onBlur={() => onCommitObjectRemovalRegion?.(normalizedObjectRemovalRegion, true)} addonAfter="%" /></label>
          <label>区域宽度<InputNumber min={5} max={100} step={1} value={Math.round(normalizedObjectRemovalRegion.width * 100)} onChange={(width) => onObjectRemovalRegionChange?.({ ...normalizedObjectRemovalRegion, width: Number(width ?? 20) / 100 })} onBlur={() => onCommitObjectRemovalRegion?.(normalizedObjectRemovalRegion, true)} addonAfter="%" /></label>
          <label>区域高度<InputNumber min={5} max={100} step={1} value={Math.round(normalizedObjectRemovalRegion.height * 100)} onChange={(height) => onObjectRemovalRegionChange?.({ ...normalizedObjectRemovalRegion, height: Number(height ?? 20) / 100 })} onBlur={() => onCommitObjectRemovalRegion?.(normalizedObjectRemovalRegion, true)} addonAfter="%" /></label>
        </div>
        <small>适合固定水印、台标和小范围静态物体。移动物体请先使用画面区域跟踪；复杂背景仍可能留下修复痕迹。</small>
      </section> : null}
      {clip.kind === 'video' ? <section className="aicut-person-segmentation"><h3>AI 人物抠像</h3><small>首次使用才加载本地轻量模型；生成灰度人物蒙版后，预览、保存和云端导出使用同一份结果。</small>{personSegmentationStale ? <strong className="aicut-person-matte-stale">蒙版已过期：片段裁剪、时长、倍速、倒放或定格发生了变化，请重新生成。</strong> : null}<Checkbox checked={personSegmentationCurrent && personSegmentation.enabled} disabled={!personSegmentationCurrent || personSegmentationBusy} onChange={(event) => onSetVisualProcessing({ ...visualProcessing, personSegmentation: { ...personSegmentation, enabled: event.target.checked } })}>启用人物抠像</Checkbox><label>人物阈值 <small>{Math.round(personSegmentation.threshold * 100)}%</small><Slider min={0} max={1} step={.01} value={personSegmentation.threshold} disabled={!personSegmentation.matteAssetId || personSegmentationBusy} onChange={(threshold) => setDraft(current => ({ ...current, visualProcessing: { ...visualProcessing, personSegmentation: { ...personSegmentation, threshold } } }))} onChangeComplete={(threshold) => onSetVisualProcessing({ ...visualProcessing, personSegmentation: { ...personSegmentation, threshold } })} /></label><label>边缘柔化 <small>{Math.round(personSegmentation.feather * 100)}%</small><Slider min={0} max={.5} step={.01} value={personSegmentation.feather} disabled={!personSegmentation.matteAssetId || personSegmentationBusy} onChange={(feather) => setDraft(current => ({ ...current, visualProcessing: { ...visualProcessing, personSegmentation: { ...personSegmentation, feather } } }))} onChangeComplete={(feather) => onSetVisualProcessing({ ...visualProcessing, personSegmentation: { ...personSegmentation, feather } })} /></label>{personSegmentationBusy ? <><Button danger block onClick={onCancelPersonSegmentation}>取消生成</Button><small>正在本机处理：{Math.round(personSegmentationProgress * 100)}%</small></> : <Button type="primary" block onClick={onRunPersonSegmentation}>{personSegmentation.matteAssetId ? '重新生成人物蒙版' : '生成人物蒙版'}</Button>}<small>倒放、定格、曲线变速或超过 5 分钟会明确拒绝；取消时不会把半成品写入时间线。</small></section> : null}
      {['video', 'image'].includes(clip.kind) ? <section><h3>位置与大小</h3><div className="aicut-inspector-grid"><label>水平位置<InputNumber value={draft.canvasTransform.x} min={-1} max={1} step={.01} onChange={(value) => patchTransform({ x: value })} onBlur={() => patchTransform({}, true)} /></label><label>垂直位置<InputNumber value={draft.canvasTransform.y} min={-1} max={1} step={.01} onChange={(value) => patchTransform({ y: value })} onBlur={() => patchTransform({}, true)} /></label><label>横向缩放<InputNumber value={draft.canvasTransform.scaleX} min={.05} max={20} step={.05} onChange={(value) => patchTransform({ scaleX: value })} onBlur={() => patchTransform({}, true)} /></label><label>纵向缩放<InputNumber value={draft.canvasTransform.scaleY} min={.05} max={20} step={.05} onChange={(value) => patchTransform({ scaleY: value })} onBlur={() => patchTransform({}, true)} /></label><label>旋转<InputNumber value={draft.canvasTransform.rotation} min={-180} max={180} onChange={(value) => patchTransform({ rotation: value })} onBlur={() => patchTransform({}, true)} addonAfter="°" /></label></div><label>不透明度<Slider min={0} max={1} step={.01} value={draft.canvasTransform.opacity} onChange={(value) => patchTransform({ opacity: value })} onPointerUp={() => patchTransform({}, true)} /></label></section> : null}
	  {['video', 'image'].includes(clip.kind) ? <section className="aicut-blend-mode"><h3>画面合成</h3><label>混合模式<Select aria-label="混合模式" value={clip.blendMode || 'normal'} options={[
		{ value: 'normal', label: '正常' },
		{ value: 'multiply', label: '正片叠底' },
		{ value: 'screen', label: '滤色' },
		{ value: 'overlay', label: '叠加' },
		{ value: 'soft-light', label: '柔光' },
		{ value: 'difference', label: '差值' },
		{ value: 'add', label: '添加' }
	  ]} onChange={onSetBlendMode} /></label><small>与下方画面重叠时生效，预览和云端导出使用同一模式。</small></section> : null}
	  {['video', 'image'].includes(clip.kind) ? <section className="aicut-geometry-panel"><h3>裁剪与翻转</h3><div className="aicut-inspector-grid">{[
		['左裁剪', 'cropLeft', visualProcessing.geometry.cropRight],
		['上裁剪', 'cropTop', visualProcessing.geometry.cropBottom],
		['右裁剪', 'cropRight', visualProcessing.geometry.cropLeft],
		['下裁剪', 'cropBottom', visualProcessing.geometry.cropTop]
	  ].map(([label, key, opposite]) => <label key={key}>{label}<InputNumber aria-label={label} min={0} max={Math.max(0, Math.floor((.94 - opposite) * 100))} step={1} value={Math.round(visualProcessing.geometry[key] * 100)} onChange={(value) => patchVisual('geometry', { [key]: Number(value || 0) / 100 })} onBlur={() => patchVisual('geometry', {}, true)} onPressEnter={() => patchVisual('geometry', {}, true)} addonAfter="%" /></label>)}</div><div className="aicut-geometry-flips"><Checkbox checked={visualProcessing.geometry.flipHorizontal} onChange={(event) => patchVisual('geometry', { flipHorizontal: event.target.checked }, true)}>水平翻转</Checkbox><Checkbox checked={visualProcessing.geometry.flipVertical} onChange={(event) => patchVisual('geometry', { flipVertical: event.target.checked }, true)}>垂直翻转</Checkbox></div></section> : null}
      {['video', 'image'].includes(clip.kind) ? <section className="aicut-advanced-color"><h3>高级调色</h3>{[['曝光', 'exposure', -2, 2, .1], ['色温', 'temperature', -1, 1, .05], ['色调', 'tint', -1, 1, .05], ['高光', 'highlights', -1, 1, .05], ['阴影', 'shadows', -1, 1, .05], ['伽马', 'gamma', .5, 2, .05]].map(([label, key, min, max, step]) => <label key={key}>{label} <small>{Number(visualProcessing.color[key]).toFixed(2)}</small><Slider min={min} max={max} step={step} value={visualProcessing.color[key]} onChange={(value) => patchVisual('color', { [key]: value })} onChangeComplete={(value) => patchVisual('color', { [key]: value }, true)} /></label>)}<ColorCurveEditor curvePoints={visualProcessing.color.curvePoints} onChange={(curvePoints) => patchVisual('color', { curvePoints })} onCommit={(curvePoints) => patchVisual('color', { curvePoints }, true)} /></section> : null}
      {['video', 'image'].includes(clip.kind) ? <section className="aicut-visual-details"><h3>画面细节</h3>{[['锐化', 'sharpen'], ['暗角', 'vignette'], ['颗粒', 'grain']].map(([label, key]) => <div className="aicut-detail-control" key={key}><label>{label}<InputNumber aria-label={`${label}百分比`} size="small" min={0} max={100} step={5} value={Math.round(visualProcessing.details[key] * 100)} onChange={(value) => patchVisual('details', { [key]: Number(value ?? 0) / 100 }, true)} addonAfter="%" /></label><Slider min={0} max={1} step={.05} value={visualProcessing.details[key]} onChange={(value) => patchVisual('details', { [key]: value })} onChangeComplete={(value) => patchVisual('details', { [key]: value }, true)} /></div>)}<small>WebGL实时预览，FFmpeg成片使用对应的锐化、暗角和噪点滤镜。</small></section> : null}
      {['video', 'image'].includes(clip.kind) ? <section className="aicut-chroma-key"><h3>色键抠像</h3><div className="aicut-chroma-actions"><Button size="small" onClick={onDetectChroma}>自动识别幕布</Button><Button size="small" type={chromaPickerActive ? 'primary' : 'default'} onClick={onStartChromaPicker}>{chromaPickerActive ? '取消吸管取色' : '吸管取色'}</Button></div>{chromaPickerActive ? <small className="aicut-chroma-hint">请在播放器画面中点击要去除的颜色，按 Esc 也可取消。</small> : null}<Checkbox checked={visualProcessing.chromaKey.enabled} onChange={(event) => patchVisual('chromaKey', { enabled: event.target.checked }, true)}>启用色键</Checkbox><Checkbox checked={visualProcessing.chromaKey.removeBlack} onChange={(event) => patchVisual('chromaKey', { removeBlack: event.target.checked }, true)}>Remove Black（去黑底）</Checkbox><label>色键颜色<input type="color" value={visualProcessing.chromaKey.color} disabled={visualProcessing.chromaKey.removeBlack} onChange={(event) => patchVisual('chromaKey', { color: event.target.value, enabled: true, removeBlack: false }, true)} /></label><label>相似度 <small>{Math.round(visualProcessing.chromaKey.similarity * 100)}%</small><Slider min={.01} max={1} step={.01} value={visualProcessing.chromaKey.similarity} onChange={(similarity) => patchVisual('chromaKey', { similarity })} onChangeComplete={(similarity) => patchVisual('chromaKey', { similarity }, true)} /></label><label>柔化 <small>{Math.round(visualProcessing.chromaKey.blend * 100)}%</small><Slider min={0} max={1} step={.01} value={visualProcessing.chromaKey.blend} onChange={(blend) => patchVisual('chromaKey', { blend })} onChangeComplete={(blend) => patchVisual('chromaKey', { blend }, true)} /></label>{selectionCount > 1 ? <Button block onClick={() => onBatchApplyVisualProcessing?.({ chromaKey: visualProcessing.chromaKey })}>批量应用色键到已选 {selectionCount} 个片段</Button> : null}</section> : null}
      {['video', 'image'].includes(clip.kind) ? <section className="aicut-mask-panel"><h3>蒙版</h3><label>蒙版类型<Select value={visualProcessing.mask.type} options={[{ value: 'none', label: '无蒙版' }, { value: 'rectangle', label: '矩形' }, { value: 'ellipse', label: '椭圆' }, { value: 'polygon', label: '手绘多边形' }, { value: 'bezier', label: '贝塞尔曲线' }]} onChange={(type) => patchVisual('mask', { type, points: type === 'polygon' ? defaultPolygonMaskPoints.map(point => ({ ...point })) : [], bezierPoints: type === 'bezier' ? defaultBezierMaskPoints.map(point => ({ ...point })) : [] }, true)} /></label>{visualProcessing.mask.type === 'polygon' ? <PolygonMaskEditor maskPoints={visualProcessing.mask.points} onChange={(points) => patchVisual('mask', { points })} onCommit={(points) => patchVisual('mask', { points }, true)} /> : visualProcessing.mask.type === 'bezier' ? <BezierMaskEditor bezierPoints={visualProcessing.mask.bezierPoints} onChange={(bezierPoints) => patchVisual('mask', { bezierPoints })} onCommit={(bezierPoints) => patchVisual('mask', { bezierPoints }, true)} /> : <div className="aicut-inspector-grid"><label>中心X<InputNumber min={0} max={1} step={.01} value={visualProcessing.mask.x} onChange={(x) => patchVisual('mask', { x: Number(x ?? .5) }, true)} /></label><label>中心Y<InputNumber min={0} max={1} step={.01} value={visualProcessing.mask.y} onChange={(y) => patchVisual('mask', { y: Number(y ?? .5) }, true)} /></label><label>宽度<InputNumber min={.05} max={1} step={.01} value={visualProcessing.mask.width} onChange={(width) => patchVisual('mask', { width: Number(width ?? 1) }, true)} /></label><label>高度<InputNumber min={.05} max={1} step={.01} value={visualProcessing.mask.height} onChange={(height) => patchVisual('mask', { height: Number(height ?? 1) }, true)} /></label></div>}<label>羽化 <small>{Math.round(visualProcessing.mask.feather * 100)}%</small><Slider min={0} max={.5} step={.01} value={visualProcessing.mask.feather} onChange={(feather) => patchVisual('mask', { feather })} onChangeComplete={(feather) => patchVisual('mask', { feather }, true)} /></label><Checkbox checked={visualProcessing.mask.invert} onChange={(event) => patchVisual('mask', { invert: event.target.checked }, true)}>反转蒙版</Checkbox></section> : null}
      {clip.kind === 'subtitle' ? <section><h3>文字与字幕</h3><label>文字内容<Input.TextArea value={draft.text || ''} rows={4} onChange={(event) => setDraft(current => ({ ...current, text: event.target.value }))} onBlur={() => onPatchClip({ text: draft.text || '' })} /></label><label>字体<Select value={textStyle.fontFamily} options={aicutFontOptions} onChange={(fontFamily) => onPatchTrack({ fontFamily })} /></label><div className="aicut-inspector-grid"><label>字号<InputNumber min={12} max={240} value={textStyle.fontSize} onChange={(fontSize) => onPatchTrack({ fontSize: Number(fontSize || 54) })} /></label><label>描边<InputNumber min={0} max={12} value={textStyle.strokeWidth} onChange={(strokeWidth) => onPatchTrack({ strokeWidth: Number(strokeWidth || 0) })} /></label><label>字距<InputNumber min={-10} max={40} value={textStyle.letterSpacing} onChange={(letterSpacing) => onPatchTrack({ letterSpacing: Number(letterSpacing || 0) })} /></label><label>行高<InputNumber min={.8} max={3} step={.1} value={textStyle.lineHeight} onChange={(lineHeight) => onPatchTrack({ lineHeight: Number(lineHeight || 1.3) })} /></label></div><label>文字颜色<input type="color" value={textStyle.textColor.slice(0, 7)} onChange={(event) => onPatchTrack({ textColor: event.target.value })} /></label><label>描边颜色<input type="color" value={textStyle.strokeColor.slice(0, 7)} onChange={(event) => onPatchTrack({ strokeColor: event.target.value })} /></label><label>对齐<Select value={textStyle.textAlign} options={[{ value: 'left', label: '左对齐' }, { value: 'center', label: '居中' }, { value: 'right', label: '右对齐' }]} onChange={(textAlign) => onPatchTrack({ textAlign })} /></label><label>逐字动画<Select value={selection?.track?.settings?.subtitleAnimation || 'none'} options={subtitleAnimationCatalog.map((item) => ({ value: item.id, label: item.label }))} onChange={(subtitleAnimation) => onPatchTrack({ subtitleAnimation })} /></label></section> : null}
      <section><h3>关键帧</h3><KeyframeEasingEditor keyframes={clip.keyframes} onSet={onSetKeyframeEasing} onBatchEditKeyframes={onBatchEditKeyframes} onCopyKeyframes={onCopyKeyframes} onPasteKeyframes={onPasteKeyframes} keyframeClipboardCount={keyframeClipboardCount} /></section>
      {['video', 'audio'].includes(clip.kind) ? <section className="aicut-speed-panel"><h3>速度</h3><label>固定倍速<Select value={Number(clip.playbackRate) || 1} options={[.25, .5, .75, 1, 1.25, 1.5, 2, 3, 4].map(value => ({ value, label: `${value}x` }))} onChange={onSetPlaybackRate} /></label>{clip.kind === 'video' ? <><div className="aicut-speed-heading"><strong>曲线变速</strong><small>真实改变预览、源时间和导出结果</small></div><SpeedCurveEditor points={clip.speedCurve} rate={Number(clip.playbackRate) || 1} onChange={onSetSpeedCurve} /><div className="aicut-speed-presets">{speedCurvePresets.map(preset => <Button key={preset.id} size="small" type={JSON.stringify(clip.speedCurve || []) === JSON.stringify(preset.points) ? 'primary' : 'default'} onClick={() => onSetSpeedCurve(preset.points)}>{preset.label}</Button>)}</div><div className="aicut-speed-summary"><span>平均倍率 {(Number(clip.playbackRate) || 1).toFixed(2)}x</span><span>输出 {(Number(clip.durationMs) / 1000).toFixed(2)} 秒</span></div></> : null}</section> : null}
      {clip.kind === 'audio' ? <section><h3>音频</h3><label>基础音量<Slider min={0} max={2} step={.05} value={Number(draft.volume ?? 1)} onChange={(volume) => setDraft(current => ({ ...current, volume }))} onPointerUp={() => onPatchClip({ volume: Number(draft.volume ?? 1) })} /></label><div className="aicut-audio-fades"><strong>音频淡入淡出</strong><div className="aicut-inspector-grid"><label>淡入时长<InputNumber aria-label="淡入时长" min={0} max={Number(clip.durationMs) || 0} step={100} value={Number(draft.fadeInMs) || 0} onChange={(fadeInMs) => updateAudioFade('fadeInMs', fadeInMs)} addonAfter="ms" /></label><label>淡出时长<InputNumber aria-label="淡出时长" min={0} max={Number(clip.durationMs) || 0} step={100} value={Number(draft.fadeOutMs) || 0} onChange={(fadeOutMs) => updateAudioFade('fadeOutMs', fadeOutMs)} addonAfter="ms" /></label></div><small>也可以直接拖动时间线音频片段两端的圆形手柄。</small></div><VolumeEnvelopeEditor clip={clip} onUpsert={onVolumePointUpsert} onMove={onVolumePointMove} onRemove={onVolumePointRemove} /></section> : null}
      {clip.kind === 'audio' ? <section className="aicut-audio-processing"><h3>音频处理</h3><label>音频效果<Select aria-label="音频效果" showSearch optionFilterProp="label" value={audioProcessing.voiceEffect} options={audioEffectOptions} onChange={(voiceEffect) => onSetAudioProcessing({ ...audioProcessing, voiceEffect })} /></label><label>音调<InputNumber aria-label="音调" min={-12} max={12} step={.5} value={audioProcessing.pitchSemitones} onChange={(pitchSemitones) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), pitchSemitones: Number(pitchSemitones ?? 0) } }))} onBlur={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} onPressEnter={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} addonAfter="半音" /></label><small>浏览器使用轻量调制近似试听；最终导出由 FFmpeg 精确升降调并尽量保持片段时长。</small><label>降噪 <small>{Math.round(audioProcessing.denoise * 100)}%</small><Slider min={0} max={1} step={.05} value={audioProcessing.denoise} onChange={(denoise) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), denoise } }))} onChangeComplete={(denoise) => onSetAudioProcessing({ ...audioProcessing, denoise })} /></label><div className="aicut-inspector-grid"><label>低频<InputNumber min={-12} max={12} value={audioProcessing.eqLowDB} onChange={(eqLowDB) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), eqLowDB: Number(eqLowDB || 0) } }))} onBlur={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} addonAfter="dB" /></label><label>中频<InputNumber min={-12} max={12} value={audioProcessing.eqMidDB} onChange={(eqMidDB) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), eqMidDB: Number(eqMidDB || 0) } }))} onBlur={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} addonAfter="dB" /></label><label>高频<InputNumber min={-12} max={12} value={audioProcessing.eqHighDB} onChange={(eqHighDB) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), eqHighDB: Number(eqHighDB || 0) } }))} onBlur={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} addonAfter="dB" /></label></div><label>左右声像 <small>{audioProcessing.pan === 0 ? '居中' : audioProcessing.pan < 0 ? '偏左' : '偏右'} {audioProcessing.pan.toFixed(2)}</small><Slider aria-label="左右声像" min={-1} max={1} step={.01} value={audioProcessing.pan} onChange={(pan) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), pan } }))} onChangeComplete={(pan) => onSetAudioProcessing({ ...audioProcessing, pan })} /><Button size="small" onClick={() => onSetAudioProcessing({ ...audioProcessing, pan: 0 })}>居中</Button></label><small>中间保留原始左右声道；向一侧移动时只减小另一侧音量。</small><label>声道混音<Select value={audioProcessing.channelMode} options={[{ value: 'stereo', label: '立体声' }, { value: 'mono', label: '混合为单声道' }, { value: 'left', label: '仅左声道' }, { value: 'right', label: '仅右声道' }, { value: 'swap', label: '左右交换' }]} onChange={(channelMode) => onSetAudioProcessing({ ...audioProcessing, channelMode })} /></label><Checkbox checked={audioProcessing.loudnessEnabled} onChange={(event) => onSetAudioProcessing({ ...audioProcessing, loudnessEnabled: event.target.checked })}>响度自动匹配</Checkbox><label>目标响度<InputNumber min={-24} max={-8} step={1} disabled={!audioProcessing.loudnessEnabled} value={audioProcessing.targetLUFS} onChange={(targetLUFS) => setDraft(current => ({ ...current, audioProcessing: { ...normalizeAudioProcessing(current.audioProcessing), targetLUFS: Number(targetLUFS ?? -16) } }))} onBlur={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} onPressEnter={() => onSetAudioProcessing(normalizeAudioProcessing(draft.audioProcessing))} addonAfter="LUFS" /></label><small>浏览器使用低耗动态处理近似试听；云端导出使用 FFmpeg 响度标准化得到准确结果。</small></section> : null}
      {clip.transition ? <section><h3>转场</h3><label>持续时间<InputNumber min={100} max={3000} value={draft.transitionMs || 500} onChange={(transitionMs) => setDraft(current => ({ ...current, transitionMs }))} onBlur={() => onPatchTransition({ transitionMs: Number(draft.transitionMs || 500) })} addonAfter="ms" /></label></section> : null}
      {clip.kind === 'effect' ? <section><h3>特效参数</h3>{resource?.parameters?.length ? resource.parameters.map((parameter) => <label key={parameter.key}>{parameter.labelZh}<InputNumber min={parameter.min} max={parameter.max} step={parameter.step} value={draft.parameters?.[parameter.key] ?? parameter.defaultValue} onChange={(value) => setDraft(current => ({ ...current, parameters: { ...(current.parameters || {}), [parameter.key]: Number(value ?? parameter.defaultValue) } }))} onBlur={() => onPatchEffect({ parameters: draft.parameters })} onPressEnter={() => onPatchEffect({ parameters: draft.parameters })} addonAfter={parameter.key.endsWith('Ms') ? 'ms' : undefined} /></label>) : <label>强度<Slider min={0} max={1} step={.05} value={Number(draft.parameters?.intensity ?? .5)} onChange={(intensity) => setDraft(current => ({ ...current, parameters: { ...(current.parameters || {}), intensity } }))} onChangeComplete={(intensity) => onPatchEffect({ parameters: { ...(draft.parameters || {}), intensity } })} /></label>}</section> : null}
      {resourceId && !resource ? <section><h3>未知资源</h3><p>资源 ID：{resourceId}</p><small>当前版本只读保留该资源，不会擅自删除工程数据。</small></section> : null}
    </div>}
  </aside>;
}
