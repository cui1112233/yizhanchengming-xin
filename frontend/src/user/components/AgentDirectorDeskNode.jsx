import { Camera, ImagePlus, Move, UserRound, X } from 'lucide-react';
import { useMemo, useRef, useState } from 'react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { directorDeskConfig, directorDeskSceneStyle, moveDirectorDeskActor } from '../pages/agent-director-desk';

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

export function AgentDirectorDeskNode({ node, onConfigChange, onPickReferenceAsset, onRemoveReference, onSaveScreenshot }) {
  const desk = directorDeskConfig(node?.data?.config?.directorDesk);
  const references = Array.isArray(node?.data?.config?.referenceAssets) ? node.data.config.referenceAssets.filter(asset => asset?.kind === 'image') : [];
  const referenceByID = useMemo(() => new Map(references.map(asset => [asset.id, asset])), [references]);
  const surface = useRef(null);
  const [savingScreenshot, setSavingScreenshot] = useState(false);
  const update = next => onConfigChange?.(node.id, { directorDesk: next });
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
  return <section className="agent-director-desk nodrag nopan" aria-label="3D 导演台预演参考" onMouseDown={event => event.stopPropagation()}>
    <header><span><Camera size={14} />导演台</span><small>预演参考</small></header>
    <div ref={surface} className="agent-director-desk__stage" style={desk.sceneAsset?.url ? { backgroundImage: `url("${desk.sceneAsset.url}")` } : undefined}>
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
