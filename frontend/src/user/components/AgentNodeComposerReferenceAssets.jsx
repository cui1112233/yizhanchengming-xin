import { Button, Popover } from 'antd';
import { FileText, Film, ImagePlus, Music2, Plus, X } from 'lucide-react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';

function stopCanvasEvent(event) {
  event.preventDefault();
  event.stopPropagation();
}

function ReferencePreview({ asset, index }) {
  if (asset?.url && asset.kind === 'video') return <AuthenticatedAgentVideo source={asset.url} aria-label={`视频参考素材 ${index + 1}`} preload="metadata" muted />;
  if (asset?.url && asset.kind === 'image') return <AuthenticatedAgentImage source={asset.url} alt={`图片参考素材 ${index + 1}`} />;
  const TypeIcon = asset?.kind === 'video' ? Film : asset?.kind === 'audio' ? Music2 : asset?.kind === 'image' ? ImagePlus : FileText;
  return <span className="agent-node-composer__reference-fallback" data-kind={asset?.kind || 'text'}><TypeIcon size={15} /></span>;
}

export function AgentNodeComposerReferenceStrip({ references = [], disabled = false, onRemove }) {
  const items = Array.isArray(references) ? references : [];
  if (!items.length) return null;
  return <div className="agent-node-composer__selected-references" aria-label="已引用资产">
    {items.map((asset, index) => <div className="agent-node-composer__selected-reference" key={asset.id || asset.assetId || `${asset.url}-${index}`}>
      <ReferencePreview asset={asset} index={index} />
      <span title={asset.name || asset.id || asset.assetId}>{asset.name || `参考素材 ${index + 1}`}</span>
      <button
        type="button"
        aria-label={`移除资产 ${asset.name || index + 1}`}
        title="移除资产"
        disabled={disabled}
        onClick={event => {
          stopCanvasEvent(event);
          onRemove?.(asset, event);
        }}
      ><X size={13} /></button>
    </div>)}
  </div>;
}

export function AgentNodeComposerReferenceAssets({ references = [], disabled = false, onAdd, onRemove }) {
  const items = Array.isArray(references) ? references : [];
  const content = <div className="agent-node-composer__reference-popover" aria-label="已选参考素材">
    {items.length ? <div className="agent-node-composer__reference-list">
      {items.map((asset, index) => <div className="agent-node-composer__reference-item" key={asset.id || `${asset.url}-${index}`}>
        <ReferencePreview asset={asset} index={index} />
        <span title={asset.name || asset.id}>{asset.name || `参考素材 ${index + 1}`}</span>
        <button
          type="button"
          aria-label={`移除第 ${index + 1} 个参考素材`}
          title="移除参考素材"
          disabled={disabled}
          onClick={event => {
            stopCanvasEvent(event);
            onRemove?.(asset, event);
          }}
        ><X size={14} /></button>
      </div>)}
    </div> : <small>暂未添加参考素材</small>}
    <Button
      type="text"
      size="small"
      icon={<Plus size={14} />}
      disabled={disabled}
      onClick={event => {
        stopCanvasEvent(event);
        onAdd?.(event);
      }}
    >添加参考素材</Button>
  </div>;

  return <Popover content={content} trigger="click" placement="bottomLeft" getPopupContainer={trigger => trigger.parentElement}>
    <button type="button" className="agent-node-composer__reference-action" disabled={disabled} onClick={stopCanvasEvent}>
      <ImagePlus size={15} /><span>参考素材</span><i>{items.length}</i>
    </button>
  </Popover>;
}
