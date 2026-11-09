import { Camera } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import {
  DIRECTOR_ASPECT_RATIOS,
  addDirectorActor,
  addDirectorCamera,
  addDirectorProp,
  directorCameraSemantics,
  normalizeDirectorProject,
  removeDirectorObject,
  setDirectorActiveCamera,
  setDirectorAspectRatio,
  setDirectorSceneAsset
} from '../../pages/agent-director-desk.js';
import DirectorAssetPanel from './DirectorAssetPanel';
import DirectorBottomDock from './DirectorBottomDock';
import { createDirectorCaptureTransaction, queueDirectorCaptureMetadata } from './DirectorDeskHostAdapter.js';
import DirectorDeskShell from './DirectorDeskShell';
import DirectorInspector from './DirectorInspector';
import { DirectorScene } from './DirectorScene';
import DirectorSceneTree from './DirectorSceneTree';
import DirectorToolRail from './DirectorToolRail';
import { createDirectorSaveQueue } from './director-save-queue.js';

const CAMERA_PRESETS = Object.freeze([
  { id: 'current', label: '当前视角' },
  { id: 'front-medium', label: '正面中景', position: [0, 2.2, 7], target: [0, 1.4, 0], fov: 42 },
  { id: 'front-close', label: '正面特写', position: [0, 2, 4.2], target: [0, 1.6, 0], fov: 34 },
  { id: 'side-follow', label: '侧面跟拍', position: [6.5, 2.1, 1.5], target: [0, 1.3, 0], fov: 46 },
  { id: 'overhead', label: '俯拍', position: [0, 8.5, 1.5], target: [0, 0, 0], fov: 48 },
  { id: 'low-angle', label: '低角度', position: [0, 0.65, 5.2], target: [0, 1.8, 0], fov: 40 },
  { id: 'over-shoulder', label: '过肩', position: [3.2, 2.1, 4.6], target: [-0.4, 1.35, 0], fov: 38 },
  { id: 'dutch-angle', label: '荷兰角', position: [-3.8, 3.1, 5.4], target: [0, 1.2, 0], fov: 44, rotation: [0, 0, -0.22] }
]);

