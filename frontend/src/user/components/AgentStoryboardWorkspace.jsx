import { Button, Empty, Input, Segmented, Select, Spin, Tag, Tooltip, message } from 'antd';
import { ArrowDown, ArrowUp, Check, ChevronDown, ChevronRight, FileText, Film, Folder, FolderOpen, GripVertical, Image as ImageIcon, ListChecks, LocateFixed, MessageSquarePlus, Music, Plus, Save, Search, Sparkles, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { updateAgentScriptShot, uploadAgentAsset } from '../../shared/api/agent';
import { listAvailableModels } from '../../shared/api/modelCatalog';
import { buildStoryboardLibrary, findStoryboardSelection, storyboardItemMetaLabel, storyboardSelectionKey } from '../pages/agent-storyboard-model';
import { groupResultLibrary, locateShotNode } from '../pages/agent-result-lineage';
import { buildShotBatchPayload } from '../pages/agent-script-shots';
import { writeStoryboardAssetDrag } from '../pages/agent-storyboard-conversation-drag';
import { STORYBOARD_SPLIT_DEFAULTS, storyboardSplitRatioFromPointer } from '../pages/agent-storyboard-split';
import { filterStoryboardLibrary, moveStoryboardMaterial, normalizeStoryboardState, reorderStoryboardMaterial, storyboardBatchAssets, storyboardBatchCategoryPatch, storyboardCategoryOptions, storyboardItemCategory, storyboardPointerDropTarget, toggleStoryboardBatchSelection } from '../pages/agent-storyboard-state';
import { audioOperationNeedsModel } from '../pages/agent-canvas-audio-jobs';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { generationSummaryText } from '../pages/agent-generation-state';
import { AgentBatchModelPicker } from './AgentBatchModelPicker';
import { AgentModelPicker } from './AgentModelPicker';
import { AgentNodeParameterPopover } from './AgentNodeParameterPopover';
import { AgentNodeSettingsPanel } from './AgentNodeSettingsPanel';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';

const statusLabels = { draft: '草稿', pending: '等待中', running: '生成中', done: '已完成', failed: '失败' };

function KindIcon({ kind, size = 14 }) {
  if (kind === 'image') return <ImageIcon size={size} />;
  if (kind === 'video') return <Film size={size} />;
  if (kind === 'audio') return <Music size={size} />;
  return <FileText size={size} />;
}

function StoryboardThumbnail({ item, large = false }) {
  if (item.kind === 'image' && item.url) return <AuthenticatedAgentImage source={item.url} alt={item.title} className={large ? 'is-large' : ''} />;
  if (item.kind === 'video' && item.url) return <AuthenticatedAgentVideo source={item.url} aria-label={item.title} className={large ? 'is-large' : ''} preload="metadata" muted controls={large} />;
  return <span className={`agent-storyboard-placeholder is-${item.kind}${large ? ' is-large' : ''}`}><KindIcon kind={item.kind} size={large ? 34 : 17} /></span>;
}

export function AgentStoryboardWorkspace({
  taskId = '', nodes = [], assets = [], shots = [], generations = [], generationSummary = null, loading = false, error = '', shotsLoading = false, shotsError = '', conversationAssetIds = [], storyboardState = {}, onStoryboardStateChange,
  onOpenNodeInWorkflow, onSelectionChange, onNodeConfigChange, onShotUpdated, onCreateShotBatch, onAddToConversation, onAddToWorkflow, onAssetsCreated
}) {
  const [query, setQuery] = useState('');
  const [collapsedKinds, setCollapsedKinds] = useState([]);
  const [batchMode, setBatchMode] = useState(false);
  const [batchSelectedKeys, setBatchSelectedKeys] = useState([]);
  const [assetUploading, setAssetUploading] = useState(false);
  const [assetUploadError, setAssetUploadError] = useState('');
  const [modelsByKind, setModelsByKind] = useState({ text: [], image: [], video: [], audio: [] });
  const [modelsLoading, setModelsLoading] = useState(true);
  const [modelLoadError, setModelLoadError] = useState('');
  const [dragTargetKey, setDragTargetKey] = useState('');
  const [draggingKey, setDraggingKey] = useState('');
  const [conversationDraggingKey, setConversationDraggingKey] = useState('');
  const [gridResizeAxis, setGridResizeAxis] = useState('');
  const [sectionDraggingKind, setSectionDraggingKind] = useState('');
  const [sectionDragTargetKind, setSectionDragTargetKind] = useState('');
  const [shotDraft, setShotDraft] = useState(null);
  const [shotSaving, setShotSaving] = useState(false);
  const [shotActionError, setShotActionError] = useState('');
  const [shotBatchKind, setShotBatchKind] = useState('image');
  const [shotBatchModels, setShotBatchModels] = useState({
    image: { provider: 'personal_api', modelId: '' },
    video: { provider: 'personal_api', modelId: '' },
    audio: { provider: 'personal_api', modelId: '' }
  });
  const [shotBatchBusy, setShotBatchBusy] = useState(false);
  const storyboard = useMemo(() => normalizeStoryboardState(storyboardState), [storyboardState]);
  const assetInputRef = useRef(null);
  const itemDragRef = useRef(null);
  const gridResizeRef = useRef(null);
  const sectionDragRef = useRef(null);
  const sectionClickSuppressionRef = useRef(false);
  const selectedKey = storyboard.selectedKey;
  const conversationAssetIDSet = useMemo(() => new Set(conversationAssetIds.map(value => String(value))), [conversationAssetIds]);
  const library = useMemo(() => buildStoryboardLibrary(nodes, assets, shots), [nodes, assets, shots]);
  const visibleLibrary = useMemo(() => filterStoryboardLibrary(library, storyboard, query), [library, query, storyboard]);
  const shotSection = useMemo(() => visibleLibrary.find(section => section.kind === 'shot') || { kind: 'shot', label: '镜头', items: [] }, [visibleLibrary]);
  const materialLibrary = useMemo(() => visibleLibrary.filter(section => section.kind !== 'shot'), [visibleLibrary]);
  const orderedMaterialLibrary = useMemo(() => {
    const sections = new Map(materialLibrary.map(section => [section.kind, section]));
    return storyboard.layout.sectionOrder.map(kind => sections.get(kind)).filter(Boolean);
  }, [materialLibrary, storyboard.layout.sectionOrder]);
  const visibleItems = useMemo(() => materialLibrary.flatMap(section => section.items || []), [materialLibrary]);
  const resultGroups = useMemo(() => groupResultLibrary({ shots, nodes, generations, assets }), [assets, generations, nodes, shots]);
  const selectedBatchItems = useMemo(() => {
    const selected = new Set(batchSelectedKeys);
    return visibleItems.filter(item => selected.has(item.key));
  }, [batchSelectedKeys, visibleItems]);
  const selectedBatchKind = selectedBatchItems.length && selectedBatchItems.every(item => item.kind === selectedBatchItems[0].kind)
    ? selectedBatchItems[0].kind
    : '';
  const selectedBatchAssets = useMemo(
    () => storyboardBatchAssets(visibleItems, batchSelectedKeys),
    [batchSelectedKeys, visibleItems]
  );
  const selected = useMemo(() => {
    const item = findStoryboardSelection(library, selectedKey);
    return item ? { ...item, category: storyboardItemCategory(item, storyboard) } : null;
  }, [library, selectedKey, storyboard]);
  const selectedShotVersions = useMemo(() => selected?.source === 'shot'
    ? resultGroups.find(group => group.id === selected.id)?.items || []
    : [], [resultGroups, selected]);

  useEffect(() => {
    if (selected?.source !== 'shot') {
      setShotDraft(null);
      setShotActionError('');
      return;
    }
    setShotDraft({ ...selected.shot });
    setShotActionError('');
  }, [selected?.id, selected?.shot?.revision]);

  useEffect(() => {
    const next = findStoryboardSelection(library, selectedKey);
    const nextKey = storyboardSelectionKey(next);
    if (nextKey !== selectedKey) onStoryboardStateChange?.({ selectedKey: nextKey });
  }, [library, onStoryboardStateChange, selectedKey]);

  useEffect(() => {
    onSelectionChange?.(selected?.nodeId || '');
  }, [onSelectionChange, selected?.nodeId]);

  useEffect(() => {
    let active = true;
    setModelsLoading(true);
    setModelLoadError('');
    Promise.all(['text', 'image', 'video', 'audio'].map(kind => listAvailableModels(kind)))
      .then(([text, image, video, audio]) => {
        if (active) setModelsByKind({ text, image, video, audio });
      })
      .catch(loadError => {
        if (active) setModelLoadError(loadError.message || '读取模型配置失败');
      })
      .finally(() => { if (active) setModelsLoading(false); });
    return () => { active = false; };
  }, []);

  function selectItem(item) {
    if (batchMode) {
      setBatchSelectedKeys(current => toggleStoryboardBatchSelection(current, item.key));
      return;
    }
    onStoryboardStateChange?.({ selectedKey: storyboardSelectionKey(item) });
  }

  function toggleBatchMode() {
    setBatchMode(current => {
      if (current) setBatchSelectedKeys([]);
      return !current;
    });
  }

  function toggleAllVisible() {
    const visibleKeys = visibleItems.map(item => item.key);
    const selected = new Set(batchSelectedKeys);
    setBatchSelectedKeys(visibleKeys.length && visibleKeys.every(key => selected.has(key)) ? [] : visibleKeys);
  }

  function batchAddToConversation() {
    for (const asset of selectedBatchAssets) onAddToConversation?.(asset);
    if (selectedBatchAssets.length) message.success(`已将 ${selectedBatchAssets.length} 项素材加入对话`);
  }

  function batchChangeCategory(category) {
    const categories = storyboardBatchCategoryPatch(visibleItems, batchSelectedKeys, category);
    if (categories) onStoryboardStateChange?.({ categories });
  }

  async function uploadProjectAsset(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || !taskId) return;
    setAssetUploading(true);
    setAssetUploadError('');
    try {
      const result = await uploadAgentAsset(taskId, file);
      if (!result?.asset?.id) throw new Error('上传结果没有返回素材');
      onAssetsCreated?.([result.asset]);
      onStoryboardStateChange?.({ selectedKey: `asset:${result.asset.id}` });
      message.success('素材已加入当前项目');
    } catch (uploadError) {
      const text = uploadError?.message || '上传素材失败';
      setAssetUploadError(text);
      message.error(text);
    } finally {
      setAssetUploading(false);
    }
  }

  async function saveSelectedShot() {
    if (!taskId || selected?.source !== 'shot' || !shotDraft) return;
    setShotSaving(true);
    setShotActionError('');
    try {
      const result = await updateAgentScriptShot(taskId, selected.id, {
        scriptNodeId: selected.shot.scriptNodeId,
        revision: selected.shot.revision,
        scene: shotDraft.scene || '',
        characters: shotDraft.characters || '',
        dialogue: shotDraft.dialogue || '',
        narration: shotDraft.narration || '',
        imagePrompt: shotDraft.imagePrompt || '',
        videoPrompt: shotDraft.videoPrompt || '',
        ttsText: shotDraft.ttsText || ''
      });
      if (!result?.shot?.id) throw new Error('镜头保存结果无效');
      onShotUpdated?.(result.shot);
      setShotDraft({ ...result.shot });
      message.success('镜头已保存');
    } catch (saveError) {
      setShotActionError(saveError?.message || '保存镜头失败');
    } finally {
      setShotSaving(false);
    }
  }

  async function createSelectedShotBatch() {
    if (!taskId || selected?.source !== 'shot') return;
    const selection = shotBatchModels[shotBatchKind] || {};
    const modelId = selection.modelId;
    const payload = buildShotBatchPayload({ scriptNodeId: selected.shot.scriptNodeId, selectedIDs: [selected.id], kind: shotBatchKind, provider: selection.provider, modelId });
    if (!modelId) {
      setShotActionError('请先选择当前类型的模型');
      return;
    }
    setShotBatchBusy(true);
    setShotActionError('');
    try {
      const result = await onCreateShotBatch?.(payload);
      if (result?.ignored) return;
      const skipped = Array.isArray(result?.skipped) ? result.skipped : [];
      if (skipped.length) setShotActionError(skipped.map(item => item.reason).filter(Boolean).join('；') || '部分镜头未创建');
      else message.success('镜头生成任务已提交');
    } catch (batchError) {
      setShotActionError(batchError?.message || '提交镜头生成失败');
    } finally {
      setShotBatchBusy(false);
    }
  }

  function toggleSection(kind) {
    setCollapsedKinds(current => current.includes(kind) ? current.filter(value => value !== kind) : [...current, kind]);
  }

  function moveSelected(direction) {
    if (!selected) return;
    const section = visibleLibrary.find(item => item.kind === selected.kind);
    const next = moveStoryboardMaterial(storyboard, selected.key, direction, section?.items.map(item => item.key) || []);
    onStoryboardStateChange?.({ order: next.order });
  }

  function startItemDrag(event, item, sectionItems) {
    if (batchMode || event.button !== 0) return;
    itemDragRef.current = {
      pointerId: event.pointerId,
      key: item.key,
      kind: item.kind,
      sectionKeys: sectionItems.map(entry => entry.key),
      startX: event.clientX,
      startY: event.clientY,
      targetKey: '',
      moved: false
    };
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.stopPropagation();
  }

  function moveItemDrag(event) {
    const source = itemDragRef.current;
    if (!source || source.pointerId !== event.pointerId) return;
    if (!source.moved && Math.hypot(event.clientX - source.startX, event.clientY - source.startY) < 5) return;
    source.moved = true;
    const candidates = document.elementsFromPoint(event.clientX, event.clientY).map(element => {
      const item = element.closest?.('[data-storyboard-key]');
      return { key: item?.dataset?.storyboardKey || '', kind: item?.dataset?.storyboardKind || '' };
    });
    source.targetKey = storyboardPointerDropTarget(source.key, source.kind, candidates);
    setDraggingKey(source.key);
    setDragTargetKey(source.targetKey);
    event.preventDefault();
  }

  function finishItemDrag(event, cancelled = false) {
    const source = itemDragRef.current;
    if (!source || source.pointerId !== event.pointerId) return;
    itemDragRef.current = null;
    event.currentTarget.releasePointerCapture?.(event.pointerId);
    setDraggingKey('');
    setDragTargetKey('');
    if (cancelled || !source.moved || !source.targetKey) return;
    const next = reorderStoryboardMaterial(storyboard, source.key, source.targetKey, source.sectionKeys);
    onStoryboardStateChange?.({ order: next.order });
  }

  function startGridResize(event, eventType = 'pointer') {
    const axis = event.currentTarget.dataset.axis;
    const split = event.currentTarget.dataset.split;
    const bounds = event.currentTarget.parentElement?.getBoundingClientRect();
    if (gridResizeRef.current || event.button !== 0 || !bounds || (axis !== 'x' && axis !== 'y') || !Object.hasOwn(STORYBOARD_SPLIT_DEFAULTS, split)) return;
    gridResizeRef.current = { axis, split, eventType, pointerId: eventType === 'pointer' ? event.pointerId : null, bounds };
    setGridResizeAxis(split);
    if (eventType === 'pointer') event.currentTarget.setPointerCapture?.(event.pointerId);
    event.preventDefault();
  }

  function moveGridResize(event) {
    const resize = gridResizeRef.current;
    if (!resize || (resize.pointerId !== null && resize.pointerId !== event.pointerId)) return;
    const ratio = storyboardSplitRatioFromPointer(
      resize.axis === 'x' ? event.clientX : event.clientY,
      resize.bounds,
      resize.axis,
      18,
      STORYBOARD_SPLIT_DEFAULTS[resize.split]
    );
    onStoryboardStateChange?.({ layout: { [resize.split]: ratio } });
    event.preventDefault();
  }

  function resetGridSplit(event) {
    const split = event.currentTarget.dataset.split;
    if (!Object.hasOwn(STORYBOARD_SPLIT_DEFAULTS, split)) return;
    onStoryboardStateChange?.({ layout: { [split]: STORYBOARD_SPLIT_DEFAULTS[split] } });
  }

  function finishGridResize(event) {
    const resize = gridResizeRef.current;
    if (!resize || (resize.pointerId !== null && resize.pointerId !== event.pointerId)) return;
    gridResizeRef.current = null;
    setGridResizeAxis('');
    if (resize.pointerId !== null) event.currentTarget?.releasePointerCapture?.(event.pointerId);
  }

  useEffect(() => {
    if (!gridResizeAxis) return undefined;
    const resize = gridResizeRef.current;
    if (!resize) return undefined;
    const moveEvent = resize.eventType === 'pointer' ? 'pointermove' : 'mousemove';
    const endEvent = resize.eventType === 'pointer' ? 'pointerup' : 'mouseup';
    window.addEventListener(moveEvent, moveGridResize);
    window.addEventListener(endEvent, finishGridResize);
    return () => {
      window.removeEventListener(moveEvent, moveGridResize);
      window.removeEventListener(endEvent, finishGridResize);
    };
  }, [gridResizeAxis]);

  function startSectionDrag(event, kind) {
    if (event.button !== 0) return;
    sectionDragRef.current = {
      pointerId: event.pointerId,
      sourceKind: kind,
      targetKind: '',
      startX: event.clientX,
      startY: event.clientY,
      moved: false
    };
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function moveSectionDrag(event) {
    const drag = sectionDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    if (!drag.moved && Math.hypot(event.clientX - drag.startX, event.clientY - drag.startY) < 5) return;
    drag.moved = true;
    const targetKind = document.elementsFromPoint(event.clientX, event.clientY)
      .map(element => element.closest?.('[data-section-kind]')?.dataset?.sectionKind || '')
      .find(kind => kind && kind !== drag.sourceKind) || '';
    drag.targetKind = targetKind;
    setSectionDraggingKind(drag.sourceKind);
    setSectionDragTargetKind(targetKind);
    event.preventDefault();
  }

  function finishSectionDrag(event, cancelled = false) {
    const drag = sectionDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    sectionDragRef.current = null;
    event.currentTarget.releasePointerCapture?.(event.pointerId);
    setSectionDraggingKind('');
    setSectionDragTargetKind('');
    if (!drag.moved) return;
    sectionClickSuppressionRef.current = true;
    if (cancelled || !drag.targetKind) return;
    const sectionOrder = [...storyboard.layout.sectionOrder];
    const sourceIndex = sectionOrder.indexOf(drag.sourceKind);
    const targetIndex = sectionOrder.indexOf(drag.targetKind);
    if (sourceIndex < 0 || targetIndex < 0) return;
    [sectionOrder[sourceIndex], sectionOrder[targetIndex]] = [sectionOrder[targetIndex], sectionOrder[sourceIndex]];
    onStoryboardStateChange?.({ layout: { sectionOrder } });
  }

  function handleSectionTitleClick(kind) {
    if (sectionClickSuppressionRef.current) {
      sectionClickSuppressionRef.current = false;
      return;
    }
    toggleSection(kind);
  }

  function startConversationAssetDrag(event, item) {
    if (batchMode || !writeStoryboardAssetDrag(event.dataTransfer, item.asset)) return;
    setConversationDraggingKey(item.key);
  }

  const total = library.reduce((sum, section) => sum + section.items.length, 0);
  const showShotSection = shotsLoading || Boolean(shotsError) || shotSection.items.length > 0;
  const selectedModels = selected ? (modelsByKind[selected.kind] || []) : [];
  const selectedModel = selected ? reconcileNodeModel(selected.config, selectedModels, selected.kind).model : null;
  const selectedAudioNeedsModel = selected?.kind === 'audio' && audioOperationNeedsModel(selected.config?.parameters?.operation || selected.config?.operation);
  const updateSelectedConfig = patch => {
    if (selected?.nodeId) onNodeConfigChange?.(selected.nodeId, patch);
  };
  function renderMaterialSection(section) {
    if (!section) return null;
    const collapsed = collapsedKinds.includes(section.kind);
    const selectedKeys = new Set(batchSelectedKeys);
    return <section
      className={`agent-storyboard-section is-${section.kind}${collapsed ? ' is-collapsed' : ''}${sectionDraggingKind === section.kind ? ' is-section-dragging' : ''}${sectionDragTargetKind === section.kind ? ' is-section-drop-target' : ''}`}
      key={section.kind}
      data-section-kind={section.kind}
    >
      <button
        type="button"
        className="agent-storyboard-section__title"
        aria-expanded={!collapsed}
        aria-label={`拖动调整位置，点击${collapsed ? '展开' : '收起'}${section.label}`}
        onPointerDown={event => startSectionDrag(event, section.kind)}
        onPointerMove={moveSectionDrag}
        onPointerUp={event => finishSectionDrag(event)}
        onPointerCancel={event => finishSectionDrag(event, true)}
        onClick={() => handleSectionTitleClick(section.kind)}
      >
        <span><GripVertical size={12} /><span className="agent-storyboard-section__folder" aria-hidden="true">{collapsed ? <Folder size={13} /> : <FolderOpen size={13} />}</span><KindIcon kind={section.kind} /><strong>{section.label}</strong><i>{section.items.length}</i></span>{collapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
      </button>
      {!collapsed ? <>
        <Segmented
          className="agent-storyboard-section__filters"
          aria-label={`${section.label}分类`}
          size="small"
          value={storyboard.filters[section.kind]}
          options={storyboardCategoryOptions(section.kind)}
          onChange={category => onStoryboardStateChange?.({ filters: { [section.kind]: category } })}
        />
        <div className="agent-storyboard-section__items">
          {section.items.map(item => {
            const conversationAssetID = String(item.asset?.id || item.asset?.assetId || '');
            const addedToConversation = conversationAssetIDSet.has(conversationAssetID);
            const batchSelected = selectedKeys.has(item.key);
            return <article
              key={item.key}
              data-storyboard-key={item.key}
              data-storyboard-kind={item.kind}
              title={conversationAssetID ? '拖到右侧对话框即可引用素材' : undefined}
              className={`agent-storyboard-item${selectedKey === item.key && !batchMode ? ' is-selected' : ''}${batchSelected ? ' is-batch-selected' : ''}${draggingKey === item.key ? ' is-dragging' : ''}${conversationDraggingKey === item.key ? ' is-conversation-dragging' : ''}${dragTargetKey === item.key ? ' is-drop-target' : ''}`}
              draggable={!batchMode && Boolean(conversationAssetID)}
              onDragStart={event => startConversationAssetDrag(event, item)}
              onDragEnd={() => setConversationDraggingKey('')}
            >
              {!batchMode ? <button
                type="button"
                className="agent-storyboard-item__drag"
                aria-label={`拖动排序 ${item.title}`}
                onClick={event => event.stopPropagation()}
                onPointerDown={event => startItemDrag(event, item, section.items)}
                onPointerMove={moveItemDrag}
                onPointerUp={event => finishItemDrag(event)}
                onPointerCancel={event => finishItemDrag(event, true)}
                onDragStart={event => event.preventDefault()}
              ><GripVertical size={13} /></button> : null}
              <button type="button" className="agent-storyboard-item__select" aria-pressed={batchMode ? batchSelected : selectedKey === item.key} onClick={() => selectItem(item)}>
                <StoryboardThumbnail item={item} />
                <span><strong>{item.title}</strong><small>{storyboardItemMetaLabel(item) || item.summary || statusLabels[item.status] || '暂无内容'}</small></span>
              </button>
              {batchMode ? <span className="agent-storyboard-item__check" aria-hidden="true">{batchSelected ? <Check size={13} /> : null}</span> : null}
              {!batchMode && conversationAssetID ? <Tooltip title={addedToConversation ? '已加入对话' : '加入对话'}><button
                type="button"
                className={`agent-storyboard-item__conversation${addedToConversation ? ' is-added' : ''}`}
                aria-label={`${addedToConversation ? '已加入对话' : '加入对话'} ${item.title}`}
                onClick={() => onAddToConversation?.(item.asset)}
              ><MessageSquarePlus size={13} /></button></Tooltip> : null}
            </article>;
          })}
          {!section.items.length ? <p>当前分类暂无{section.label}</p> : null}
        </div>
      </> : null}
    </section>;
  }
  return <section
    className={`agent-storyboard-workspace${selected ? ' has-selection' : ''}`}
    aria-label="故事板工作区"
  >
    <aside className="agent-storyboard-library" aria-label="故事板项目素材">
      <header>
        <div className="agent-storyboard-library__heading"><div><strong>项目素材</strong><span>{total}</span></div><div className="agent-storyboard-library__actions">
          <Button size="small" icon={<Plus size={13} />} loading={assetUploading} disabled={!taskId} onClick={() => assetInputRef.current?.click()}>添加素材</Button>
          <Button size="small" type={batchMode ? 'primary' : 'default'} icon={<ListChecks size={13} />} onClick={toggleBatchMode}>批量操作</Button>
        </div></div>
        {generationSummary ? <span className="agent-storyboard-generation-summary" role="status" aria-live="polite">{generationSummaryText(generationSummary)}</span> : null}
        <Input allowClear prefix={<Search size={14} />} placeholder="搜索名称" value={query} onChange={event => setQuery(event.target.value)} />
        <input ref={assetInputRef} className="agent-storyboard-upload-input" type="file" accept="image/*,video/*,audio/*" onChange={uploadProjectAsset} />
        {assetUploadError ? <span className="agent-storyboard-upload-error">{assetUploadError}</span> : null}
      </header>
      {batchMode ? <div className="agent-storyboard-batchbar" aria-label="故事板批量操作">
        <strong>已选 {batchSelectedKeys.length}</strong>
        <Button size="small" onClick={toggleAllVisible}>全选当前结果</Button>
        <Select
          size="small"
          aria-label="批量素材分类"
          placeholder={selectedBatchKind ? '批量分类' : '请只选同类型'}
          disabled={!selectedBatchKind}
          options={selectedBatchKind ? storyboardCategoryOptions(selectedBatchKind, { includeAll: false }) : []}
          onChange={batchChangeCategory}
        />
        <Button size="small" icon={<MessageSquarePlus size={13} />} disabled={!selectedBatchAssets.length} onClick={batchAddToConversation}>批量加入对话</Button>
        <Tooltip title="完成批量操作"><Button size="small" type="text" aria-label="完成批量操作" icon={<X size={14} />} onClick={toggleBatchMode} /></Tooltip>
      </div> : null}
      {showShotSection ? <section className="agent-storyboard-shots" aria-label="故事板镜头">
        <header><span><FileText size={14} /><strong>镜头</strong><i>{shotSection.items.length}</i></span></header>
        {shotsLoading ? <span className="agent-storyboard-shots__state"><Spin size="small" />正在读取镜头</span> : null}
        {shotsError ? <span className="agent-storyboard-shots__state is-error">{shotsError}</span> : null}
        <div className="agent-storyboard-shots__list">
          {shotSection.items.map(item => <button type="button" className={selectedKey === item.key ? 'is-selected' : ''} key={item.key} disabled={batchMode} onClick={() => selectItem(item)}>
            <strong>{item.title}</strong><small>{item.summary || '待补充镜头内容'}</small>
          </button>)}
        </div>
      </section> : null}
      <div
        className={`agent-storyboard-library__body${gridResizeAxis ? ' is-resizing-grid' : ''}`}
        style={{
          '--storyboard-column-split': `${storyboard.layout.column}%`,
          '--storyboard-left-row-split': `${storyboard.layout.leftRow}%`,
          '--storyboard-right-row-split': `${storyboard.layout.rightRow}%`
        }}
      >
        {loading ? <div className="agent-storyboard-state"><Spin size="small" />正在读取项目素材</div> : null}
        {error ? <div className="agent-storyboard-state is-error">{error}</div> : null}
        {!loading && !error && !total ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前项目还没有内容" /> : null}
        <div className="agent-storyboard-material-column agent-storyboard-material-column--left">
          {renderMaterialSection(orderedMaterialLibrary[0])}
          {renderMaterialSection(orderedMaterialLibrary[1])}
          <button
            type="button"
            className="agent-storyboard-grid-splitter agent-storyboard-grid-splitter--left-row"
            data-axis="y"
            data-split="leftRow"
            aria-label="拖动调整左列高度，双击恢复默认"
            onPointerDown={event => startGridResize(event, 'pointer')}
            onMouseDown={event => startGridResize(event, 'mouse')}
            onPointerCancel={finishGridResize}
            onDoubleClick={resetGridSplit}
          />
        </div>
        <div className="agent-storyboard-material-column agent-storyboard-material-column--right">
          {renderMaterialSection(orderedMaterialLibrary[2])}
          {renderMaterialSection(orderedMaterialLibrary[3])}
          <button
            type="button"
            className="agent-storyboard-grid-splitter agent-storyboard-grid-splitter--right-row"
            data-axis="y"
            data-split="rightRow"
            aria-label="拖动调整右列高度，双击恢复默认"
            onPointerDown={event => startGridResize(event, 'pointer')}
            onMouseDown={event => startGridResize(event, 'mouse')}
            onPointerCancel={finishGridResize}
            onDoubleClick={resetGridSplit}
          />
        </div>
        <button
          type="button"
          className="agent-storyboard-grid-splitter agent-storyboard-grid-splitter--column"
          data-axis="x"
          data-split="column"
          aria-label="拖动调整素材列宽，双击恢复默认"
          onPointerDown={event => startGridResize(event, 'pointer')}
          onMouseDown={event => startGridResize(event, 'mouse')}
          onPointerCancel={finishGridResize}
          onDoubleClick={resetGridSplit}
        />
      </div>
    </aside>

    {selected ? <aside className="agent-storyboard-inspector" aria-label="故事板详情">
      <>
        <header>
          <div><KindIcon kind={selected.kind} /><span><strong>{selected.title}</strong><small>{selected.source === 'shot' ? '结构化镜头' : selected.source === 'asset' ? '项目素材' : '画布节点'}</small></span></div>
          <div className="agent-storyboard-inspector__actions">
            {selected.source !== 'shot' ? <><Tooltip title="上移素材"><Button type="text" aria-label="上移素材" icon={<ArrowUp size={15} />} onClick={() => moveSelected(-1)} /></Tooltip>
            <Tooltip title="下移素材"><Button type="text" aria-label="下移素材" icon={<ArrowDown size={15} />} onClick={() => moveSelected(1)} /></Tooltip></> : null}
            {selected.source === 'shot' ? <Button type="text" icon={<LocateFixed size={15} />} onClick={() => onOpenNodeInWorkflow?.(locateShotNode(nodes, selected.id, { ...selected.shot, nodeId: selected.nodeId })?.id)}>定位来源</Button> : selected.nodeId ? <Button type="text" icon={<LocateFixed size={15} />} onClick={() => onOpenNodeInWorkflow?.(selected.nodeId)}>在工作流中打开</Button> : null}
            <Tooltip title="关闭详情"><Button type="text" aria-label="关闭素材详情" icon={<X size={15} />} onClick={() => onStoryboardStateChange?.({ selectedKey: '' })} /></Tooltip>
          </div>
        </header>
        <div className="agent-storyboard-inspector__body">
          {selected.source === 'shot' ? <>
            <section className="agent-storyboard-shot-editor" aria-label="编辑镜头">
              <div className="agent-storyboard-shot-editor__heading"><strong>编辑镜头</strong><Button size="small" type="primary" icon={<Save size={13} />} loading={shotSaving} onClick={saveSelectedShot}>保存</Button></div>
              <Input value={shotDraft?.scene || ''} placeholder="场景" aria-label="镜头场景" onChange={event => setShotDraft(current => ({ ...current, scene: event.target.value }))} />
              <Input value={shotDraft?.characters || ''} placeholder="人物" aria-label="镜头人物" onChange={event => setShotDraft(current => ({ ...current, characters: event.target.value }))} />
              <Input.TextArea value={shotDraft?.dialogue || ''} placeholder="对白" aria-label="镜头对白" autoSize={{ minRows: 2, maxRows: 5 }} onChange={event => setShotDraft(current => ({ ...current, dialogue: event.target.value }))} />
              <Input.TextArea value={shotDraft?.narration || ''} placeholder="旁白" aria-label="镜头旁白" autoSize={{ minRows: 2, maxRows: 5 }} onChange={event => setShotDraft(current => ({ ...current, narration: event.target.value }))} />
              <Input.TextArea value={shotDraft?.imagePrompt || ''} placeholder="图片生成要求" aria-label="镜头图片生成要求" autoSize={{ minRows: 2, maxRows: 5 }} onChange={event => setShotDraft(current => ({ ...current, imagePrompt: event.target.value }))} />
              <Input.TextArea value={shotDraft?.videoPrompt || ''} placeholder="视频生成要求" aria-label="镜头视频生成要求" autoSize={{ minRows: 2, maxRows: 5 }} onChange={event => setShotDraft(current => ({ ...current, videoPrompt: event.target.value }))} />
              <Input.TextArea value={shotDraft?.ttsText || ''} placeholder="配音文字" aria-label="镜头配音文字" autoSize={{ minRows: 2, maxRows: 5 }} onChange={event => setShotDraft(current => ({ ...current, ttsText: event.target.value }))} />
            </section>
            <section className="agent-storyboard-shot-batch" aria-label="选中镜头批量生成">
              <strong>生成节点</strong>
              <Segmented size="small" value={shotBatchKind} options={[{ value: 'image', label: '图片' }, { value: 'video', label: '视频' }, { value: 'audio', label: '音频' }]} onChange={setShotBatchKind} />
              <AgentBatchModelPicker
                kind={shotBatchKind}
                config={shotBatchModels[shotBatchKind]}
                models={modelsByKind[shotBatchKind] || []}
                onChange={config => setShotBatchModels(current => ({ ...current, [shotBatchKind]: config }))}
              />
              <Button type="primary" icon={<Sparkles size={14} />} loading={shotBatchBusy} onClick={createSelectedShotBatch}>生成当前镜头</Button>
              {shotActionError ? <span className="agent-storyboard-inspector__hint is-error">{shotActionError}</span> : null}
            </section>
            <section className="agent-storyboard-shot-versions" aria-label="镜头版本">
              <div className="agent-storyboard-shot-editor__heading"><strong>镜头版本</strong><span>{selectedShotVersions.length}</span></div>
              {!selectedShotVersions.length ? <p>当前镜头还没有生成结果</p> : selectedShotVersions.map(item => <button type="button" className={item.active ? 'is-active' : ''} key={item.id} onClick={() => onOpenNodeInWorkflow?.(item.nodeId)}>
                <span><strong>{item.label}</strong><small>{item.title}</small></span><Tag>{item.active ? '当前' : '历史'}</Tag>
              </button>)}
            </section>
          </> : <>
          <div className="agent-storyboard-inspector__preview"><StoryboardThumbnail item={selected} large /></div>
          <div className="agent-storyboard-inspector__meta">
            <div><span>类型</span><strong>{library.find(section => section.kind === selected.kind)?.label || selected.kind}</strong></div>
            <div><span>状态</span><Tag>{statusLabels[selected.status] || selected.status || '已入库'}</Tag></div>
            <label><span>素材分类</span><Select aria-label="素材分类" value={selected.category} options={storyboardCategoryOptions(selected.kind, { includeAll: false })} onChange={category => onStoryboardStateChange?.({ categories: { [selected.key]: category } })} /></label>
            {selected.nodeId ? <div><span>节点</span><strong>{selected.nodeId}</strong></div> : null}
            {selected.asset?.byteSize ? <div><span>大小</span><strong>{Math.max(1, Math.round(selected.asset.byteSize / 1024))} KB</strong></div> : null}
          </div>
          {selected.summary ? <section className="agent-storyboard-inspector__section"><strong>内容</strong><p>{selected.summary}</p></section> : null}
          {selected.nodeId ? <section className="agent-storyboard-inspector__editor" aria-label="故事板节点编辑">
            <strong>节点编辑</strong>
            <Input.TextArea
              aria-label={`${selected.kind === 'text' ? '文字' : selected.kind === 'image' ? '图片' : selected.kind === 'video' ? '视频' : '音频'}生成要求`}
              autoSize={{ minRows: 3, maxRows: 7 }}
              value={selected.config?.prompt || ''}
              placeholder="输入生成要求"
              onChange={event => updateSelectedConfig({ prompt: event.target.value })}
            />
            <div className="agent-storyboard-inspector__toolbar">
              {selected.kind !== 'audio' || selectedAudioNeedsModel ? <AgentModelPicker kind={selected.kind} config={selected.config} models={selectedModels} disabled={modelsLoading} onChange={updateSelectedConfig} /> : <span className="agent-node-composer__local-processor">本地音频处理</span>}
              <AgentNodeParameterPopover kind={selected.kind} config={selected.config} capabilities={selectedModel?.capabilities} disabled={modelsLoading} onChange={updateSelectedConfig} />
              <AgentNodeSettingsPanel kind={selected.kind} config={selected.config} models={selectedModels} disabled={modelsLoading} onChange={updateSelectedConfig} />
            </div>
            {modelsLoading ? <span className="agent-storyboard-inspector__hint"><Spin size="small" />正在读取可用模型</span> : null}
            {modelLoadError ? <span className="agent-storyboard-inspector__hint is-error">{modelLoadError}</span> : null}
            <span className="agent-storyboard-inspector__hint">这里修改的就是工作流原节点，保存后两边同步。</span>
          </section> : <section className="agent-storyboard-inspector__section"><strong>项目素材</strong><span>此素材尚未加入工作流。</span><Button type="primary" size="small" icon={<Plus size={14} />} onClick={() => onAddToWorkflow?.(selected.asset)}>加入工作流</Button></section>}
          </>}
        </div>
      </>
    </aside> : null}
  </section>;
}
