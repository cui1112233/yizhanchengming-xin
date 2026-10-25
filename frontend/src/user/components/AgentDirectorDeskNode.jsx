import { Camera, ImagePlus, Move, Send, UserRound, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { createDirectorPromptDraft, directorDeskConfig, directorDeskSceneStyle, discardDirectorPromptDraft, editDirectorPromptDraft, mergeDirectorDeskConfig, moveDirectorDeskActor, submitDirectorPromptDraft, syncDirectorPromptDraft } from '../pages/agent-director-desk';

function imageForAsset(surface, assetID) {
  return [...(surface?.querySelectorAll?.('img[data-director-asset-id]') || [])]
    .find(image => image.dataset.directorAssetId === String(assetID || '')) || null;
}

function drawCover(context, image, width, height) {
  const sourceWidth = Number(image?.naturalWidth) || 1;
  const sourceHeight = Number(image?.naturalHeight) || 1;
  const scale = Math.max(width / sourceWidth, height / sourceHeight);
  const drawWidth = sourceWidth * scale;
  const drawHeight = sourceHeight * scale;
  context.drawImage(image, (width - drawWidth) / 2, (height - drawHeight) / 2, drawWidth, drawHeight);
}

export async function captureDirectorDeskScreenshot(surface, config) {
  const desk = directorDeskConfig(config);
  if (!surface || !desk.sceneAsset?.id) throw new Error('请先选择场景图');
  const scene = imageForAsset(surface, desk.sceneAsset.id);
  if (!scene) throw new Error('场景图尚未加载完成');
  await scene.decode?.().catch(() => undefined);
  const canvas = document.createElement('canvas');
  canvas.width = 1280;
  canvas.height = 720;
  const context = canvas.getContext('2d');
  if (!context) throw new Error('当前浏览器无法生成预演图');
  drawCover(context, scene, canvas.width, canvas.height);
  context.save();
  context.translate(canvas.width / 2, canvas.height / 2);
  context.scale(desk.camera.zoom, desk.camera.zoom);
  context.translate(canvas.width * desk.camera.x / 100, canvas.height * desk.camera.y / 100);
  context.translate(-canvas.width / 2, -canvas.height / 2);
  for (const actor of desk.actors) {
    const image = imageForAsset(surface, actor.assetId);
    if (!image) continue;
    await image.decode?.().catch(() => undefined);
    const height = 280 * actor.scale;
    const width = height * ((Number(image.naturalWidth) || 1) / Math.max(Number(image.naturalHeight) || 1, 1));
    context.save();
    context.translate(canvas.width * actor.x / 100, canvas.height * actor.y / 100);
    context.rotate(actor.rotation * Math.PI / 180);
    context.drawImage(image, -width / 2, -height, width, height);
    context.restore();
  }
  context.restore();
  const blob = await new Promise(resolve => canvas.toBlob(resolve, 'image/png'));
  if (!blob) throw new Error('预演图生成失败');
  return new File([blob], `director-desk-${Date.now()}.png`, { type: 'image/png' });
}

export function AgentDirectorDeskWorkspace({ node, onConfigChange, onPickReferenceAsset, onRemoveReference, onSaveScreenshot }) {
  const desk = directorDeskConfig(node?.data?.config?.directorDesk);
  const references = Array.isArray(node?.data?.config?.referenceAssets) ? node.data.config.referenceAssets.filter(asset => asset?.kind === 'image') : [];
  const referenceByID = useMemo(() => new Map(references.map(asset => [asset.id, asset])), [references]);
  const surface = useRef(null);
  const [savingScreenshot, setSavingScreenshot] = useState(false);
  const update = next => onConfigChange?.(node.id, { directorDesk: mergeDirectorDeskConfig(node?.data?.config?.directorDesk, next) });
  const setScene = asset => update({ ...desk, sceneAsset: asset ? { id: asset.id, url: asset.url, name: asset.name } : null });
  const addActor = asset => update({ ...desk, actors: [...desk.actors, { id: `actor-${asset.id}`, assetId: asset.id, x: 50, y: 68, scale: 1 }] });
  const pointerMove = (event, actor) => {
    if (!surface.current || event.buttons !== 1) return;
    const bounds = surface.current.getBoundingClientRect();
    update(moveDirectorDeskActor(desk, actor.id, {
      x: (event.clientX - bounds.left) * 100 / Math.max(bounds.width, 1),
      y: (event.clientY - bounds.top) * 100 / Math.max(bounds.height, 1)
    }));
  };
  const saveScreenshot = async () => {
    if (savingScreenshot) return;
    setSavingScreenshot(true);
    try {
      await onSaveScreenshot?.(node.id, await captureDirectorDeskScreenshot(surface.current, desk));
    } finally {
      setSavingScreenshot(false);
    }
  };
  return <section className="agent-director-desk" aria-label="导演台预演参考">
    <header><span><Camera size={14} />导演台</span><small>预演参考</small></header>
    <div ref={surface} className="agent-director-desk__stage" style={desk.sceneAsset?.url ? { backgroundImage: `url("${desk.sceneAsset.url}")` } : undefined}>
      {desk.sceneAsset?.url ? <AuthenticatedAgentImage
        source={desk.sceneAsset.url}
        data-director-asset-id={desk.sceneAsset.id}
        className="agent-director-desk__scene-source"
        alt=""
      /> : null}
      <div className="agent-director-desk__camera" style={directorDeskSceneStyle(desk)}>
        {desk.actors.map(actor => {
          const asset = referenceByID.get(actor.assetId);
          return asset?.url ? <button key={actor.id} type="button" className="agent-director-desk__actor" style={{ left: `${actor.x}%`, top: `${actor.y}%`, transform: `translate(-50%, -100%) scale(${actor.scale}) rotate(${actor.rotation}deg)` }} onPointerDown={event => event.currentTarget.setPointerCapture?.(event.pointerId)} onPointerMove={event => pointerMove(event, actor)} title={`拖动 ${asset.name}`}>
            <AuthenticatedAgentImage source={asset.url} data-director-asset-id={asset.id} alt={asset.name || '角色'} />
          </button> : null;
        })}
      </div>
      {!desk.sceneAsset ? <span className="agent-director-desk__empty">添加场景图后开始预演</span> : null}
    </div>
    <div className="agent-director-desk__toolbar">
      <button type="button" onClick={event => onPickReferenceAsset?.(node.id, event)}><ImagePlus size={13} />添加场景/角色</button>
      <label><Move size={13} />取景 <input aria-label="导演台镜头缩放" type="range" min="0.6" max="2" step="0.1" value={desk.camera.zoom} onChange={event => update({ ...desk, camera: { ...desk.camera, zoom: Number(event.target.value) } })} /></label>
      <button type="button" disabled={!desk.sceneAsset || savingScreenshot} onClick={saveScreenshot}><Camera size={13} />{savingScreenshot ? '正在生成' : '生成预演图'}</button>
    </div>
    {references.length ? <div className="agent-director-desk__references">{references.map(asset => <span key={asset.id} className={desk.sceneAsset?.id === asset.id ? 'is-scene' : ''}>
      <button type="button" title={`设为场景：${asset.name}`} onClick={() => setScene(asset)}><AuthenticatedAgentImage source={asset.url} data-director-asset-id={asset.id} alt="" /></button>
      <button type="button" title={`添加角色：${asset.name}`} onClick={() => addActor(asset)}><UserRound size={12} /></button>
      <button type="button" title={`移除素材：${asset.name}`} onClick={() => onRemoveReference?.(node.id, asset.id)}><X size={12} /></button>
    </span>)}</div> : null}
  </section>;
}

// The canvas only needs a compact, draggable entry point. Composition belongs
// in the dedicated workspace above so controls never fight React Flow dragging.
export function AgentDirectorDeskNode({ node, onOpen, onConfigChange, onPickReferenceAsset }) {
  const desk = directorDeskConfig(node?.data?.config?.directorDesk);
  const actorCount = desk.actors.length;
  const [draftState, setDraftState] = useState(() => ({ nodeID: node.id, draft: createDirectorPromptDraft(desk.scenePrompt) }));
  const draft = draftState.nodeID === node.id ? syncDirectorPromptDraft(draftState.draft, desk.scenePrompt) : createDirectorPromptDraft(desk.scenePrompt);
  useEffect(() => {
    setDraftState(previous => {
      const next = previous.nodeID === node.id ? syncDirectorPromptDraft(previous.draft, desk.scenePrompt) : createDirectorPromptDraft(desk.scenePrompt);
      if (previous.nodeID === node.id && previous.draft.value === next.value && previous.draft.base === next.base && previous.draft.conflict === next.conflict && previous.draft.pending === next.pending) return previous;
      return { nodeID: node.id, draft: next };
    });
  }, [node.id, desk.scenePrompt]);
  const saveDraft = event => {
    event?.preventDefault();
    event?.stopPropagation();
    if (draft.conflict || !onConfigChange) return;
    const next = mergeDirectorDeskConfig(node?.data?.config?.directorDesk, { ...desk, scenePrompt: draft.value });
    setDraftState({ nodeID: node.id, draft: submitDirectorPromptDraft(draft, next.scene.prompt) });
    onConfigChange(node.id, { directorDesk: next });
  };
  const stopCanvasEvent = event => event.stopPropagation();
  const hasDraft = Boolean(desk.scenePrompt || desk.sceneAsset?.id);
  return <section className="agent-director-desk-node" aria-label="导演台节点">
    <div className="agent-director-desk-node__preview" style={desk.sceneAsset?.url ? { backgroundImage: `url("${desk.sceneAsset.url}")` } : undefined}>
      {desk.sceneAsset?.url ? <AuthenticatedAgentImage source={desk.sceneAsset.url} className="agent-director-desk-node__source" alt="" /> : <Camera size={24} />}
      {actorCount ? <span>{actorCount} 位角色</span> : null}
    </div>
    <div className="agent-director-desk-node__summary">
      <strong>{desk.sceneAsset?.name || '未设置场景'}</strong>
      <small>{actorCount ? `${actorCount} 位角色 · 3D 导演台` : hasDraft ? '构图草稿已保存，可进入 3D 导演台继续搭建' : '填写场景并添加参考图后开始预演'}</small>
    </div>
    <div className="agent-director-desk-node__composer nodrag nopan" onMouseDown={stopCanvasEvent}>
      <textarea
        aria-label="导演台场景描述"
        value={draft.value}
        maxLength={1000}
        placeholder="描述想要搭建的场景，支持通过参考图创建"
        onChange={event => setDraftState({ nodeID: node.id, draft: editDirectorPromptDraft(draft, event.target.value, desk.scenePrompt) })}
        onPointerDown={stopCanvasEvent}
        onKeyDown={stopCanvasEvent}
      />
      {draft.conflict ? <div role="alert">场景描述已在其他位置更新。继续编辑草稿或放弃草稿后再保存。<button type="button" onClick={() => setDraftState({ nodeID: node.id, draft: discardDirectorPromptDraft(draft, desk.scenePrompt) })}>放弃草稿</button></div> : null}
      <div>
        <button type="button" onClick={event => { stopCanvasEvent(event); onPickReferenceAsset?.(node.id, event); }}><ImagePlus size={13} />上传图片</button>
        <button type="button" disabled={draft.conflict || (!draft.value.trim() && !desk.sceneAsset?.id)} onClick={saveDraft}><Send size={13} />保存构图</button>
      </div>
    </div>
    <button type="button" className="nodrag nopan" onMouseDown={stopCanvasEvent} onClick={event => { stopCanvasEvent(event); onOpen?.(node.id); }}><Camera size={13} />打开 3D 导演台</button>
  </section>;
}