export default function DirectorDeskApp({ adapter }) {
  const [project, setProject] = useState(() => normalizeDirectorProject(adapter.readState()));
  const [images, setImages] = useState([]);
  const [activeTool, setActiveTool] = useState('scene');
  const [selectedObject, setSelectedObject] = useState({ kind: 'scene', id: 'scene' });
  const [viewMode, setViewMode] = useState('director');
  const [transformMode, setTransformMode] = useState('translate');
  const [promptDraft, setPromptDraft] = useState(project.scene.prompt);
  const [shotPromptDraft, setShotPromptDraft] = useState(project.shotPrompt);
  const [saveStatus, setSaveStatus] = useState('saved');
  const [saveError, setSaveError] = useState('');
  const [operationError, setOperationError] = useState('');
  const [sceneReady, setSceneReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [leftCollapsed, setLeftCollapsed] = useState(false);
  const [rightCollapsed, setRightCollapsed] = useState(false);
  const sceneRef = useRef(null);
  const projectRef = useRef(project);
  const saveQueueRef = useRef(null);
  const captureTransactionRef = useRef(null);

  if (!saveQueueRef.current) {
    saveQueueRef.current = createDirectorSaveQueue(next => adapter.writeState(next), {
      onStatus: (status, cause) => {
        setSaveStatus(status);
        if (status === 'error') setSaveError(cause?.message || '导演台保存失败');
        if (status === 'saved') setSaveError('');
      }
    });
  }
  const saveQueue = saveQueueRef.current;
  if (!captureTransactionRef.current) {
    captureTransactionRef.current = createDirectorCaptureTransaction({
      capture: async () => {
        await saveQueue.flush();
        return sceneRef.current.capture();
      },
      storeCapture: file => adapter.storeCapture(file),
      persistMetadata: asset => queueDirectorCaptureMetadata({
        saveQueue,
        project: projectRef.current,
        asset,
        updateProject: next => {
          projectRef.current = next;
          setProject(next);
        }
      }),
      insertCapture: asset => adapter.insertCapture(asset)
    });
  }
  const captureTransaction = captureTransactionRef.current;

  useEffect(() => {
    let active = true;
    adapter.listProjectImages()
      .then(next => { if (active) setImages(next); })
      .catch(cause => { if (active) setOperationError(cause?.message || '项目图片读取失败'); });
    return () => { active = false; };
  }, [adapter]);

  const applyProject = next => {
    const normalized = normalizeDirectorProject(next);
    projectRef.current = normalized;
    setProject(normalized);
    saveQueue.schedule(normalized);
    return normalized;
  };

  const selectScene = asset => {
    applyProject(setDirectorSceneAsset(projectRef.current, asset));
    setSelectedObject({ kind: 'scene', id: 'scene' });
  };
  const addActor = asset => {
    const next = applyProject(addDirectorActor(projectRef.current, { assetId: asset.id }));
    const actor = next.actors.at(-1);
    if (actor) setSelectedObject({ kind: 'actor', id: actor.id });
  };
  const addProp = asset => {
    const next = applyProject(addDirectorProp(projectRef.current, { assetId: asset.id }));
    const prop = next.props.at(-1);
    if (prop) setSelectedObject({ kind: 'prop', id: prop.id });
  };
  const addCamera = preset => {
    const camera = preset.id === 'current' ? projectRef.current.directorCamera : preset;
    const next = applyProject(addDirectorCamera(projectRef.current, { ...camera, id: '', name: preset.label }));
    const added = next.cameras.at(-1);
    if (added) setSelectedObject({ kind: 'camera', id: added.id });
  };
  const removeObject = selection => {
    applyProject(removeDirectorObject(projectRef.current, selection.id));
    if (selectedObject?.kind === selection.kind && selectedObject?.id === selection.id) {
      setSelectedObject({ kind: 'scene', id: 'scene' });
    }
  };
  const uploadReference = async file => {
    setBusy(true);
    try {
      await adapter.uploadProjectImage(file);
      setImages(await adapter.listProjectImages());
      setActiveTool('assets');
      setOperationError('');
    } catch (cause) {
      setOperationError(cause?.message || '参考图上传失败');
    } finally {
      setBusy(false);
    }
  };
  const savePrompt = () => applyProject({
    ...projectRef.current,
    scenePrompt: promptDraft,
    shotPrompt: shotPromptDraft,
    scene: { ...projectRef.current.scene, prompt: promptDraft, scenePrompt: promptDraft, shotPrompt: shotPromptDraft }
  });
  const close = async () => {
    try {
      await captureTransaction.flush();
      await saveQueue.flush();
      adapter.close();
    } catch (cause) {
      setSaveError(cause?.message || '导演台保存失败');
      setSaveStatus('error');
    }
  };
  const capture = async () => {
    if (!sceneReady || busy) return;
    setBusy(true);
    setOperationError('');
    try {
      await captureTransaction.start();
      setSaveError('');
    } catch (cause) {
      if (saveQueue.getStatus() === 'error') setSaveError(cause?.message || '导演台保存失败');
      else setOperationError(cause?.message || '3D 构图保存失败');
    } finally {
      setBusy(false);
    }
  };

  let toolPanel = <DirectorSceneTree
    project={project}
    selectedObject={selectedObject}
    onSelect={setSelectedObject}
    onLocate={selection => sceneRef.current?.focusObject(selection)}
    onRemove={removeObject}
  />;
  if (activeTool === 'assets' || activeTool === 'actors') {
    toolPanel = <div className={`agent-director-tool-panel${activeTool === 'actors' ? ' is-actors' : ''}`}><DirectorAssetPanel
      images={images}
       onSetScene={selectScene}
       onAddActor={addActor}
       onAddProp={addProp}
     /></div>;
  } else if (activeTool === 'cameras') {
    toolPanel = <section className="agent-director-choice-panel" aria-label="添加机位"><h2>添加机位</h2>{CAMERA_PRESETS.map(preset => <button
      key={preset.id} type="button" onClick={() => addCamera(preset)}>{preset.label}</button>)}</section>;
  } else if (activeTool === 'ratio') {
    toolPanel = <section className="agent-director-choice-panel" aria-label="画幅比例"><h2>画幅比例</h2>{DIRECTOR_ASPECT_RATIOS.map(ratio => <button
      key={ratio} type="button" aria-pressed={project.aspectRatio === ratio}
      onClick={() => applyProject(setDirectorAspectRatio(projectRef.current, ratio))}>{ratio === 'adaptive' ? '自适应' : ratio}</button>)}</section>;
  }

  const top = <div className="agent-director-shell__top-content">
    <span className="agent-director-shell__title"><Camera size={17} /><strong>3D导演台</strong></span>
    <div className="agent-director-shell__view-switch" aria-label="查看视角">
      <button type="button" aria-pressed={viewMode === 'director'} onClick={() => setViewMode('director')}>导演视角</button>
      <button type="button" aria-pressed={viewMode === 'camera'} onClick={() => setViewMode('camera')}>机位视角</button>
    </div>
  </div>;
  const viewport = <>
    <DirectorScene
      ref={sceneRef}
      project={project}
      images={images}
      selection={selectedObject}
      viewMode={viewMode}
      transformMode={transformMode}
      onSelectionChange={setSelectedObject}
      onProjectChange={applyProject}
      onReadyChange={setSceneReady}
    />
    {saveError || operationError ? <p className="agent-director-desk-fullscreen__error" role="alert">{saveError || operationError}</p> : null}
  </>;
  const inspector = <DirectorInspector
    project={project}
    images={images}
    selection={selectedObject}
    onProjectChange={applyProject}
    onSetActiveCamera={cameraID => {
      applyProject(setDirectorActiveCamera(projectRef.current, cameraID));
      setViewMode('camera');
    }}
    onChangeScene={() => setActiveTool('assets')}
    onSemanticCameraChange={patch => applyProject(directorCameraSemantics(projectRef.current, patch, selectedObject?.kind === 'camera' ? selectedObject.id : ''))}
  />;
  const bottom = <DirectorBottomDock
    prompt={promptDraft}
    shotPrompt={shotPromptDraft}
    saveStatus={saveStatus}
    canCapture={sceneReady && !busy}
    transformMode={transformMode}
    onTransformModeChange={setTransformMode}
    onPromptChange={setPromptDraft}
    onShotPromptChange={setShotPromptDraft}
    onSavePrompt={savePrompt}
    onOpenAssets={() => setActiveTool('assets')}
    onUploadReference={uploadReference}
    onResetView={() => sceneRef.current?.resetDirectorView()}
    onCapture={capture}
  />;

  return <section className="agent-director-desk-fullscreen" data-director-desk-open
    onPointerDown={event => event.stopPropagation()} onWheel={event => event.stopPropagation()} onContextMenu={event => event.stopPropagation()}>
    <DirectorDeskShell
      left={<><DirectorToolRail activeTool={activeTool} onToolChange={setActiveTool} />{toolPanel}</>}
      viewport={viewport}
      right={inspector}
      top={top}
      bottom={bottom}
      leftCollapsed={leftCollapsed}
      rightCollapsed={rightCollapsed}
      onToggleLeft={() => setLeftCollapsed(current => !current)}
      onToggleRight={() => setRightCollapsed(current => !current)}
      onClose={close}
    />
  </section>;
}
