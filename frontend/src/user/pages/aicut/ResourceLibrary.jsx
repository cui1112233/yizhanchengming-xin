import { Button, Dropdown, Empty, Input, InputNumber, Upload } from 'antd';
import { Captions, Clapperboard, Film, Image as ImageIcon, MoreHorizontal, Music, PanelLeftClose, PanelLeftOpen, Plus, Search, Sparkles, Star, Type, WandSparkles, SlidersHorizontal, Sticker, Package, Puzzle } from 'lucide-react';
import { useMemo, useState } from 'react';
import { catalogForTab, normalizeCatalogParameters } from './editorCatalog.js';
import { assetDurationLabel, assetToolActions } from './resourceLibraryModel.js';
import './resource-library-redesign.css';
import MediaArtwork from './MediaArtwork.jsx';
import AssetFolderDirectory from './AssetFolderDirectory.jsx';
import { projectTemplateCatalog } from './projectTemplates.js';

export const resourceTabs = [
  ['素材', 'media', ImageIcon], ['音频', 'audio', Music], ['文本', 'text', Type],
  ['贴纸', 'sticker', Sticker], ['特效', 'effect', Sparkles], ['转场', 'transition', Film],
  ['字幕', 'caption', Captions], ['智能包装', 'packaging', Package], ['滤镜', 'filter', WandSparkles],
  ['调节', 'adjustment', SlidersHorizontal], ['模板', 'template', Clapperboard], ['插件', 'plugin', Puzzle],
];

const templates = projectTemplateCatalog.map(template => ({ ...template, parameters: [] }));

function dragPayload(event, type, value) {
  event.dataTransfer.effectAllowed = type === 'application/x-aicut-asset' ? 'copyMove' : 'copy';
  event.dataTransfer.setData(type, typeof value === 'string' ? value : JSON.stringify(value));
}

function cardKeyboard(event, preview, apply) {
  if (event.target !== event.currentTarget) return;
  if (event.key === 'Enter') { event.preventDefault(); event.stopPropagation(); apply(); }
  if (event.key === ' ') { event.preventDefault(); event.stopPropagation(); preview(); }
}

function stopCardAction(event) {
  event.stopPropagation();
}

function MediaThumbnail({ asset, projectId, used }) {
  return <span className={`aicut-resource-thumb media-${asset.kind}`}>
    <MediaArtwork asset={asset} projectId={projectId} />
    {used ? <small className="aicut-media-used">已添加</small> : null}
    <small className="aicut-media-duration">{assetDurationLabel(asset)}</small>
  </span>;
}

