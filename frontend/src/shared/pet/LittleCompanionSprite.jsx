import { useEffect, useLayoutEffect, useMemo, useReducer, useRef, useState } from 'react';
import {
  LITTLE_COMPANION_POSES,
  littleCompanionFrame,
  littleCompanionIdleFrameAfterWalk,
  littleCompanionPoseStep,
  littleCompanionTimeline,
  littleCompanionWalkFrame,
  nextLittleCompanionStep
} from './littleCompanionMotion.js';
import { littleCompanionEdgeSpriteOffset } from './littleCompanionRoaming.js';

// 先加载待机与小跑，其余动作随后补齐。这样散步首帧不会临时触发一次
// 网络请求，避免“刚起步就只剩名字”的水合空窗。
const PRELOAD_POSE_ORDER = ['idle', 'walk', 'peek', 'dance', 'peace', 'candy', 'think', 'sleep'];
const IMAGE_RETRY_DELAY_MS = 6000;
const imageStates = new Map();
const imageListeners = new Set();

function notifyImageState() {
  imageListeners.forEach(listener => listener());
}

function subscribeImageState(listener) {
  imageListeners.add(listener);
  return () => imageListeners.delete(listener);
}

export function littleCompanionImageStatus(src) {
  return imageStates.get(src)?.status || 'idle';
}

export function loadLittleCompanionImage(src) {
  if (!src || typeof window === 'undefined' || typeof window.Image !== 'function') return;
  const current = imageStates.get(src);
  if (current && (current.status === 'loading' || current.status === 'loaded')) return;
  const entry = { status: 'loading', retryTimer: null };
  imageStates.set(src, entry);
  const image = new window.Image();
  image.decoding = 'async';
  image.onload = () => {
    if (imageStates.get(src) !== entry) return;
    entry.status = 'loaded';
    notifyImageState();
  };
  image.onerror = () => {
    if (imageStates.get(src) !== entry) return;
    entry.status = 'failed';
    notifyImageState();
    entry.retryTimer = window.setTimeout(() => {
      if (imageStates.get(src) !== entry) return;
      imageStates.delete(src);
      loadLittleCompanionImage(src);
      notifyImageState();
    }, IMAGE_RETRY_DELAY_MS);
  };
  image.src = src;
}

