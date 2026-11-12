import { normalizePersonSegmentation } from './aicutPersonSegmentation.js';
import { normalizeHSLAdjustments } from './aicut/hslColor.js';
import { createColorGradeLUTAtlas, hasColorGradeAdjustments, normalizeRGBColorCurves, normalizeColorWheels } from './aicut/colorGrade.js';

// One bounded cache avoids rebuilding an unchanged LUT when the renderer restarts.
let lastHSLAtlas = null;
function hslAtlas(color) {
  const signature = JSON.stringify({ hsl: normalizeHSLAdjustments(color.hsl), rgbCurves: normalizeRGBColorCurves(color.rgbCurves), wheels: normalizeColorWheels(color.wheels) });
  if (lastHSLAtlas?.signature !== signature) lastHSLAtlas = { signature, atlas: createColorGradeLUTAtlas(color) };
  return lastHSLAtlas.atlas;
}

function clamp(value, minimum, maximum, fallback) {
  const number = Number(value);
  return Number.isFinite(number) ? Math.min(maximum, Math.max(minimum, number)) : fallback;
}

const defaultCurvePoints = [{ x: 0, y: 0 }, { x: .25, y: .25 }, { x: .5, y: .5 }, { x: .75, y: .75 }, { x: 1, y: 1 }];
export const defaultPolygonMaskPoints = [{ x: .2, y: .2 }, { x: .8, y: .2 }, { x: .8, y: .8 }, { x: .2, y: .8 }];
export const defaultBezierMaskPoints = [
  { x: .25, y: .15, inX: .12, inY: .25, outX: .38, outY: .05 },
  { x: .75, y: .15, inX: .62, inY: .05, outX: .88, outY: .25 },
  { x: .75, y: .85, inX: .88, inY: .75, outX: .62, outY: .95 },
  { x: .25, y: .85, inX: .38, inY: .95, outX: .12, outY: .75 }
];

export function polygonArea(points = []) {
  if (!Array.isArray(points) || points.length < 3) return 0;
  return Math.abs(points.reduce((sum, point, index) => {
    const next = points[(index + 1) % points.length];
    return sum + Number(point.x) * Number(next.y) - Number(next.x) * Number(point.y);
  }, 0)) / 2;
}

export function normalizePolygonMaskPoints(points) {
  if (!Array.isArray(points) || points.length < 3 || points.length > 12) return defaultPolygonMaskPoints.map(point => ({ ...point }));
  const normalized = points.map(point => ({ x: Number(point?.x), y: Number(point?.y) }));
  const valid = normalized.every(point => Number.isFinite(point.x) && Number.isFinite(point.y) && point.x >= 0 && point.x <= 1 && point.y >= 0 && point.y <= 1)
    && polygonArea(normalized) >= .0001;
  return valid ? normalized : defaultPolygonMaskPoints.map(point => ({ ...point }));
}

export function normalizeBezierMaskPoints(points) {
  if (!Array.isArray(points) || points.length < 3 || points.length > 8) return defaultBezierMaskPoints.map(point => ({ ...point }));
  const keys = ['x', 'y', 'inX', 'inY', 'outX', 'outY'];
  const normalized = points.map(point => Object.fromEntries(keys.map(key => [key, Number(point?.[key])])));
  const valid = normalized.every(point => keys.every(key => Number.isFinite(point[key]) && point[key] >= 0 && point[key] <= 1));
  return valid ? normalized : defaultBezierMaskPoints.map(point => ({ ...point }));
}

function cubicCoordinate(start, controlA, controlB, end, ratio) {
  const inverse = 1 - ratio;
  return inverse ** 3 * start + 3 * inverse ** 2 * ratio * controlA + 3 * inverse * ratio ** 2 * controlB + ratio ** 3 * end;
}

export function sampleBezierMaskPoints(points, samplesPerSegment = 6) {
  const anchors = normalizeBezierMaskPoints(points);
  const samples = Math.max(2, Math.min(8, Math.round(Number(samplesPerSegment) || 6)));
  return anchors.flatMap((anchor, index) => {
    const next = anchors[(index + 1) % anchors.length];
    return Array.from({ length: samples }, (_, step) => {
      const ratio = step / samples;
      return {
        x: cubicCoordinate(anchor.x, anchor.outX, next.inX, next.x, ratio),
        y: cubicCoordinate(anchor.y, anchor.outY, next.inY, next.y, ratio)
      };
    });
  });
}

export function normalizeColorCurvePoints(value = {}) {
  const legacy = Array.isArray(value.curve) && value.curve.length === 5
    ? value.curve.map((y, index) => ({ x: index / 4, y: clamp(y, 0, 1, index / 4) }))
    : null;
  if (!Array.isArray(value.curvePoints) || value.curvePoints.length < 2 || value.curvePoints.length > 16) return legacy || defaultCurvePoints.map(point => ({ ...point }));
  const points = value.curvePoints.map(point => ({ x: Number(point?.x), y: Number(point?.y) }));
  const valid = points.every((point, index) => Number.isFinite(point.x) && Number.isFinite(point.y) && point.x >= 0 && point.x <= 1 && point.y >= 0 && point.y <= 1 && (index === 0 || point.x > points[index - 1].x))
    && points[0].x === 0 && points.at(-1).x === 1;
  return valid ? points : (legacy || defaultCurvePoints.map(point => ({ ...point })));
}

