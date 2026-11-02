const pose = (name, frames, durationMs, columns, rows, extension = 'png') => Object.freeze({
  pose: name,
  src: `/pets/little-companion/${name}.${extension}`,
  frames,
  durationMs,
  columns,
  rows
});

export const LITTLE_COMPANION_POSES = Object.freeze({
  idle: pose('idle', 24, 3600, 8, 4),
  walk: pose('walk', 12, 720, 8, 2),
  sleep: pose('sleep', 48, 4600, 8, 7),
  think: pose('think', 64, 12000, 8, 9),
  candy: pose('candy', 40, 4800, 8, 6),
  dance: pose('dance', 12, 610, 8, 2),
  peace: pose('peace', 24, 2000, 8, 4),
  peek: pose('peek', 24, 4000, 8, 4)
});

// Frames 7 and 11 share an identical head position while alternating feet.
// Keeping that pair prevents the source atlas' head-turn frames from shaking.
const WALK_FRAME_SEQUENCE = Object.freeze([6, 10]);
const WALK_TO_IDLE_FRAME = Object.freeze([4, 5]);

const IDLE_TIMELINE = Object.freeze(['idle']);
const TASK_TIMELINES = Object.freeze({
  working: Object.freeze(['think']),
  success: Object.freeze(['peace', 'dance']),
  error: Object.freeze(['idle'])
});

const timelineFor = state => TASK_TIMELINES[state] || IDLE_TIMELINE;

export function littleCompanionTimeline(state) {
  return timelineFor(state).map(name => Object.freeze({
    pose: name,
    durationMs: LITTLE_COMPANION_POSES[name].durationMs
  }));
}

export function littleCompanionPoseStep(state, stepIndex, preview = false) {
  const timeline = littleCompanionTimeline(preview ? 'idle' : state);
  if (preview) return timeline[0];
  const rawStep = Number.isFinite(Number(stepIndex)) ? Math.trunc(Number(stepIndex)) : 0;
  const index = ((rawStep % timeline.length) + timeline.length) % timeline.length;
  return timeline[index];
}

export function nextLittleCompanionStep(state, currentStep) {
  const length = timelineFor(state).length;
  const rawStep = Number.isFinite(Number(currentStep)) ? Math.trunc(Number(currentStep)) : 0;
  return ((rawStep + 1) % length + length) % length;
}

export function littleCompanionFrame(poseName, frameNumber) {
  const definition = LITTLE_COMPANION_POSES[poseName] || LITTLE_COMPANION_POSES.idle;
  const rawFrame = Number.isFinite(Number(frameNumber)) ? Math.trunc(Number(frameNumber)) : 0;
  const frame = ((rawFrame % definition.frames) + definition.frames) % definition.frames;
  return {
    pose: definition.pose,
    src: definition.src,
    column: frame % definition.columns,
    row: Math.floor(frame / definition.columns),
    columns: definition.columns,
    rows: definition.rows,
    frameCount: definition.frames,
    frameDelayMs: Math.round(definition.durationMs / definition.frames)
  };
}

export function littleCompanionWalkFrame(frameNumber) {
  const rawFrame = Number.isFinite(Number(frameNumber)) ? Math.trunc(Number(frameNumber)) : 0;
  const index = ((rawFrame % WALK_FRAME_SEQUENCE.length) + WALK_FRAME_SEQUENCE.length) % WALK_FRAME_SEQUENCE.length;
  const frame = littleCompanionFrame('walk', WALK_FRAME_SEQUENCE[index]);
  return {
    ...frame,
    frameCount: WALK_FRAME_SEQUENCE.length,
    frameDelayMs: 120
  };
}

export function littleCompanionIdleFrameAfterWalk(frameNumber) {
  const rawFrame = Number.isFinite(Number(frameNumber)) ? Math.trunc(Number(frameNumber)) : 0;
  const index = ((rawFrame % WALK_TO_IDLE_FRAME.length) + WALK_TO_IDLE_FRAME.length) % WALK_TO_IDLE_FRAME.length;
  return WALK_TO_IDLE_FRAME[index];
}
