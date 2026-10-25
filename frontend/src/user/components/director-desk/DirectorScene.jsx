import { OrbitControls } from '@react-three/drei';
import { Canvas, useThree } from '@react-three/fiber';
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';
import * as THREE from 'three';
import { createAuthenticatedMediaSession } from '../../pages/agent-canvas-media.js';
import { updateDirectorActorTransform } from '../../pages/agent-director-desk.js';
import { actorPlaneProps, cameraFromControls, updateCamera } from './director-scene-state.js';

export { actorPlaneProps, cameraFromControls, updateCamera } from './director-scene-state.js';

function CameraFrame({ cameraState }) {
  const { camera } = useThree();
  useEffect(() => {
    camera.position.fromArray(cameraState.position);
    camera.fov = cameraState.fov;
    camera.updateProjectionMatrix();
    camera.lookAt(...cameraState.target);
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

export function ImageActorPlane({ actor, asset, onTransform, selected, onSelect }) {
  const meshRef = useRef(null);
  const dragRef = useRef(false);
  const texture = useAuthenticatedTexture(asset?.url);
  const props = actorPlaneProps(actor);
  const finishDrag = event => {
    if (!dragRef.current || !meshRef.current) return;
    dragRef.current = false;
    event.stopPropagation();
    onTransform?.({
      position: meshRef.current.position.toArray(),
      rotation: meshRef.current.rotation.toArray(),
      scale: meshRef.current.scale.x
    });
  };
  return <mesh
    ref={meshRef}
    position={props.position}
    rotation={props.rotation}
    scale={props.scale}
    onClick={event => { event.stopPropagation(); onSelect?.(actor.id); }}
    onPointerDown={event => { event.stopPropagation(); dragRef.current = true; event.target.setPointerCapture?.(event.pointerId); onSelect?.(actor.id); }}
    onPointerMove={event => {
      if (!dragRef.current || !meshRef.current) return;
      event.stopPropagation();
      meshRef.current.position.set(event.point.x, Math.max(0, event.point.y), event.point.z);
    }}
    onPointerUp={finishDrag}
    onPointerCancel={finishDrag}
  >
    <planeGeometry args={[2.25, 3.5]} />
    {texture ? <meshBasicMaterial map={texture} transparent toneMapped={false} /> : <meshBasicMaterial color={selected ? '#d97706' : '#55606b'} />}
    {selected ? <lineSegments><edgesGeometry args={[new THREE.PlaneGeometry(2.25, 3.5)]} /><lineBasicMaterial color="#fbbf24" /></lineSegments> : null}
  </mesh>;
}

export const DirectorScene = forwardRef(function DirectorScene({ project, images, onProjectChange }, ref) {
  const glRef = useRef(null);
  const [selectedActorID, setSelectedActorID] = useState('');
  useImperativeHandle(ref, () => ({
    capture: () => new Promise((resolve, reject) => {
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
    })
  }), []);

  return <Canvas
    className="agent-director-3d-scene"
    camera={{ position: project.camera.position, fov: project.camera.fov }}
    onCreated={({ gl }) => { glRef.current = gl; gl.setClearColor('#0b1017'); }}
  >
    <CameraFrame cameraState={project.camera} />
    <ambientLight intensity={1.2} />
    <directionalLight position={[4, 8, 5]} intensity={1.8} />
    <SceneBackground scene={project.scene} />
    <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]}><planeGeometry args={[30, 30]} /><meshStandardMaterial color="#17212b" roughness={1} /></mesh>
    {project.actors.map(actor => <ImageActorPlane
      key={actor.id}
      actor={actor}
      asset={(images || []).find(image => image.id === actor.assetId)}
      selected={selectedActorID === actor.id}
      onSelect={setSelectedActorID}
      onTransform={transform => onProjectChange(updateDirectorActorTransform(project, actor.id, transform))}
    />)}
    <OrbitControls
      target={project.camera.target}
      onEnd={event => onProjectChange(updateCamera(project, cameraFromControls(event.target.object, event.target.target)))}
    />
  </Canvas>;
});