export function sampleColorCurve(points, x) {
  const value = Math.min(1, Math.max(0, Number(x) || 0));
  for (let index = 1; index < points.length; index += 1) {
    const left = points[index - 1];
    const right = points[index];
    if (value <= right.x) {
      const ratio = (value - left.x) / Math.max(.000001, right.x - left.x);
      return left.y + (right.y - left.y) * ratio;
    }
  }
  return points.at(-1)?.y ?? value;
}

function normalizeColor(value = {}) {
  const curvePoints = normalizeColorCurvePoints(value);
  const curve = [0, .25, .5, .75, 1].map(x => sampleColorCurve(curvePoints, x));
  return {
    exposure: clamp(value.exposure, -2, 2, 0),
    temperature: clamp(value.temperature, -1, 1, 0),
    tint: clamp(value.tint, -1, 1, 0),
    highlights: clamp(value.highlights, -1, 1, 0),
    shadows: clamp(value.shadows, -1, 1, 0),
    gamma: clamp(value.gamma, .5, 2, 1),
    hsl: normalizeHSLAdjustments(value.hsl),
    ...(value.rgbCurves === undefined ? {} : { rgbCurves: normalizeRGBColorCurves(value.rgbCurves) }),
    ...(value.wheels === undefined ? {} : { wheels: normalizeColorWheels(value.wheels) }),
    curve,
    curvePoints
  };
}

export function normalizeVisualProcessing(value = {}) {
  const chroma = value?.chromaKey || {};
  const mask = value?.mask || {};
  const objectRemoval = value?.objectRemoval || {};
  const hasObjectRemovalRegion = Boolean(objectRemoval.enabled) || (Number(objectRemoval.width) > 0 && Number(objectRemoval.height) > 0);
  const geometry = value?.geometry || {};
  const details = value?.details || {};
  const stabilization = value?.stabilization || {};
  const color = /^#[0-9a-f]{6}$/i.test(String(chroma.color || '')) ? String(chroma.color).toLowerCase() : '#00ff00';
  const maskType = ['none', 'rectangle', 'ellipse', 'polygon', 'bezier'].includes(mask.type) ? mask.type : 'none';
  const bezierPoints = maskType === 'bezier' ? normalizeBezierMaskPoints(mask.bezierPoints) : [];
  const frameInterpolation = ['off', 'blend', 'optical-flow'].includes(value?.frameInterpolation) ? value.frameInterpolation : 'off';
  return {
    color: normalizeColor(value?.color),
    chromaKey: {
      enabled: Boolean(chroma.enabled),
      removeBlack: Boolean(chroma.removeBlack),
      color,
      similarity: clamp(chroma.similarity, .01, 1, .2),
      blend: clamp(chroma.blend, 0, 1, .08)
    },
    mask: {
      type: maskType,
      x: clamp(mask.x, 0, 1, .5),
      y: clamp(mask.y, 0, 1, .5),
      width: clamp(mask.width, .05, 1, 1),
      height: clamp(mask.height, .05, 1, 1),
      feather: clamp(mask.feather, 0, .5, 0),
      invert: Boolean(mask.invert),
      points: maskType === 'polygon' ? normalizePolygonMaskPoints(mask.points) : maskType === 'bezier' ? sampleBezierMaskPoints(bezierPoints) : [],
      bezierPoints
    },
    geometry: {
      cropLeft: clamp(geometry.cropLeft, 0, .9, 0),
      cropTop: clamp(geometry.cropTop, 0, .9, 0),
      cropRight: clamp(geometry.cropRight, 0, .9, 0),
      cropBottom: clamp(geometry.cropBottom, 0, .9, 0),
      flipHorizontal: Boolean(geometry.flipHorizontal),
      flipVertical: Boolean(geometry.flipVertical)
    },
    details: {
      sharpen: clamp(details.sharpen, 0, 1, 0),
      vignette: clamp(details.vignette, 0, 1, 0),
      grain: clamp(details.grain, 0, 1, 0)
    },
    frameInterpolation,
    stabilization: {
      enabled: Boolean(stabilization.enabled),
      smoothness: Math.round(clamp(stabilization.smoothness, 1, 100, 15)),
      zoom: clamp(stabilization.zoom, 0, 30, 0),
      tripod: Boolean(stabilization.tripod)
    },
    objectRemoval: {
      enabled: Boolean(objectRemoval.enabled),
      x: clamp(hasObjectRemovalRegion ? objectRemoval.x : undefined, 0, 1, .5),
      y: clamp(hasObjectRemovalRegion ? objectRemoval.y : undefined, 0, 1, .5),
      width: clamp(hasObjectRemovalRegion ? objectRemoval.width : undefined, .05, 1, .2),
      height: clamp(hasObjectRemovalRegion ? objectRemoval.height : undefined, .05, 1, .2)
    },
    personSegmentation: normalizePersonSegmentation(value?.personSegmentation)
  };
}

