import { Button, Empty, Input, Upload } from 'antd';
import { Clapperboard, Film, Image as ImageIcon, Plus, Sparkles, Star, Type, WandSparkles } from 'lucide-react';
import { useMemo, useState } from 'react';
import { catalogForTab } from './editorCatalog.js';

export const resourceTabs = [
  ['媒体', 'media', ImageIcon],
  ['文字', 'text', Type],
  ['转场', 'transition', Film],
  ['特效', 'effect', Sparkles],
  ['滤镜', 'filter', WandSparkles],
  ['模板', 'template', Clapperboard],
];

const templates = [
  { id: 'template-short-drama', labelZh: '短剧节奏', category: 'template', description: '字幕、转场和节奏的基础组合' },
  { id: 'template-commentary', labelZh: '解说视频', category: 'template', description: '旁白字幕和画面衔接组合' },
  { id: 'template-interview', labelZh: '人物访谈', category: 'template', description: '下三分之一和人声增强组合' },
  { id: 'template-product', labelZh: '产品展示', category: 'template', description: '清晰滤镜和重点放大组合' },
];

function dragPayload(event, type, value) {
  event.dataTransfer.effectAllowed = 'copy';
  event.dataTransfer.setData(type, typeof value === 'string' ? value : JSON.stringify(value));
}

export default function ResourceLibrary({
  activeTab,
  assets = [],
  favorites = [],
  recent = [],
  uploading = false,
  onTabChange,
  onUpload,
  onPreviewAsset,
  onAddAsset,
  onOpenASR,
  onPreviewResource,
  onApplyResource,
  onToggleFavorite,
}) {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState('全部');
  const tab = resourceTabs.find(([label]) => label === activeTab) || resourceTabs[0];
  const category = tab[1];
  const resources = useMemo(() => {
    const base = category === 'template' ? templates : catalogForTab(category);
    const keyword = query.trim().toLowerCase();
    return base.filter((item) => {
      if (filter === '收藏' && !favorites.includes(item.id)) return false;
      if (filter === '最近使用' && !recent.includes(item.id)) return false;
      return !keyword || `${item.labelZh} ${item.description}`.toLowerCase().includes(keyword);
    });
  }, [category, favorites, filter, query, recent]);
  const visibleAssets = assets.filter((asset) => !query.trim() || String(asset.name || '').toLowerCase().includes(query.trim().toLowerCase()));

  return <div className="aicut-resource-library">
    <nav className="aicut-resource-tabs" aria-label="资源分类">{resourceTabs.map(([label, , Icon]) => <button type="button" key={label} className={activeTab === label ? 'active' : ''} onClick={() => onTabChange(label)}><Icon size={17} /><span>{label}</span></button>)}</nav>
    <section className="aicut-resource-panel">
      <div className="aicut-resource-search"><Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索资源" />{category === 'media' ? <Upload multiple showUploadList={false} beforeUpload={onUpload} accept="image/*,video/*,audio/*"><Button loading={uploading} icon={<Plus size={14} />}>添加</Button></Upload> : null}</div>
      <div className="aicut-resource-filters">{['全部', '收藏', '最近使用'].map((item) => <button type="button" key={item} className={filter === item ? 'active' : ''} onClick={() => setFilter(item)}>{item}</button>)}</div>
      {category === 'media' ? <div className="aicut-resource-list">{visibleAssets.length ? visibleAssets.map((asset) => <article key={asset.id} className="aicut-resource-card" draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-asset', asset.id)} onClick={() => onPreviewAsset(asset)} onDoubleClick={() => onAddAsset(asset)}><span className="aicut-resource-thumb"><ImageIcon size={19} /></span><span><strong>{asset.name}</strong><small>{asset.kind === 'audio' ? '音频素材' : asset.kind === 'image' ? '图片素材' : '视频素材'}</small></span><div>{['audio', 'video'].includes(asset.kind) ? <button type="button" onClick={(event) => { event.stopPropagation(); onOpenASR(asset); }}>识别字幕</button> : null}<button type="button" onClick={(event) => { event.stopPropagation(); onAddAsset(asset); }}>加入</button></div></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有素材，点击添加上传" />}</div>
        : <div className="aicut-resource-list">{resources.length ? resources.map((resource) => <article key={resource.id} className="aicut-resource-card" draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-resource', resource)} onClick={() => onPreviewResource(resource)} onDoubleClick={() => onApplyResource(resource)}><span className={`aicut-resource-thumb category-${resource.category}`}><Sparkles size={19} /></span><span><strong>{resource.labelZh}</strong><small>{resource.description}</small></span><button type="button" className={favorites.includes(resource.id) ? 'favorite active' : 'favorite'} aria-label={`${favorites.includes(resource.id) ? '取消收藏' : '收藏'}${resource.labelZh}`} onClick={(event) => { event.stopPropagation(); onToggleFavorite(resource.id); }}><Star size={13} fill={favorites.includes(resource.id) ? 'currentColor' : 'none'} /></button></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的资源" />}</div>}
      <p className="aicut-resource-hint">单击预览，双击应用，也可以拖到时间线。</p>
    </section>
  </div>;
}
