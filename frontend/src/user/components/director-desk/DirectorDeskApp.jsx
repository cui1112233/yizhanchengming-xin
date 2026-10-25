import { Camera, ImagePlus, UserRound, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { directorProjectCaptureMetadata, normalizeDirectorProject } from '../../pages/agent-director-desk.js';
import { DirectorScene } from './DirectorScene';
import { updateCamera } from './director-scene-state.js';

function assetForScene(asset) {
  return { assetId: asset.id, url: asset.url || '', name: asset.name || '', prompt: '' };
}

export default function DirectorDeskApp({ adapter, onCapture }) {
  const [project, setProject] = useState(() => normalizeDirectorProject(adapter.readState()));
  const [images, setImages] = useState([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const sceneRef = useRef(null);

  useEffect(() => {
    let active = true;
    adapter.listProjectImages()
      .then(next => { if (active) setImages(next); })
      .catch(cause => { if (active) setError(cause?.message || '项目图片读取失败'); });
    return () => { active = false; };
  }, [adapter]);

  const saveProject = async next => {
    const normalized = normalizeDirectorProject(next);
    setProject(normalized);
    setError('');
    try {
      await adapter.writeState(normalized);
      return true;
    } catch (cause) {
      setError(cause?.message || '导演台保存失败');
      return false;
    }
  };
  const addActor = asset => saveProject({
    ...project,
    actors: [...project.actors, { id: `actor-${asset.id}-${Date.now()}`, assetId: asset.id, position: [0, 0, 0], rotation: [0, 0, 0], scale: 1 }]
  });
  const capture = async () => {
    if (busy) return;
    setBusy(true);
    setError('');
    try {
      const file = await sceneRef.current?.capture();
      if (!file) throw new Error('3D 构图尚未准备完成');
      const asset = await adapter.storeCapture(file);
      const next = directorProjectCaptureMetadata(project, asset);
      const saved = await saveProject(next);
      if (!saved) throw new Error('3D 构图状态保存失败');
      await onCapture?.(asset);
    } catch (cause) {
      setError(cause?.message || '3D 构图保存失败');
    } finally {
      setBusy(false);
    }
  };
  const directorView = () => saveProject(updateCamera(project, { position: [5.5, 3.5, 8], target: [0, 1.4, 0], fov: 38 }));

  return <section className="agent-director-desk-fullscreen" data-director-desk-open onPointerDown={event => event.stopPropagation()} onWheel={event => event.stopPropagation()} onContextMenu={event => event.stopPropagation()}>
    <header className="agent-director-desk-fullscreen__header">
      <span><Camera size={18} /><strong>3D导演台</strong></span>
      <div><button type="button" onClick={directorView}>导演视角</button><button type="button" aria-label="关闭导演台" onClick={() => adapter.close()}><X size={16} />关闭</button></div>
    </header>
    <main className="agent-director-desk-fullscreen__body">
      <aside className="agent-director-desk-fullscreen__assets" aria-label="项目图片素材">
        <header><ImagePlus size={15} />项目图片</header>
        {!images.length ? <p>项目中还没有图片素材</p> : images.map(asset => <article key={asset.id}>
          {asset.url ? <img src={asset.url} alt="" /> : null}
          <strong>{asset.name || '未命名图片'}</strong>
          <div><button type="button" onClick={() => saveProject({ ...project, scene: assetForScene(asset) })}>设为场景</button><button type="button" onClick={() => addActor(asset)}><UserRound size={13} />加入角色</button></div>
        </article>)}
      </aside>
      <section className="agent-director-desk-fullscreen__viewport" aria-label="3D 场景">
        <DirectorScene ref={sceneRef} project={project} images={images} onProjectChange={saveProject} />
        {error ? <p className="agent-director-desk-fullscreen__error" role="alert">{error}</p> : null}
      </section>
      <aside className="agent-director-desk-fullscreen__actions">
        <strong>{project.scene.name || '未选择场景图'}</strong>
        <span>{project.actors.length} 个角色立面</span>
        <button type="button" disabled={busy} onClick={capture}><Camera size={15} />{busy ? '正在保存构图' : '截取到画布'}</button>
      </aside>
    </main>
  </section>;
}