export default function ResourceLibrary({
  projectId,
  usedAssetIds = [],
  activeTab,
  workspaces = {},
  personalTemplateWorkspace,
  audioRecorder,
  assets = [],
  assetFolders = [],
  selectedFolderId = 'all',
  onSelectFolder,
  onCreateFolder,
  onRenameFolder,
  onDeleteFolder,
  onMoveAssetFolder,
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
  const [previewAssetId, setPreviewAssetId] = useState('');
  const [templateSource, setTemplateSource] = useState('builtin');
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState('全部');
  const [mediaKind, setMediaKind] = useState('all');
  const [recentAssetIds, setRecentAssetIds] = useState([]);
  const [selectedResource, setSelectedResource] = useState(null);
  const [parameterValues, setParameterValues] = useState({});
  const displayTab = activeTab === '媒体' ? '素材' : activeTab === '文字' ? '文本' : activeTab;
  const tab = resourceTabs.find(([label]) => label === displayTab) || resourceTabs[0];
  const category = tab[1];
  const customWorkspace = category === 'template' && templateSource === 'personal' ? personalTemplateWorkspace : workspaces[category];
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
	  if (category === 'media' && selectedFolderId !== 'all' && (asset.folderId || '') !== selectedFolderId) return false;
	  if (category === 'media' && (!['video', 'image', 'audio'].includes(asset.kind) || (mediaKind !== 'all' && asset.kind !== mediaKind))) return false;
	  if (category === 'audio' && asset.kind !== 'audio') return false;
	  if (category === 'audio' && filter === '收藏' && !favorites.includes(asset.id)) return false;
	  if (category === 'audio' && filter === '最近使用' && !recentAssetIds.includes(asset.id)) return false;
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

  const previewAsset = (asset) => { setPreviewAssetId(asset.id); setRecentAssetIds(current => [asset.id, ...current.filter(id => id !== asset.id)].slice(0, 50)); onPreviewAsset(asset); };
  const toolHandlers = { asr: onOpenASR, proxy: onGenerateProxy, background: onRemoveBackground, analyze: onAnalyzeAudio, clone: onOpenVoiceClone };
  const assetMenu = (asset) => ({
    items: [...(onMoveAssetFolder ? [{ key: 'move-folder', label: '移动到文件夹', children: [{ key: 'folder:', label: '未分类' }, ...assetFolders.map(folder => ({ key: `folder:${folder.id}`, label: folder.name }))] }] : []), { key: 'favorite', label: favorites.includes(asset.id) ? '取消收藏素材' : '收藏素材' }, ...assetToolActions(asset, { proxyBusyId, backgroundBusyId }).map((action) => ({
      ...action,
      label: action.key === 'analyze' ? <span>分析节拍</span> : action.label,
    }))],
    onClick: ({ key, domEvent }) => { domEvent.stopPropagation(); if (key.startsWith('folder:')) onMoveAssetFolder?.(asset.id, key.slice(7)); else if (key === 'favorite') onToggleFavorite(asset.id); else toolHandlers[key]?.(asset); },
  });

  return <div className="aicut-resource-library aicut-library-redesign">
    <nav className="aicut-resource-tabs" aria-label="资源分类"><div className="aicut-resource-tab-scroll">{resourceTabs.map(([label, , Icon]) => <button type="button" key={label} aria-pressed={displayTab === label} className={displayTab === label ? 'active' : ''} onClick={() => onTabChange(label === '素材' ? '媒体' : label === '文本' ? '文字' : label)}><Icon size={16} /><span>{label}</span></button>)}</div><button type="button" className="aicut-resource-collapse" aria-label={collapsed ? '展开素材区' : '收起素材区'} title={collapsed ? '展开素材区' : '收起素材区'} onClick={onToggleCollapsed}>{collapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}<span>{collapsed ? '展开' : '收起'}</span></button></nav>
    <div className={`aicut-library-workspace${customWorkspace ? ' aicut-custom-workspace' : ''}`}>
      <nav className="aicut-resource-directory" aria-label="素材目录">
        <strong>{category === 'media' ? '项目素材' : activeTab}</strong>
        {customWorkspace ? <span className="aicut-directory-note">当前工程</span> : category === 'media' ? [['all', '全部素材'], ['video', '视频'], ['image', '图片'], ['audio', '音频']].map(([kind, label]) => <button type="button" key={kind} aria-pressed={mediaKind === kind && selectedFolderId === 'all'} className={mediaKind === kind && selectedFolderId === 'all' ? 'active' : ''} onClick={() => { setMediaKind(kind); onSelectFolder?.('all'); }}><span>{label}</span><small>{assets.filter(asset => kind === 'all' ? ['video', 'image', 'audio'].includes(asset.kind) : asset.kind === kind).length}</small></button>) : ['全部', '收藏', '最近使用'].map(item => <button type="button" key={item} aria-pressed={filter === item} className={filter === item ? 'active' : ''} onClick={() => setFilter(item)}>{item}</button>)}
        {category === 'media' && !customWorkspace && onCreateFolder ? <AssetFolderDirectory folders={assetFolders} assets={assets.filter(asset => ['video', 'image', 'audio'].includes(asset.kind))} selectedId={selectedFolderId} onSelect={id => { setMediaKind('all'); onSelectFolder?.(id); }} onCreate={onCreateFolder} onRename={onRenameFolder} onDelete={onDeleteFolder} onMove={onMoveAssetFolder} /> : null}
      </nav>
    <section className="aicut-resource-panel" aria-label={`${displayTab || '素材'}资源`}>
      {category === 'template' && personalTemplateWorkspace ? <div className="aicut-resource-create" aria-label="模板来源"><Button size="small" type={templateSource === 'builtin' ? 'primary' : 'default'} onClick={() => setTemplateSource('builtin')}>内置模板</Button><Button size="small" type={templateSource === 'personal' ? 'primary' : 'default'} onClick={() => setTemplateSource('personal')}>我的模板</Button></div> : null}
      {customWorkspace || <>
      <div className="aicut-resource-search"><Input allowClear prefix={<Search size={14} />} aria-label="搜索资源" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索资源" />{category === 'media' || category === 'audio' ? <Upload multiple showUploadList={false} beforeUpload={onUpload} accept={category === 'audio' ? 'audio/*' : 'image/*,video/*,audio/*'}><Button loading={uploading} icon={<Plus size={14} />}>导入</Button></Upload> : null}</div>
      {category === 'audio' || category === 'text' ? <div className="aicut-resource-create">{category === 'audio' ? <><Button size="small" icon={<WandSparkles size={13} />} onClick={onOpenTTS}>文字转语音</Button><Button size="small" icon={<Music size={13} />} onClick={() => onOpenVoiceClone()}>声音克隆</Button>{audioRecorder}</> : <><Upload showUploadList={false} beforeUpload={onUpload} accept=".ttf,.otf"><Button size="small" loading={uploading}>上传字体</Button></Upload><Button size="small" icon={<Plus size={13} />} onClick={onOpenTextEditor}>新建自定义文字</Button></>}</div> : null}
      <div className="aicut-resource-heading"><span>{category === 'media' || category === 'audio' ? `素材 · ${visibleAssets.length}` : `${activeTab} · ${resources.length}`}</span><small>双击添加</small></div>
      <div className="aicut-resource-content">
        {category === 'media' || category === 'audio' ? <div className={`aicut-resource-list aicut-media-grid${category === 'audio' ? ' is-audio' : ''}`}>
          {visibleAssets.length ? visibleAssets.map((asset) => <article key={asset.id} className={`aicut-resource-card aicut-media-card${previewAssetId === asset.id ? ' selected' : ''}`} tabIndex={0} aria-label={`${asset.name}，空格预览，回车加入时间线`} draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-asset', asset.id)} onKeyDown={(event) => cardKeyboard(event, () => previewAsset(asset), () => onAddAsset(asset))} onClick={() => previewAsset(asset)} onDoubleClick={() => onAddAsset(asset)}>
            <MediaThumbnail asset={asset} projectId={projectId} used={usedAssetIds.includes(asset.id)} />
            <span className="aicut-media-title"><strong title={asset.name}>{asset.name}</strong><small>{asset.kind === 'audio' ? '音频素材' : asset.kind === 'image' ? '图片素材' : asset.proxyUrl ? '视频素材 · 代理已启用' : '视频素材'}</small></span>
            <footer className="aicut-media-actions" onClick={stopCardAction} onDoubleClick={stopCardAction} onKeyDown={stopCardAction} onDragStart={(event) => event.preventDefault()}>
              <button type="button" className="aicut-media-add" aria-label={`将${asset.name}加入时间线`} onClick={() => onAddAsset(asset)}><Plus size={13} /><span>加入时间线</span></button>
              <Dropdown trigger={['click']} overlayClassName="aicut-resource-menu" menu={assetMenu(asset)}><button type="button" className="aicut-media-more" aria-label={`${asset.name}的更多操作`} title="更多操作"><MoreHorizontal size={17} /></button></Dropdown>
            </footer>
          </article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={query ? '没有匹配的素材' : category === 'audio' ? '还没有音频素材' : '导入图片、视频，开始剪辑'} />}
        </div> : <>
          {category === 'text' && visibleAssets.length ? <div className="aicut-resource-list aicut-font-assets">{visibleAssets.map((asset) => <article key={asset.id} className="aicut-resource-card" tabIndex={0} aria-label={`使用字体${asset.fontFamily || asset.name}`} onKeyDown={(event) => cardKeyboard(event, () => onUseFont(asset), () => onUseFont(asset))} onClick={() => onUseFont(asset)}><span className="aicut-resource-thumb"><Type size={19} /></span><span><strong>{asset.fontFamily || asset.name}</strong><small>{asset.name} · 项目用户字体</small></span><button type="button" onDoubleClick={stopCardAction} onClick={(event) => { event.stopPropagation(); onUseFont(asset); }}>使用</button></article>)}</div> : null}
          {selectedResource && selectedResource.category === category ? <div className="aicut-resource-parameters"><header><span><strong>{selectedResource.labelZh}</strong><small>{selectedResource.description}</small></span><Button size="small" onClick={() => { setSelectedResource(null); setParameterValues({}); onPreviewResource(null, {}); }}>关闭预览</Button></header>{selectedResource.parameters.length ? <div className="aicut-resource-parameter-grid">{selectedResource.parameters.map((parameter) => <label key={parameter.key}><span>{parameter.labelZh}</span><InputNumber min={parameter.min} max={parameter.max} step={parameter.step} value={parameterValues[parameter.key] ?? parameter.defaultValue} onChange={(value) => { const next = normalizeCatalogParameters(selectedResource.id, { ...parameterValues, [parameter.key]: value }); setParameterValues(next); onPreviewResource(selectedResource, next); }} addonAfter={parameter.key.endsWith('Ms') ? 'ms' : undefined} /></label>)}</div> : <p>该资源没有额外参数，应用后仍可在右侧属性栏继续调整。</p>}<Button type="primary" block onClick={() => onApplyResource(selectedResource, normalizedParameters)}>应用到时间线</Button></div> : null}
          <div className="aicut-resource-list aicut-catalog-list">{resources.length ? resources.map((resource) => <article key={resource.id} className={`aicut-resource-card${selectedResource?.id === resource.id ? ' selected' : ''}`} tabIndex={0} aria-label={`${resource.labelZh}，空格预览，回车应用`} draggable onDragStart={(event) => dragPayload(event, 'application/x-aicut-resource', resource)} onKeyDown={(event) => cardKeyboard(event, () => selectResource(resource), () => onApplyResource(resource, normalizeCatalogParameters(resource.id, {})))} onClick={() => selectResource(resource)} onDoubleClick={() => onApplyResource(resource, normalizeCatalogParameters(resource.id, {}))}><span className={`aicut-resource-thumb category-${resource.category}`}><Sparkles size={19} /></span><span><strong>{resource.labelZh}</strong><small title={resource.description}>{resource.description}</small></span><button type="button" className={favorites.includes(resource.id) ? 'favorite active' : 'favorite'} aria-pressed={favorites.includes(resource.id)} aria-label={`${favorites.includes(resource.id) ? '取消收藏' : '收藏'}${resource.labelZh}`} onDoubleClick={stopCardAction} onClick={(event) => { event.stopPropagation(); onToggleFavorite(resource.id); }}><Star size={14} fill={favorites.includes(resource.id) ? 'currentColor' : 'none'} /></button></article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的资源" />}</div>
        </>}
      </div>
      <p className="aicut-resource-hint">单击预览 · 双击或拖入时间线</p>
      </>}
    </section>
    </div>
  </div>;
}