function byteHex(value) {
  return Math.min(255, Math.max(0, Math.round(Number(value) || 0))).toString(16).padStart(2, '0');
}

export function sampleChromaColorFromPixels(pixels, width, height, x, y) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  const column = Math.floor(Number(x));
  const row = Math.floor(Number(y));
  if (!pixels || columns < 1 || rows < 1 || column < 0 || row < 0 || column >= columns || row >= rows) return null;
  const index = (row * columns + column) * 4;
  if (pixels.length < index + 4 || Number(pixels[index + 3]) === 0) return null;
  return `#${byteHex(pixels[index])}${byteHex(pixels[index + 1])}${byteHex(pixels[index + 2])}`;
}

export function detectChromaScreenColor(pixels, width, height) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  if (!pixels || columns < 1 || rows < 1 || pixels.length < columns * rows * 4) return null;
  const edgeX = Math.max(1, Math.ceil(columns * .2));
  const edgeY = Math.max(1, Math.ceil(rows * .2));
  const buckets = new Map();
  for (let y = 0; y < rows; y += 1) {
    for (let x = 0; x < columns; x += 1) {
      if (x >= edgeX && x < columns - edgeX && y >= edgeY && y < rows - edgeY) continue;
      const index = (y * columns + x) * 4;
      if (Number(pixels[index + 3]) < 128) continue;
      const rgb = [pixels[index], pixels[index + 1], pixels[index + 2]].map(value => Math.floor(Number(value) / 16) * 16);
      const key = rgb.join(',');
      const current = buckets.get(key) || { count: 0, rgb };
      current.count += 1;
      buckets.set(key, current);
    }
  }
  const ranked = [...buckets.values()].sort((left, right) => {
    const leftSaturation = Math.max(...left.rgb) - Math.min(...left.rgb);
    const rightSaturation = Math.max(...right.rgb) - Math.min(...right.rgb);
    return right.count - left.count || rightSaturation - leftSaturation;
  });
  const winner = ranked[0];
  if (!winner) return null;
  return `#${winner.rgb.map(byteHex).join('')}`;
}

function hexRGB(value) {
  return [1, 3, 5].map(index => Number.parseInt(value.slice(index, index + 2), 16) / 255);
}

export function visualProcessingPreviewPlan(value = {}) {
  const processing = normalizeVisualProcessing(value);
  return {
    ...processing.color,
    curve: processing.color.curve,
    curvePoints: processing.color.curvePoints,
    exposureMultiplier: 2 ** processing.color.exposure,
    chromaKey: processing.chromaKey,
    keyRGB: hexRGB(processing.chromaKey.removeBlack ? '#000000' : processing.chromaKey.color),
    mask: processing.mask,
    geometry: processing.geometry,
    details: processing.details,
    frameInterpolation: processing.frameInterpolation,
    stabilization: processing.stabilization,
    objectRemoval: processing.objectRemoval
  };
}

export function estimateFrameTranslation(previous, current, width, height, maxShift = 4) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  const radius = Math.max(0, Math.min(12, Math.floor(Number(maxShift) || 0)));
  if (!previous || !current || columns < 4 || rows < 4 || previous.length < columns * rows || current.length < columns * rows) {
    return { x: 0, y: 0, confidence: 0, score: Number.POSITIVE_INFINITY };
  }
  let best = { x: 0, y: 0, score: Number.POSITIVE_INFINITY };
  let secondScore = Number.POSITIVE_INFINITY;
  for (let offsetY = -radius; offsetY <= radius; offsetY += 1) {
    for (let offsetX = -radius; offsetX <= radius; offsetX += 1) {
      let difference = 0;
      let samples = 0;
      const startX = Math.max(radius, -offsetX);
      const endX = Math.min(columns - radius, columns - offsetX);
      const startY = Math.max(radius, -offsetY);
      const endY = Math.min(rows - radius, rows - offsetY);
      for (let y = startY; y < endY; y += 2) {
        for (let x = startX; x < endX; x += 2) {
          difference += Math.abs(Number(previous[y * columns + x]) - Number(current[(y + offsetY) * columns + x + offsetX]));
          samples += 1;
        }
      }
      const score = samples ? difference / samples : Number.POSITIVE_INFINITY;
      const distance = Math.abs(offsetX) + Math.abs(offsetY);
      const bestDistance = Math.abs(best.x) + Math.abs(best.y);
      if (score < best.score - .000001 || (Math.abs(score - best.score) <= .000001 && distance < bestDistance)) {
        secondScore = best.score;
        best = { x: offsetX, y: offsetY, score };
      } else if (score < secondScore) secondScore = score;
    }
  }
  const confidence = Number.isFinite(secondScore) ? Math.max(0, (secondScore - best.score) / Math.max(1, secondScore)) : 0;
  return { ...best, confidence };
}

