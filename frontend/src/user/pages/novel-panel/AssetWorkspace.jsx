import { Alert, Button, Empty, Image, Tag } from 'antd';
import { useMemo } from 'react';
import { buildAssetCards } from './assetWorkspaceModel';
import './asset-workspace.css';

export function AssetWorkspace({ project, onRetry }) {
  const assets = project?.data?.assets || project?.data?.referenceAssets || [];
  const cards = useMemo(() => buildAssetCards(assets), [assets]);
  if (!cards.length) return <Empty description="暂无资产" />;
  return (
    <section className="novel-panel-assets" aria-label="资产设置">
      <div className="novel-panel-assets__grid">
        {cards.map(card => (
          <article className="novel-panel-assets__card" key={`${card.assetType}-${card.assetId}`}>
            <div className="novel-panel-assets__preview">
              {card.preview.kind === 'image'
                ? <Image src={card.preview.src} alt={card.name} preview />
                : <span>{card.preview.label}</span>}
            </div>
            <div className="novel-panel-assets__meta">
              <strong>{card.name || '未命名资产'}</strong>
              <Tag>{card.assetType || 'asset'}</Tag>
            </div>
            {card.error && <Alert type="error" showIcon message={card.error} action={card.retryable && <Button size="small" onClick={() => onRetry?.(card)}>重试</Button>} />}
          </article>
        ))}
      </div>
    </section>
  );
}

export default AssetWorkspace;
