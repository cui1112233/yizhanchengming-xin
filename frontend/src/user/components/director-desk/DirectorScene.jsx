import { OrbitControls, TransformControls } from '@react-three/drei';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';
import * as THREE from 'three';
import { createAuthenticatedMediaSession } from '../../pages/agent-canvas-media.js';
import { normalizeDirectorProject, updateDirectorActorTransform, updateDirectorCamera, updateDirectorViewCamera } from '../../pages/agent-director-desk.js';
import { actorPlaneProps, applyCameraFrame, aspectRatioValue, cameraForView, cameraFromControls, canTransformActor, focusDirectorObject, resolveSceneSelection, updateCamera } from './director-scene-state.js';

export { actorPlaneProps, cameraFromControls, updateCamera } from './director-scene-state.js';

function CameraFrame({ cameraState }) {
  const { camera } = useThree();
  useEffect(() => {
    applyCameraFrame(camera, cameraState);
  }, [camera, cameraState]);
  return null;
}

function useAuthenticatedTexture(source) {
  const [texture, setTexture] = useState(null);
  useEffect(() => {
    let active = true;
    let loadedTexture = null;
    const session = createAuthenticatedMediaSession(source, {
      onState: state => {
        if (!active || !state.url) return;
        const image = new Image();
        image.onload = () => {
          if (!active) return;
          loadedTexture = new THREE.Texture(image);
          loadedTexture.colorSpace = THREE.SRGBColorSpace;
          loadedTexture.needsUpdate = true;
          setTexture(loadedTexture);
        };
        image.src = state.url;
      }
    });
    return () => {
      active = false;
      session.dispose();
      loadedTexture?.dispose();
    };
  }, [source]);
  return texture;
}

export function SceneBackground({ scene }) {
  const texture = useAuthenticatedTexture(scene?.url);
  if (!scene?.assetId || !texture) {
    return <mesh position={[0, 3.2, -3.6]}><planeGeometry args={[16, 9]} /><meshBasicMaterial color="#28323c" /></mesh>;
  }
  return <mesh position={[0, 3.2, -3.6]}><planeGeometry args={[16, 9]} /><meshBasicMaterial map={texture} toneMapped={false} /></mesh>;
}

export function ImageActorPlane({ actor, asset, selected, onSelect }) {
  const texture = useAuthenticatedTexture(asset?.url);
  const props = actorPlaneProps(actor);
  return <mesh
    position={props.position}
    rotation={props.rotation}
    scale={props.scale}
    onClick={event => { event.stopPropagation(); onSelect?.(actor.id); }}
  >
    <planeGeometry args={[2.25, 3.5]} />
    {texture ? <meshBasicMaterial map={texture} transparent toneMapped={false} /> : <meshBasicMaterial color={selected ? '#d97706' : '#55606b'} />}
    {selected ? <lineSegments><edgesGeometry args={[new THREE.PlaneGeometry(2.25, 3.5)]} /><lineBasicMaterial color="#fbbf24" /></lineSegments> : null}
  </mesh>;
}

function FrameReady({ onReady }) {
  const sent = useRef(false);
  const frame = useRef(null);
  useEffect(() => () => cancelAnimationFrame(frame.current), []);
  useFrame(() => {
    if (sent.current) return;
    sent.current = true;
    frame.current = requestAnimationFrame(onReady);
  });
  return null;
}

function Movable({ selected, onFinish, onDragging, position, rotation, scale, children }) {
  const control = useRef(null);
  if (!selected) return <group position={position} rotation={rotation} scale={scale}>{children}</group>;
  return <TransformControls ref={control} mode="translate" onMouseDown={() => onDragging(true)} onMouseUp={() => {
    onDragging(false);
    const object = control.current?.object;
    if (object) onFinish(object);
  }} position={position} rotation={rotation} scale={scale}>{children}</TransformControls>;
}

function CameraMarker({ camera, selected, onSelect }) {
  return <group onClick={event => { event.stopPropagation(); onSelect(); }}>
    <mesh><boxGeometry args={[0.55, 0.38, 0.34]} /><meshBasicMaterial color={selected ? '#ffbd38' : '#58c8df'} /></mesh>
    <mesh position={[0, 0, -0.3]} rotation={[Math.PI / 2, 0, 0]}><coneGeometry args={[0.2, 0.3, 4]} /><meshBasicMaterial color={selected ? '#ffbd38' : '#58c8df'} /></mesh>
  </group>;
}