export function estimateRegionTranslation(previous, current, width, height, region = {}, maxShift = 8) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  const radius = Math.max(1, Math.min(24, Math.floor(Number(maxShift) || 8)));
  if (!previous || !current || columns < 4 || rows < 4 || previous.length < columns * rows || current.length < columns * rows) {
    return { x: 0, y: 0, confidence: 0, score: Number.POSITIVE_INFINITY };
  }
  const left = Math.max(0, Math.min(columns - 2, Math.round(clamp(region.x, 0, 1, .25) * columns)));
  const top = Math.max(0, Math.min(rows - 2, Math.round(clamp(region.y, 0, 1, .25) * rows)));
  const regionWidth = Math.max(2, Math.min(columns - left, Math.round(clamp(region.width, .02, 1, .5) * columns)));
  const regionHeight = Math.max(2, Math.min(rows - top, Math.round(clamp(region.height, .02, 1, .5) * rows)));
  let best = { x: 0, y: 0, score: Number.POSITIVE_INFINITY };
  let secondScore = Number.POSITIVE_INFINITY;
  for (let offsetY = -radius; offsetY <= radius; offsetY += 1) {
    for (let offsetX = -radius; offsetX <= radius; offsetX += 1) {
      if (left + offsetX < 0 || top + offsetY < 0 || left + regionWidth + offsetX > columns || top + regionHeight + offsetY > rows) continue;
      let difference = 0;
      let samples = 0;
      for (let y = top; y < top + regionHeight; y += 1) {
        for (let x = left; x < left + regionWidth; x += 1) {
          difference += Math.abs(Number(previous[y * columns + x]) - Number(current[(y + offsetY) * columns + x + offsetX]));
          samples += 1;
        }
      }
      const score = samples ? difference / samples : Number.POSITIVE_INFINITY;
      const distance = Math.abs(offsetX) + Math.abs(offsetY);
      const bestDistance = Math.abs(best.x) + Math.abs(best.y);
      if (score < best.score - .000001 || (Math.abs(score - best.score) <= .000001 && distance < bestDistance)) {
        secondScore = best.score;
        best = { x: offsetX, y: offsetY, score };
      } else if (score < secondScore) secondScore = score;
    }
  }
  const confidence = Number.isFinite(secondScore) ? Math.max(0, (secondScore - best.score) / Math.max(1, secondScore)) : 0;
  return { ...best, confidence };
}

export function trackRegionSequence(frames = [], width, height, region = {}, options = {}) {
  const columns = Math.floor(Number(width));
  const rows = Math.floor(Number(height));
  if (!Array.isArray(frames) || frames.length < 2 || columns < 4 || rows < 4) return [];
  const firstTime = Number(frames[0]?.timeMs);
  if (!Number.isFinite(firstTime)) return [];
  const maxShift = Math.max(1, Math.min(24, Math.floor(Number(options.maxShift) || 8)));
  const minimumConfidence = clamp(options.minimumConfidence, 0, 1, .01);
  const currentRegion = {
    x: clamp(region.x, 0, 1, .25),
    y: clamp(region.y, 0, 1, .25),
    width: clamp(region.width, .02, 1, .5),
    height: clamp(region.height, .02, 1, .5)
  };
  let cumulativeX = 0;
  let cumulativeY = 0;
  const result = [{ offsetMs: 0, x: 0, y: 0, confidence: 1 }];
  for (let index = 1; index < frames.length; index += 1) {
    const previous = frames[index - 1]?.pixels;
    const current = frames[index]?.pixels;
    const timeMs = Number(frames[index]?.timeMs);
    if (!Number.isFinite(timeMs) || timeMs < firstTime || !previous || !current) continue;
    const motion = estimateRegionTranslation(previous, current, columns, rows, currentRegion, maxShift);
    const accepted = motion.confidence >= minimumConfidence || motion.score === 0;
    if (accepted) {
      cumulativeX += motion.x / columns;
      cumulativeY += motion.y / rows;
      currentRegion.x = clamp(currentRegion.x + motion.x / columns, 0, Math.max(0, 1 - currentRegion.width), currentRegion.x);
      currentRegion.y = clamp(currentRegion.y + motion.y / rows, 0, Math.max(0, 1 - currentRegion.height), currentRegion.y);
    }
    result.push({
      offsetMs: Math.round(timeMs - firstTime),
      x: Math.max(-1, Math.min(1, cumulativeX)),
      y: Math.max(-1, Math.min(1, cumulativeY)),
      confidence: accepted ? motion.confidence : 0
    });
  }
  return result;
}

export function createPreviewStabilizer({ smoothness = 15, tripod = false } = {}) {
  const smoothing = Math.max(1, Math.min(100, Math.round(Number(smoothness) || 15)));
  const alpha = 2 / (smoothing + 1);
  let pathX = 0;
  let pathY = 0;
  let smoothX = 0;
  let smoothY = 0;
  return {
    update(motion = {}) {
      pathX += Number(motion.x) || 0;
      pathY += Number(motion.y) || 0;
      if (!tripod) {
        smoothX += (pathX - smoothX) * alpha;
        smoothY += (pathY - smoothY) * alpha;
      }
      return { x: pathX - smoothX, y: pathY - smoothY };
    },
    reset() {
      pathX = 0;
      pathY = 0;
      smoothX = 0;
      smoothY = 0;
    }
  };
}

