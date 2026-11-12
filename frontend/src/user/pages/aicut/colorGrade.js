import { applyHSLToRGB, createHSLRGBTransformer, hasHSLAdjustments } from './hslColor.js';

export const COLOR_CURVE_CHANNELS = ['red', 'green', 'blue'];
export const COLOR_WHEEL_RANGES = ['shadows', 'midtones', 'highlights'];
const identity = () => [{ x: 0, y: 0 }, { x: 1, y: 1 }];
const clamp = (value, min, max) => Math.min(max, Math.max(min, value));
const finite = value => typeof value === 'number' && Number.isFinite(value);
const bounded = (value, min, max) => finite(value) ? clamp(value, min, max) : 0;

export function normalizeRGBColorCurves(value) {
  return Object.fromEntries(COLOR_CURVE_CHANNELS.map(key => {
    const points = value?.[key];
    const valid = Array.isArray(points) && points.length >= 2 && points.length <= 16
      && points.every((point, i) => finite(point?.x) && finite(point?.y) && point.x >= 0 && point.x <= 1 && point.y >= 0 && point.y <= 1 && (!i || point.x > points[i - 1].x))
      && points[0].x === 0 && points.at(-1).x === 1;
    return [key, valid ? points.map(({ x, y }) => ({ x, y })) : identity()];
  }));
}

export function normalizeColorWheels(value) {
  return Object.fromEntries(COLOR_WHEEL_RANGES.map(key => [key, {
    hue: bounded(value?.[key]?.hue, -180, 180),
    saturation: bounded(value?.[key]?.saturation, 0, 100),
    luminance: bounded(value?.[key]?.luminance, -100, 100),
  }]));
}

export function hasColorGradeAdjustments(color = {}) {
  return hasHSLAdjustments(color.hsl)
    || Object.values(normalizeRGBColorCurves(color.rgbCurves)).some(points => points.some(point => point.x !== point.y))
    || Object.values(normalizeColorWheels(color.wheels)).some(wheel => wheel.saturation || wheel.luminance);
}

function sample(points, value) {
  for (let i = 1; i < points.length; i++) if (value <= points[i].x) {
    const a = points[i - 1], b = points[i];
    return a.y + (b.y - a.y) * (value - a.x) / (b.x - a.x);
  }
  return points.at(-1).y;
}

function compile(color = {}) {
  return { curves: normalizeRGBColorCurves(color.rgbCurves), wheels: normalizeColorWheels(color.wheels), hslTransform: createHSLRGBTransformer(color.hsl) };
}

function transform(pixel, grade) {
  if (pixel[3] === 0) return [...pixel];
  let channels = COLOR_CURVE_CHANNELS.map((key, i) => sample(grade.curves[key], pixel[i] / 255));
  const y = channels[0] * .2126 + channels[1] * .7152 + channels[2] * .0722;
  const weights = [(1 - y) ** 2, 2 * y * (1 - y), y ** 2];
  const bias = [0, 0, 0];
  let lift = 0;
  COLOR_WHEEL_RANGES.forEach((key, index) => {
    const wheel = grade.wheels[key], weight = weights[index];
    const hue = ((wheel.hue % 360) + 360) % 360;
    const x = 1 - Math.abs((hue / 60) % 2 - 1);
    const vector = hue < 60 ? [1, x, 0] : hue < 120 ? [x, 1, 0] : hue < 180 ? [0, 1, x] : hue < 240 ? [0, x, 1] : hue < 300 ? [x, 0, 1] : [1, 0, x];
    const vectorY = vector[0] * .2126 + vector[1] * .7152 + vector[2] * .0722;
    for (let i = 0; i < 3; i++) bias[i] += (vector[i] - vectorY) * wheel.saturation / 100 * .25 * weight;
    lift += wheel.luminance / 100 * weight;
  });
  channels = channels.map((channel, i) => {
    const biased = clamp(channel + bias[i], 0, 1);
    return clamp(biased + (lift >= 0 ? (1 - biased) * lift : biased * lift), 0, 1) * 255;
  });
  const output = pixel.length === 4 ? [...channels, pixel[3]] : channels;
  grade.hslTransform(channels[0], channels[1], channels[2], output);
  return output.map((value, i) => i < 3 ? Math.round(value) : value);
}

/** Existing global curve precedes this transform; input/alpha never mutate. */
export function applyColorGradeToRGB(pixel, color) {
  applyHSLToRGB(pixel); // Shared strict RGB(A) boundary validation.
  if (!hasColorGradeAdjustments(color)) return Array.from(pixel);
  return transform(pixel, compile(color));
}

/** Same atlas layout/quantization as createHSLLUTAtlas, now RGB -> wheels -> HSL. */
export function createColorGradeLUTAtlas(color) {
  const size = 33, width = size * size, height = size;
  const pixels = new Uint8Array(width * height * 4);
  const grade = compile(color);
  for (let b = 0; b < size; b++) for (let g = 0; g < size; g++) for (let r = 0; r < size; r++) {
    const offset = (g * width + b * size + r) * 4;
    pixels.set(transform([r * 255 / 32, g * 255 / 32, b * 255 / 32, 255], grade), offset);
  }
  return { size, width, height, pixels };
}