export function preloadLittleCompanionPoses() {
  PRELOAD_POSE_ORDER.forEach(name => {
    const definition = LITTLE_COMPANION_POSES[name];
    if (definition) loadLittleCompanionImage(definition.src);
  });
}

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
  const [, refreshImageState] = useReducer(value => value + 1, 0);
  const timeline = useMemo(() => {
    const entries = littleCompanionTimeline(state);
    return preview ? entries.slice(0, 1) : entries;
  }, [preview, state]);
  const [stepIndex, setStepIndex] = useState(0);
  const [frameIndex, setFrameIndex] = useState(0);
  const displayedPoseRef = useRef(null);
  const lastWalkFrameRef = useRef(0);
  const timelineStep = littleCompanionPoseStep(state, stepIndex, preview);
  const step = state === 'idle' && idleStep ? idleStep : timelineStep;
  const frame = step.pose === 'walk'
    ? littleCompanionWalkFrame(frameIndex)
    : littleCompanionFrame(step.pose, frameIndex);
  const frameStatus = littleCompanionImageStatus(frame.src);
  // 记录最后一次成功加载过的动作帧。后续动作图加载失败时继续显示它，
  // 不再退化成只有名字的文字占位。
  const [lastGoodFrame, setLastGoodFrame] = useState(() => (
    littleCompanionImageStatus(frame.src) === 'loaded' ? frame : null
  ));

  useEffect(() => subscribeImageState(refreshImageState), []);

  useEffect(() => {
    preloadLittleCompanionPoses();
  }, []);

  useEffect(() => {
    loadLittleCompanionImage(frame.src);
  }, [frame.src]);

  useEffect(() => {
    setStepIndex(0);
    setFrameIndex(0);
  }, [preview, state]);

  useLayoutEffect(() => {
    if (frameStatus !== 'loaded') return;
    setLastGoodFrame(current => (current && current.src === frame.src ? current : frame));
  }, [frame, frameStatus]);

  useLayoutEffect(() => {
    if (step.pose === 'walk') lastWalkFrameRef.current = frameIndex;
  }, [frameIndex, step.pose]);

  useLayoutEffect(() => {
    const previousPose = displayedPoseRef.current;
    setFrameIndex(step.pose === 'idle' && previousPose === 'walk'
      ? littleCompanionIdleFrameAfterWalk(lastWalkFrameRef.current)
      : 0);
    displayedPoseRef.current = step.pose;
  }, [step.pose]);

  useEffect(() => {
    onPoseChange?.(step);
  }, [onPoseChange, step.durationMs, step.pose]);

  useEffect(() => {
    if (!reduced) return;
    setFrameIndex(0);
  }, [reduced]);

  useEffect(() => {
    if (reduced) return undefined;
    const timer = window.setInterval(() => {
      setFrameIndex(current => (current + 1) % frame.frameCount);
    }, frame.frameDelayMs);
    return () => window.clearInterval(timer);
  }, [frame.frameCount, frame.frameDelayMs, reduced]);

  useEffect(() => {
    if (reduced || preview || state === 'idle' || timeline.length <= 1) return undefined;
    const timer = window.setTimeout(() => {
      setStepIndex(current => nextLittleCompanionStep(state, current));
    }, step.durationMs);
    return () => window.clearTimeout(timer);
  }, [preview, reduced, state, step.durationMs, timeline.length]);

  const visualDirection = peekSide === 'left' ? -1 : peekSide === 'right' ? 1 : direction;
  const edgeOffset = littleCompanionEdgeSpriteOffset(peekSide);
  // 当前动作图已就绪时直接播放当前帧；加载中/失败时沿用上一个可用帧。
  // 只有连一个可用帧都没有时才进入占位图，占位图始终带图形，名字只作为附属。
  const displayedFrame = frameStatus === 'loaded'
    ? frame
    : lastGoodFrame || (frameStatus === 'failed' ? null : frame);
  const spriteUnavailable = !displayedFrame;

  useEffect(() => {
    onLoadError?.(spriteUnavailable);
  }, [onLoadError, spriteUnavailable]);

  return (
    <span className={`little-companion-sprite${visualDirection < 0 ? ' is-facing-left' : ''}${peekSide ? ` is-peeking-${peekSide}` : ''}${spriteUnavailable ? ' is-placeholder' : ''}`} style={{ '--little-edge-offset': `${edgeOffset}px`, '--little-facing-scale': visualDirection < 0 ? -1 : 1 }} aria-hidden="true">
      {displayedFrame ? (
        <span
          className="little-companion-frame"
          style={{
            backgroundImage: `url("${displayedFrame.src}")`,
            backgroundSize: `${displayedFrame.columns * 120}px ${displayedFrame.rows * 120}px`,
            backgroundPosition: `${-displayedFrame.column * 120}px ${-displayedFrame.row * 120}px`
          }}
        />
      ) : (
        <span className="little-companion-placeholder">
          <svg className="little-companion-placeholder-art" viewBox="0 0 32 32" focusable="false">
            <rect fill="currentColor" x="5" y="5" width="4" height="4" rx="1" />
            <rect fill="currentColor" x="23" y="5" width="4" height="4" rx="1" />
            <rect fill="currentColor" x="6" y="7" width="20" height="18" rx="6" />
            <rect fill="var(--legacy-card, #fff)" x="10" y="12" width="3" height="4" rx="1" />
            <rect fill="var(--legacy-card, #fff)" x="19" y="12" width="3" height="4" rx="1" />
            <rect fill="var(--legacy-card, #fff)" x="11" y="20" width="10" height="4" rx="2" opacity="0.55" />
          </svg>
          {spriteUnavailable ? <span className="little-companion-placeholder-name">{label}</span> : null}
        </span>
      )}
    </span>
  );
}

export default LittleCompanionSprite;