export function visualProcessingCSS(value = {}) {
  const plan = visualProcessingPreviewPlan(value);
  const brightness = plan.exposureMultiplier * (1 + plan.shadows * .12 + plan.highlights * .08);
  const contrast = 1 + plan.highlights * .18 - plan.shadows * .12;
  const sepia = Math.abs(plan.temperature) * .18 + Math.abs(plan.tint) * .08;
  const hue = plan.temperature * -12 + plan.tint * 20;
  const style = { filter: `brightness(${brightness}) contrast(${contrast}) sepia(${sepia}) hue-rotate(${hue}deg)` };
  if (plan.mask.type === 'rectangle') {
    const left = Math.max(0, (plan.mask.x - plan.mask.width / 2) * 100);
    const right = Math.max(0, (1 - plan.mask.x - plan.mask.width / 2) * 100);
    const top = Math.max(0, (plan.mask.y - plan.mask.height / 2) * 100);
    const bottom = Math.max(0, (1 - plan.mask.y - plan.mask.height / 2) * 100);
    style.clipPath = `inset(${top}% ${right}% ${bottom}% ${left}%)`;
  } else if (plan.mask.type === 'ellipse') {
    style.clipPath = `ellipse(${plan.mask.width * 50}% ${plan.mask.height * 50}% at ${plan.mask.x * 100}% ${plan.mask.y * 100}%)`;
  } else if (['polygon', 'bezier'].includes(plan.mask.type)) {
    style.clipPath = `polygon(${plan.mask.points.map(point => `${point.x * 100}% ${point.y * 100}%`).join(', ')})`;
  }
  return style;
}

function compileShader(gl, type, source) {
  const shader = gl.createShader(type);
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    gl.deleteShader(shader);
    return null;
  }
  return shader;
}

