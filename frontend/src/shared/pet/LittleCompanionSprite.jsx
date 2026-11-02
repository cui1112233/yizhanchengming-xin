import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import {
  littleCompanionFrame,
  littleCompanionIdleFrameAfterWalk,
  littleCompanionPoseStep,
  littleCompanionTimeline,
  littleCompanionWalkFrame,
  nextLittleCompanionStep
} from './littleCompanionMotion.js';
import { littleCompanionEdgeSpriteOffset } from './littleCompanionRoaming.js';

function useReducedMotion() {
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return undefined;
    const query = window.matchMedia('(prefers-reduced-motion: reduce)');
    const update = () => setReduced(query.matches);
    update();
    if (typeof query.addEventListener === 'function') query.addEventListener('change', update);
    else query.addListener?.(update);
    return () => {
      if (typeof query.removeEventListener === 'function') query.removeEventListener('change', update);
      else query.removeListener?.(update);
    };
  }, []);

  return reduced;
}

export function LittleCompanionSprite({
  state = 'idle',
  preview = false,
  label = '小家伙',
  direction = 1,
  idleStep,
  peekSide = null,
  onPoseChange,
  onLoadError
}) {
  const reduced = useReducedMotion();
  const timeline = useMemo(() => {
    const entries = littleCompanionTimeline(state);
    return preview ? entries.slice(0, 1) : entries;
  }, [preview, state]);
  const [stepIndex, setStepIndex] = useState(0);
  const [frameIndex, setFrameIndex] = useState(0);
  const [failed, setFailed] = useState(false);
  const displayedPoseRef = useRef(null);
  const lastWalkFrameRef = useRef(0);
  const timelineStep = littleCompanionPoseStep(state, stepIndex, preview);
  const step = state === 'idle' && idleStep ? idleStep : timelineStep;
  const frame = step.pose === 'walk'
    ? littleCompanionWalkFrame(frameIndex)
    : littleCompanionFrame(step.pose, frameIndex);

  useEffect(() => {
    setStepIndex(0);
    setFrameIndex(0);
  }, [preview, state]);

  useLayoutEffect(() => {
    if (step.pose === 'walk') lastWalkFrameRef.current = frameIndex;
  }, [frameIndex, step.pose]);

  useLayoutEffect(() => {
    const previousPose = displayedPoseRef.current;
    setFrameIndex(step.pose === 'idle' && previousPose === 'walk'
      ? littleCompanionIdleFrameAfterWalk(lastWalkFrameRef.current)
      : 0);
    displayedPoseRef.current = step.pose;
    setFailed(false);
    onLoadError?.(false);
  }, [step.pose]);

  useEffect(() => {
    onPoseChange?.(step);
  }, [onPoseChange, step.durationMs, step.pose]);

  useEffect(() => {
    if (!reduced) return;
    setFrameIndex(0);
  }, [reduced]);

  useEffect(() => {
    if (reduced || failed) return undefined;
    const timer = window.setInterval(() => {
      setFrameIndex(current => (current + 1) % frame.frameCount);
    }, frame.frameDelayMs);
    return () => window.clearInterval(timer);
  }, [failed, frame.frameCount, frame.frameDelayMs, reduced]);

  useEffect(() => {
    if (reduced || preview || state === 'idle' || timeline.length <= 1) return undefined;
    const timer = window.setTimeout(() => {
      setStepIndex(current => nextLittleCompanionStep(state, current));
    }, step.durationMs);
    return () => window.clearTimeout(timer);
  }, [preview, reduced, state, step.durationMs, timeline.length]);

  const visualDirection = peekSide === 'left' ? -1 : peekSide === 'right' ? 1 : direction;
  const edgeOffset = littleCompanionEdgeSpriteOffset(peekSide);

  return (
    <span className={`little-companion-sprite${visualDirection < 0 ? ' is-facing-left' : ''}${peekSide ? ` is-peeking-${peekSide}` : ''}${failed ? ' is-failed' : ''}`} style={{ '--little-edge-offset': `${edgeOffset}px`, '--little-facing-scale': visualDirection < 0 ? -1 : 1 }} aria-hidden="true">
      {failed ? <span className="little-companion-fallback">{label}</span> : (
        <>
          <span
            className="little-companion-frame"
            style={{
              backgroundImage: `url("${frame.src}")`,
              backgroundSize: `${frame.columns * 120}px ${frame.rows * 120}px`,
              backgroundPosition: `${-frame.column * 120}px ${-frame.row * 120}px`
            }}
          />
          <img
            className="little-companion-preload"
            src={frame.src}
            alt=""
            draggable="false"
            onError={() => {
              setFailed(true);
              onLoadError?.(true);
            }}
          />
        </>
      )}
    </span>
  );
}

export default LittleCompanionSprite;
