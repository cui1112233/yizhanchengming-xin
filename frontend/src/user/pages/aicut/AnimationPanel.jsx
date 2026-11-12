import { Button, InputNumber, Segmented } from 'antd';
import { useState } from 'react';
import { clipAnimationPresets, clipAnimationDurations, normalizeClipAnimation } from '../aicutClipAnimation.js';
import './animation-panel.css';

export default function AnimationPanel({ clip, disabled = false, onChange }) {
  const [phase, setPhase] = useState('entrance');
  const settings = normalizeClipAnimation(clip?.animation);
  const selected = settings[phase];
  const effective = clipAnimationDurations(settings, clip?.durationMs);
  const update = patch => onChange?.({ ...settings, [phase]: { ...selected, ...patch } });
  const isDisabled = disabled || !clip || !['video', 'image'].includes(clip.kind);
  return <section className="aicut-animation-panel" aria-label="画面动画">
    <Segmented block value={phase} onChange={setPhase} options={[{ value: 'entrance', label: '入场动画' }, { value: 'exit', label: '出场动画' }, { value: 'loop', label: '循环动画' }]} />
    <div className="aicut-animation-presets">{clipAnimationPresets[phase].map(preset => <Button key={preset.value} disabled={isDisabled} type={selected.preset === preset.value ? 'primary' : 'default'} aria-pressed={selected.preset === preset.value} onClick={() => update({ preset: preset.value })}>{preset.label}</Button>)}</div>
    <label>{phase === 'loop' ? '循环周期' : '动画时长'}<InputNumber aria-label={phase === 'loop' ? '循环周期（秒）' : '动画时长（秒）'} disabled={isDisabled || selected.preset === 'none'} min={.05} max={10} step={.05} precision={2} value={selected.durationMs / 1000} suffix="秒" onChange={value => value !== null && update({ durationMs: Math.round(value * 1000) })} /></label>
    {effective.entranceMs + effective.exitMs < (settings.entrance.preset === 'none' ? 0 : settings.entrance.durationMs) + (settings.exit.preset === 'none' ? 0 : settings.exit.durationMs) && <small>片段较短，入场和出场时长已按比例缩短，避免重叠。</small>}
    <small>入场、出场和循环可分别设置，预览与导出使用相同的片段时间。</small>
    <Button size="small" disabled={isDisabled} onClick={() => onChange?.(null)}>清除全部动画</Button>
  </section>;
}