export function startVisualProcessingRenderer(canvas, mediaElement, value = {}) {
  const gl = canvas?.getContext?.('webgl', { alpha: true, antialias: false, preserveDrawingBuffer: false });
  if (!gl || !mediaElement) return null;
  const vertex = compileShader(gl, gl.VERTEX_SHADER, `
    attribute vec2 a_position;
    varying vec2 v_uv;
    void main(){ v_uv=(a_position+1.0)*0.5; gl_Position=vec4(a_position,0.0,1.0); }
  `);
  const fragment = compileShader(gl, gl.FRAGMENT_SHADER, `
    precision highp float;
    varying vec2 v_uv;
    uniform sampler2D u_texture;
    uniform sampler2D u_hsl_lut;
    uniform float u_hsl_enabled;
    uniform vec3 u_key;
    uniform float u_chroma;
    uniform float u_similarity;
    uniform float u_blend;
    uniform float u_exposure;
    uniform float u_temperature;
    uniform float u_tint;
    uniform float u_highlights;
    uniform float u_shadows;
    uniform float u_gamma;
    uniform vec2 u_texel;
    uniform float u_sharpen;
    uniform float u_vignette;
    uniform float u_grain;
    uniform vec2 u_curve_points[16];
    uniform float u_curve_count;
    uniform vec4 u_crop;
	uniform vec2 u_stabilization_offset;
	uniform float u_stabilization_scale;
	uniform float u_flip_horizontal;
	uniform float u_flip_vertical;
    uniform float u_mask_type;
    uniform vec2 u_mask_center;
    uniform vec2 u_mask_size;
    uniform float u_mask_feather;
    uniform float u_mask_invert;
    uniform vec2 u_mask_points[64];
    uniform float u_mask_point_count;
    uniform float u_object_removal;
    uniform vec2 u_object_center;
    uniform vec2 u_object_size;
    vec3 applyHSL(vec3 value){
      if(u_hsl_enabled<0.5) return value;
      vec3 index=clamp(value,0.0,1.0)*32.0;
      float lower=floor(index.b);
      float upper=min(32.0,lower+1.0);
      vec2 uv0=vec2((index.r+lower*33.0+0.5)/1089.0,(index.g+0.5)/33.0);
      vec2 uv1=vec2((index.r+upper*33.0+0.5)/1089.0,(index.g+0.5)/33.0);
      return mix(texture2D(u_hsl_lut,uv0).rgb,texture2D(u_hsl_lut,uv1).rgb,index.b-lower);
    }
    vec2 sourceUV(vec2 viewUV){
      vec2 mapped=(viewUV-vec2(0.5))/max(1.0,u_stabilization_scale)+vec2(0.5)+u_stabilization_offset;
      if(u_flip_horizontal>0.5) mapped.x=1.0-mapped.x;
      if(u_flip_vertical>0.5) mapped.y=1.0-mapped.y;
      mapped.x=u_crop.x+mapped.x*(1.0-u_crop.x-u_crop.z);
      mapped.y=u_crop.w+mapped.y*(1.0-u_crop.y-u_crop.w);
      return mapped;
    }
    float applyCurve(float value){
      value=clamp(value,0.0,1.0);
      vec2 left=u_curve_points[0];
      for(int i=1;i<16;i++){
        if(float(i)>=u_curve_count) break;
        vec2 right=u_curve_points[i];
        if(value<=right.x) return mix(left.y,right.y,(value-left.x)/max(0.000001,right.x-left.x));
        left=right;
      }
      return left.y;
    }
    float segmentDistance(vec2 point, vec2 start, vec2 end){
      vec2 segment=end-start;
      float ratio=clamp(dot(point-start,segment)/max(0.000001,dot(segment,segment)),0.0,1.0);
      return length(point-(start+segment*ratio));
    }
    float polygonMask(vec2 point){
      float inside=0.0;
      float distanceToEdge=10.0;
      vec2 first=u_mask_points[0];
      vec2 previous=first;
      for(int i=1;i<64;i++){
        if(float(i)>=u_mask_point_count) break;
        vec2 current=u_mask_points[i];
        bool crosses=((current.y>point.y)!=(previous.y>point.y)) && (point.x<(previous.x-current.x)*(point.y-current.y)/(previous.y-current.y)+current.x);
        if(crosses) inside=1.0-inside;
        distanceToEdge=min(distanceToEdge,segmentDistance(point,previous,current));
        previous=current;
      }
      bool closes=((first.y>point.y)!=(previous.y>point.y)) && (point.x<(previous.x-first.x)*(point.y-first.y)/(previous.y-first.y)+first.x);
      if(closes) inside=1.0-inside;
      distanceToEdge=min(distanceToEdge,segmentDistance(point,previous,first));
      if(u_mask_feather<=0.0001) return inside;
      float signedDistance=inside>0.5?distanceToEdge:-distanceToEdge;
      return smoothstep(-u_mask_feather,u_mask_feather,signedDistance);
    }
    void main(){
	  vec2 sampleUV=sourceUV(v_uv);
	  vec4 pixel=texture2D(u_texture,sampleUV);
      vec3 sourceColor=pixel.rgb;
      vec2 objectHalf=u_object_size*0.5;
      vec2 objectDelta=abs(v_uv-u_object_center);
      if(u_object_removal>0.5 && objectDelta.x<objectHalf.x && objectDelta.y<objectHalf.y){
        float left=u_object_center.x-objectHalf.x;
        float right=u_object_center.x+objectHalf.x;
        float bottom=u_object_center.y-objectHalf.y;
        float top=u_object_center.y+objectHalf.y;
        float horizontalRatio=clamp((v_uv.x-left)/max(0.0001,u_object_size.x),0.0,1.0);
        float verticalRatio=clamp((v_uv.y-bottom)/max(0.0001,u_object_size.y),0.0,1.0);
        vec3 horizontal=mix(texture2D(u_texture,sourceUV(vec2(max(0.0,left-u_texel.x*2.0),v_uv.y))).rgb,texture2D(u_texture,sourceUV(vec2(min(1.0,right+u_texel.x*2.0),v_uv.y))).rgb,horizontalRatio);
        vec3 vertical=mix(texture2D(u_texture,sourceUV(vec2(v_uv.x,max(0.0,bottom-u_texel.y*2.0)))).rgb,texture2D(u_texture,sourceUV(vec2(v_uv.x,min(1.0,top+u_texel.y*2.0)))).rgb,verticalRatio);
        sourceColor=mix(horizontal,vertical,0.5);
        pixel.rgb=sourceColor;
      }
      vec3 original=sourceColor;
      if(u_sharpen>0.0001){
        vec3 neighbors=texture2D(u_texture,sampleUV+vec2(u_texel.x,0.0)).rgb+texture2D(u_texture,sampleUV-vec2(u_texel.x,0.0)).rgb+texture2D(u_texture,sampleUV+vec2(0.0,u_texel.y)).rgb+texture2D(u_texture,sampleUV-vec2(0.0,u_texel.y)).rgb;
        original=clamp(sourceColor*(1.0+2.4*u_sharpen)-neighbors*(0.6*u_sharpen),0.0,1.0);
      }
      float luma=dot(original,vec3(0.2126,0.7152,0.0722));
      vec3 color=original*pow(2.0,u_exposure);
      color.r+=u_temperature*0.10; color.b-=u_temperature*0.10; color.g+=u_tint*0.08;
      color+=u_shadows*0.18*(1.0-smoothstep(0.0,0.55,luma));
      color+=u_highlights*0.15*smoothstep(0.45,1.0,luma);
      color=pow(max(color,vec3(0.0)),vec3(1.0/max(0.5,u_gamma)));
      color=vec3(applyCurve(color.r),applyCurve(color.g),applyCurve(color.b));
      if(pixel.a>0.0) color=applyHSL(color);
      float vignetteDistance=length((v_uv-vec2(0.5))*vec2(1.0,0.78));
      color*=1.0-u_vignette*smoothstep(0.22,0.62,vignetteDistance)*0.82;
      float grainNoise=fract(sin(dot(v_uv*vec2(1733.0,941.0),vec2(12.9898,78.233)))*43758.5453);
      color+=(grainNoise-0.5)*u_grain*0.18;
      float alpha=pixel.a;
      if(u_chroma>0.5){ float d=distance(sourceColor,u_key); alpha*=smoothstep(u_similarity,u_similarity+max(0.001,u_blend),d); }
      float inside=1.0;
      if(u_mask_type>0.5 && u_mask_type<1.5){
        vec2 halfSize=u_mask_size*0.5;
        float edge=min(halfSize.x-abs(v_uv.x-u_mask_center.x),halfSize.y-abs(v_uv.y-u_mask_center.y));
        inside=u_mask_feather>0.0001?smoothstep(-u_mask_feather,u_mask_feather,edge):step(0.0,edge);
      } else if(u_mask_type>=1.5 && u_mask_type<2.5){
        vec2 halfSize=max(u_mask_size*0.5,vec2(0.001));
        float radius=length((v_uv-u_mask_center)/halfSize);
        float edge=(1.0-radius)*min(halfSize.x,halfSize.y);
        inside=u_mask_feather>0.0001?smoothstep(-u_mask_feather,u_mask_feather,edge):step(radius,1.0);
      } else if(u_mask_type>=2.5){
        inside=polygonMask(v_uv);
      }
      if(u_mask_invert>0.5) inside=1.0-inside;
      gl_FragColor=vec4(clamp(color,0.0,1.0),alpha*inside);
    }
  `);
  if (!vertex || !fragment) return null;
  const program = gl.createProgram();
  gl.attachShader(program, vertex);
  gl.attachShader(program, fragment);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) return null;
  gl.useProgram(program);
  const buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1,-1,1,-1,-1,1,-1,1,1,-1,1,1]), gl.STATIC_DRAW);
  const position = gl.getAttribLocation(program, 'a_position');
  gl.enableVertexAttribArray(position);
  gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0);
  const texture = gl.createTexture();
  gl.activeTexture(gl.TEXTURE0);
  gl.bindTexture(gl.TEXTURE_2D, texture);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);
  const plan = visualProcessingPreviewPlan(value);
  const uniform1 = (name, number) => gl.uniform1f(gl.getUniformLocation(program, name), number);
  const hslTexture = gl.createTexture();
  gl.activeTexture(gl.TEXTURE1);
  gl.bindTexture(gl.TEXTURE_2D, hslTexture);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  const hslEnabled = hasColorGradeAdjustments(plan);
  // Typed atlas rows already have their intended orientation; media flips separately.
  gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, false);
  if (hslEnabled) {
    const atlas = hslAtlas(plan);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, atlas.width, atlas.height, 0, gl.RGBA, gl.UNSIGNED_BYTE, atlas.pixels);
  } else {
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, new Uint8Array([0, 0, 0, 255]));
  }
  gl.uniform1i(gl.getUniformLocation(program, 'u_texture'), 0);
  gl.uniform1i(gl.getUniformLocation(program, 'u_hsl_lut'), 1);
  uniform1('u_hsl_enabled', hslEnabled ? 1 : 0);
  gl.activeTexture(gl.TEXTURE0);
  gl.bindTexture(gl.TEXTURE_2D, texture);
  gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);
  gl.uniform3fv(gl.getUniformLocation(program, 'u_key'), plan.keyRGB);
  uniform1('u_chroma', plan.chromaKey.enabled || plan.chromaKey.removeBlack ? 1 : 0);
  uniform1('u_similarity', plan.chromaKey.similarity);
  uniform1('u_blend', plan.chromaKey.blend);
  uniform1('u_exposure', plan.exposure);
  uniform1('u_temperature', plan.temperature);
  uniform1('u_tint', plan.tint);
  uniform1('u_highlights', plan.highlights);
  uniform1('u_shadows', plan.shadows);
  uniform1('u_gamma', plan.gamma);
  uniform1('u_sharpen', plan.details.sharpen);
  uniform1('u_vignette', plan.details.vignette);
  uniform1('u_grain', plan.details.grain);
  const texelLocation = gl.getUniformLocation(program, 'u_texel');
  gl.uniform2f(texelLocation, 1 / 1280, 1 / 720);
  const paddedCurvePoints = Array.from({ length: 16 }, (_, index) => plan.curvePoints[Math.min(index, plan.curvePoints.length - 1)]);
  gl.uniform2fv(gl.getUniformLocation(program, 'u_curve_points[0]'), new Float32Array(paddedCurvePoints.flatMap(point => [point.x, point.y])));
  uniform1('u_curve_count', plan.curvePoints.length);
	gl.uniform4f(gl.getUniformLocation(program, 'u_crop'), plan.geometry.cropLeft, plan.geometry.cropTop, plan.geometry.cropRight, plan.geometry.cropBottom);
	const stabilizationOffsetLocation = gl.getUniformLocation(program, 'u_stabilization_offset');
	gl.uniform2f(stabilizationOffsetLocation, 0, 0);
	uniform1('u_stabilization_scale', plan.stabilization.enabled ? 1 + plan.stabilization.zoom / 100 : 1);
	uniform1('u_flip_horizontal', plan.geometry.flipHorizontal ? 1 : 0);
	uniform1('u_flip_vertical', plan.geometry.flipVertical ? 1 : 0);
  uniform1('u_mask_type', plan.mask.type === 'rectangle' ? 1 : plan.mask.type === 'ellipse' ? 2 : ['polygon', 'bezier'].includes(plan.mask.type) ? 3 : 0);
  gl.uniform2f(gl.getUniformLocation(program, 'u_mask_center'), plan.mask.x, 1 - plan.mask.y);
  gl.uniform2f(gl.getUniformLocation(program, 'u_mask_size'), plan.mask.width, plan.mask.height);
  uniform1('u_mask_feather', plan.mask.feather);
  uniform1('u_mask_invert', plan.mask.invert ? 1 : 0);
  const polygonPoints = ['polygon', 'bezier'].includes(plan.mask.type) ? plan.mask.points : defaultPolygonMaskPoints;
  const paddedMaskPoints = Array.from({ length: 64 }, (_, index) => polygonPoints[Math.min(index, polygonPoints.length - 1)]);
  gl.uniform2fv(gl.getUniformLocation(program, 'u_mask_points[0]'), new Float32Array(paddedMaskPoints.flatMap(point => [point.x, 1 - point.y])));
  uniform1('u_mask_point_count', polygonPoints.length);
  uniform1('u_object_removal', plan.objectRemoval.enabled ? 1 : 0);
  gl.uniform2f(gl.getUniformLocation(program, 'u_object_center'), plan.objectRemoval.x, 1 - plan.objectRemoval.y);
  gl.uniform2f(gl.getUniformLocation(program, 'u_object_size'), plan.objectRemoval.width, plan.objectRemoval.height);
  let frame = 0;
  let sampledFrame = 0;
  let stopped = false;
  let previousLuma = null;
  let previousMediaTime = null;
  const stabilizer = createPreviewStabilizer(plan.stabilization);
  const analysisCanvas = plan.stabilization.enabled && mediaElement.tagName === 'VIDEO' && typeof document !== 'undefined' ? document.createElement('canvas') : null;
  const analysisContext = analysisCanvas?.getContext('2d', { alpha: false, willReadFrequently: true }) || null;
  const resetStabilization = () => {
    previousLuma = null;
    previousMediaTime = null;
    stabilizer.reset();
    gl.uniform2f(stabilizationOffsetLocation, 0, 0);
  };
  const updateStabilization = (width, height) => {
    if (!analysisContext || !analysisCanvas || mediaElement.readyState < 2) return;
    const mediaTime = Number(mediaElement.currentTime) || 0;
    if (previousMediaTime !== null && (mediaTime < previousMediaTime || Math.abs(mediaTime - previousMediaTime) > .35)) resetStabilization();
    previousMediaTime = mediaTime;
    const analysisWidth = 96;
    const analysisHeight = Math.max(32, Math.round(analysisWidth * height / width));
    if (analysisCanvas.width !== analysisWidth || analysisCanvas.height !== analysisHeight) {
      analysisCanvas.width = analysisWidth;
      analysisCanvas.height = analysisHeight;
      previousLuma = null;
    }
    analysisContext.drawImage(mediaElement, 0, 0, analysisWidth, analysisHeight);
    const rgba = analysisContext.getImageData(0, 0, analysisWidth, analysisHeight).data;
    const luma = new Uint8Array(analysisWidth * analysisHeight);
    for (let index = 0, pixel = 0; index < rgba.length; index += 4, pixel += 1) {
      luma[pixel] = Math.round(rgba[index] * .2126 + rgba[index + 1] * .7152 + rgba[index + 2] * .0722);
    }
    if (previousLuma) {
      const motion = estimateFrameTranslation(previousLuma, luma, analysisWidth, analysisHeight, 4);
      const correction = motion.confidence > .002 ? stabilizer.update(motion) : stabilizer.update({ x: 0, y: 0 });
      const maximum = Math.min(.12, .01 + plan.stabilization.zoom / 200);
      const offsetX = Math.max(-maximum, Math.min(maximum, correction.x / analysisWidth));
      const offsetY = Math.max(-maximum, Math.min(maximum, -correction.y / analysisHeight));
      gl.uniform2f(stabilizationOffsetLocation, offsetX, offsetY);
    }
    previousLuma = luma;
  };
  const draw = () => {
    if (stopped) return;
    const width = Number(mediaElement.videoWidth || mediaElement.naturalWidth || 0);
    const height = Number(mediaElement.videoHeight || mediaElement.naturalHeight || 0);
    if (width > 0 && height > 0 && (mediaElement.tagName !== 'VIDEO' || mediaElement.readyState >= 2)) {
      const scale = Math.min(1, 1280 / width, 720 / height);
      const targetWidth = Math.max(2, Math.round(width * scale));
      const targetHeight = Math.max(2, Math.round(height * scale));
      if (canvas.width !== targetWidth || canvas.height !== targetHeight) { canvas.width = targetWidth; canvas.height = targetHeight; gl.viewport(0, 0, targetWidth, targetHeight); }
      gl.uniform2f(texelLocation, 1 / targetWidth, 1 / targetHeight);
      gl.bindTexture(gl.TEXTURE_2D, texture);
      gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, mediaElement);
      if (plan.stabilization.enabled && sampledFrame % 3 === 0) updateStabilization(width, height);
      sampledFrame += 1;
      gl.drawArrays(gl.TRIANGLES, 0, 6);
    }
    frame = requestAnimationFrame(draw);
  };
  draw();
  return { stop: () => { stopped = true; cancelAnimationFrame(frame); gl.deleteTexture(texture); gl.deleteTexture(hslTexture); gl.deleteBuffer(buffer); gl.deleteProgram(program); gl.deleteShader(vertex); gl.deleteShader(fragment); } };
}
