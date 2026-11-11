import { Button, Empty, Input, InputNumber, Upload } from 'antd';
import { Captions, Clapperboard, Film, Image as ImageIcon, Music, PanelLeftClose, PanelLeftOpen, Plus, Sparkles, Star, Type, WandSparkles } from 'lucide-react';
import { useMemo, useState } from 'react';
import { catalogForTab, normalizeCatalogParameters } from './editorCatalog.js';

export const resourceTabs = [
  ['媒体', 'media', ImageIcon],
  ['音频', 'audio', Music],
  ['文字', 'text', Type],
  ['字幕', 'caption', Captions],
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
  proxyBusyId = '',
  backgroundBusyId = '',
  collapsed = false,
  onTabChange,
  onUpload,
	onUseFont,
  onPreviewAsset,
  onAddAsset,
  onOpenASR,
  onAnalyzeAudio,
  onOpenTTS,
  onOpenVoiceClone,
  onGenerateProxy,
  onRemoveBackground,
  onOpenTextEditor,
  onPreviewResource,
  onApplyResource,
  onToggleFavorite,
  onToggleCollapsed,
}) {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState('全部');
  const [selectedResource, setSelectedResource] = useState(null);
  const [parameterValues, setParameterValues] = useState({});
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
	const visibleAssets = assets.filter((asset) => {
	  if (category === 'media' && !['video', 'image'].includes(asset.kind)) return false;
	  if (category === 'audio' && asset.kind !== 'audio') return false;
	  if (category === 'text' && asset.kind !== 'font') return false;
	  return !query.trim() || `${asset.name || ''} ${asset.fontFamily || ''}`.toLowerCase().includes(query.trim().toLowerCase());
	});
  const selectResource = (resource) => {
    const normalized = normalizeCatalogParameters(resource.id, {});
    setSelectedResource(resource);
    setParameterValues(normalized);
    onPreviewResource(resource, normalized);
  };
  const normalizedParameters = selectedResource ? normalizeCatalogParameters(selectedResource.id, parameterValues) : {};

  return <div className="aicut-resource-library">
    <nav className="aicut-resource-tabs" aria-label="资源分类">{resourceTabs.map(([label, , Icon]) => <button type="button" key={label} className={activeTab === label ? 'active' : ''} onClick={() => onTabChange(label)}><Icon size={17} /><span>{label}</span></button>)}<button type="button" className="aicut-resource-collapse" aria-label={collapsed ? '展开素材区' : '收起素材区'} onClick={onToggleCollapsed}>{collapsed ? <PanelLeftOpen size={17} /> : <PanelLeftClose size={17} />}<span>{collapsed ? '展开' : '收起'}</span></button></nav>
    <section className="aicut-resource-panel">
	  <div className="aicut-resource-search"><Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索资源" />{category === 'media' ? <Upload multiple showUploadList={false} beforeUpload={onUpload} accept="image/*,video/*,audio/*"><Button loading={uploading} icon={<Plus size={14} />}>添加</Button></Upload> : null}{category === 'audio' ? <><Button icon={<WandSparkles size={14} />} onClick={onOpenTTS}>文字转语音</Button><Button icon={<Music size={14} />} onClick={() => onOpenVoiceClone()}>声音克隆</Button></> : null}{category === 'text' ? <><Upload showUploadList={false} beforeUpload={onUpload} accept=".ttf,.otf"><Button loading={uploading}>上传字体</Button></Upload><Button icon={<Plus size={14} />} onClick={onOpenTextEditor}>新建自定义文字</Button></> : null}</div>
      <div className="aicut-resource-filters">{['全部', '收藏', '最近使用'].map((item) => <button type="button" key={item} className={filter === item ? 'active' : ''} onClick={() => setFilter(item)}>{item}</button>)}</div>
      {category === 'media' || category === 'audio' ? <div className="aicut-resource-list">{visibleAssets.length ? visibleAssets.map((asset) => <article key={asset.id} className="aicut-resource-card" draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-asset', asset.id)} onClick={() => onPreviewAsset(asset)} onDoubleClick={() => onAddAsset(asset)}><span className="aicut-resource-thumb">{asset.kind === 'audio' ? <Music size={19} /> : <ImageIcon size={19} />}</span><span><strong>{asset.name}</strong><small>{asset.kind === 'audio' ? '音频素材' : asset.kind === 'image' ? '图片素材' : asset.proxyUrl ? '视频素材 · 代理已启用' : '视频素材'}</small></span><div>{category === 'audio' ? <><button type="button" onClick={(event) => { event.stopPropagation(); onAnalyzeAudio(asset); }}>分析节拍</button><button type="button" onClick={(event) => { event.stopPropagation(); onOpenVoiceClone(asset); }}>克隆音色</button></> : ['audio', 'video'].includes(asset.kind) ? <button type="button" onClick={(event) => { event.stopPropagation(); onOpenASR(asset); }}>识别字幕</button> : null}{asset.kind === 'image' ? <button type="button" onClick={(event) => { event.stopPropagation(); onRemoveBackground(asset); }}>{backgroundBusyId === asset.id ? '取消去背' : '背景移除'}</button> : null}{asset.kind === 'video' ? <button type="button" disabled={proxyBusyId === asset.id} onClick={(event) => { event.stopPropagation(); onGenerateProxy(asset); }}>{proxyBusyId === asset.id ? '生成中…' : asset.proxyUrl ? '重新生成代理' : '生成代理'}</button> : null}<button type="button" onClick={(event) => { event.stopPropagation(); onAddAsset(asset); }}>加入</button></div></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={category === 'audio' ? '还没有音频素材' : '还没有素材，点击添加上传'} />}</div>
		: <>{category === 'text' && visibleAssets.length ? <div className="aicut-resource-list aicut-font-assets">{visibleAssets.map((asset) => <article key={asset.id} className="aicut-resource-card" onClick={() => onUseFont(asset)}><span className="aicut-resource-thumb"><Type size={19} /></span><span><strong>{asset.fontFamily || asset.name}</strong><small>{asset.name} · 项目用户字体</small></span><button type="button" onClick={(event) => { event.stopPropagation(); onUseFont(asset); }}>用于文字</button></article>)}</div> : null}<div className="aicut-resource-list">{resources.length ? resources.map((resource) => <article key={resource.id} className={`aicut-resource-card${selectedResource?.id === resource.id ? ' selected' : ''}`} draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-resource', resource)} onClick={() => selectResource(resource)} onDoubleClick={() => onApplyResource(resource, normalizeCatalogParameters(resource.id, {}))}><span className={`aicut-resource-thumb category-${resource.category}`}><Sparkles size={19} /></span><span><strong>{resource.labelZh}</strong><small>{resource.description}</small></span><button type="button" className={favorites.includes(resource.id) ? 'favorite active' : 'favorite'} aria-label={`${favorites.includes(resource.id) ? '取消收藏' : '收藏'}${resource.labelZh}`} onClick={(event) => { event.stopPropagation(); onToggleFavorite(resource.id); }}><Star size={13} fill={favorites.includes(resource.id) ? 'currentColor' : 'none'} /></button></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的资源" />}</div>{selectedResource && selectedResource.category === category ? <div className="aicut-resource-parameters"><header><span><strong>{selectedResource.labelZh}</strong><small>{selectedResource.description}</small></span><Button size="small" onClick={() => { setSelectedResource(null); setParameterValues({}); onPreviewResource(null, {}); }}>关闭预览</Button></header>{selectedResource.parameters.length ? <div className="aicut-resource-parameter-grid">{selectedResource.parameters.map((parameter) => <label key={parameter.key}><span>{parameter.labelZh}</span><InputNumber min={parameter.min} max={parameter.max} step={parameter.step} value={parameterValues[parameter.key] ?? parameter.defaultValue} onChange={(value) => { const next = normalizeCatalogParameters(selectedResource.id, { ...parameterValues, [parameter.key]: value }); setParameterValues(next); onPreviewResource(selectedResource, next); }} addonAfter={parameter.key.endsWith('Ms') ? 'ms' : undefined} /></label>)}</div> : <p>该资源没有额外参数，应用后仍可在右侧属性栏继续调整。</p>}<Button type="primary" block onClick={() => onApplyResource(selectedResource, normalizedParameters)}>应用到时间线</Button></div> : null}</>}
      <p className="aicut-resource-hint">单击预览，双击应用，也可以拖到时间线。</p>
    </section>
  </div>;
}
