import { Crop, FlipHorizontal, Grid2X2, RotateCw } from 'lucide-react';
import { useState } from 'react';
import { InputNumber } from 'antd';

function assetID(asset) {
  return String(asset?.id || asset?.assetId || '').trim();
}

export function AgentImageTools({ sourceAsset, busy = false, onApply }) {
  const [crop, setCrop] = useState({ x: 0, y: 0, width: 512, height: 512 });
  const [cropOpen, setCropOpen] = useState(false);
  const sourceAssetId = assetID(sourceAsset);
  const disabled = busy || !sourceAssetId || typeof onApply !== 'function';
  const apply = (operation, settings) => onApply?.({ operation, sourceAssetId, settings });
  return <section className="agent-image-tools" aria-label="图片编辑工具">
    <header className="agent-image-tools__header"><div><strong>图片处理</strong><span>处理结果会保存为新图片</span></div><small>{sourceAssetId ? '当前图片已就绪' : '请先添加图片'}</small></header>
    <div className="agent-image-tools__action-grid">
      <button type="button" disabled={disabled} onClick={() => apply('rotate', { degrees: 90 })}><RotateCw size={15} /><span>旋转 90°</span></button>
      <button type="button" disabled={disabled} onClick={() => apply('mirror', { direction: 'horizontal' })}><FlipHorizontal size={15} /><span>水平镜像</span></button>
      <button type="button" disabled={disabled} onClick={() => apply('grid_split', { columns: 2, rows: 2 })}><Grid2X2 size={15} /><span>2 x 2 拆图</span></button>
      <button type="button" className={cropOpen ? 'is-active' : ''} aria-expanded={cropOpen} disabled={disabled} onClick={() => setCropOpen(open => !open)}><Crop size={15} /><span>裁剪</span></button>
    </div>
    {cropOpen ? <div className="agent-image-tools__crop-panel">
      <div className="agent-image-tools__crop-heading"><strong>裁剪区域</strong><span>单位：像素</span></div>
      <div className="agent-image-tools__crop-fields">
        <label><span>起点 X</span><InputNumber aria-label="裁剪起点 X" size="small" min={0} value={crop.x} onChange={x => setCrop(current => ({ ...current, x: Number(x) || 0 }))} /></label>
        <label><span>起点 Y</span><InputNumber aria-label="裁剪起点 Y" size="small" min={0} value={crop.y} onChange={y => setCrop(current => ({ ...current, y: Number(y) || 0 }))} /></label>
        <label><span>宽度</span><InputNumber aria-label="裁剪宽度" size="small" min={1} value={crop.width} onChange={width => setCrop(current => ({ ...current, width: Number(width) || 0 }))} /></label>
        <label><span>高度</span><InputNumber aria-label="裁剪高度" size="small" min={1} value={crop.height} onChange={height => setCrop(current => ({ ...current, height: Number(height) || 0 }))} /></label>
      </div>
      <button className="agent-image-tools__apply-crop" type="button" disabled={disabled || crop.width < 1 || crop.height < 1} onClick={() => apply('crop', crop)}><Crop size={14} /><span>生成裁剪结果</span></button>
    </div> : null}
  </section>;
}
