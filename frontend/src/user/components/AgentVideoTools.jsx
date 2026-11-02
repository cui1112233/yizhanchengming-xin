import { Crop, Film, Scissors, Volume2 } from 'lucide-react';
import { Input, InputNumber } from 'antd';
import { useState } from 'react';

function assetID(asset) {
  return String(asset?.id || asset?.assetId || '').trim();
}

export function AgentVideoTools({ sourceAsset, busy = false, onApply }) {
  const [trim, setTrim] = useState({ startMs: 0, endMs: 5000 });
  const [crop, setCrop] = useState({ x: 0, y: 0, width: 720, height: 1280 });
  const [timestamps, setTimestamps] = useState('0, 1, 2');
  const sourceAssetId = assetID(sourceAsset);
  const disabled = busy || !sourceAssetId || typeof onApply !== 'function';
  const apply = (operation, settings) => onApply?.({ operation, sourceAssetId, settings });
  const frameTimes = timestamps.split(',').map(value => Number(value.trim())).filter(value => Number.isFinite(value) && value >= 0).sort((left, right) => left - right);
  const uniqueFrameTimes = [...new Set(frameTimes)];
  return <section className="agent-video-tools" aria-label="视频处理工具">
    <header><strong>处理当前视频</strong><span>{sourceAssetId ? '输出到新节点' : '请先确认视频'}</span></header>
    <div className="agent-video-tools__quick-actions">
      <button type="button" title="提取音频" aria-label="提取音频" disabled={disabled} onClick={() => apply('separate_audio', {})}><Volume2 size={14} /></button>
      <button type="button" title="截取视频" aria-label="截取视频" disabled={disabled || trim.endMs <= trim.startMs} onClick={() => apply('trim', trim)}><Scissors size={14} /></button>
    </div>
    <div className="agent-video-tools__row">
      <Scissors size={14} aria-hidden="true" />
      <InputNumber aria-label="剪辑开始毫秒" size="small" min={0} value={trim.startMs} onChange={startMs => setTrim(current => ({ ...current, startMs: Number(startMs) || 0 }))} />
      <InputNumber aria-label="剪辑结束毫秒" size="small" min={1} value={trim.endMs} onChange={endMs => setTrim(current => ({ ...current, endMs: Number(endMs) || 0 }))} />
    </div>
    <div className="agent-video-tools__row agent-video-tools__crop">
      <Crop size={14} aria-hidden="true" />
      <InputNumber aria-label="裁剪起点 X" size="small" min={0} value={crop.x} onChange={x => setCrop(current => ({ ...current, x: Number(x) || 0 }))} />
      <InputNumber aria-label="裁剪起点 Y" size="small" min={0} value={crop.y} onChange={y => setCrop(current => ({ ...current, y: Number(y) || 0 }))} />
      <InputNumber aria-label="裁剪宽度" size="small" min={1} value={crop.width} onChange={width => setCrop(current => ({ ...current, width: Number(width) || 0 }))} />
      <InputNumber aria-label="裁剪高度" size="small" min={1} value={crop.height} onChange={height => setCrop(current => ({ ...current, height: Number(height) || 0 }))} />
      <button type="button" title="裁剪到新视频" aria-label="裁剪到新视频" disabled={disabled || crop.width < 1 || crop.height < 1} onClick={() => apply('crop', crop)}><Crop size={14} /></button>
    </div>
    <div className="agent-video-tools__row agent-video-tools__frames">
      <Film size={14} aria-hidden="true" />
      <Input aria-label="抽帧秒数" size="small" value={timestamps} onChange={event => setTimestamps(event.target.value)} placeholder="0, 1.5, 3" />
      <button type="button" title="抽取画面帧" aria-label="抽取画面帧" disabled={disabled || !uniqueFrameTimes.length} onClick={() => apply('extract_frames', { timestampsMs: uniqueFrameTimes.map(value => Math.round(value * 1000)) })}><Film size={14} /></button>
    </div>
  </section>;
}