export const DirectorScene = forwardRef(function DirectorScene({ project, images, selection, viewMode = 'director', transformMode = 'translate', onSelectionChange, onProjectChange, onReadyChange }, ref) {
  const glRef = useRef(null);
  const controlsRef = useRef(null);
  const readyRef = useRef(false);
  const onReadyChangeRef = useRef(onReadyChange);
  onReadyChangeRef.current = onReadyChange;
  const [localSelection, setLocalSelection] = useState(null);
  const [transformDragging, setTransformDragging] = useState(false);
  const selected = resolveSceneSelection(selection, localSelection);
  const current = normalizeDirectorProject(project);
  const cameraState = cameraForView(current, viewMode);
  const ratio = aspectRatioValue(current.aspectRatio);
  const select = value => { setLocalSelection(value); onSelectionChange?.(value); };
  useEffect(() => () => { readyRef.current = false; onReadyChangeRef.current?.(false); }, []);
  useImperativeHandle(ref, () => ({
    capture: () => new Promise((resolve, reject) => {
      if (!readyRef.current) { reject(new Error('3D 构图尚未准备完成')); return; }
      const canvas = glRef.current?.domElement;
      if (!canvas?.toBlob) {
        reject(new Error('当前浏览器无法截取 3D 构图'));
        return;
      }
      canvas.toBlob(blob => {
        if (!blob) {
          reject(new Error('3D 构图截图失败'));
          return;
        }
        resolve(new File([blob], `director-desk-${Date.now()}.png`, { type: 'image/png' }));
      }, 'image/png');
    }),
    resetDirectorView: () => onProjectChange(updateDirectorViewCamera(current, normalizeDirectorProject().directorCamera)),
    focusObject: target => onProjectChange(focusDirectorObject(current, target))
  }), [current, onProjectChange]);

  return <div className={`agent-director-viewport-frame ${ratio ? '' : 'is-adaptive'}`} style={ratio ? { aspectRatio: String(ratio), '--director-aspect-ratio': ratio } : undefined}><Canvas
    className="agent-director-3d-scene"
    camera={{ position: cameraState.position, fov: cameraState.fov }}
    gl={{ preserveDrawingBuffer: true }}
    onCreated={({ gl }) => { glRef.current = gl; gl.setClearColor('#0b1017'); }}
  >
    <FrameReady onReady={() => { readyRef.current = true; onReadyChangeRef.current?.(true); }} />
    <CameraFrame cameraState={cameraState} />
    <ambientLight intensity={1.2} />
    <directionalLight position={[4, 8, 5]} intensity={1.8} />
    <SceneBackground scene={current.scene} />
    <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]}><planeGeometry args={[30, 30]} /><meshStandardMaterial color="#17212b" roughness={1} /></mesh>
    {current.actors.map(actor => <Movable key={actor.id} {...actorPlaneProps(actor)} selected={canTransformActor(selected, actor.id, transformMode, viewMode)} onDragging={setTransformDragging} onFinish={object => onProjectChange(updateDirectorActorTransform(current, actor.id, { position: object.position.toArray(), rotation: [object.rotation.x, object.rotation.y, object.rotation.z], scale: object.scale.toArray() }))}><ImageActorPlane
      actor={{ ...actor, position: [0, 0, 0], rotation: [0, 0, 0], scale: [1, 1, 1] }}
      asset={(images || []).find(image => image.id === actor.assetId)}
      selected={selected?.kind === 'actor' && selected.id === actor.id}
      onSelect={() => select({ kind: 'actor', id: actor.id })}
    /></Movable>)}
    {viewMode === 'director' ? current.cameras.map(camera => <Movable key={camera.id} position={camera.position} selected={transformMode === 'translate' && selected?.kind === 'camera' && selected.id === camera.id} onDragging={setTransformDragging} onFinish={object => onProjectChange(updateDirectorCamera(current, camera.id, { position: object.position.toArray() }))}><CameraMarker camera={camera} selected={selected?.kind === 'camera' && selected.id === camera.id} onSelect={() => select({ kind: 'camera', id: camera.id })} /></Movable>) : null}
    {viewMode === 'director' ? <OrbitControls ref={controlsRef} enabled={!transformDragging} target={current.directorCamera.target}
      onEnd={event => onProjectChange(updateCamera(current, cameraFromControls(event.target.object, event.target.target)))} /> : null}
  </Canvas></div>;
});
