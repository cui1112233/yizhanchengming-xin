const PET_WIDTH = 120;
const PET_HEIGHT = 130;
const SAFE_PADDING = 8;
const MIN_TOP = 58;
const BOTTOM_ZONE_HEIGHT = 220;
const EDGE_TRIGGER_DISTANCE = 32;
const MIN_EDGE_VIEWPORT_WIDTH = 200;
const EDGE_SPRITE_OFFSET = 22;

const finite = (value, fallback = 0) => {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
};

const clamp = (value, min, max) => Math.max(min, Math.min(max, value));

function horizontalBounds(width) {
  if (width < PET_WIDTH + SAFE_PADDING * 2) {
    const centered = Math.max(0, (width - PET_WIDTH) / 2);
    return { minLeft: centered, maxLeft: centered };
  }
  return { minLeft: SAFE_PADDING, maxLeft: width - PET_WIDTH - SAFE_PADDING };
}

function viewportVerticalBounds(height) {
  if (height < PET_HEIGHT + MIN_TOP + SAFE_PADDING) {
    const centered = Math.max(0, (height - PET_HEIGHT) / 2);
    return { minTop: centered, maxTop: centered };
  }
  return { minTop: MIN_TOP, maxTop: height - PET_HEIGHT - SAFE_PADDING };
}

export function littleCompanionActivityBounds(viewport = {}) {
  const width = Math.max(0, finite(viewport.width));
  const height = Math.max(0, finite(viewport.height));
  const horizontal = horizontalBounds(width);
  const visibleVertical = viewportVerticalBounds(height);
  if (visibleVertical.minTop === visibleVertical.maxTop) {
    return { ...horizontal, ...visibleVertical };
  }
  return {
    ...horizontal,
    minTop: Math.max(visibleVertical.minTop, height - BOTTOM_ZONE_HEIGHT),
    maxTop: visibleVertical.maxTop
  };
}

export function clampLittleCompanionToViewport(position = {}, viewport = {}) {
  const width = Math.max(0, finite(viewport.width));
  const height = Math.max(0, finite(viewport.height));
  const horizontal = horizontalBounds(width);
  const vertical = viewportVerticalBounds(height);
  return {
    left: clamp(finite(position.left, horizontal.maxLeft), horizontal.minLeft, horizontal.maxLeft),
    top: clamp(finite(position.top, vertical.maxTop), vertical.minTop, vertical.maxTop)
  };
}

export function littleCompanionWalkTarget(position = {}, viewport = {}, distanceRoll = 0.5, directionRoll = 0.5) {
  const bounds = littleCompanionActivityBounds(viewport);
  const left = clamp(finite(position.left, bounds.minLeft), bounds.minLeft, bounds.maxLeft);
  const distance = 160 + clamp(finite(distanceRoll, 0.5), 0, 1) * 260;
  const direction = clamp(finite(directionRoll, 0.5), 0, 1) < 0.5 ? -1 : 1;
  let targetLeft = clamp(left + direction * distance, bounds.minLeft, bounds.maxLeft);
  if (targetLeft === left && bounds.minLeft !== bounds.maxLeft) {
    targetLeft = clamp(left - direction * distance, bounds.minLeft, bounds.maxLeft);
  }
  return {
    left: targetLeft,
    top: bounds.maxTop,
    direction: targetLeft >= left ? 1 : -1
  };
}

export function littleCompanionWalkDuration(from = {}, to = {}, speedRoll = 0.5) {
  const deltaLeft = finite(to.left) - finite(from.left);
  const deltaTop = finite(to.top) - finite(from.top);
  const distance = Math.hypot(deltaLeft, deltaTop);
  const speed = 90 + clamp(finite(speedRoll, 0.5), 0, 1) * 50;
  return Math.round(distance / speed * 1000);
}

export function littleCompanionEdgeSide(position = {}, viewport = {}) {
  const width = Math.max(0, finite(viewport.width));
  if (width < MIN_EDGE_VIEWPORT_WIDTH) return null;
  const left = finite(position.left, width - PET_WIDTH - SAFE_PADDING);
  const leftGap = left;
  const rightGap = width - (left + PET_WIDTH);
  const nearLeft = leftGap <= EDGE_TRIGGER_DISTANCE;
  const nearRight = rightGap <= EDGE_TRIGGER_DISTANCE;
  if (nearLeft && nearRight) return leftGap <= rightGap ? 'left' : 'right';
  if (nearLeft) return 'left';
  if (nearRight) return 'right';
  return null;
}

function nearestEdgeSide(position = {}, viewport = {}) {
  const width = Math.max(0, finite(viewport.width));
  const left = finite(position.left, width - PET_WIDTH - SAFE_PADDING);
  return left + PET_WIDTH / 2 <= width / 2 ? 'left' : 'right';
}

export function littleCompanionEdgeTarget(position = {}, viewport = {}, requestedSide) {
  const width = Math.max(0, finite(viewport.width));
  const height = Math.max(0, finite(viewport.height));
  const horizontal = horizontalBounds(width);
  const vertical = viewportVerticalBounds(height);
  const top = clamp(finite(position.top, vertical.maxTop), vertical.minTop, vertical.maxTop);
  const side = requestedSide === 'left' || requestedSide === 'right'
    ? requestedSide
    : nearestEdgeSide(position, viewport);
  return {
    side,
    left: side === 'left' ? horizontal.minLeft : horizontal.maxLeft,
    top
  };
}

export function littleCompanionEdgeSpriteOffset(side) {
  if (side === 'left') return -EDGE_SPRITE_OFFSET;
  if (side === 'right') return EDGE_SPRITE_OFFSET;
  return 0;
}

export function littleCompanionPeekTarget(position = {}, viewport = {}) {
  return littleCompanionEdgeTarget(
    position,
    viewport,
    littleCompanionEdgeSide(position, viewport) || nearestEdgeSide(position, viewport)
  );
}

export function littleCompanionInwardTarget(position = {}, viewport = {}, requestedSide, requestedDistance = 180) {
  const bounds = littleCompanionActivityBounds(viewport);
  const side = requestedSide === 'left' || requestedSide === 'right'
    ? requestedSide
    : nearestEdgeSide(position, viewport);
  const distance = Math.max(0, finite(requestedDistance, 180));
  const left = side === 'left'
    ? clamp(bounds.minLeft + distance, bounds.minLeft, bounds.maxLeft)
    : clamp(bounds.maxLeft - distance, bounds.minLeft, bounds.maxLeft);
  return { left, top: clamp(finite(position.top, bounds.maxTop), bounds.minTop, bounds.maxTop), direction: side === 'left' ? 1 : -1 };
}

export function interpolateLittleCompanionPosition(from = {}, to = {}, progress = 0) {
  const targetLeft = finite(to.left);
  const targetTop = finite(to.top);
  const startLeft = finite(from.left, targetLeft);
  const startTop = finite(from.top, targetTop);
  const amount = clamp(finite(progress), 0, 1);
  return {
    left: startLeft + (targetLeft - startLeft) * amount,
    top: startTop + (targetTop - startTop) * amount
  };
}

export function canLittleCompanionRoam(context = {}) {
  return context.petId === 'little-companion'
    && context.state === 'idle'
    && context.chatOpen !== true
    && context.dragging !== true
    && context.pageVisible !== false
    && context.reducedMotion !== true
    && context.spriteFailed !== true;
}
